package handlers

import (
	"embed"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"strings"

	"github.com/hrapovd1/pmetrics/internal/storage"
)

//go:embed templates/index.html
var index embed.FS

const (
	metricName    = 3
	metricVal     = 4
	minPathLen    = 5
	getMetricType = 2
	getMetricName = 3
	getPathLen    = 4
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

	metricKey := splitedPath[metricName]
	metricValue, err := storage.StrToGauge(splitedPath[metricVal])
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}
	ms.Storage.Rewrite(metricKey, metricValue)

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

	metricKey := splitedPath[metricName]
	metricValue, err := storage.StrToCounter(splitedPath[metricVal])
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	ms.Storage.Append(metricKey, metricValue)

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

	metricType := splitedPath[getMetricType]
	metric := splitedPath[getMetricName]
	var metricValue string

	if metricType == "gauge" || metricType == "counter" {
		metricVal := ms.Storage.Get(metric)
		if metricVal == nil {
			errMsg := fmt.Sprint("Error when get ", metric)
			http.Error(rw, errMsg, http.StatusNotFound)
			return
		}
		switch metricVal := metricVal.(type) {
		case int64:
			metricValue = fmt.Sprint(metricVal)
		case float64:
			metricValue = fmt.Sprint(metricVal)
		}
	} else {
		http.Error(rw, "Metric is't implemented yet.", http.StatusNotImplemented)
		return
	}

	rw.Header().Set("Content-Type", "text/plain")
	rw.WriteHeader(http.StatusOK)
	_, err := rw.Write([]byte(metricValue))
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

	type table struct {
		Key string
		Val string
	}
	outTable := make([]table, 0)

	for k, v := range ms.Storage.GetAll() {
		switch value := v.(type) {
		case int64:
			outTable = append(outTable, table{Key: k, Val: fmt.Sprint(value)})
		case float64:
			outTable = append(outTable, table{Key: k, Val: fmt.Sprint(value)})
		}
	}

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
