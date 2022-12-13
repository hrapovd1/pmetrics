package types

import "database/sql"

const DBtablePrefix = "pmetric_"

type Metric struct {
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
	Hash  string   `json:"hash,omitempty"`  // значение хеш-функции
}

type Repository interface {
	Append(key string, value int64)
	Get(key string) interface{}
	GetAll() map[string]interface{}
	Rewrite(key string, value float64)
	StoreAll(*[]Metric)
}

type MetricModel struct {
	Timestamp int64 `gorm:"primaryKey;autoCreateTime"`
	ID        string
	Mtype     string
	Value     sql.NullFloat64
	Delta     sql.NullInt64
}
