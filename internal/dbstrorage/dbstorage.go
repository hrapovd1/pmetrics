package dbstorage

import (
	"context"
	"database/sql"
	"log"
	"strings"
	"time"

	"github.com/hrapovd1/pmetrics/internal/config"
	"github.com/hrapovd1/pmetrics/internal/filestorage"
	"github.com/hrapovd1/pmetrics/internal/types"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type DBStorage struct {
	dbConnect *sql.DB
	buff      map[string]interface{}
	logger    *log.Logger
	fileStor  *filestorage.FileStorage
}

func (ds *DBStorage) Append(ctx context.Context, key string, value int64) {
	if ds.dbConnect == nil {
		return
	}
	var val int64
	_, ok := ds.buff[key]
	if ok {
		val = ds.buff[key].(int64) + value
	} else {
		val = value
	}
	metric := types.MetricModel{
		ID:    key,
		Mtype: "counter",
		Delta: sql.NullInt64{Int64: val, Valid: true},
	}
	if err := ds.store(ctx, &metric); err != nil {
		if ds.logger != nil {
			ds.logger.Println(err)
		}
	}
}

func (ds *DBStorage) Get(ctx context.Context, key string) interface{} {
	return nil
}

func (ds *DBStorage) GetAll(ctx context.Context) map[string]interface{} {
	return nil
}

func (ds *DBStorage) Rewrite(ctx context.Context, key string, value float64) {
	if ds.dbConnect != nil {
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
}

func (ds *DBStorage) StoreAll(ctx context.Context, metrics *[]types.Metric) {
	metricsDB := make([]types.MetricModel, 0)
	for _, m := range *metrics {
		metricDB := types.MetricModel{ID: m.ID, Mtype: m.MType}
		switch m.MType {
		case "counter":
			var val int64
			_, ok := ds.buff[m.ID]
			if ok {
				val = ds.buff[m.ID].(int64) + *m.Delta
			} else {
				val = *m.Delta
			}
			metricDB.Delta = sql.NullInt64{Int64: val, Valid: true}
		case "gauge":
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

func NewDBStorage(conf config.Config, logger *log.Logger, buff map[string]interface{}, fs *filestorage.FileStorage) (*DBStorage, error) {
	db := DBStorage{}
	db.logger = logger
	db.buff = buff
	db.fileStor = fs
	if conf.DatabaseDSN == "" {
		return &db, nil
	}
	dbConnect, err := sql.Open("pgx", conf.DatabaseDSN)
	db.dbConnect = dbConnect
	return &db, err
}

func (ds *DBStorage) Storing(ctx context.Context, logger log.Logger) {}

func (ds *DBStorage) Close() {
	if ds.dbConnect != nil {
		ds.dbConnect.Close()
	}
}

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
		if !db.Migrator().HasTable(tableName) {
			if err := db.Table(tableName).Migrator().CreateTable(&types.MetricModel{}); err != nil {
				return err
			}
		}
		db.Table(tableName).Create(metric)
		return nil
	}
}

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
				if !tx.Migrator().HasTable(tableName) {
					if err := tx.Table(tableName).Migrator().CreateTable(&types.MetricModel{}); err != nil {
						return err
					}
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
