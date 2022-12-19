package filestorage

import (
	"bufio"
	"context"
	"encoding/json"
	"log"
	"os"
	"time"

	"github.com/hrapovd1/pmetrics/internal/config"
	"github.com/hrapovd1/pmetrics/internal/storage"
	"github.com/hrapovd1/pmetrics/internal/types"
)

type FileStorage struct {
	file   *os.File
	writer *bufio.Writer
	ms     *storage.MemStorage
}

func (fs *FileStorage) Append(ctx context.Context, key string, value int64) {
	fs.ms.Append(ctx, key, value)
}

func (fs *FileStorage) Get(ctx context.Context, key string) interface{} {
	return fs.ms.Get(ctx, key)
}

func (fs *FileStorage) GetAll(ctx context.Context) map[string]interface{} {
	return fs.ms.GetAll(ctx)
}

func (fs *FileStorage) Rewrite(ctx context.Context, key string, value float64) {
	log.Printf("Rewrite FileStorage gauge: %v, %v", key, value)
	fs.ms.Rewrite(ctx, key, value)
}

func (fs *FileStorage) StoreAll(ctx context.Context, metrics *[]types.Metric) {
	fs.ms.StoreAll(ctx, metrics)
}

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

func (fs *FileStorage) Close() error {
	return fs.file.Close()
}

func (fs *FileStorage) Ping(ctx context.Context) bool {
	return false
}

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
		err = json.Unmarshal(data, &metrics)
		fs.ms.StoreAll(ctx, &metrics)
		return err
	}
}

func (fs *FileStorage) Store(ctx context.Context) error {
	metrics := make([]types.Metric, 0)
	select {
	case <-ctx.Done():
		return nil
	default:
		buff := fs.ms.GetAll(ctx)
		log.Printf("FileStorage.Store.buff = %v", buff)
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

func (fs *FileStorage) Storing(ctx context.Context, logger *log.Logger, interval time.Duration, restore bool) {
	defer fs.Close()
	if restore {
		if err := fs.Restore(ctx); err != nil {
			logger.Println(err)
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
				logger.Println(err)
			}
		}
	}
}
