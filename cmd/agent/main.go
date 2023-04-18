package main

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	crand "crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"syscall"
	"time"

	"github.com/hrapovd1/pmetrics/internal/config"
	pb "github.com/hrapovd1/pmetrics/internal/proto"
	"github.com/hrapovd1/pmetrics/internal/types"
	"github.com/hrapovd1/pmetrics/internal/usecase"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

var (
	buildVersion string
	buildDate    string
	buildCommit  string
)

type (
	gauge    float64
	counter  int64
	mmetrics struct {
		mu          sync.Mutex
		pollCounter counter
		mtrcs       map[string]interface{}
	}
)

func main() {
	logger := log.New(os.Stdout, "AGENT\t", log.Ldate|log.Ltime)
	agentConf, err := config.NewAgentConf(config.GetAgentFlags())
	if err != nil {
		logger.Fatalln(err)
	}

	nctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	defer stop()

	metrics := mmetrics{
		pollCounter: counter(0),
		mtrcs:       make(map[string]interface{}, 29),
	}

	localAddr := getLocalAddr(*agentConf, logger)
	conn, err := grpc.Dial(agentConf.ServerAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Fatalf("when grpc.Dial got error: %v\n", err)
	}
	defer conn.Close()

	client := pb.NewMetricsClient(conn)

	md := metadata.New(map[string]string{"X-Real-IP": localAddr})
	ctx := metadata.NewOutgoingContext(nctx, md)

	if buildVersion == "" {
		buildVersion = "N/A"
	}
	if buildDate == "" {
		buildDate = "N/A"
	}
	if buildCommit == "" {
		buildCommit = "N/A"
	}

	logger.Println("Agent has started")
	logger.Printf("\tBuild version: %s\n", buildVersion)
	logger.Printf("\tBuild date: %s\n", buildDate)
	logger.Printf("\tBuild commit: %s\n", buildCommit)
	defer logger.Println("stopped")

	wg := &sync.WaitGroup{}
	wg.Add(3)

	go pollMetrics(ctx, wg, &metrics, agentConf.PollInterval)
	go pollHwMetrics(ctx, wg, &metrics, agentConf.PollInterval, logger)

	go reportMetrics(ctx, wg, &metrics, *agentConf, client, logger)

	wg.Wait()
}

func reportMetrics(ctx context.Context, w *sync.WaitGroup, metrics *mmetrics, cfg config.Config, clnt pb.MetricsClient, logger *log.Logger) {
	defer w.Done()
	var (
		pubKey     *rsa.PublicKey
		dataEnc    string
		encDataKey string
		err        error
		stream     pb.Metrics_ReportMetricsClient
		encrypt    = struct {
			enc     bool
			symmKey []byte
			stream  pb.Metrics_ReportEncMetricsClient
		}{enc: false}
	)

	if cfg.CryptoKey != "" {
		if pubKey, err = getPubKey(cfg.CryptoKey, logger); err != nil {
			logger.Fatalf("getPubKey got error: %v", err)
		}
		encrypt.enc = true
	}

	reportTick := time.NewTicker(cfg.ReportInterval)
	defer reportTick.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-reportTick.C:
			if encrypt.enc {
				// prepare for transmition
				encrypt.stream, err = clnt.ReportEncMetrics(ctx)
				if err != nil {
					logger.Println(err)
					break
				}
				// gen symm key
				encrypt.symmKey, err = genSymmKey(24)
				if err != nil {
					logger.Println(err)
					break
				}
				encDataKey, err = symmKeyToEnc(pubKey, encrypt.symmKey)
				if err != nil {
					logger.Println(err)
					break
				}
			} else {
				stream, err = clnt.ReportMetrics(ctx)
				if err != nil {
					logger.Println(err)
					break
				}
			}

			// send metrics in stream
			metrics.mu.Lock()
			for mKey, mVal := range metrics.mtrcs {
				data, err := metricToJSON(mKey, mVal, cfg.Key)
				if err != nil {
					logger.Println(err)
				}
				if encrypt.enc {
					dataEnc, err = dataToEnc(encrypt.symmKey, data)
					if err != nil {
						logger.Println(err)
						break
					}
					req := pb.EncMetricRequest{
						Data: &pb.EncMetric{
							Data0: encDataKey,
							Data:  dataEnc,
						},
					}
					if err := encrypt.stream.Send(&req); err != nil {
						logger.Println(err)
						break
					}
				} else {
					req := pb.MetricRequest{
						Metric: data,
					}
					if err := stream.Send(&req); err != nil {
						logger.Println(err)
						break
					}
				}
			}
			metrics.mu.Unlock()

			// close opened stream
			if encrypt.enc {
				if err := encrypt.stream.CloseSend(); err != nil {
					logger.Printf("when close encrypt stream got err: %v", err)
				}
			} else {
				if err := stream.CloseSend(); err != nil {
					logger.Printf("when close stream got err: %v", err)
				}
			}
		}
	}
}

