package main

import (
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
