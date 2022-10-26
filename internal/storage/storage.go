package storage

import "strconv"

type gauge float64
type counter int64

type Repository interface {
	Rewrite(key string, value interface{})
	Append(value interface{})
}

type MemStorage struct {
	GaugeBuff   map[string]gauge
	CounterBuff []counter
}

func (ms *MemStorage) Rewrite(key string, value interface{}) {
	ms.GaugeBuff[key] = value.(gauge)
}

func (ms *MemStorage) Append(value interface{}) {
	ms.CounterBuff = append(ms.CounterBuff, value.(counter))
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		GaugeBuff:   make(map[string]gauge),
		CounterBuff: make([]counter, 0),
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