func pollMetrics(ctx context.Context, w *sync.WaitGroup, metrics *mmetrics, pollIntvl time.Duration) {
	defer w.Done()
	pollTick := time.NewTicker(pollIntvl)
	defer pollTick.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-pollTick.C:
			var rtm runtime.MemStats
			runtime.ReadMemStats(&rtm)
			metrics.mu.Lock()

			metrics.pollCounter++
			metrics.mtrcs["PollCount"] = metrics.pollCounter
			metrics.mtrcs["Alloc"] = gauge(rtm.Alloc)
			metrics.mtrcs["TotalAlloc"] = gauge(rtm.TotalAlloc)
			metrics.mtrcs["Sys"] = gauge(rtm.Sys)
			metrics.mtrcs["Lookups"] = gauge(rtm.Lookups)
			metrics.mtrcs["Mallocs"] = gauge(rtm.Mallocs)
			metrics.mtrcs["Frees"] = gauge(rtm.Frees)
			metrics.mtrcs["HeapAlloc"] = gauge(rtm.HeapAlloc)
			metrics.mtrcs["HeapSys"] = gauge(rtm.HeapSys)
			metrics.mtrcs["HeapIdle"] = gauge(rtm.HeapIdle)
			metrics.mtrcs["HeapInuse"] = gauge(rtm.HeapInuse)
			metrics.mtrcs["HeapReleased"] = gauge(rtm.HeapReleased)
			metrics.mtrcs["HeapObjects"] = gauge(rtm.HeapObjects)
			metrics.mtrcs["StackInuse"] = gauge(rtm.StackInuse)
			metrics.mtrcs["StackSys"] = gauge(rtm.StackSys)
			metrics.mtrcs["MSpanInuse"] = gauge(rtm.MSpanInuse)
			metrics.mtrcs["MSpanSys"] = gauge(rtm.MSpanSys)
			metrics.mtrcs["MCacheInuse"] = gauge(rtm.MCacheInuse)
			metrics.mtrcs["MCacheSys"] = gauge(rtm.MCacheSys)
			metrics.mtrcs["BuckHashSys"] = gauge(rtm.BuckHashSys)
			metrics.mtrcs["GCSys"] = gauge(rtm.GCSys)
			metrics.mtrcs["OtherSys"] = gauge(rtm.OtherSys)
			metrics.mtrcs["NextGC"] = gauge(rtm.NextGC)
			metrics.mtrcs["LastGC"] = gauge(rtm.LastGC)
			metrics.mtrcs["PauseTotalNs"] = gauge(rtm.PauseTotalNs)
			metrics.mtrcs["NumGC"] = gauge(rtm.NumGC)
			metrics.mtrcs["NumForcedGC"] = gauge(rtm.NumForcedGC)
			metrics.mtrcs["GCCPUFraction"] = gauge(rtm.GCCPUFraction)
			metrics.mtrcs["RandomValue"] = gauge(rand.Float64())

			metrics.mu.Unlock()
		}
	}
}

