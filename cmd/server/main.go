package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/hrapovd1/pmetrics/internal/config"
	"github.com/hrapovd1/pmetrics/internal/handlers"
	"github.com/hrapovd1/pmetrics/internal/types"
)

var (
	buildVersion string
	buildDate    string
	buildCommit  string
)

func main() {
	logger := log.New(os.Stdout, "SERVER\t", log.Ldate|log.Ltime)
	// Чтение флагов и установка конфигурации сервера
	serverConf, err := config.NewServerConf(config.GetServerFlags())
	if err != nil {
		logger.Fatalln(err)
	}

	if buildVersion == "" {
		buildVersion = "N/A"
	}
	if buildDate == "" {
		buildDate = "N/A"
	}
	if buildCommit == "" {
		buildCommit = "N/A"
	}

	logger.Printf("\tBuild version: %s\n", buildVersion)
	logger.Printf("\tBuild date: %s\n", buildDate)
	logger.Printf("\tBuild commit: %s\n", buildCommit)
	logger.Println("Server start on ", serverConf.ServerAddress)

	handlerMetrics := handlers.NewMetricsHandler(*serverConf, logger)
	handlerStorage := handlerMetrics.Storage.(types.Storager)
	defer func() {
		if err := handlerStorage.Close(); err != nil {
			logger.Print(err)
		}
	}()

	go handlerStorage.Storing(context.Background(), logger, serverConf.StoreInterval, serverConf.IsRestore)

	router := chi.NewRouter()
	router.Use(handlerMetrics.DecryptMiddle)
	router.Use(handlerMetrics.GzipMiddle)
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
	router.Mount("/debug", middleware.Profiler())

	logger.Fatal(http.ListenAndServe(serverConf.ServerAddress, router))
}
