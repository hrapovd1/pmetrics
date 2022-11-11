package handlers

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/hrapovd1/pmetrics/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMetricStorage_UpdateHandler(t *testing.T) {
	tests := []struct {
		name        string
		data        string
		contentType string
		statusCode  int
	}{
		{
			name:        "Alloc1",
			data:        `{"id":"Alloc1","type":"gauge","value":-4.5}`,
			contentType: "application/json",
			statusCode:  http.StatusOK,
		},
		{
			name:        "Count1",
			data:        `{"id":"Count1","type":"counter","delta":5}`,
			contentType: "application/json",
			statusCode:  http.StatusOK,
		},
		{
			name:        "Empty data",
			data:        `{}`,
			contentType: "application/json",
			statusCode:  http.StatusBadRequest,
		},
	}

	ms := MetricStorage{
		Storage: storage.NewMemStorage(),
	}

	for _, test := range tests {
		reqst := httptest.NewRequest(http.MethodPost, "/update/", strings.NewReader(test.data))
		reqst.Header.Set("Content-Type", test.contentType)
		rec := httptest.NewRecorder()
		hndl := http.HandlerFunc(ms.UpdateHandler)
		// qeury server
		hndl.ServeHTTP(rec, reqst)

		t.Run(test.name, func(t *testing.T) {
			result := rec.Result()
			defer result.Body.Close()
			body, err := io.ReadAll(result.Body)
			assert.Nil(t, err)
			if test.statusCode == http.StatusOK {
				assert.Equal(t, []byte(test.data), body)
			} else {
				assert.Equal(t, test.statusCode, result.StatusCode)
			}
		})
	}
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
	ms := MetricStorage{
		Storage: locStorage,
	}

	tests := []struct {
		name        string
		data        string
		contentType string
		statusCode  int
	}{
		{
			name:        "PollCount",
			data:        `{"id":"PollCount","type":"counter","delta":4}`,
			contentType: "application/json",
			statusCode:  http.StatusOK,
		},
		{
			name:        "Sys",
			data:        `{"id":"Sys","type":"gauge","value":0}`,
			contentType: "application/json",
			statusCode:  http.StatusOK,
		},
		{
			name:        "Wrong data",
			data:        `{}`,
			contentType: "application/json",
			statusCode:  http.StatusBadRequest,
		},
	}
	hndl := http.HandlerFunc(ms.GetMetricHandler)

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			reqst := httptest.NewRequest(http.MethodPost, "/value/", strings.NewReader(test.data))
			// qeury server
			hndl.ServeHTTP(rec, reqst)
			result := rec.Result()
			defer result.Body.Close()
			body, err := io.ReadAll(result.Body)
			require.NoError(t, err)

			if test.statusCode == http.StatusOK {
				assert.Equal(t, test.contentType, result.Header.Get("Content-Type"))
				assert.Equal(t, []byte(test.data), body)
			} else {
				assert.Equal(t, test.statusCode, result.StatusCode)
			}
		})
	}
}
