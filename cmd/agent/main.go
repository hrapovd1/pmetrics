package main

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"os"
	"runtime"
	"time"
)

const pollInterval = 2 * time.Second
const reportInterval = 10 * time.Second
const serverAddress = "127.0.0.1"
const serverPort = "8080"

type gauge float64
type counter int64

var PollCount counter

func reportClient(client *http.Client, logger *log.Logger) {
	serverURL := "http://" + serverAddress + ":" + serverPort
	data := []byte("")
	req, err := http.NewRequest(http.MethodPost, serverURL, bytes.NewBuffer(data))
	if err != nil {
		logger.Fatal("Error reading request. ", err)
	}
	req.Header.Set("Content-Type", "text/plain")
	resp, err := client.Do(req)
	if err != nil {
		logger.Fatal("Error reading response. ", err)
	}
	defer resp.Body.Close()
	_, err = io.Copy(io.Discard, resp.Body)
	if err != nil {
		logger.Fatal("Error reading body. ", err)
	}
}

func pollMetrics(metrics map[string]gauge) {
	var rtm runtime.MemStats
	runtime.ReadMemStats(&rtm)
	metrics["Alloc"] = gauge(rtm.Alloc)
}

func main() {
	logger := log.New(os.Stdout, "AGENT\t", log.Ldate|log.Ltime)
	pollTick := time.NewTicker(pollInterval)
	reportTick := time.NewTicker(reportInterval)
	httpClient := &http.Client{}
	PollCount = counter(0)
	metrics := make(map[string]gauge, 28)

	logger.Println("started")
	defer logger.Println("stopped")
	for {
		select {
		case <-pollTick.C:
			PollCount++
			pollMetrics(metrics)
			// poll metrics
		case <-reportTick.C:
			// report to server
			reportClient(httpClient, logger)
		}
	}
}
