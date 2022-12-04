package storage

import (
	"context"
	"database/sql"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/hrapovd1/pmetrics/internal/config"
	"github.com/hrapovd1/pmetrics/internal/types"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type DBStorage struct {
	dbConnect *sql.DB
	buffer    map[string]interface{}
	config    config.Config
}

func NewDBStorage(backConf config.Config) (*DBStorage, error) {
	db := DBStorage{}
	db.config = backConf
	if backConf.DatabaseDSN == "" {
		return &db, nil
	}
	dbConnect, err := sql.Open("pgx", db.config.DatabaseDSN)
	db.dbConnect = dbConnect
	return &db, err
}

func (ds *DBStorage) Close() {
	if ds.dbConnect != nil {
		ds.dbConnect.Close()
	}
}

func (ds *DBStorage) IsOK() bool {
	if ds.dbConnect == nil {
		return false
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := ds.dbConnect.PingContext(ctx); err != nil {
		return false
	}
	return true
}

func (ds *DBStorage) PingDB(rw http.ResponseWriter, r *http.Request) {
	if !ds.IsOK() {
		http.Error(rw, "DB connect is NOT ok", http.StatusInternalServerError)
		return
	}
	rw.Header().Set("Content-Type", "text/html")
	rw.WriteHeader(http.StatusOK)
}

func (ds *DBStorage) Restore() error {
	db, err := gorm.Open(postgres.New(postgres.Config{Conn: ds.dbConnect}), &gorm.Config{})
	if err != nil {
		return err
	}
	tables, err := db.Migrator().GetTables()
	if err != nil {
		return err
	}
	metricTable, err := regexp.Compile("^" + types.DBtablePrefix)
	if err != nil {
		return err
	}

	for _, table := range tables {
		if !metricTable.MatchString(table) {
			continue
		}
		dbMetric := types.MetricModel{}
		db.Table(table).Last(&dbMetric)

		metric := types.Metric{}
		metric.ID = dbMetric.ID
		switch dbMetric.Mtype {
		case "gauge":
			ds.buffer[metric.ID] = dbMetric.Value.Float64
		case "counter":
			ds.buffer[metric.ID] = dbMetric.Delta.Int64
		}
	}
	return nil
}

func (ds *DBStorage) store(metric *types.MetricModel) error {
	db, err := gorm.Open(postgres.New(postgres.Config{Conn: ds.dbConnect}), &gorm.Config{})
	if err != nil {
		return err
	}
	tableName := strings.ToLower(types.DBtablePrefix + metric.ID)
	if !db.Migrator().HasTable(tableName) {
		if err := db.Table(tableName).Migrator().CreateTable(&types.MetricModel{}); err != nil {
			return err
		}
	}
	db.Table(tableName).Create(metric)
	return nil
}

func (ds *DBStorage) storeBatch(metrics *[]types.MetricModel) error {
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
