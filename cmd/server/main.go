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

var hanlersStorage = handlers.MetricStorage{
	Storage: storage.NewMemStorage(),
}

func main() {
	serverAddr := fmt.Sprint(serverHost, ":", serverPort)

	http.HandleFunc("/update/gauge/", hanlersStorage.GaugeHandler)
	http.HandleFunc("/update/counter/", hanlersStorage.CounterHandler)
	http.Handle("/", http.NotFoundHandler())
	log.Println("Server start on ", serverAddr)
	log.Fatal(http.ListenAndServe(serverAddr, nil))
}
