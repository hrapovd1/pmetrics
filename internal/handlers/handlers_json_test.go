package handlers

import (
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/hrapovd1/pmetrics/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMetricsHandler_UpdateHandler(t *testing.T) {
	tests := []struct {
		name        string
		data        string
		contentType string
		key         string
		statusCode  int
	}{
		{
			name:        "Alloc1",
			data:        `{"id":"Alloc1","type":"gauge","value":-4.5}`,
			contentType: "application/json",
			key:         "",
			statusCode:  http.StatusOK,
		},
		{
			name:        "Count1",
			data:        `{"id":"Count1","type":"counter","delta":5}`,
			contentType: "application/json",
			key:         "",
			statusCode:  http.StatusOK,
		},
		{
			name:        "Empty data",
			data:        `{}`,
			contentType: "application/json",
			key:         "",
			statusCode:  http.StatusBadRequest,
		},
		{
			name:        "Bad data",
			data:        `{"}`,
			contentType: "application/json",
			key:         "",
			statusCode:  http.StatusInternalServerError,
		},
		{
			name:        "Sign data",
			data:        `{"id":"Count1","type":"counter","delta":5}`,
			contentType: "application/json",
			key:         "1234rewq",
			statusCode:  http.StatusBadRequest,
		},
	}

	ms := MetricsHandler{
		Storage: storage.NewMemStorage(),
		logger:  log.New(os.Stderr, "test", log.Default().Flags()),
	}

	for _, test := range tests {
		ms.Config.Key = test.key
		reqst := httptest.NewRequest(http.MethodPost, "/update/", strings.NewReader(test.data))
		reqst.Header.Set("Content-Type", test.contentType)
		rec := httptest.NewRecorder()
		hndl := http.HandlerFunc(ms.UpdateHandler)
		// qeury server
		hndl.ServeHTTP(rec, reqst)

		t.Run(test.name, func(t *testing.T) {
			result := rec.Result()
			defer assert.Nil(t, result.Body.Close())
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

func TestMetricsHandler_UpdatesHandler(t *testing.T) {
	tests := []struct {
		name        string
		data        string
		contentType string
		key         string
		statusCode  int
	}{
		{
			name:        "Alloc",
			data:        `[{"id":"Alloc1","type":"gauge","value":-4.5}]`,
			key:         "1234rewq",
			contentType: "application/json",
			statusCode:  http.StatusBadRequest,
		},
		{
			name:        "Count1",
			data:        `[{"id":"Count1","type":"counter","delta":5}]`,
			key:         "",
			contentType: "application/json",
			statusCode:  http.StatusOK,
		},
		{
			name:        "Empty data",
			data:        `{}`,
			key:         "",
			contentType: "application/json",
			statusCode:  http.StatusInternalServerError,
		},
	}

	ms := MetricsHandler{
		Storage: storage.NewMemStorage(),
		logger:  log.New(os.Stderr, "test", log.Default().Flags()),
	}

	for _, test := range tests {
		ms.Config.Key = test.key
		reqst := httptest.NewRequest(http.MethodPost, "/updates/", strings.NewReader(test.data))
		reqst.Header.Set("Content-Type", test.contentType)
		rec := httptest.NewRecorder()
		hndl := http.HandlerFunc(ms.UpdatesHandler)
		// qeury server
		hndl.ServeHTTP(rec, reqst)

		t.Run(test.name, func(t *testing.T) {
			result := rec.Result()
			defer assert.Nil(t, result.Body.Close())
			_, err := io.ReadAll(result.Body)
			assert.Nil(t, err)
			assert.Equal(t, test.statusCode, result.StatusCode)
		})
	}
}

func TestMetricsHandler_GetMetricJSONHandler(t *testing.T) {
	stor := make(map[string]interface{})
	locStorage := storage.NewMemStorage(storage.WithBuffer(stor))
	stor["PollCount"] = int64(4)
	stor["Sys"] = float64(0.0)
	ms := MetricsHandler{
		Storage: locStorage,
		logger:  log.New(os.Stderr, "test", log.Default().Flags()),
	}

	tests := []struct {
		name        string
		data        string
		contentType string
		key         string
		statusCode  int
	}{
		{
			name:        "PollCount",
			data:        `{"id":"PollCount","type":"counter","delta":4}`,
			contentType: "application/json",
			key:         "",
			statusCode:  http.StatusOK,
		},
		{
			name:        "Sys",
			data:        `{"id":"Sys","type":"gauge","value":0,"hash":"4f289c7282738ac5473752e362b02d41c598f7424fff27e8f0563fea6a183ac9"}`,
			contentType: "application/json",
			key:         "1234rewq",
			statusCode:  http.StatusOK,
		},
		{
			name:        "Empty data",
			data:        `{}`,
			contentType: "application/json",
			key:         "",
			statusCode:  http.StatusNotFound,
		},
		{
			name:        "Wrong data",
			data:        `{"}`,
			contentType: "application/json",
			key:         "",
			statusCode:  http.StatusInternalServerError,
		},
	}
	hndl := http.HandlerFunc(ms.GetMetricJSONHandler)

	for _, test := range tests {
		ms.Config.Key = test.key
		t.Run(test.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			reqst := httptest.NewRequest(http.MethodPost, "/value/", strings.NewReader(test.data))
			// qeury server
			hndl.ServeHTTP(rec, reqst)
			result := rec.Result()
			defer assert.Nil(t, result.Body.Close())
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
