package main

import (
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"

	"github.com/hrapovd1/pmetrics/internal/config"
	"github.com/hrapovd1/pmetrics/internal/handlers"
	"github.com/hrapovd1/pmetrics/internal/storage"
)

func main() {
	var serverConf config.Config
	logger := log.New(os.Stdout, "SERVER\t", log.Ldate|log.Ltime)
	if err := serverConf.NewServer(); err != nil {
		logger.Fatalln(err)
	}
	backendStorage := storage.NewBackend(serverConf)
	handlersStorage := handlers.MetricStorage{
		Storage: storage.NewMemStorage(storage.WithBackend(&backendStorage)),
	}

	donech := make(chan struct{})
	defer close(donech)

	go backendStorage.Storing(donech, logger)

	router := chi.NewRouter()
	router.Get("/", handlersStorage.GetAllHandler)
	router.Get("/value/*", handlersStorage.GetMetricHandler)
	router.Post("/value/", handlersStorage.GetMetricJSONHandler)

	update := chi.NewRouter()
	update.Post("/gauge/*", handlersStorage.GaugeHandler)
	update.Post("/counter/*", handlersStorage.CounterHandler)
	update.Post("/", handlersStorage.UpdateHandler)
	update.Post("/*", handlers.NotImplementedHandler)

	router.Mount("/update", update)

	log.Println("Server start on ", serverConf.ServerAddress)
	log.Fatal(http.ListenAndServe(serverConf.ServerAddress, router))
}
