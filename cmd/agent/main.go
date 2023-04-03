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
	"os"
	"os/signal"
	"runtime"
	"sync"
	"syscall"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/hrapovd1/pmetrics/internal/config"
	"github.com/hrapovd1/pmetrics/internal/types"
	"github.com/hrapovd1/pmetrics/internal/usecase"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
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

	httpClient := resty.New()

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
	ctx := context.WithValue(nctx, types.Waitgrp("WG"), wg)

	go pollMetrics(ctx, &metrics, agentConf.PollInterval)
	go pollHwMetrics(ctx, &metrics, agentConf.PollInterval, logger)

	go reportMetrics(ctx, &metrics, *agentConf, httpClient, logger)

	wg.Wait()
}

func reportMetrics(ctx context.Context, metrics *mmetrics, cfg config.Config, httpClnt *resty.Client, logger *log.Logger) {
	metricURL := "http://" + cfg.ServerAddress + "/updates/"
	var (
		pubKey  *rsa.PublicKey
		dataEnc []byte
		err     error
	)
	waitGroup := ctx.Value(types.Waitgrp("WG")).(*sync.WaitGroup)
	defer waitGroup.Done()
	encrypt := false
	if cfg.CryptoKey != "" {
		if pubKey, err = getPubKey(cfg.CryptoKey, logger); err != nil {
			logger.Fatalf("getPubKey got error: %v", err)
		}
		encrypt = true
	}

	reportTick := time.NewTicker(cfg.ReportInterval)
	defer reportTick.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-reportTick.C:
			metrics.mu.Lock()
			data, err := metricsToJSON(metrics.mtrcs, cfg.Key)
			metrics.mu.Unlock()
			if err != nil {
				logger.Println(err)
			}

			if encrypt {
				dataEnc, err = dataToEncJSON(pubKey, data)
				if err != nil {
					logger.Println(err)
				}
			}

			restyClient := httpClnt.R().SetHeader("Content-Type", "application/json")
			if encrypt {
				restyClient.SetHeader("Encrypt-Type", "1").SetBody(dataEnc)
			} else {
				restyClient.SetBody(data)
			}

			_, err = restyClient.Post(metricURL)
			if err != nil {
				logger.Print("Error when sent metrics. ", err)
			}
		}
	}
}

func pollMetrics(ctx context.Context, metrics *mmetrics, pollIntvl time.Duration) {
	waitGroup := ctx.Value(types.Waitgrp("WG")).(*sync.WaitGroup)
	defer waitGroup.Done()
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

func pollHwMetrics(ctx context.Context, metrics *mmetrics, pollIntvl time.Duration, logger *log.Logger) {
	waitGroup := ctx.Value(types.Waitgrp("WG")).(*sync.WaitGroup)
	defer waitGroup.Done()
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

func metricsToJSON(mtrcs map[string]interface{}, key string) ([]byte, error) {
	metrics := make([]types.Metric, 0)
	for k, v := range mtrcs {
		var value float64
		var delta int64
		data := types.Metric{ID: k}
		switch val := v.(type) {
		case gauge:
			value = float64(val)
			data.MType = "gauge"
			data.Value = &value
			if key != "" {
				if err := usecase.SignData(&data, key); err != nil {
					return nil, err
				}
			}
			metrics = append(metrics, data)
		case counter:
			delta = int64(val)
			data.MType = "counter"
			data.Delta = &delta
			if key != "" {
				if err := usecase.SignData(&data, key); err != nil {
					return nil, err
				}
			}
			metrics = append(metrics, data)
		}
	}
	return json.Marshal(metrics)
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

func dataToEncJSON(key *rsa.PublicKey, data []byte) ([]byte, error) {
	// gen symm key
	symmKey, err := genSymmKey(24)
	if err != nil {
		return nil, err
	}

	// encrypt symm key
	keyEnc, err := rsa.EncryptPKCS1v15(crand.Reader, key, symmKey)
	if err != nil {
		return nil, err
	}
	// encrypt data
	cphr, err := aes.NewCipher(symmKey)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(cphr)
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(crand.Reader, nonce); err != nil {
		return nil, err
	}
	dataEnc := gcm.Seal(nonce, nonce, data, nil)

	// json marshal
	return json.Marshal(
		types.EncData{
			Data0: base64.StdEncoding.EncodeToString(keyEnc),
			Data:  base64.StdEncoding.EncodeToString(dataEnc),
		},
	)
}

func genSymmKey(n int) ([]byte, error) {
	out := make([]byte, n)
	n1, err := crand.Read(out)
	if err != nil || n1 != n {
		return out, err
	}
	return out, nil
}
