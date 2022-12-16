package main

import (
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"

	"github.com/hrapovd1/pmetrics/internal/config"
	dbstorage "github.com/hrapovd1/pmetrics/internal/dbstrorage"
	"github.com/hrapovd1/pmetrics/internal/filestorage"
	"github.com/hrapovd1/pmetrics/internal/handlers"
	"github.com/hrapovd1/pmetrics/internal/storage"
)

func main() {
	logger := log.New(os.Stdout, "SERVER\t", log.Ldate|log.Ltime)
	// Чтение флагов и установка конфигурации сервера
	serverConf, err := config.NewServerConf(config.GetServerFlags())
	if err != nil {
		logger.Fatalln(err)
	}

	memBuff := make(map[string]interface{}) // Основной буфер хранения метрик

	fileStorage := filestorage.NewFileStorage(*serverConf, memBuff) // Файловый бекенд хранилища метрик
	defer fileStorage.Close()

	dbStorage, err := dbstorage.NewDBStorage(*serverConf, logger, memBuff) // БД для метрик
	if err != nil {
		logger.Fatalln(err)
	}
	defer dbStorage.Close()

	handlerMetrics := handlers.MetricsHandler{ // Хранилище метрик
		MemStor: storage.NewMemStorage(
			storage.WithBuffer(memBuff),
		),
		FileStor: fileStorage,
		DBStor:   dbStorage,
		Config:   *serverConf,
	}

	go fileStorage.Storing(logger)

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
