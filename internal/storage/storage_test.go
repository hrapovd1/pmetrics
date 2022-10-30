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
	ms := NewMemStorage()
	t.Run("Rewrite values.", func(t *testing.T) {
		for _, tt := range tests {
			ms.Rewrite(tt.key, gauge(tt.value))
			assert.Equal(t, tt.value, ms.Buffer[tt.key])
		}
	})
	t.Run("Count values.", func(t *testing.T) {
		assert.Equal(t, 2, len(ms.Buffer))
	})
}

func TestMemStorage_Append(t *testing.T) {
	tests := []int64{1, 2, 3}
	ms := NewMemStorage()
	t.Run("Append values", func(t *testing.T) {
		for _, val := range tests {
			ms.Append(counter(val))
			last := len(ms.Buffer["PollCount"].([]int64)) - 1
			assert.Equal(t, val, ms.Buffer["PollCount"].([]int64)[last])
		}
	})
	t.Run("Count values", func(t *testing.T) {
		pollCount := ms.Buffer["PollCount"].([]int64)
		assert.Equal(t, 3, len(pollCount))
	})
}

func TestNewMemStorage(t *testing.T) {
	want := &MemStorage{
		Buffer: map[string]interface{}{
			"PollCount": make([]int64, 0),
		},
	}
	assert.True(t, cmp.Equal(NewMemStorage(), want))
}

func TestMemStorage_Get(t *testing.T) {
	ms := NewMemStorage()
	pollCount, _ := ms.Buffer["PollCount"].([]int64)
	ms.Buffer["PollCount"] = append(pollCount, int64(1))
	ms.Buffer["Alloc"] = float64(3.0)
	ms.Buffer["TotalAlloc"] = float64(-3.0)

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
				metric, ok := ms.Get(tt.name)
				require.True(t, ok)
				if tt.pcount {
					assert.Equal(t, tt.want2, metric.(int64))
				} else {
					assert.Equal(t, tt.want1, metric.(float64))
				}
			} else {
				_, ok := ms.Get(tt.name)
				require.False(t, ok)
			}
		})
	}
}

func TestMemStorage_GetAll(t *testing.T) {
	ms := NewMemStorage()
	ms.Buffer["PollCount"] = []int64{4}
	ms.Buffer["Alloc"] = float64(3.0)
	ms.Buffer["TotalAlloc"] = float64(-3.0)
	t.Run("Check GetAll", func(t *testing.T) {
		out := ms.GetAll()
		assert.True(t, cmp.Equal(ms.Buffer, out))
	})
}

func TestStrToGauge(t *testing.T) {
	tests := []struct {
		name     string
		positive bool
		value    string
		want     gauge
	}{
		{
			name:     "Float value",
			positive: true,
			value:    "1.0",
			want:     gauge(1.0),
		},
		{
			name:     "Int value",
			positive: true,
			value:    "1",
			want:     gauge(1.0),
		},
		{
			name:     "Wrong value",
			positive: false,
			value:    "a",
			want:     gauge(1.0),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.positive {
				value, err := StrToGauge(test.value)
				assert.Nil(t, err)
				assert.Equal(t, test.want, value)
			} else {
				_, err := StrToGauge(test.value)
				assert.NotNil(t, err)
			}
		})
	}
}

func TestStrToCounter(t *testing.T) {
	tests := []struct {
		name     string
		positive bool
		value    string
		want     counter
	}{
		{
			name:     "Float value",
			positive: false,
			value:    "1.0",
			want:     counter(1),
		},
		{
			name:     "Int value",
			positive: true,
			value:    "-0",
			want:     counter(0),
		},
		{
			name:     "Wrong value",
			positive: false,
			value:    "a",
			want:     counter(0),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.positive {
				value, err := StrToCounter(test.value)
				assert.Nil(t, err)
				assert.Equal(t, test.want, value)
			} else {
				_, err := StrToCounter(test.value)
				assert.NotNil(t, err)
			}
		})
	}
}
