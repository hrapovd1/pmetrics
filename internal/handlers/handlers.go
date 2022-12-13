package handlers

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"strings"

	"github.com/hrapovd1/pmetrics/internal/config"
	dbstorage "github.com/hrapovd1/pmetrics/internal/dbstrorage"
	"github.com/hrapovd1/pmetrics/internal/types"
	"github.com/hrapovd1/pmetrics/internal/usecase"
	"github.com/hrapovd1/pmetrics/templates/core"
)

const (
	minPathLen = 5
	getPathLen = 4
)

type MetricStorage struct {
	MemStor  types.Repository
	FileStor types.Repository
	DBStor   types.Repository
	Config   config.Config
}

func (ms *MetricStorage) UpdateHandler(rw http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
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
	if ms.Config.Key != "" {
		if !usecase.IsSignEqual(data, ms.Config.Key) {
			http.Error(rw, "sign metric is bad", http.StatusBadRequest)
			return
		}
	}

	// Write new metrics value
	err = usecase.WriteJSONMetric(
		data,
		ms.MemStor,
		ms.DBStor,
	)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	// Get metric value for response
	err = usecase.GetJSONMetric(ms.MemStor, &data)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	// sign metric with hash in data.
	if ms.Config.Key != "" {
		err := usecase.SignData(&data, ms.Config.Key)
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

func (ms *MetricStorage) UpdatesHandler(rw http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
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
	if ms.Config.Key != "" {
		for _, item := range data {
			if !usecase.IsSignEqual(item, ms.Config.Key) {
				http.Error(rw, "sign metric is bad", http.StatusBadRequest)
				return
			}
		}
	}

	// Write new metrics value
	err = usecase.WriteJSONMetrics(
		&data,
		ms.MemStor,
		ms.DBStor,
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

func (ms *MetricStorage) GetMetricJSONHandler(rw http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
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
	if err = usecase.GetJSONMetric(ms.MemStor, &data); err != nil {
		http.Error(rw, err.Error(), http.StatusNotFound)
		return
	}

	// sign metric with hash in data.
	if ms.Config.Key != "" {
		err := usecase.SignData(&data, ms.Config.Key)
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

func (ms *MetricStorage) GaugeHandler(rw http.ResponseWriter, r *http.Request) {
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
		splitedPath,
		ms.MemStor,
		ms.DBStor,
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

func (ms *MetricStorage) CounterHandler(rw http.ResponseWriter, r *http.Request) {
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
		splitedPath,
		ms.MemStor,
		ms.DBStor,
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

func (ms *MetricStorage) GetMetricHandler(rw http.ResponseWriter, r *http.Request) {
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

	metricVal, err := usecase.GetMetric(ms.MemStor, splitedPath)
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

func (ms *MetricStorage) GetAllHandler(rw http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(rw, "Only GET requests are allowed.", http.StatusMethodNotAllowed)
		return
	}

	outTable := usecase.GetTableMetrics(ms.MemStor)

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

func (ms *MetricStorage) PingDB(rw http.ResponseWriter, r *http.Request) {
	dbstor := ms.DBStor.(*dbstorage.DBStorage)
	if !dbstor.Ping() {
		http.Error(rw, "DB connect is NOT ok", http.StatusInternalServerError)
		return
	}
	rw.Header().Set("Content-Type", "text/html")
	rw.WriteHeader(http.StatusOK)
}

func NotImplementedHandler(rw http.ResponseWriter, r *http.Request) {
	rw.WriteHeader(http.StatusNotImplemented)
	_, err := rw.Write([]byte("It's not implemented yet."))
	if err != nil {
		http.Error(rw, err.Error(), http.StatusNotImplemented)
		return
	}
}
