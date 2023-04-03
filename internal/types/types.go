// Модуль types содержит общие для проетка типы и интерфейсы.
package types

import (
	"context"
	"database/sql"
	"log"
	"time"
)

// Префикс в названиях таблиц базы
const DBtablePrefix = "pmetric_"

// Metric тип JSON формата метрики
type Metric struct {
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
	Hash  string   `json:"hash,omitempty"`  // значение хеш-функции
}

type EncData struct {
	Data0 string `json:"data0"` // зашифрованные данные
	Data  string `json:"data1"` // зашифрованные данные
}

// Repository основной интерфейс хранилища метрик
type Repository interface {
	Append(ctx context.Context, key string, value int64)
	Get(ctx context.Context, key string) interface{}
	GetAll(ctx context.Context) map[string]interface{}
	Rewrite(ctx context.Context, key string, value float64)
	StoreAll(ctx context.Context, metrics *[]Metric)
}

// Storager вспомогательный интерфейс хранилища метрик
type Storager interface {
	Close() error
	Ping(ctx context.Context) bool
	Restore(ctx context.Context) error
	Storing(ctx context.Context, logger *log.Logger, interval time.Duration, restore bool)
}

// MetricModel модель таблицы для хранения метрики в базе
type MetricModel struct {
	Timestamp int64 `gorm:"primaryKey;autoCreateTime"`
	ID        string
	Mtype     string
	Value     sql.NullFloat64
	Delta     sql.NullInt64
}

// Waitgrp тип для передачи wait group в контексте
type Waitgrp string
