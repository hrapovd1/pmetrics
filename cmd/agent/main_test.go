package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_pollMetrics(t *testing.T) {
	testMetrics := make(map[string]interface{})
	test := struct {
		name string
		args map[string]interface{}
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

func Test_metricsJSON(t *testing.T) {
	tests := []struct {
		name    string
		metrics map[string]interface{}
		want    []byte
	}{
		{
			name:    "Check gauge",
			metrics: map[string]interface{}{"M1": counter(345), "M2": gauge(63.689)},
			want:    []byte(`[{"id":"M1","type":"counter","delta":345},{"id":"M2","type":"gauge","value":63.689}]`),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := metricsToJSON(tt.metrics, "")
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
