package storage

import (
	"context"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/hrapovd1/pmetrics/internal/types"
	"golang.org/x/exp/maps"
)

// Option тип для модификации хранилища MemStorage
type Option func(mem *MemStorage) *MemStorage

// MemStorage тип реализации хранения в памяти
type MemStorage struct {
	buffer map[string]interface{}
}

// NewMemStorage создает хранилище MemStorage
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

// Append сохраняет новое значение типа counter с дозаписью к старому
func (ms *MemStorage) Append(ctx context.Context, key string, value int64) {
	select {
	case <-ctx.Done():
		return
	default:
		var val int64
		_, ok := ms.buffer[key]
		if ok {
			val = ms.buffer[key].(int64) + value
		} else {
			val = value
		}
		ms.buffer[key] = val
	}
}

// Get возвращает значение метрики переданной через key
func (ms *MemStorage) Get(ctx context.Context, key string) interface{} {
	select {
	case <-ctx.Done():
		return nil
	default:
		val, ok := ms.buffer[key]
		if ok {
			return val
		}
		return nil
	}
}

// GetAll возвращает все метрики
func (ms *MemStorage) GetAll(ctx context.Context) map[string]interface{} {
	return maps.Clone(ms.buffer)
}

// Rewrite перезаписывает значение метрики типа gauge
func (ms *MemStorage) Rewrite(ctx context.Context, key string, value float64) {
	ms.buffer[key] = value
}

// StoreAll сохраняет все полученные метрики через слайс metrics
func (ms *MemStorage) StoreAll(ctx context.Context, metrics *[]types.Metric) {
	select {
	case <-ctx.Done():
		return
	default:
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
}

// Close для реализации интерфейса Storager
func (ms *MemStorage) Close() error { return nil }

// Ping для реализации интерфейса Storager
func (ms *MemStorage) Ping(ctx context.Context) bool { return false }

// Restore для реализации интерфейса Storager
func (ms *MemStorage) Restore(ctx context.Context) error { return nil }

// Storing для реализации интерфейса Storager
func (ms *MemStorage) Storing(ctx context.Context, w *sync.WaitGroup, logger *log.Logger, interval time.Duration, restore bool) {
}

// WitBuffer модифицирует MemStorage позволяя передать внешний буфер
// как внутреннее хранилище, используется в тестах
func WithBuffer(buffer map[string]interface{}) Option {
	return func(mem *MemStorage) *MemStorage {
		mem.buffer = buffer
		return mem
	}
}

// StrToFloat64 преобразует строку в float64
func StrToFloat64(input string) (float64, error) {
	out, err := strconv.ParseFloat(input, 64)
	return out, err
}

// StrToInt64 преобразование строки в int64
func StrToInt64(input string) (int64, error) {
	out, err := strconv.ParseInt(input, 10, 64)
	return out, err
}
