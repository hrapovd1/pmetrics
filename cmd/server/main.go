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
	logger := log.New(os.Stdout, "SERVER\t", log.Ldate|log.Ltime)
	// Чтение флагов и установка конфигурации сервера
	serverConf, err := config.NewServer(config.GetServerFlags())
	if err != nil {
		logger.Fatalln(err)
	}
	backendStorage := storage.NewBackend(*serverConf) // Файловый бекенд хранилища метрик
	handlersStorage := handlers.MetricStorage{        // Хранилище метрик
		Storage: storage.NewMemStorage(storage.WithBackend(&backendStorage)),
	}

	donech := make(chan struct{})
	defer close(donech)

	go backendStorage.Storing(donech, logger)

	router := chi.NewRouter()
	router.Use(handlers.GzipMiddle)
	router.Get("/", handlersStorage.GetAllHandler)
	router.Get("/value/*", handlersStorage.GetMetricHandler)
	router.Post("/value/", handlersStorage.GetMetricJSONHandler)

	update := chi.NewRouter()
	update.Post("/gauge/*", handlersStorage.GaugeHandler)
	update.Post("/counter/*", handlersStorage.CounterHandler)
	update.Post("/", handlersStorage.UpdateHandler)
	update.Post("/*", handlers.NotImplementedHandler)

	router.Mount("/update", update)

	logger.Println("Server start on ", serverConf.ServerAddress)
	logger.Fatal(http.ListenAndServe(serverConf.ServerAddress, router))
}
