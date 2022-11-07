package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"runtime"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/hrapovd1/pmetrics/internal/config"
)

type gauge float64
type counter int64

func main() {
	logger := log.New(os.Stdout, "AGENT\t", log.Ldate|log.Ltime)
	pollTick := time.NewTicker(config.AgentConfig.PollInterval)
	reportTick := time.NewTicker(config.AgentConfig.ReportInterval)
	httpClient := resty.New()
	PollCount := counter(0)
	metrics := make(map[string]gauge, 28)

	logger.Println("started")
	defer logger.Println("stopped")
	for {
		select {
		case <-pollTick.C:
			PollCount++
			pollMetrics(metrics)
		case <-reportTick.C:
			for k, v := range metrics {
				metricURL := fmt.Sprint("http://", config.AgentConfig.ServerAddress, ":", config.AgentConfig.ServerPort, "/update/gauge/", k, "/", v)
				_, err := httpClient.R().SetHeader("Content-Type", "text/plain").Post(metricURL)
				if err != nil {
					logger.Print("Error when sent metric. ", err)
					return
				}
			}
			metricURL := fmt.Sprint("http://", config.AgentConfig.ServerAddress, ":", config.AgentConfig.ServerPort, "/update/counter/PollCount/", PollCount)
			_, err := httpClient.R().SetHeader("Content-Type", "text/plain").Post(metricURL)
			if err != nil {
				logger.Print("Error when sent metric. ", err)
				return
			}
		}
	}
}

func pollMetrics(metrics map[string]gauge) {
	var rtm runtime.MemStats
	runtime.ReadMemStats(&rtm)
	metrics["Alloc"] = gauge(rtm.Alloc)
	metrics["TotalAlloc"] = gauge(rtm.TotalAlloc)
	metrics["Sys"] = gauge(rtm.Sys)
	metrics["Lookups"] = gauge(rtm.Lookups)
	metrics["Mallocs"] = gauge(rtm.Mallocs)
	metrics["Frees"] = gauge(rtm.Frees)
	metrics["HeapAlloc"] = gauge(rtm.HeapAlloc)
	metrics["HeapSys"] = gauge(rtm.HeapSys)
	metrics["HeapIdle"] = gauge(rtm.HeapIdle)
	metrics["HeapInuse"] = gauge(rtm.HeapInuse)
	metrics["HeapReleased"] = gauge(rtm.HeapReleased)
	metrics["HeapObjects"] = gauge(rtm.HeapObjects)
	metrics["StackInuse"] = gauge(rtm.StackInuse)
	metrics["StackSys"] = gauge(rtm.StackSys)
	metrics["MSpanInuse"] = gauge(rtm.MSpanInuse)
	metrics["MSpanSys"] = gauge(rtm.MSpanSys)
	metrics["MCacheInuse"] = gauge(rtm.MCacheInuse)
	metrics["MCacheSys"] = gauge(rtm.MCacheSys)
	metrics["BuckHashSys"] = gauge(rtm.BuckHashSys)
	metrics["GCSys"] = gauge(rtm.GCSys)
	metrics["OtherSys"] = gauge(rtm.OtherSys)
	metrics["NextGC"] = gauge(rtm.NextGC)
	metrics["LastGC"] = gauge(rtm.LastGC)
	metrics["PauseTotalNs"] = gauge(rtm.PauseTotalNs)
	metrics["NumGC"] = gauge(rtm.NumGC)
	metrics["NumForcedGC"] = gauge(rtm.NumForcedGC)
	metrics["GCCPUFraction"] = gauge(rtm.GCCPUFraction)
	metrics["RandomValue"] = gauge(rand.Float64())
}
