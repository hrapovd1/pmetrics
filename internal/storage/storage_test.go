package storage

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hrapovd1/pmetrics/internal/types"
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
	ctx := context.Background()
	t.Run("Rewrite values.", func(t *testing.T) {
		for _, tt := range tests {
			ms.Rewrite(ctx, tt.key, tt.value)
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
	ctx := context.Background()
	for _, test := range tests {
		t.Run("Append values", func(t *testing.T) {
			ms.Append(ctx, test.key, test.value)
			assert.Equal(t, test.value, stor[test.key].(int64))
		})
	}
	t.Run("Count values", func(t *testing.T) {
		assert.Equal(t, 3, len(stor))
	})
}

func TestMemStorage_Get(t *testing.T) {
	stor := make(map[string]interface{})
	ctx := context.Background()
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
				metric := ms.Get(ctx, tt.name)
				require.NotNil(t, metric)
				if tt.pcount {
					assert.Equal(t, tt.want2, metric.(int64))
				} else {
					assert.Equal(t, tt.want1, metric.(float64))
				}
			} else {
				metric := ms.Get(ctx, tt.name)
				require.Nil(t, metric)
			}
		})
	}
}

func TestMemStorage_GetAll(t *testing.T) {
	stor := make(map[string]interface{})
	ms := NewMemStorage(WithBuffer(stor))
	ctx := context.Background()
	stor["PollCount"] = []int64{4}
	stor["Alloc"] = float64(3.0)
	stor["TotalAlloc"] = float64(-3.0)
	t.Run("Check GetAll", func(t *testing.T) {
		out := ms.GetAll(ctx)
		assert.True(t, cmp.Equal(stor, out))
	})
}

func TestMemStorage_StoreAll(t *testing.T) {
	type args struct {
		ctx     context.Context
		metrics *[]types.Metric
	}
	var m1 = int64(4567)
	var m2 = float64(45.67)
	metrics := []types.Metric{
		{ID: "M1", MType: "counter", Delta: &m1},
		{ID: "M2", MType: "gauge", Value: &m2},
	}
	buff := make(map[string]interface{})

	tests := []struct {
		name   string
		args   args
		result map[string]interface{}
	}{
		{
			"first test",
			args{
				ctx:     context.Background(),
				metrics: &metrics,
			},
			map[string]interface{}{"M1": int64(4567), "M2": float64(45.67)},
		},
		{
			"second test",
			args{
				ctx:     context.Background(),
				metrics: &metrics,
			},
			map[string]interface{}{"M1": int64(9134), "M2": float64(45.67)},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ms := &MemStorage{
				buffer: buff,
			}
			ms.StoreAll(tt.args.ctx, tt.args.metrics)
			assert.Equal(t, tt.result, buff)
		})
	}
}

func TestStrToFloat64(t *testing.T) {
	tests := []struct {
		name   string
		inVal  string
		outVal float64
	}{
		{
			"empty",
			"",
			float64(0),
		},
		{
			"positive",
			"3.1",
			float64(3.1),
		},
		{
			"negative",
			"-2.0",
			float64(-2),
		},
		{
			"less zero",
			"0.001",
			float64(0.001),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := StrToFloat64(tt.inVal)
			if tt.inVal != "" {
				require.NoError(t, err)
				assert.Equal(t, tt.outVal, res)
			} else {
				require.Error(t, err)
			}

		})
	}
}

func TestStrToInt64(t *testing.T) {
	tests := []struct {
		name   string
		inVal  string
		outVal int64
	}{
		{
			"empty",
			"",
			int64(0),
		},
		{
			"positive",
			"3",
			int64(3),
		},
		{
			"negative",
			"-2",
			int64(-2),
		},
		{
			"zero",
			"0",
			int64(0),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := StrToInt64(tt.inVal)
			if tt.inVal != "" {
				require.NoError(t, err)
				assert.Equal(t, tt.outVal, res)
			} else {
				require.Error(t, err)
			}
		})
	}
}
