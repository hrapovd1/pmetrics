package handlers

import (
	"io"
	"net/http"
	"strings"

	"github.com/hrapovd1/pmetrics/internal/storage"
)

const (
	metricName = 2
	metricVal  = 3
)

type MetricStorage struct {
	Storage storage.Repository
}

func (ms MetricStorage) GaugeHandler(rw http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(rw, "Only POST requests are allowed.", http.StatusMethodNotAllowed)
		return
	}
	if r.Header.Get("Content-Type") != "text/plain" {
		http.Error(rw, "Content-type must be text/plain", http.StatusUnsupportedMediaType)
		return
	}
	_, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	splitedPath := strings.Split(r.URL.Path, "/")
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

func (ms MetricStorage) CounterHandler(rw http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(rw, "Only POST requests are allowed.", http.StatusMethodNotAllowed)
		return
	}
	if r.Header.Get("Content-Type") != "text/plain" {
		http.Error(rw, "Content-type must be text/plain", http.StatusUnsupportedMediaType)
		return
	}
	_, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	splitedPath := strings.Split(r.URL.Path, "/")
	metricValue, err := storage.StrToCounter(splitedPath[3])
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}
	ms.Storage.Append(metricValue)

	rw.WriteHeader(http.StatusOK)
	_, err = rw.Write([]byte(""))
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}
}
