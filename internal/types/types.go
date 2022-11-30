package types

import "database/sql"

type Metric struct {
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
	Hash  string   `json:"hash,omitempty"`  // значение хеш-функции
}

type MetricModel struct {
	Timestamp int64 `gorm:"primaryKey;autoCreateTime"`
	Mtype     string
	Value     sql.NullFloat64
	Delta     sql.NullInt64
}
