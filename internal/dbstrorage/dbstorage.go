// Модуль dbstorage содержит типы и методы для хранения метрик
// в базе postgresql.
package dbstorage

import (
	"context"
	"database/sql"
	"log"
	"strings"
	"time"

	"github.com/hrapovd1/pmetrics/internal/types"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// DBStorage тип для хранения метрик в базу
type DBStorage struct {
	dbConnect  *sql.DB
	logger     *log.Logger
	backStor   types.Repository
	tableNames map[string]struct{}
}

// NewDBStorage возвращает тип DBStorage по полученному конфигу
func NewDBStorage(dsn string, logger *log.Logger, backStor types.Repository) (*DBStorage, error) {
	db := DBStorage{
		logger:     logger,
		backStor:   backStor,
		tableNames: make(map[string]struct{}),
	}
	dbConnect, err := sql.Open("pgx", dsn)
	db.dbConnect = dbConnect
	return &db, err
}

// Append сохраняет новое значение типа counter с дозаписью к старому
func (ds *DBStorage) Append(ctx context.Context, key string, value int64) {
	ds.backStor.Append(ctx, key, value)
	metric := types.MetricModel{
		ID:    key,
		Mtype: "counter",
		Delta: sql.NullInt64{
			Int64: ds.backStor.Get(ctx, key).(int64),
			Valid: true,
		},
	}
	if err := ds.store(ctx, &metric); err != nil {
		if ds.logger != nil {
			ds.logger.Println(err)
		}
	}
}

// Get возвращает значение метрики переданной через key
func (ds *DBStorage) Get(ctx context.Context, key string) interface{} {
	return ds.backStor.Get(ctx, key)
}

// GetAll возвращает все метрики
func (ds *DBStorage) GetAll(ctx context.Context) map[string]interface{} {
	return ds.backStor.GetAll(ctx)
}

// Rewrite перезаписывает значение метрики типа gauge
func (ds *DBStorage) Rewrite(ctx context.Context, key string, value float64) {
	ds.backStor.Rewrite(ctx, key, value)
	metric := types.MetricModel{
		ID:    key,
		Mtype: "gauge",
		Value: sql.NullFloat64{Float64: value, Valid: true},
	}
	if err := ds.store(ctx, &metric); err != nil {
		if ds.logger != nil {
			ds.logger.Println(err)
		}
	}
}

// StoreAll сохраняет все полученные метрики через слайс metrics
func (ds *DBStorage) StoreAll(ctx context.Context, metrics *[]types.Metric) {
	metricsDB := make([]types.MetricModel, 0)
	for _, m := range *metrics {
		metricDB := types.MetricModel{ID: m.ID, Mtype: m.MType}
		switch m.MType {
		case "counter":
			ds.backStor.Append(ctx, m.ID, *m.Delta)
			metricDB.Delta = sql.NullInt64{
				Int64: ds.backStor.Get(ctx, m.ID).(int64),
				Valid: true,
			}
		case "gauge":
			ds.backStor.Rewrite(ctx, m.ID, *m.Value)
			metricDB.Value = sql.NullFloat64{Float64: *m.Value, Valid: true}
		}
		metricsDB = append(metricsDB, metricDB)
	}
	if ds.dbConnect != nil {
		if err := ds.storeBatch(ctx, &metricsDB); err != nil {
			ds.logger.Print(err)
		}
	}
}

// Storing запускается в отдельной go routine для сохранения метрик в файл
func (ds *DBStorage) Storing(ctx context.Context, logger *log.Logger, interval time.Duration, restore bool) {
	stor := ds.backStor.(types.Storager)
	stor.Storing(ctx, logger, interval, restore)
}

// Close закрывает подключение к БД, необходимо запускать в defer
func (ds *DBStorage) Close() error {
	stor := ds.backStor.(types.Storager)
	defer ds.logger.Println(stor.Close())
	return ds.dbConnect.Close()
}

// Ping используется для проверки доступности базы
func (ds *DBStorage) Ping(ctx context.Context) bool {
	if ds.dbConnect == nil {
		return false
	}
	ctxT, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()
	if err := ds.dbConnect.PingContext(ctxT); err != nil {
		return false
	}
	return true
}

// Restore восстанавливае значение метрик при запуске из файла
func (ds *DBStorage) Restore(ctx context.Context) error {
	stor := ds.backStor.(types.Storager)
	return stor.Restore(ctx)
}

// store внутренняя функция сохранения метрики в базу
func (ds *DBStorage) store(ctx context.Context, metric *types.MetricModel) error {
	db, err := gorm.Open(postgres.New(postgres.Config{Conn: ds.dbConnect}), &gorm.Config{})
	if err != nil {
		return err
	}
	tableName := strings.ToLower(types.DBtablePrefix + metric.ID)
	select {
	case <-ctx.Done():
		return nil
	default:
		if _, ok := ds.tableNames[tableName]; !ok {
			if !db.Migrator().HasTable(tableName) {
				if err := db.Table(tableName).Migrator().CreateTable(&types.MetricModel{}); err != nil {
					return err
				}
			}
			ds.tableNames[tableName] = struct{}{}
		}
		db.Table(tableName).Create(metric)
		return nil
	}
}

// storeBatch внутренняя функция сохранения нескольких метрик в базу
func (ds *DBStorage) storeBatch(ctx context.Context, metrics *[]types.MetricModel) error {
	select {
	case <-ctx.Done():
		return nil
	default:
		db, err := gorm.Open(postgres.New(postgres.Config{Conn: ds.dbConnect}), &gorm.Config{})
		if err != nil {
			return err
		}

		if err := db.Transaction(func(tx *gorm.DB) error {
			for _, metric := range *metrics {
				tableName := strings.ToLower(types.DBtablePrefix + metric.ID)
				if _, ok := ds.tableNames[tableName]; !ok {
					if !db.Migrator().HasTable(tableName) {
						if err := tx.Table(tableName).Migrator().CreateTable(&types.MetricModel{}); err != nil {
							return err
						}
					}
					ds.tableNames[tableName] = struct{}{}
				}
				tx.Table(tableName).Create(&metric)
			}
			return nil

		}); err != nil {
			return err
		}
		return nil
	}
}
