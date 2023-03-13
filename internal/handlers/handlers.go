// Модуль handlers содержит типы, методы и константы для
// API handlers
package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/hrapovd1/pmetrics/internal/config"
	dbstorage "github.com/hrapovd1/pmetrics/internal/dbstrorage"
	"github.com/hrapovd1/pmetrics/internal/filestorage"
	"github.com/hrapovd1/pmetrics/internal/storage"
	"github.com/hrapovd1/pmetrics/internal/types"
	"github.com/hrapovd1/pmetrics/internal/usecase"
	"github.com/hrapovd1/pmetrics/templates/core"
)

const (
	// минимально ожидаемая длина url метрики для POST
	minPathLen = 5
	// минимально ожидаемая длина url метрики для GET
	getPathLen = 4
)

// MetricsHandler тип обработчиков API
// содержит конфигурацию и хранилище
type MetricsHandler struct {
	Storage types.Repository
	Config  config.Config
	logger  *log.Logger
}

// NewMetricsHandler возвращает обработчик API
func NewMetricsHandler(conf config.Config, logger *log.Logger) *MetricsHandler {
	mh := &MetricsHandler{Config: conf, logger: logger}
	var fs *filestorage.FileStorage
	// Have mem, fs and db storage
	if mh.Config.StoreFile != "" && mh.Config.DatabaseDSN != "" {
		db, err := dbstorage.NewDBStorage(
			conf.DatabaseDSN,
			logger,
			filestorage.NewFileStorage(conf, make(map[string]interface{})),
		)
		if err != nil {
			logger.Fatal(err)
		}
		mh.Storage = db
	}
	// Have mem and db storage
	if mh.Config.DatabaseDSN != "" && mh.Config.StoreFile == "" {
		db, err := dbstorage.NewDBStorage(
			conf.DatabaseDSN,
			logger,
			storage.NewMemStorage(),
		)
		if err != nil {
			logger.Fatal(err)
		}
		mh.Storage = db
	}
	// Have mem and fs storage
	if mh.Config.StoreFile != "" && mh.Config.DatabaseDSN == "" {
		fs = filestorage.NewFileStorage(conf, make(map[string]interface{}))
		mh.Storage = fs
	}
	// Have mem storage
	if mh.Config.DatabaseDSN == "" && mh.Config.StoreFile == "" {
		ms := storage.NewMemStorage()
		mh.Storage = ms
	}
	return mh
}

// UpdateHandler POST обработчик обновления одной метрики в JSON формате
func (mh *MetricsHandler) UpdateHandler(rw http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()
	defer mh.logger.Println(r.Body.Close())
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	var data types.Metric
	if err := json.Unmarshal(body, &data); err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	// check metric hash in data.
	if mh.Config.Key != "" {
		if !usecase.IsSignEqual(data, mh.Config.Key) {
			http.Error(rw, "sign metric is bad", http.StatusBadRequest)
			return
		}
	}

	// Write new metrics value
	err = usecase.WriteJSONMetric(
		ctx,
		data,
		mh.Storage,
	)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	// Get metric value for response
	err = usecase.GetJSONMetric(ctx, mh.Storage, &data)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	// sign metric with hash in data.
	if mh.Config.Key != "" {
		err := usecase.SignData(&data, mh.Config.Key)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	resp, err := json.Marshal(data)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	rw.WriteHeader(http.StatusOK)
	_, err = rw.Write(resp)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}
}

// UpdatesHandler POST обработчик обновления нескольких метрик в JSON формате
func (mh *MetricsHandler) UpdatesHandler(rw http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()
	defer mh.logger.Println(r.Body.Close())
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	var data []types.Metric
	if err := json.Unmarshal(body, &data); err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	// check metric hash in data.
	if mh.Config.Key != "" {
		for _, item := range data {
			if !usecase.IsSignEqual(item, mh.Config.Key) {
				http.Error(rw, "sign metric is bad", http.StatusBadRequest)
				return
			}
		}
	}

	// Write new metrics value
	usecase.WriteJSONMetrics(
		ctx,
		&data,
		mh.Storage,
	)

	rw.WriteHeader(http.StatusOK)
	_, err = rw.Write([]byte(""))
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

}

