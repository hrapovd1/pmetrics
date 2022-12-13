package storage

import (
	"strconv"

	"github.com/hrapovd1/pmetrics/internal/types"
	"golang.org/x/exp/maps"
)

type MemStorage struct {
	buffer map[string]interface{}
}

type Option func(mem *MemStorage) *MemStorage

func (ms *MemStorage) Append(key string, value int64) {
	var val int64
	_, ok := ms.buffer[key]
	if ok {
		val = ms.buffer[key].(int64) + value
	} else {
		val = value
	}
	ms.buffer[key] = val
}

func (ms *MemStorage) Get(key string) interface{} {
	val, ok := ms.buffer[key]
	if ok {
		return val
	}
	return nil
}

func (ms *MemStorage) GetAll() map[string]interface{} {
	return maps.Clone(ms.buffer)
}

func (ms *MemStorage) Rewrite(key string, value float64) {
	ms.buffer[key] = value
}

func (ms *MemStorage) StoreAll(metrics *[]types.Metric) {
	for _, m := range *metrics {
		switch m.MType {
		case "counter":
			var val int64
			_, ok := ms.buffer[m.ID]
			if ok {
				val = ms.buffer[m.ID].(int64) + *m.Delta
			} else {
				val = *m.Delta
			}
			ms.buffer[m.ID] = val
		case "gauge":
			ms.buffer[m.ID] = *m.Value
		}
	}
}

func NewMemStorage(opts ...Option) *MemStorage {
	buffer := make(map[string]interface{})
	ms := &MemStorage{
		buffer: buffer,
	}

	for _, opt := range opts {
		ms = opt(ms)
	}

	return ms
}

func WithBuffer(buffer map[string]interface{}) Option {
	return func(mem *MemStorage) *MemStorage {
		mem.buffer = buffer
		return mem
	}
}

func StrToFloat64(input string) (float64, error) {
	out, err := strconv.ParseFloat(input, 64)
	return out, err
}

func StrToInt64(input string) (int64, error) {
	out, err := strconv.ParseInt(input, 10, 64)
	return out, err
}
