package dbstorage

import (
	"context"
	"fmt"
	"log"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hrapovd1/pmetrics/internal/storage"
	"github.com/hrapovd1/pmetrics/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewDBStorage(t *testing.T) {
	ds, err := NewDBStorage("", log.Default(), storage.NewMemStorage())
	require.NoError(t, err)
	assert.IsType(t, &DBStorage{}, ds)
}

func TestDBStorage_Append(t *testing.T) {
	buff := make(map[string]interface{})
	ds, err := NewDBStorage("", log.Default(), storage.NewMemStorage(storage.WithBuffer(buff)))
	require.NoError(t, err)
	ds.Append(context.Background(), "M1", int64(123))
	assert.Equal(t, int64(123), buff["M1"])
}

func TestDBStorage_Get(t *testing.T) {
	buff := map[string]interface{}{"M1": int64(321)}
	ds, err := NewDBStorage("", log.Default(), storage.NewMemStorage(storage.WithBuffer(buff)))
	require.NoError(t, err)
	result := ds.Get(context.Background(), "M1")
	assert.Equal(t, int64(321), result)

}

func TestDBStorage_GetAll(t *testing.T) {
	buff := map[string]interface{}{"M1": int64(321)}
	ds, err := NewDBStorage("", log.Default(), storage.NewMemStorage(storage.WithBuffer(buff)))
	require.NoError(t, err)
	buff["PollCount"] = []int64{4}
	buff["Alloc"] = float64(3.0)
	buff["TotalAlloc"] = float64(-3.0)
	t.Run("Check GetAll", func(t *testing.T) {
		out := ds.GetAll(context.Background())
		assert.True(t, cmp.Equal(buff, out))
	})
}

func TestDBStorage_Rewrite(t *testing.T) {
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

	buff := make(map[string]interface{})
	ds, err := NewDBStorage("", log.Default(), storage.NewMemStorage(storage.WithBuffer(buff)))
	require.NoError(t, err)
	t.Run("Rewrite values.", func(t *testing.T) {
		for _, tt := range tests {
			ds.Rewrite(context.Background(), tt.key, tt.value)
			assert.Equal(t, tt.value, buff[tt.key])
		}
	})
	t.Run("Count values.", func(t *testing.T) {
		assert.Equal(t, 1, len(buff))
	})
}

func TestDBStorage_StoreAll(t *testing.T) {
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
			ds, err := NewDBStorage("", log.Default(), storage.NewMemStorage(storage.WithBuffer(buff)))
			require.NoError(t, err)
			ds.StoreAll(tt.args.ctx, tt.args.metrics)
			assert.Equal(t, tt.result, buff)
		})
	}
}

func TestDBStorage_Ping(t *testing.T) {
	ds, err := NewDBStorage("", log.Default(), storage.NewMemStorage())
	require.NoError(t, err)
	assert.False(t, ds.Ping(context.Background()))
}

func TestDBStorage_Close(t *testing.T) {
	ds, err := NewDBStorage("", log.Default(), storage.NewMemStorage())
	require.NoError(t, err)
	assert.NoError(t, ds.Close())
}

func TestDBStorage_Restore(t *testing.T) {
	ds, err := NewDBStorage("", log.Default(), storage.NewMemStorage())
	require.NoError(t, err)
	assert.NoError(t, ds.Restore(context.Background()))
}

func TestDBStorage_Storing(t *testing.T) {
	buff := make(map[string]interface{})
	ds, err := NewDBStorage("", log.Default(), storage.NewMemStorage(storage.WithBuffer(buff)))
	require.NoError(t, err)
	ds.Storing(
		context.Background(),
		log.Default(),
		0,
		true,
	)
	assert.Empty(t, buff)
}

func ExampleNewDBStorage() {
	dsn := "postgres://postgres:postgres@localhost:5432/postgres"
	storage := storage.NewMemStorage()

	dbStorage, _ := NewDBStorage(dsn, log.Default(), storage)

	fmt.Println(dbStorage)
}
