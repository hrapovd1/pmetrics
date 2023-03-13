package dbstorage

import (
	"fmt"
	"log"

	"github.com/hrapovd1/pmetrics/internal/storage"
)

func ExampleNewDBStorage() {
	dsn := "postgres://postgres:postgres@localhost:5432/postgres"
	storage := storage.NewMemStorage()

	dbStorage, _ := NewDBStorage(dsn, log.Default(), storage)

	fmt.Println(dbStorage)
}
