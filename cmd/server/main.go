package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"

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

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	defer stop()

	wg := sync.WaitGroup{}

	wg.Add(1)
	go handlerStorage.Storing(ctx, &wg, logger, serverConf.StoreInterval, serverConf.IsRestore)

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

	server := http.Server{
		Addr:    serverConf.ServerAddress,
		Handler: router,
	}

	wg.Add(1)
	go func(c context.Context, w *sync.WaitGroup, s *http.Server, l *log.Logger) {
		defer wg.Done()
		<-c.Done()
		l.Println("got signal to stop")
		if err := s.Shutdown(c); err != nil {
			l.Println(err)
		}

	}(ctx, &wg, &server, logger)

	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		logger.Fatal(err)
	}

	wg.Wait()
	logger.Println("server stoped gracefully")
}
