package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/hrapovd1/pmetrics/internal/handlers"
	"github.com/hrapovd1/pmetrics/internal/storage"
)

const (
	serverHost = "127.0.0.1"
	serverPort = "8080"
)

var handlersStorage = handlers.MetricStorage{
	Storage: storage.NewMemStorage(),
}

func main() {
	serverAddr := fmt.Sprint(serverHost, ":", serverPort)

	http.HandleFunc("/update/gauge/", handlersStorage.GaugeHandler)
	http.HandleFunc("/update/counter/", handlersStorage.CounterHandler)
	http.HandleFunc("/update/", handlers.NotImplementedHandler)
	http.HandleFunc("/value/", handlersStorage.GetMetricHandler)
	http.HandleFunc("/", handlersStorage.GetAllHandler)
	log.Println("Server start on ", serverAddr)
	log.Fatal(http.ListenAndServe(serverAddr, nil))
}
