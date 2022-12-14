package main

import (
	"context"
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

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	memBuff := make(map[string]interface{}) // Основной буфер хранения метрик

	fileStorage := filestorage.NewFileStorage(ctx, *serverConf, memBuff) // Файловый бекенд хранилища метрик
	defer fileStorage.Close()

	dbStorage, err := dbstorage.NewDBStorage(ctx, *serverConf, logger, memBuff) // БД для метрик
	if err != nil {
		logger.Fatalln(err)
	}
	defer dbStorage.Close()

	handlersStorage := handlers.MetricStorage{ // Хранилище метрик
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
	router.Get("/", handlersStorage.GetAllHandler)
	router.Get("/value/*", handlersStorage.GetMetricHandler)
	router.Get("/ping", handlersStorage.PingDB)
	router.Post("/value/", handlersStorage.GetMetricJSONHandler)
	router.Post("/updates/", handlersStorage.UpdatesHandler)

	update := chi.NewRouter()
	update.Post("/gauge/*", handlersStorage.GaugeHandler)
	update.Post("/counter/*", handlersStorage.CounterHandler)
	update.Post("/", handlersStorage.UpdateHandler)
	update.Post("/*", handlers.NotImplementedHandler)

	router.Mount("/update", update)

	logger.Println("Server start on ", serverConf.ServerAddress)
	logger.Fatal(http.ListenAndServe(serverConf.ServerAddress, router))
}
