package usecase

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hrapovd1/pmetrics/internal/storage"
	"github.com/hrapovd1/pmetrics/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWriteJSONMetric(t *testing.T) {
	M1 := int64(5)
	M2 := float64(-4.65)
	tests := []struct {
		name string
		data types.Metric
		want string
	}{
		{
			name: "M1",
			data: types.Metric{ID: "M1", MType: "counter", Delta: &M1},
			want: "5",
		},
		{
			name: "M2",
			data: types.Metric{ID: "M2", MType: "gauge", Value: &M2},
			want: "-4.65",
		},
	}
	stor := make(map[string]interface{})
	ctx := context.Background()
	locStorage := storage.NewMemStorage(storage.WithBuffer(stor))
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := WriteJSONMetric(ctx, tt.data, locStorage)
			require.NoError(t, err)
			switch result := locStorage.Get(ctx, tt.data.ID).(type) {
			case int64:
				assert.Equal(t, tt.want, fmt.Sprint(result))
			case float64:
				assert.Equal(t, tt.want, fmt.Sprint(result))
			}
		})
	}
}

func TestGetJSONMetric(t *testing.T) {
	tests := []struct {
		name    string
		data    types.Metric
		withErr bool
		want    string
	}{
		{
			name:    "M1",
			data:    types.Metric{ID: "M1", MType: "counter"},
			withErr: false,
			want:    "5",
		},
		{
			name:    "M2",
			data:    types.Metric{ID: "M2", MType: "gauge"},
			withErr: false,
			want:    "-4.65",
		},
		{
			name:    "M3",
			data:    types.Metric{ID: "M3", MType: "type"},
			withErr: true,
			want:    "<nil>",
		},
	}
	stor := make(map[string]interface{})
	stor["M1"] = int64(5)
	stor["M2"] = float64(-4.65)
	locStorage := storage.NewMemStorage(storage.WithBuffer(stor))
	ctx := context.Background()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := GetJSONMetric(ctx, locStorage, &tt.data)
			if tt.withErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			assert.Equal(t, tt.want, fmt.Sprint(stor[tt.data.ID]))
		})
	}
}

func TestWriteMetric(t *testing.T) {
	tests := []struct {
		name       string
		path       []string
		metricName string
		want       string
	}{
		{
			name:       "M1",
			path:       []string{"", "update", "counter", "M1", "5"},
			metricName: "M1",
			want:       "5",
		},
		{
			name:       "M2",
			path:       []string{"", "update", "gauge", "M2", "0"},
			metricName: "M2",
			want:       "0",
		},
		{
			name:       "M1_1",
			path:       []string{"", "update", "counter", "M1", "3"},
			metricName: "M1",
			want:       "8",
		},
		{
			name:       "M2_1",
			path:       []string{"", "update", "gauge", "M2", "-3.3"},
			metricName: "M2",
			want:       "-3.3",
		},
	}
	stor := make(map[string]interface{})
	locStorage := storage.NewMemStorage(storage.WithBuffer(stor))
	ctx := context.Background()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := WriteMetric(ctx, tt.path, locStorage)
			require.NoError(t, err)
			switch result := locStorage.Get(ctx, tt.metricName).(type) {
			case int64:
				assert.Equal(t, tt.want, fmt.Sprint(result))
			case float64:
				assert.Equal(t, tt.want, fmt.Sprint(result))
			}
		})
	}
}

func TestGetMetric(t *testing.T) {
	tests := []struct {
		name       string
		path       []string
		metricName string
		withErr    bool
		want       string
	}{
		{
			name:       "M1",
			path:       []string{"", "value", "counter", "M1"},
			metricName: "M1",
			withErr:    false,
			want:       "5",
		},
		{
			name:       "M2",
			path:       []string{"", "update", "gauge", "M2"},
			metricName: "M2",
			withErr:    false,
			want:       "0",
		},
		{
			name:       "M1_1",
			path:       []string{"", "update", "simple", "M1"},
			metricName: "M1",
			withErr:    true,
			want:       "",
		},
		{
			name:       "M3",
			path:       []string{"", "update", "gauge", "M3"},
			metricName: "M3",
			withErr:    false,
			want:       "",
		},
	}
	stor := make(map[string]interface{})
	stor["M1"] = int64(5)
	stor["M2"] = float64(0)
	locStorage := storage.NewMemStorage(storage.WithBuffer(stor))
	ctx := context.Background()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetMetric(ctx, locStorage, tt.path)
			if tt.withErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGetTableMetrics(t *testing.T) {
	test := struct {
		name string
		want map[string]string
	}{
		name: "Check table",
		want: map[string]string{"M1": "5", "M2": "0"},
	}
	stor := make(map[string]interface{})
	stor["M1"] = int64(5)
	stor["M2"] = float64(0)
	locStorage := storage.NewMemStorage(storage.WithBuffer(stor))
	ctx := context.Background()
	result := GetTableMetrics(ctx, locStorage)

	t.Run(test.name, func(t *testing.T) {
		assert.True(t, cmp.Equal(test.want, result))
	})
}

func TestSignData(t *testing.T) {
	const key = "wersdjfl23.w3"
	counter := int64(34567)
	gauge := float64(8723.098)
	tests := []struct {
		name string
		data types.Metric
		want string
	}{
		{
			name: "counter value",
			data: types.Metric{ID: "test1", MType: "counter", Delta: &counter},
			want: "4003975ccfa11fdd45fc8ad03202a1f1dab466c3d118b2c7200e45de6f90da37",
		},
		{
			name: "gauge value",
			data: types.Metric{ID: "test2", MType: "gauge", Value: &gauge},
			want: "de0a02dd05ed708397ab730155bdf233ef415eebe8778c921b1aeff8333570e6",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.NoError(t, SignData(&test.data, key))
			assert.Equal(t, test.want, test.data.Hash)
		})
	}
}

func TestIsSignEqual(t *testing.T) {
	const key = "wersdjfl23.w3"
	value := int64(34567)
	tests := []struct {
		isPositive bool
		data       types.Metric
		want       bool
	}{
		{
			isPositive: true,
			data: types.Metric{
				ID:    "test",
				MType: "counter",
				Delta: &value,
				Hash:  "bf9f25fcb26f7df011969f39a83d81c7746a07ea400bf93e4758fd565fce32f8",
			},
			want: true,
		},
		{
			isPositive: false,
			data: types.Metric{
				ID:    "test",
				MType: "counter",
				Delta: &value,
				Hash:  "af9f25fcb26f7df011969f39a83d81c7746a07ea400bf93e4758fd565fce32f8",
			},
			want: true,
		},
	}
	for _, test := range tests {
		if test.isPositive {
			t.Run("positive", func(t *testing.T) {
				assert.True(t, IsSignEqual(test.data, key))
			})
		} else {
			t.Run("negative", func(t *testing.T) {
				assert.False(t, IsSignEqual(test.data, key))
			})

		}

	}

}