// GetMetricJSONHandler GET обработчик чтения одной метрики в JSON формате
func (mh *MetricsHandler) GetMetricJSONHandler(rw http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()
	defer mh.logger.Println(r.Body.Close())
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	var data types.Metric
	if err := json.Unmarshal(body, &data); err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	// Get metric value for response
	if err = usecase.GetJSONMetric(ctx, mh.Storage, &data); err != nil {
		http.Error(rw, err.Error(), http.StatusNotFound)
		return
	}

	// sign metric with hash in data.
	if mh.Config.Key != "" {
		err := usecase.SignData(&data, mh.Config.Key)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	resp, err := json.Marshal(data)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(http.StatusOK)
	_, err = rw.Write(resp)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}
}

// GaugeHandler POST обработчик обновления gauge метрики в url формате
func (mh *MetricsHandler) GaugeHandler(rw http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()
	if r.Method != http.MethodPost {
		http.Error(rw, "Only POST requests are allowed.", http.StatusMethodNotAllowed)
		return
	}
	_, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	splitedPath := strings.Split(r.URL.Path, "/")
	if len(splitedPath) < minPathLen {
		errMsg := fmt.Sprint("URL - ", r.URL.Path, " - not found.")
		http.Error(rw, errMsg, http.StatusNotFound)
		return
	}

	err = usecase.WriteMetric(
		ctx,
		splitedPath,
		mh.Storage,
	)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	rw.WriteHeader(http.StatusOK)
	_, err = rw.Write([]byte(""))
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}
}

// CounterHandler POST обработчик обновления counter метрики в url формате
func (mh *MetricsHandler) CounterHandler(rw http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()
	if r.Method != http.MethodPost {
		http.Error(rw, "Only POST requests are allowed.", http.StatusMethodNotAllowed)
		return
	}
	_, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	splitedPath := strings.Split(r.URL.Path, "/")
	if len(splitedPath) < minPathLen {
		errMsg := fmt.Sprint("URL - ", r.URL.Path, " - not found.")
		http.Error(rw, errMsg, http.StatusNotFound)
		return
	}

	err = usecase.WriteMetric(
		ctx,
		splitedPath,
		mh.Storage,
	)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	rw.WriteHeader(http.StatusOK)
	_, err = rw.Write([]byte(""))
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}
}

// GetMetricHandler GET обработчик получения метрики в url формате
func (mh *MetricsHandler) GetMetricHandler(rw http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()
	if r.Method != http.MethodGet {
		http.Error(rw, "Only GET requests are allowed.", http.StatusMethodNotAllowed)
		return
	}

	splitedPath := strings.Split(r.URL.Path, "/")
	if len(splitedPath) < getPathLen {
		errMsg := fmt.Sprint("URL - ", r.URL.Path, " - not found.")
		http.Error(rw, errMsg, http.StatusNotFound)
		return
	}

	metricVal, err := usecase.GetMetric(ctx, mh.Storage, splitedPath)
	if err != nil {
		http.Error(rw, "Metric is't implemented yet.", http.StatusNotImplemented)
		return
	}
	if metricVal == "" {
		http.Error(rw, "Error when get metric", http.StatusNotFound)
		return
	}

	rw.Header().Set("Content-Type", "text/plain")
	rw.WriteHeader(http.StatusOK)
	_, err = rw.Write([]byte(metricVal))
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}
}

// GetAllHandler GET обработчик получения всех метрик в HTML формате
func (mh *MetricsHandler) GetAllHandler(rw http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()
	if r.Method != http.MethodGet {
		http.Error(rw, "Only GET requests are allowed.", http.StatusMethodNotAllowed)
		return
	}

	outTable := usecase.GetTableMetrics(ctx, mh.Storage)

	indexTmplt, err := template.ParseFS(core.Index, "index.html")
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	rw.Header().Set("Content-Type", "text/html")
	rw.WriteHeader(http.StatusOK)
	err = indexTmplt.Execute(rw, outTable)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}
}

// PingDB GET обработчик проверки доступности базы
func (mh *MetricsHandler) PingDB(rw http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()
	dbstor := mh.Storage.(types.Storager)
	if !dbstor.Ping(ctx) {
		http.Error(rw, "DB connect is NOT ok", http.StatusInternalServerError)
		return
	}
	rw.Header().Set("Content-Type", "text/html")
	rw.WriteHeader(http.StatusOK)
}

// NotImplementedHandler обработчик для ответа на не реализованные url
func NotImplementedHandler(rw http.ResponseWriter, r *http.Request) {
	rw.WriteHeader(http.StatusNotImplemented)
	_, err := rw.Write([]byte("It's not implemented yet."))
	if err != nil {
		http.Error(rw, err.Error(), http.StatusNotImplemented)
		return
	}
}
