package usecase

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hrapovd1/pmetrics/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := WriteMetric(locStorage, tt.path)
			require.NoError(t, err)
			switch result := locStorage.Get(tt.metricName).(type) {
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

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetMetric(locStorage, tt.path)
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
	result := GetTableMetrics(locStorage)

	t.Run(test.name, func(t *testing.T) {
		assert.True(t, cmp.Equal(test.want, result))
	})
}