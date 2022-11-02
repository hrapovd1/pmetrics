package storage

import (
	"strconv"

	"golang.org/x/exp/maps"
)

type gauge float64
type counter int64

type Repository interface {
	Append(key string, value counter)
	Get(key string) interface{}
	GetAll() map[string]interface{}
	Rewrite(key string, value gauge)
}

type MemStorage struct {
	buffer map[string]interface{}
}

type Option func(mem *MemStorage) *MemStorage

func (ms *MemStorage) Append(key string, value counter) {
	_, ok := ms.buffer[key]
	if !ok {
		ms.buffer[key] = int64(value)
		return
	}
	val := ms.buffer[key].(int64) + int64(value)
	ms.buffer[key] = int64(val)
}

func (ms *MemStorage) Get(key string) interface{} {
	val, ok := ms.buffer[key]
	if ok {
		switch val := val.(type) {
		case float64:
			return val
		case int64:
			return val
		}
	}
	return nil
}

func (ms *MemStorage) GetAll() map[string]interface{} {
	out := make(map[string]interface{})
	for k, v := range ms.buffer {
		out[k] = v
	}
	return maps.Clone(ms.buffer)
}

func (ms *MemStorage) Rewrite(key string, value gauge) {
	ms.buffer[key] = float64(value)
}

func NewMemStorage(opts ...Option) *MemStorage {
	ms := &MemStorage{
		buffer: make(map[string]interface{}),
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

func StrToGauge(input string) (gauge, error) {
	out, err := strconv.ParseFloat(input, 64)
	return gauge(out), err
}

func StrToCounter(input string) (counter, error) {
	out, err := strconv.ParseInt(input, 10, 64)
	return counter(out), err
}
