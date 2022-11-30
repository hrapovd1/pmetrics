package storage

import (
	"context"
	"database/sql"
	"log"
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
		metric.ID = strings.Split(table, types.DBtablePrefix)[1]
		switch dbMetric.Mtype {
		case "gauge":
			ds.buffer[metric.ID] = dbMetric.Value.Float64
		case "counter":
			ds.buffer[metric.ID] = dbMetric.Delta.Int64
		}
	}
	return nil
}

func (ds *DBStorage) Store() error {
	db, err := gorm.Open(postgres.New(postgres.Config{Conn: ds.dbConnect}), &gorm.Config{})
	if err != nil {
		return err
	}
	for k, v := range ds.buffer {
		if !db.Migrator().HasTable(types.DBtablePrefix + k) {
			if err := db.Table(types.DBtablePrefix + k).Migrator().CreateTable(&types.MetricModel{}); err != nil {
				return err
			}
		}
		metricVal := types.MetricModel{}
		switch val := v.(type) {
		case float64:
			metricVal.Mtype = "gauge"
			metricVal.Value = sql.NullFloat64{Float64: val, Valid: true}
		case int64:
			metricVal.Mtype = "counter"
			metricVal.Delta = sql.NullInt64{Int64: val, Valid: true}
		}
		db.Table(types.DBtablePrefix + k).Create(&metricVal)
	}
	return nil
}

func (ds *DBStorage) Storing(donech chan struct{}, logger *log.Logger) {
	defer ds.Close()

	if ds.config.DatabaseDSN == "" {
		return
	}

	if ds.config.IsRestore {
		if err := ds.Restore(); err != nil {
			logger.Println(err)
		}
	}

	var storeInterval time.Duration
	if ds.config.StoreInterval > 0 {
		storeInterval = ds.config.StoreInterval
	} else {
		storeInterval = ds.config.ReportInterval
	}
	storeTick := time.NewTicker(storeInterval)
	defer storeTick.Stop()
	for {
		select {
		case <-donech:
			return
		case <-storeTick.C:
			if err := ds.Store(); err != nil {
				logger.Println(err)
			}
		}
	}

}