func pollHwMetrics(ctx context.Context, w *sync.WaitGroup, metrics *mmetrics, pollIntvl time.Duration, logger *log.Logger) {
	defer w.Done()
	pollTick := time.NewTicker(pollIntvl)
	defer pollTick.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-pollTick.C:
			v, _ := mem.VirtualMemory()
			cpus, err := cpu.PercentWithContext(ctx, pollIntvl, true)
			if err != nil {
				logger.Println(err)
			}
			metrics.mu.Lock()

			metrics.mtrcs["TotalMemory"] = gauge(v.Total)
			metrics.mtrcs["FreeMemory"] = gauge(v.Free)
			for c, cpuVal := range cpus {
				name := fmt.Sprintf("CPUutilization%v", c+1)
				metrics.mtrcs[name] = gauge(cpuVal)
			}

			metrics.mu.Unlock()
		}
	}
}

func metricToJSON(mKey string, mValue interface{}, key string) ([]byte, error) {
	var value float64
	var delta int64
	data := types.Metric{ID: mKey}
	switch val := mValue.(type) {
	case gauge:
		value = float64(val)
		data.MType = "gauge"
		data.Value = &value
		if key != "" {
			if err := usecase.SignData(&data, key); err != nil {
				return nil, err
			}
		}
	case counter:
		delta = int64(val)
		data.MType = "counter"
		data.Delta = &delta
		if key != "" {
			if err := usecase.SignData(&data, key); err != nil {
				return nil, err
			}
		}
	}
	return json.Marshal(data)
}

func getPubKey(fname string, logger *log.Logger) (*rsa.PublicKey, error) {
	// read public key from file
	keyFile, err := os.Open(fname)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := keyFile.Close(); err != nil {
			logger.Println(err)
		}
	}()
	pemPubKey := make([]byte, 4*1024)
	n, err := keyFile.Read(pemPubKey)
	if err != nil {
		return nil, err
	}
	pemPubKey = pemPubKey[:n]

	// decode public key from pem format
	pubKey, _ := pem.Decode(pemPubKey)
	if pubKey == nil || pubKey.Type != "PUBLIC KEY" {
		return nil, errors.New("not found PUBLIC KEY in file " + fname)
	}
	// parse public key from byte slice
	rsaPubKey, err := x509.ParsePKIXPublicKey(pubKey.Bytes)
	if err != nil {
		return nil, err
	}
	key, ok := rsaPubKey.(*rsa.PublicKey)
	if !ok {
		return key, errors.New("can't convert key to *rsa.PublicKey")
	}
	return key, nil
}

func symmKeyToEnc(key *rsa.PublicKey, symmKey []byte) (string, error) {
	// encrypt symm key
	keyEnc, err := rsa.EncryptPKCS1v15(crand.Reader, key, symmKey)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(keyEnc), nil
}

func dataToEnc(symmKey []byte, data []byte) (string, error) {
	// encrypt data
	cphr, err := aes.NewCipher(symmKey)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(cphr)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(crand.Reader, nonce); err != nil {
		return "", err
	}
	dataEnc := gcm.Seal(nonce, nonce, data, nil)

	return base64.StdEncoding.EncodeToString(dataEnc), nil
}

func genSymmKey(n int) ([]byte, error) {
	out := make([]byte, n)
	n1, err := crand.Read(out)
	if err != nil || n1 != n {
		return out, err
	}
	return out, nil
}

func getLocalAddr(conf config.Config, logger *log.Logger) string {
	confAddr := net.ParseIP(conf.TrustedSubnet)
	if confAddr != nil {
		return confAddr.String()
	}
	conn, err := net.Dial("tcp", conf.ServerAddress)
	if err != nil {
		logger.Printf("when try to dial server %v got error: %v\n", conf.ServerAddress, err)
		return ""
	}
	defer func() {
		if err := conn.Close(); err != nil {
			logger.Printf("when close connection got error: %v\n", err)
		}
	}()
	return conn.LocalAddr().(*net.TCPAddr).IP.String()
}
