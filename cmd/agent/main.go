package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/hrapovd1/pmetrics/internal/config"
	"github.com/hrapovd1/pmetrics/internal/types"
	"github.com/hrapovd1/pmetrics/internal/usecase"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
)

type gauge float64
type counter int64
type mmetrics struct {
	mu          sync.Mutex
	pollCounter counter
	mtrcs       map[string]interface{}
}

func main() {
	logger := log.New(os.Stdout, "AGENT\t", log.Ldate|log.Ltime)
	agentConf, err := config.NewAgentConf(config.GetAgentFlags())
	if err != nil {
		logger.Fatalln(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	metrics := mmetrics{
		pollCounter: counter(0),
		mtrcs:       make(map[string]interface{}, 29),
	}

	httpClient := resty.New()

	sigint := make(chan os.Signal, 1)
	defer close(sigint)
	signal.Notify(sigint, os.Interrupt)

	logger.Println("started")
	defer logger.Println("stopped")

	go pollMetrics(ctx, &metrics, agentConf.PollInterval)
	go pollHwMetrics(ctx, &metrics, agentConf.PollInterval, logger)

	go reportMetrics(ctx, &metrics, *agentConf, httpClient, logger)

	<-sigint
}

func reportMetrics(ctx context.Context, metrics *mmetrics, cfg config.Config, httpClnt *resty.Client, logger *log.Logger) {
	metricURL := "http://" + cfg.ServerAddress + "/updates/"
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
			_, err = httpClnt.R().
				SetHeader("Content-Type", "application/json").
				SetBody(data).
				Post(metricURL)
			if err != nil {
				logger.Print("Error when sent metrics. ", err)
			}
		}
	}
}

func pollMetrics(ctx context.Context, metrics *mmetrics, pollIntvl time.Duration) {
	pollTick := time.NewTicker(pollIntvl)
	defer pollTick.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-pollTick.C:
			metrics.mu.Lock()

			metrics.pollCounter++
			metrics.mtrcs["PollCount"] = metrics.pollCounter
			var rtm runtime.MemStats
			runtime.ReadMemStats(&rtm)
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
