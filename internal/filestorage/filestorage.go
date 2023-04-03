// Модуль filestorage содержит типы и методы для
// хранения метрик в файле, в формате JSON.
package filestorage

import (
	"bufio"
	"context"
	"encoding/json"
	"log"
	"os"
	"sync"
	"time"

	"github.com/hrapovd1/pmetrics/internal/config"
	"github.com/hrapovd1/pmetrics/internal/storage"
	"github.com/hrapovd1/pmetrics/internal/types"
)

// FileStorage - тип реализует types.Repository интерфейс для
// хранения метрик в файле.
type FileStorage struct {
	file   *os.File
	writer *bufio.Writer
	ms     *storage.MemStorage
}

// NewFileStorage создает объект типа FileStorage
func NewFileStorage(conf config.Config, buff map[string]interface{}) *FileStorage {
	fs := &FileStorage{
		ms: storage.NewMemStorage(
			storage.WithBuffer(buff),
		),
	}
	var err error
	fileOptions := os.O_RDWR | os.O_CREATE | os.O_APPEND
	if conf.StoreInterval == 0 {
		fileOptions = fileOptions | os.O_SYNC
	}

	fs.file, err = os.OpenFile(conf.StoreFile, fileOptions, 0777)
	if err != nil {
		panic(err)
	}
	fs.writer = bufio.NewWriter(fs.file)

	return fs
}

// Append сохраняет новое значение типа counter с дозаписью к старому
func (fs *FileStorage) Append(ctx context.Context, key string, value int64) {
	fs.ms.Append(ctx, key, value)
}

// Get возвращает значение метрики переданной через key
func (fs *FileStorage) Get(ctx context.Context, key string) interface{} {
	return fs.ms.Get(ctx, key)
}

// GetAll возвращает все метрики
func (fs *FileStorage) GetAll(ctx context.Context) map[string]interface{} {
	return fs.ms.GetAll(ctx)
}

// Rewrite перезаписывает значение метрики типа gauge
func (fs *FileStorage) Rewrite(ctx context.Context, key string, value float64) {
	fs.ms.Rewrite(ctx, key, value)
}

// StoreAll сохраняет все полученные метрики через слайс metrics
func (fs *FileStorage) StoreAll(ctx context.Context, metrics *[]types.Metric) {
	fs.ms.StoreAll(ctx, metrics)
}

// Close закрывает открытый ранее файл, необходимо запускать в defer
func (fs *FileStorage) Close() error {
	return fs.file.Close()
}

// Ping для реализации интерфейса Storager
func (fs *FileStorage) Ping(ctx context.Context) bool {
	return false
}

// Restore восстанавливае значение метрик при запуске из файла
func (fs *FileStorage) Restore(ctx context.Context) error {
	var err error
	var data []byte
	metrics := make([]types.Metric, 0)
	scan := bufio.NewScanner(fs.file)
	select {
	case <-ctx.Done():
		return nil
	default:
		for scan.Scan() {
			data = scan.Bytes()
		}
		if err = scan.Err(); err != nil {
			return err
		}
		if err = json.Unmarshal(data, &metrics); err != nil {
			return err
		}
		fs.ms.StoreAll(ctx, &metrics)
		return nil
	}
}

// Store сбрасывает значения метрик из памяти в файл в JSON формате
func (fs *FileStorage) Store(ctx context.Context) error {
	metrics := make([]types.Metric, 0)
	select {
	case <-ctx.Done():
		return nil
	default:
		buff := fs.ms.GetAll(ctx)
		for k, v := range buff {
			metric := types.Metric{ID: k}
			switch val := v.(type) {
			case int64:
				metric.MType = "counter"
				metric.Delta = &val
			case float64:
				metric.MType = "gauge"
				metric.Value = &val
			}
			metrics = append(metrics, metric)
		}
		data, err := json.Marshal(metrics)
		if err != nil {
			return err
		}
		if _, err := fs.writer.Write(data); err != nil {
			return err
		}
		if err := fs.writer.WriteByte('\n'); err != nil {
			return err
		}
		return fs.writer.Flush()
	}
}

// Storing запускается в отдельной go routine для сохранения метрик в файл
func (fs *FileStorage) Storing(ctx context.Context, logger *log.Logger, interval time.Duration, restore bool) {
	waitGroup := ctx.Value(types.Waitgrp("WG")).(*sync.WaitGroup)
	defer func() {
		if err := fs.Close(); err != nil {
			logger.Printf("fs.Close: %v", err)
		}
		waitGroup.Done()
	}()

	if restore {
		if err := fs.Restore(ctx); err != nil {
			logger.Printf("fs.Restore err: %v", err)
		}
	}
	storeTick := time.NewTicker(interval)
	defer storeTick.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-storeTick.C:
			if err := fs.Store(ctx); err != nil {
				logger.Printf("fs.Store err: %v", err)
			}
		}
	}
}
