package storage

import (
	"bufio"
	"os"
	"strconv"

	"github.com/hrapovd1/pmetrics/internal/config"
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

type fileStorage struct {
	file   *os.File
	writer *bufio.Writer
	config config.Config
	buff   map[string]interface{}
}

type MemStorage struct {
	buffer  map[string]interface{}
	backend fileStorage
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
	return maps.Clone(ms.buffer)
}

func (ms *MemStorage) Rewrite(key string, value gauge) {
	ms.buffer[key] = float64(value)
}

func NewMemStorage(storConfig config.Config, opts ...Option) *MemStorage {
	buffer := make(map[string]interface{})
	ms := &MemStorage{
		buffer:  buffer,
		backend: newBackend(storConfig, buffer),
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

func ToGauge(input float64) gauge {
	return gauge(input)
}

func ToCounter(input int64) counter {
	return counter(input)
}
