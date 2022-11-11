package main

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func Test_metricToJSON(t *testing.T) {
	type value struct {
		name    string
		isGauge bool
		val     gauge
		del     counter
	}
	tests := []struct {
		name   string
		metric value
		want   []byte
	}{
		{
			name:   "Check gauge",
			metric: value{name: "Test1", isGauge: true, val: gauge(-45.8)},
			want:   []byte(`{"id":"Test1","type":"gauge","value":-45.8}`),
		},
		{
			name:   "Check counter",
			metric: value{name: "Test2", isGauge: false, del: counter(8)},
			want:   []byte(`{"id":"Test2","type":"counter","delta":8}`),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.metric.isGauge {
				got, err := metricToJSON(tt.metric.name, tt.metric.val)
				require.NoError(t, err)
				assert.True(t, reflect.DeepEqual(got, tt.want))
			} else {
				got, err := metricToJSON(tt.metric.name, tt.metric.del)
				require.NoError(t, err)
				assert.True(t, reflect.DeepEqual(got, tt.want))
			}
		})
	}
}
