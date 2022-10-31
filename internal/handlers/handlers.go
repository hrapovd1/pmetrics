package handlers

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/hrapovd1/pmetrics/internal/storage"
)

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
	log.Println("gauge splitedPath = ", splitedPath)
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
	log.Println("counter splitedPath = ", splitedPath)
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

func (ms *MetricStorage) GetMetricHandler(rw http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(rw, "Only GET requests are allowed.", http.StatusMethodNotAllowed)
		return
	}

	splitedPath := strings.Split(r.URL.Path, "/")
	log.Println("value splitedPath = ", splitedPath)
	if len(splitedPath) < getPathLen {
		errMsg := fmt.Sprint("URL - ", r.URL.Path, " - not found.")
		http.Error(rw, errMsg, http.StatusNotFound)
		return
	}

	metricType := splitedPath[getMetricType]
	metric := splitedPath[getMetricName]
	var metricValue string

	if (metricType == "gauge" && metric == "PollCount") ||
		(metricType == "counter" && metric != "PollCount") {
		errMsg := fmt.Sprint("Wrong type ", metricType, " for ", metric)
		http.Error(rw, errMsg, http.StatusBadRequest)
		return
	}
	if metricType == "gauge" || metricType == "counter" {
		metricVal, ok := ms.Storage.Get(metric)
		if !ok {
			errMsg := fmt.Sprint("Error when get ", metric)
			http.Error(rw, errMsg, http.StatusInternalServerError)
			return
		}
		if metric == "PollCount" {
			metricValue = fmt.Sprint(metricVal.(int64))
		} else {
			metricValue = fmt.Sprint(metricVal.(float64))
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

	outPage := []string{
		"<html>",
		"<head><title> All metrics </title></head>",
		"<body>",
		"<table><thead><tr>",
		"<th>Metric name</th><th>Metric value</th>",
		"</tr></thead><tbody>",
	}

	for k, v := range ms.Storage.GetAll() {
		var outString string
		if k == "PollCount" {
			pollCount, ok := v.([]int64)
			if !ok {
				http.Error(rw, "", http.StatusInternalServerError)
				return
			}
			if len(pollCount) > 0 {
				last := len(pollCount) - 1
				outString = fmt.Sprint("<tr><td>", k, "</td><td>", pollCount[last], "</td></tr>")
			}
		} else {
			val := v.(float64)
			outString = fmt.Sprint("<tr><td>", k, "</td><td>", val, "</td></tr>")
		}
		outPage = append(
			outPage,
			outString,
		)
	}

	outPage = append(
		outPage,
		"</tbody></table></body></html>",
	)

	rw.Header().Set("Content-Type", "text/plain")
	rw.WriteHeader(http.StatusOK)
	_, err := rw.Write([]byte(strings.Join(outPage, "")))
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}
}
