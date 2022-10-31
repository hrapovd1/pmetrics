package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"

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

	router := chi.NewRouter()
	router.Get("/", handlersStorage.GetAllHandler)
	router.Get("/value/*", handlersStorage.GetMetricHandler)

	update := chi.NewRouter()
	update.Post("/gauge/*", handlersStorage.GaugeHandler)
	update.Post("/counter/*", handlersStorage.CounterHandler)
	update.Post("/*", handlers.NotImplementedHandler)

	router.Mount("/update", update)

	log.Println("Server start on ", serverAddr)
	log.Fatal(http.ListenAndServe(serverAddr, router))
}
