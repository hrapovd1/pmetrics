package handlers

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hrapovd1/pmetrics/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMetricStorage_GaugeHandler(t *testing.T) {
	tests := []struct {
		name        string
		path        string
		method      string
		contentType string
		metric      string
		want        float64
		statusCode  int
	}{
		{
			name:        "Alloc1",
			path:        "/update/gauge/Alloc/34.9",
			method:      http.MethodPost,
			contentType: "text/plain",
			metric:      "Alloc",
			want:        float64(34.9),
			statusCode:  http.StatusOK,
		},
		{
			name:        "Alloc2",
			path:        "/update/gauge/Alloc/0",
			method:      http.MethodPost,
			contentType: "text/plain",
			metric:      "Alloc",
			want:        float64(0.0),
			statusCode:  http.StatusOK,
		},
		{
			name:        "Wrong Method",
			path:        "/update/gauge/Alloc/0",
			method:      http.MethodGet,
			contentType: "text/plain",
			metric:      "Alloc",
			want:        float64(0.0),
			statusCode:  http.StatusMethodNotAllowed,
		},
		{
			name:        "Wrong Content",
			path:        "/update/gauge/Alloc/0",
			method:      http.MethodPost,
			contentType: "",
			metric:      "Alloc",
			want:        float64(0.0),
			statusCode:  http.StatusUnsupportedMediaType,
		},
		{
			name:        "Wrong URL",
			path:        "/update/gauge/Alloc/a",
			method:      http.MethodPost,
			contentType: "text/plain",
			metric:      "Alloc",
			want:        float64(0.0),
			statusCode:  http.StatusBadRequest,
		},
	}

	locStorage := storage.NewMemStorage()
	ms := MetricStorage{
		Storage: locStorage,
	}

	for _, test := range tests {
		reqst := httptest.NewRequest(test.method, test.path, nil)
		reqst.Header.Set("Content-Type", test.contentType)
		rec := httptest.NewRecorder()
		hndl := http.HandlerFunc(ms.GaugeHandler)
		// qeury server
		hndl.ServeHTTP(rec, reqst)

		t.Run(test.name, func(t *testing.T) {
			result := rec.Result()
			defer result.Body.Close()
			_, err := io.ReadAll(result.Body)
			assert.Nil(t, err)
			if test.statusCode == http.StatusOK {
				assert.Equal(t, test.want, float64(locStorage.GaugeBuff[test.metric]))
			} else {
				assert.Equal(t, test.statusCode, result.StatusCode)
			}
		})
	}
	t.Run("Check values count", func(t *testing.T) {
		assert.Equal(t, 1, len(locStorage.GaugeBuff))
	})
}

func TestMetricStorage_CounterHandler(t *testing.T) {
	tests := []struct {
		name        string
		path        string
		method      string
		contentType string
		want        int64
		statusCode  int
	}{
		{
			name:        "PollCount1",
			path:        "/update/counter/PollCount/34",
			method:      http.MethodPost,
			contentType: "text/plain",
			want:        int64(34),
			statusCode:  http.StatusOK,
		},
		{
			name:        "PollCount2",
			path:        "/update/counter/PollCount/1",
			method:      http.MethodPost,
			contentType: "text/plain",
			want:        int64(1),
			statusCode:  http.StatusOK,
		},
		{
			name:        "Wrong Method",
			path:        "/update/counter/PollCount/1",
			method:      http.MethodGet,
			contentType: "text/plain",
			want:        int64(1),
			statusCode:  http.StatusMethodNotAllowed,
		},
		{
			name:        "Wrong Content",
			path:        "/update/counter/PollCount/1",
			method:      http.MethodPost,
			contentType: "",
			want:        int64(1),
			statusCode:  http.StatusUnsupportedMediaType,
		},
		{
			name:        "Wrong URL",
			path:        "/update/counter/PollCount/a",
			method:      http.MethodPost,
			contentType: "text/plain",
			want:        int64(1),
			statusCode:  http.StatusBadRequest,
		},
	}

	locStorage := storage.NewMemStorage()
	ms := MetricStorage{
		Storage: locStorage,
	}

	for _, test := range tests {
		reqst := httptest.NewRequest(test.method, test.path, nil)
		reqst.Header.Set("Content-Type", test.contentType)
		rec := httptest.NewRecorder()
		hndl := http.HandlerFunc(ms.CounterHandler)
		// qeury server
		hndl.ServeHTTP(rec, reqst)

		t.Run(test.name, func(t *testing.T) {
			result := rec.Result()
			defer result.Body.Close()
			_, err := io.ReadAll(result.Body)
			assert.Nil(t, err)
			if test.statusCode == http.StatusOK {
				require.NotZero(t, len(locStorage.CounterBuff))
				last := len(locStorage.CounterBuff) - 1
				assert.Equal(t, test.want, int64(locStorage.CounterBuff[last]))
			} else {
				assert.Equal(t, test.statusCode, result.StatusCode)
			}
		})
	}
	t.Run("Check values count", func(t *testing.T) {
		assert.Equal(t, 2, len(locStorage.CounterBuff))
	})
}
