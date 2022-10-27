package handlers

import (
	"net/http"
	"testing"

	"github.com/hrapovd1/pmetrics/internal/storage"
)

func TestMetricStorage_GaugeHandler(t *testing.T) {
	type args struct {
		rw http.ResponseWriter
		r  *http.Request
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	ms := MetricStorage{
		Storage: storage.NewMemStorage(),
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ms.GaugeHandler(tt.args.rw, tt.args.r)
		})
	}
}
