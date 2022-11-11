package handlers

import (
	"encoding/json"
	"html/template"
	"io"
	"net/http"

	"github.com/hrapovd1/pmetrics/internal/storage"
	"github.com/hrapovd1/pmetrics/internal/types"
	"github.com/hrapovd1/pmetrics/internal/usecase"
	"github.com/hrapovd1/pmetrics/templates/core"
)

type MetricStorage struct {
	Storage storage.Repository
}

func (ms *MetricStorage) UpdateHandler(rw http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	var data types.Metrics
	if err := json.Unmarshal(body, &data); err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	// Write new metrics value
	err = usecase.WriteMetric(ms.Storage.(*storage.MemStorage), data)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	// Get metric value for response
	err = usecase.GetMetric(ms.Storage.(*storage.MemStorage), &data)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
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

func (ms *MetricStorage) GetMetricHandler(rw http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	var data types.Metrics
	if err := json.Unmarshal(body, &data); err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	// Get metric value for response
	err = usecase.GetMetric(ms.Storage.(*storage.MemStorage), &data)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
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

func (ms *MetricStorage) GetAllHandler(rw http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(rw, "Only GET requests are allowed.", http.StatusMethodNotAllowed)
		return
	}

	outTable := usecase.GetTableMetrics(ms.Storage.(*storage.MemStorage))

	indexTmplt, err := template.ParseFS(core.Index, "index.html")
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
