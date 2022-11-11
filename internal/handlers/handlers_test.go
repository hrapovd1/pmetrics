package handlers

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
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
			name:        "Wrong URL",
			path:        "/update/gauge/Alloc/a",
			method:      http.MethodPost,
			contentType: "text/plain",
			metric:      "Alloc",
			want:        float64(0.0),
			statusCode:  http.StatusBadRequest,
		},
	}

	stor := make(map[string]interface{})
	locStorage := storage.NewMemStorage(storage.WithBuffer(stor))
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
				metric := locStorage.Get(test.metric)
				require.NotNil(t, metric)
				assert.Equal(t, test.want, metric.(float64))
			} else {
				assert.Equal(t, test.statusCode, result.StatusCode)
			}
		})
	}
	t.Run("Check values count", func(t *testing.T) {
		assert.Equal(t, 1, len(stor))
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
			want:        int64(35),
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
			name:        "Wrong URL",
			path:        "/update/counter/PollCount/a",
			method:      http.MethodPost,
			contentType: "text/plain",
			want:        int64(1),
			statusCode:  http.StatusBadRequest,
		},
	}

	stor := make(map[string]interface{})
	locStorage := storage.NewMemStorage(storage.WithBuffer(stor))
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
			log.Println(test.name, " ", test.path)
			result := rec.Result()
			defer result.Body.Close()
			_, err := io.ReadAll(result.Body)
			assert.Nil(t, err)
			if test.statusCode == http.StatusOK {
				pollCount := locStorage.Get("PollCount")
				require.NotNil(t, pollCount)
				assert.Equal(t, test.want, pollCount.(int64))
			} else {
				assert.Equal(t, test.statusCode, result.StatusCode)
			}
		})
	}
	t.Run("Check values count", func(t *testing.T) {
		assert.Equal(t, 1, len(stor))
	})
}

func TestMetricStorage_GetAllHandler(t *testing.T) {
	stor := make(map[string]interface{})
	locStorage := storage.NewMemStorage(storage.WithBuffer(stor))
	stor["Sys"] = float64(0.0)
	stor["Alloc"] = float64(3.0)
	stor["TotalAlloc"] = float64(-3.0)
	ms := MetricStorage{
		Storage: locStorage,
	}

	reqst := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	hndl := http.HandlerFunc(ms.GetAllHandler)
	// qeury server
	hndl.ServeHTTP(rec, reqst)
	result := rec.Result()
	body, err := io.ReadAll(result.Body)
	assert.Nil(t, err)
	defer result.Body.Close()
	statusCode := result.StatusCode

	t.Run("Status Code", func(t *testing.T) {
		assert.Equal(t, statusCode, http.StatusOK)
	})

	val1 := strings.Split(string(body), "<tr><td>")
	val1 = val1[1:]
	values := make([]string, 0)
	for _, val := range val1 {
		for _, val := range strings.Split(val, "</td><td>") {
			for _, val := range strings.Split(val, "</td></tr>") {
				if val != "" {
					values = append(values, val)
				}
			}
		}
	}
	values = values[:len(values)-1]

	for k, v := range stor {
		t.Run(k, func(t *testing.T) {
			for i, val := range values {
				if val == k {
					want := fmt.Sprint(v)
					assert.Equal(t, want, values[i+1])
				}
			}

		})

	}
}

func TestMetricStorage_GetMetricHandler(t *testing.T) {
	stor := make(map[string]interface{})
	locStorage := storage.NewMemStorage(storage.WithBuffer(stor))
	stor["PollCount"] = int64(4)
	stor["Sys"] = float64(0.0)
	stor["Alloc"] = float64(3.0)
	stor["TotalAlloc"] = float64(-3.1)
	ms := MetricStorage{
		Storage: locStorage,
	}

	tests := []struct {
		name     string
		positive bool
		url      string
		want     string
	}{
		{
			name:     "PollCount",
			positive: true,
			url:      "/value/counter/PollCount",
			want:     "4",
		},
		{
			name:     "Sys",
			positive: true,
			url:      "/value/gauge/Sys",
			want:     "0",
		},
		{
			name:     "Alloc",
			positive: true,
			url:      "/value/gauge/Alloc",
			want:     "3",
		},
		{
			name:     "TotalAlloc",
			positive: true,
			url:      "/value/gauge/TotalAlloc",
			want:     "-3.1",
		},
		{
			name:     "Wrong url",
			positive: false,
			url:      "/value/gauge/",
			want:     "-3.1",
		},
	}
	hndl := http.HandlerFunc(ms.GetMetricHandler)

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			reqst := httptest.NewRequest(http.MethodGet, test.url, nil)
			// qeury server
			hndl.ServeHTTP(rec, reqst)
			result := rec.Result()
			defer result.Body.Close()
			body, err := io.ReadAll(result.Body)
			require.NoError(t, err)

			if test.positive {
				assert.Equal(t, http.StatusOK, result.StatusCode)
				assert.Equal(t, test.want, string(body))
			} else {
				assert.NotEqual(t, http.StatusOK, result.StatusCode)
			}
		})
	}
}
