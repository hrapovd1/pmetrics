package types

import (
	"context"
	"database/sql"
	"log"
)

const DBtablePrefix = "pmetric_"

type Metric struct {
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
	Hash  string   `json:"hash,omitempty"`  // значение хеш-функции
}

type Repository interface {
	Append(ctx context.Context, key string, value int64)
	Get(ctx context.Context, key string) interface{}
	GetAll(ctx context.Context) map[string]interface{}
	Rewrite(ctx context.Context, key string, value float64)
	StoreAll(ctx context.Context, metrics *[]Metric)
}

type Pinger interface {
	Ping(ctx context.Context) bool
}

type Storer interface {
	Restore(ctx context.Context, logger log.Logger)
	Storing(ctx context.Context, logger log.Logger)
}

type MetricModel struct {
	Timestamp int64 `gorm:"primaryKey;autoCreateTime"`
	ID        string
	Mtype     string
	Value     sql.NullFloat64
	Delta     sql.NullInt64
}
