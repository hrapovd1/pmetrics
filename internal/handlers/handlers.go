package handlers

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/hrapovd1/pmetrics/internal/storage"
)

const (
	metricName = 3
	metricVal  = 4
	minPathLen = 5
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
	var splitedPath []string
	if r.Method != http.MethodPost {
		http.Error(rw, "Only POST requests are allowed.", http.StatusMethodNotAllowed)
		return
	}
	_, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	splitedPath = strings.Split(r.URL.Path, "/")
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

	if metricKey == "PollCount" {
		ms.Storage.Append(metricValue)
	}

	rw.WriteHeader(http.StatusOK)
	_, err = rw.Write([]byte(""))
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}
}
