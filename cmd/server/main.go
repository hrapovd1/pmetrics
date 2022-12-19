package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"

	"github.com/hrapovd1/pmetrics/internal/config"
	"github.com/hrapovd1/pmetrics/internal/handlers"
	"github.com/hrapovd1/pmetrics/internal/types"
)

func main() {
	logger := log.New(os.Stdout, "SERVER\t", log.Ldate|log.Ltime)
	// Чтение флагов и установка конфигурации сервера
	serverConf, err := config.NewServerConf(config.GetServerFlags())
	if err != nil {
		logger.Fatalln(err)
	}

	logger.Printf("serverConf: %v", serverConf)

	handlerMetrics := handlers.NewMetricsHandler(*serverConf, logger)
	handlerStorage := handlerMetrics.Storage.(types.Storager)
	defer handlerStorage.Close()

	go handlerStorage.Storing(context.Background(), logger, serverConf.StoreInterval, serverConf.IsRestore)

	router := chi.NewRouter()
	router.Use(handlers.GzipMiddle)
	router.Get("/", handlerMetrics.GetAllHandler)
	router.Get("/value/*", handlerMetrics.GetMetricHandler)
	router.Get("/ping", handlerMetrics.PingDB)
	router.Post("/value/", handlerMetrics.GetMetricJSONHandler)
	router.Post("/updates/", handlerMetrics.UpdatesHandler)

	update := chi.NewRouter()
	update.Post("/gauge/*", handlerMetrics.GaugeHandler)
	update.Post("/counter/*", handlerMetrics.CounterHandler)
	update.Post("/", handlerMetrics.UpdateHandler)
	update.Post("/*", handlers.NotImplementedHandler)

	router.Mount("/update", update)

	logger.Println("Server start on ", serverConf.ServerAddress)
	logger.Fatal(http.ListenAndServe(serverConf.ServerAddress, router))
}
