package main

import (
	"encoding/json"
	"errors"
	"log"
	"math/rand"
	"os"
	"runtime"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/hrapovd1/pmetrics/internal/config"
	"github.com/hrapovd1/pmetrics/internal/types"
)

type gauge float64
type counter int64

func main() {
	logger := log.New(os.Stdout, "AGENT\t", log.Ldate|log.Ltime)
	agentConf, err := config.NewAgentConf(config.GetAgentFlags())
	if err != nil {
		logger.Fatalln(err)
	}
	pollTick := time.NewTicker(agentConf.PollInterval)
	reportTick := time.NewTicker(agentConf.ReportInterval)
	httpClient := resty.New()
	PollCount := counter(0)
	metrics := make(map[string]gauge, 28)
	metricURL := "http://" + agentConf.ServerAddress + "/update/"

	logger.Println("started")
	defer logger.Println("stopped")
	for {
		select {
		case <-pollTick.C:
			PollCount++
			pollMetrics(metrics)
		case <-reportTick.C:
			for k, v := range metrics {
				data, err := metricToJSON(k, v)
				if err != nil {
					logger.Fatalln(err)
				}
				_, err = httpClient.R().
					SetHeader("Content-Type", "application/json").
					SetBody(data).
					Post(metricURL)
				if err != nil {
					logger.Print("Error when sent metric. ", err)
				}
			}
			data, err := metricToJSON("PollCount", PollCount)
			if err != nil {
				logger.Fatalln(err)
			}
			_, err = httpClient.R().
				SetHeader("Content-Type", "application/json").
				SetBody(data).
				Post(metricURL)
			if err != nil {
				logger.Print("Error when sent metric. ", err)
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

func metricToJSON(name string, val interface{}) ([]byte, error) {
	var value float64
	var delta int64
	switch val := val.(type) {
	case gauge:
		value = float64(val)
		return json.Marshal(
			types.Metric{
				ID:    name,
				MType: "gauge",
				Value: &value,
			},
		)
	case counter:
		delta = int64(val)
		return json.Marshal(
			types.Metric{
				ID:    name,
				MType: "counter",
				Delta: &delta,
			},
		)
	}
	return nil, errors.New("got undefined metric type")
}
