package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/hrapovd1/pmetrics/internal/config"
	"github.com/hrapovd1/pmetrics/internal/handlers"
	"github.com/hrapovd1/pmetrics/internal/storage"
)

var handlersStorage = handlers.MetricStorage{
	Storage: storage.NewMemStorage(),
}

func main() {
	serverAddr := fmt.Sprint(config.ServerConfig.ServerAddress, ":", config.ServerConfig.ServerPort)

	router := chi.NewRouter()
	router.Get("/", handlersStorage.GetAllHandler)
	router.Post("/value/", handlersStorage.GetMetricHandler)
	router.Post("/update/", handlersStorage.UpdateHandler)

	log.Println("Server start on ", serverAddr)
	log.Fatal(http.ListenAndServe(serverAddr, router))
}
