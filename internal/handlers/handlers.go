package handlers

import (
	"embed"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"strings"

	"github.com/hrapovd1/pmetrics/internal/storage"
	"github.com/hrapovd1/pmetrics/internal/usecase"
)

//go:embed templates/index.html
var index embed.FS

const (
	minPathLen = 5
	getPathLen = 4
)

type MetricStorage struct {
	Storage storage.Repository
}

func NotImplementedHandler(rw http.ResponseWriter, r *http.Request) {
	rw.WriteHeader(http.StatusNotImplemented)
	_, err := rw.Write([]byte("It's not implemented yet."))
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

	err = usecase.WriteMetric(ms.Storage.(*storage.MemStorage), splitedPath)
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

	err = usecase.WriteMetric(ms.Storage.(*storage.MemStorage), splitedPath)
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

	metricVal, err := usecase.GetMetric(ms.Storage.(*storage.MemStorage), splitedPath)
	if metricVal == "" {
		http.Error(rw, "Error when get metric", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(rw, "Metric is't implemented yet.", http.StatusNotImplemented)
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

	outTable := usecase.GetTableMetrics(ms.Storage.(*storage.MemStorage))

	indexTmplt, err := template.ParseFS(index, "templates/index.html")
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	rw.WriteHeader(http.StatusOK)
	err = indexTmplt.Execute(rw, outTable)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}
}
