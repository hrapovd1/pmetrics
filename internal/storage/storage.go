package storage

import (
	"strconv"
)

type gauge float64
type counter int64

type Repository interface {
	Append(key string, value counter)
	Get(key string) (interface{}, bool)
	GetAll() map[string]interface{}
	Rewrite(key string, value gauge)
}

type MemStorage struct {
	Buffer map[string]interface{}
}

func (ms *MemStorage) Append(key string, value counter) {
	_, ok := ms.Buffer[key]
	if !ok {
		ms.Buffer[key] = int64(value)
		return
	}
	val := ms.Buffer[key].(int64) + int64(value)
	ms.Buffer[key] = int64(val)
}

func (ms *MemStorage) Get(key string) (interface{}, bool) {
	val, ok := ms.Buffer[key]
	if ok {
		switch val := val.(type) {
		case float64:
			return val, ok
		case int64:
			return val, ok
		}
	}
	return nil, false
}

func (ms *MemStorage) GetAll() map[string]interface{} {
	out := make(map[string]interface{})
	for k, v := range ms.Buffer {
		out[k] = v
	}
	return out
}

func (ms *MemStorage) Rewrite(key string, value gauge) {
	ms.Buffer[key] = float64(value)
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		Buffer: make(map[string]interface{}),
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
