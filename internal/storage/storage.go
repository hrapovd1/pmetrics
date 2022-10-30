package storage

import (
	"strconv"
)

type gauge float64
type counter int64

type Repository interface {
	Append(value counter)
	Get(key string) (interface{}, bool)
	GetAll() map[string]interface{}
	Rewrite(key string, value gauge)
}

type MemStorage struct {
	Buffer map[string]interface{}
}

func (ms *MemStorage) Append(value counter) {
	ms.Buffer["PollCount"] = append(
		ms.Buffer["PollCount"].([]int64),
		int64(value),
	)
}

func (ms *MemStorage) Get(key string) (interface{}, bool) {
	switch key {
	case "PollCount":
		pollCount, ok := ms.Buffer["PollCount"].([]int64)
		if !ok {
			return nil, ok
		}
		last := len(pollCount) - 1
		if last < 0 {
			return nil, false
		}
		return ms.Buffer["PollCount"].([]int64)[last], true
	default:
		val, ok := ms.Buffer[key]
		if ok {
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
		Buffer: map[string]interface{}{
			"PollCount": make([]int64, 0),
		},
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
