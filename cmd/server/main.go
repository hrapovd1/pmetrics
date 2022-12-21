package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"

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
	ctx, cancel := context.WithCancel(context.Background())
	server := http.Server{Addr: serverConf.ServerAddress, ErrorLog: logger}

	handlerMetrics := handlers.NewMetricsHandler(*serverConf, logger)
	handlerStorage := handlerMetrics.Storage.(types.Storager)
	defer handlerStorage.Close()

	go handlerStorage.Storing(ctx, logger, serverConf.StoreInterval, serverConf.IsRestore)

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

	server.Handler = router

	go func() {
		defer cancel()
		sigint := make(chan os.Signal, 1)
		defer close(sigint)

		signal.Notify(sigint, os.Interrupt)
		<-sigint

		if err := server.Shutdown(ctx); err != nil {
			// Error from closing listeners, or context timeout:
			logger.Printf("HTTP server Shutdown: %v", err)
		}
	}()

	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		// Error starting or closing listener:
		logger.Printf("HTTP server ListenAndServe: %v", err)
	}
}
