package main

import (
	"context"
	"log"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_pollHwMetrics(t *testing.T) {
	testMetrics := make(map[string]interface{})
	test := struct {
		name string
		args mmetrics
		want []string
	}{
		name: "Check metric names",
		args: mmetrics{pollCounter: counter(0), mtrcs: testMetrics},
		want: []string{},
	}
	t.Run(test.name, func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Microsecond*600)
		defer cancel()
		pollHwMetrics(ctx, &test.args, time.Microsecond*500, log.New(os.Stdout, "AGENT\t", log.Ldate|log.Ltime))
		for _, val := range test.want {
			_, ok := test.args.mtrcs[val]
			assert.True(t, ok)
		}
	})
}

func Test_pollMetrics(t *testing.T) {
	testMetrics := make(map[string]interface{})
	test := struct {
		name string
		args mmetrics
		want []string
	}{
		name: "Check metric names",
		args: mmetrics{pollCounter: counter(0), mtrcs: testMetrics},
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
		ctx, cancel := context.WithTimeout(context.Background(), time.Microsecond*600)
		defer cancel()
		pollMetrics(ctx, &test.args, time.Microsecond*500)
		for _, val := range test.want {
			_, ok := test.args.mtrcs[val]
			assert.True(t, ok)
		}
	})
}

func Test_metricsJSON(t *testing.T) {
	tests := []struct {
		name    string
		metrics map[string]interface{}
		want    []byte
		wantn   []byte
	}{
		{
			name:    "Check gauge",
			metrics: map[string]interface{}{"M1": counter(345), "M2": gauge(63.689)},
			want:    []byte(`[{"id":"M1","type":"counter","delta":345},{"id":"M2","type":"gauge","value":63.689}]`),
			wantn:   []byte(`[{"id":"M2","type":"gauge","value":63.689},{"id":"M1","type":"counter","delta":345}]`),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := metricsToJSON(tt.metrics, "")
			require.NoError(t, err)
			assert.True(
				t,
				reflect.DeepEqual(tt.want, got) || reflect.DeepEqual(tt.wantn, got),
			)
		})
	}
}
