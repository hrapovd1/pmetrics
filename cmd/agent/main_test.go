package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_pollMetrics(t *testing.T) {
	testMetrics := make(map[string]gauge)
	test := struct {
		name string
		args map[string]gauge
		want []string
	}{
		name: "Check metric names",
		args: testMetrics,
		want: []string{
			"Alloc",
			"TotalAlloc",
			"Sys",
			"Lookups",
			"Mallocs",
			"Frees",
			"HeapAlloc",
			"HeapSys",
			"HeapIdle",
			"HeapInuse",
			"HeapReleased",
			"HeapObjects",
			"StackInuse",
			"StackSys",
			"MSpanInuse",
			"MSpanSys",
			"MCacheInuse",
			"MCacheSys",
			"BuckHashSys",
			"GCSys",
			"OtherSys",
			"NextGC",
			"LastGC",
			"PauseTotalNs",
			"NumGC",
			"NumForcedGC",
			"GCCPUFraction",
			"RandomValue",
		},
	}
	t.Run(test.name, func(t *testing.T) {
		pollMetrics(test.args)
		for _, val := range test.want {
			_, ok := test.args[val]
			assert.True(t, ok)
		}
	})
}

func Test_reportClient(t *testing.T) {
	metrics := map[string]gauge{
		"Alloc":      gauge(45.9),
		"TotalAlloc": gauge(40),
		"Sys":        gauge(0),
		"Lookups":    gauge(-0),
	}

	want := map[string]string{
		"Alloc":      "/update/gauge/Alloc/45.9",
		"TotalAlloc": "/update/gauge/TotalAlloc/40",
		"Sys":        "/update/gauge/Sys/0",
		"Lookups":    "/update/gauge/Lookups/0",
	}

	var (
		//logBuf        bytes.Buffer
		logger        = log.New(os.Stdout, "", log.Lshortfile)
		serverRequest *http.Request
	)

	testServer := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			serverRequest = r
			fmt.Fprintln(w, "")
		}),
	)
	defer testServer.Close()

	for key, val := range metrics {
		t.Run(key, func(t *testing.T) {
			metricURL := fmt.Sprint(testServer.URL, "/update/gauge/", key, "/", val)
			reportClient(&http.Client{}, metricURL, logger)
			assert.Equal(t, want[key], serverRequest.URL.Path)
		})
	}
}
