package main

import (
	"encoding/json"
	"log"
	"math/rand"
	"os"
	"runtime"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/hrapovd1/pmetrics/internal/config"
	"github.com/hrapovd1/pmetrics/internal/types"
	"github.com/hrapovd1/pmetrics/internal/usecase"
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
	metrics := make(map[string]interface{}, 29)
	metricURL := "http://" + agentConf.ServerAddress + "/updates/"

	logger.Println("started")
	defer logger.Println("stopped")
	for {
		select {
		case <-pollTick.C:
			PollCount++
			metrics["PollCount"] = PollCount
			pollMetrics(metrics)
		case <-reportTick.C:
			data, err := metricsToJSON(metrics, agentConf.Key)
			if err != nil {
				logger.Println(err)
			}
			_, err = httpClient.R().
				SetHeader("Content-Type", "application/json").
				SetBody(data).
				Post(metricURL)
			if err != nil {
				logger.Print("Error when sent metrics. ", err)
			}
		}
	}
}

func pollMetrics(metrics map[string]interface{}) {
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
