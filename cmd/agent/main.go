package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"runtime"
	"time"
)

const (
	pollInterval   = 2 * time.Second
	reportInterval = 10 * time.Second
	serverAddress  = "127.0.0.1"
	serverPort     = "8080"
)

type gauge float64
type counter int64

func main() {
	logger := log.New(os.Stdout, "AGENT\t", log.Ldate|log.Ltime)
	pollTick := time.NewTicker(pollInterval)
	reportTick := time.NewTicker(reportInterval)
	httpClient := &http.Client{}
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
				metricURL := fmt.Sprint("http://", serverAddress, ":", serverPort, "/update/gauge/", k, "/", v)
				reportClient(httpClient, metricURL, logger)
			}
			metricURL := fmt.Sprint("http://", serverAddress, ":", serverPort, "/update/counter/PollCount/", PollCount)
			reportClient(httpClient, metricURL, logger)
		}
	}
}

func reportClient(client *http.Client, url string, logger *log.Logger) {
	data := []byte("")
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(data))
	if err != nil {
		logger.Print("Error reading request. ", err)
		return
	}
	req.Header.Set("Content-Type", "text/plain")
	resp, err := client.Do(req)
	if err != nil {
		logger.Print("Error reading response. ", err)
		return
	}
	defer resp.Body.Close()
	_, err = io.Copy(io.Discard, resp.Body)
	if err != nil {
		logger.Print("Error reading body. ", err)
		return
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
