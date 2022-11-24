package storage

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMemStorage_Rewrite(t *testing.T) {
	type args struct {
		key   string
		value float64
	}
	tests := []args{
		{
			key:   "1",
			value: -0.1,
		},
		{
			key:   "1",
			value: 1,
		},
	}

	stor := make(map[string]interface{})
	ms := NewMemStorage(WithBuffer(stor))
	t.Run("Rewrite values.", func(t *testing.T) {
		for _, tt := range tests {
			ms.Rewrite(tt.key, gauge(tt.value))
			assert.Equal(t, tt.value, stor[tt.key])
		}
	})
	t.Run("Count values.", func(t *testing.T) {
		assert.Equal(t, 1, len(stor))
	})
}

func TestMemStorage_Append(t *testing.T) {
	tests := []struct {
		key   string
		value int64
	}{
		{
			key:   "Count1",
			value: 23,
		},
		{
			key:   "Count2",
			value: -23,
		},
		{
			key:   "Count3",
			value: -0,
		},
	}
	stor := make(map[string]interface{})
	ms := NewMemStorage(WithBuffer(stor))
	for _, test := range tests {
		t.Run("Append values", func(t *testing.T) {
			ms.Append(test.key, counter(test.value))
			assert.Equal(t, test.value, stor[test.key].(int64))
		})
	}
	t.Run("Count values", func(t *testing.T) {
		assert.Equal(t, 3, len(stor))
	})
}

func TestMemStorage_Get(t *testing.T) {
	stor := make(map[string]interface{})
	ms := NewMemStorage(WithBuffer(stor))
	stor["PollCount"] = int64(1)
	stor["Alloc"] = float64(3.0)
	stor["TotalAlloc"] = float64(-3.0)

	tests := []struct {
		name     string
		positive bool
		pcount   bool
		want1    float64
		want2    int64
	}{
		{
			name:     "PollCount",
			positive: true,
			pcount:   true,
			want1:    0,
			want2:    1,
		},
		{
			name:     "Alloc",
			positive: true,
			pcount:   false,
			want1:    3.0,
			want2:    1,
		},
		{
			name:     "TotalAlloc",
			positive: true,
			pcount:   false,
			want1:    -3.0,
			want2:    1,
		},
		{
			name:     "Undefined",
			positive: false,
			pcount:   true,
			want1:    0,
			want2:    1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.positive {
				metric := ms.Get(tt.name)
				require.NotNil(t, metric)
				if tt.pcount {
					assert.Equal(t, tt.want2, metric.(int64))
				} else {
					assert.Equal(t, tt.want1, metric.(float64))
				}
			} else {
				metric := ms.Get(tt.name)
				require.Nil(t, metric)
			}
		})
	}
}

func TestMemStorage_GetAll(t *testing.T) {
	stor := make(map[string]interface{})
	ms := NewMemStorage(WithBuffer(stor))
	stor["PollCount"] = []int64{4}
	stor["Alloc"] = float64(3.0)
	stor["TotalAlloc"] = float64(-3.0)
	t.Run("Check GetAll", func(t *testing.T) {
		out := ms.GetAll()
		assert.True(t, cmp.Equal(stor, out))
	})
}

func TestToGauge(t *testing.T) {
	tests := []struct {
		name  string
		value float64
		want  gauge
	}{
		{
			name:  "Float value",
			value: float64(1.0),
			want:  gauge(1.0),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			value := ToGauge(test.value)
			assert.Equal(t, test.want, value)
		})
	}
}

func TestToCounter(t *testing.T) {
	tests := []struct {
		name  string
		value int64
		want  counter
	}{
		{
			name:  "Int value",
			value: int64(-0),
			want:  counter(0),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			value := ToCounter(test.value)
			assert.Equal(t, test.want, value)
		})
	}
}
