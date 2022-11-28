package storage

import (
	"context"
	"database/sql"
	"net/http"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/hrapovd1/pmetrics/internal/config"
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