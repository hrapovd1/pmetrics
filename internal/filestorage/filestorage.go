package filestorage

import (
	"bufio"
	"context"
	"encoding/json"
	"log"
	"os"
	"time"

	"github.com/hrapovd1/pmetrics/internal/config"
	"github.com/hrapovd1/pmetrics/internal/types"
)

type FileStorage struct {
	file   *os.File
	writer *bufio.Writer
	ctx    context.Context
	config config.Config
	buff   map[string]interface{}
}

func (fs *FileStorage) Append(key string, value int64) {
}

func (fs *FileStorage) Get(key string) interface{} {
	return nil
}

func (fs *FileStorage) GetAll() map[string]interface{} {
	return nil
}

func (fs *FileStorage) Rewrite(key string, value float64) {
}

func (fs *FileStorage) StoreAll(metric *[]types.Metric) {
}

func NewFileStorage(ctx context.Context, conf config.Config, buff map[string]interface{}) *FileStorage {
	fs := &FileStorage{}
	fs.ctx = ctx
	var err error
	fs.config = conf
	if conf.StoreFile == "" {
		fs.file = nil
		fs.writer = nil
		fs.config.IsRestore = false
		return fs
	}
	fs.buff = buff
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
	if fs.file != nil {
		return fs.file.Close()
	}
	return nil
}

func (fs *FileStorage) Restore() error {
	var err error
	var data []byte
	metrics := make([]types.Metric, 0)
	scan := bufio.NewScanner(fs.file)
	select {
	case <-fs.ctx.Done():
		return nil
	default:
		for scan.Scan() {
			data = scan.Bytes()
		}
		err = json.Unmarshal(data, &metrics)
		for _, metric := range metrics {
			switch metric.MType {
			case "gauge":
				fs.buff[metric.ID] = *metric.Value
			case "counter":
				fs.buff[metric.ID] = *metric.Delta
			}
		}
		return err
	}
}

func (fs *FileStorage) Store() error {
	metrics := make([]types.Metric, 0)
	select {
	case <-fs.ctx.Done():
		return nil
	default:
		for k, v := range fs.buff {
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

func (fs *FileStorage) Storing(logger *log.Logger) {
	defer fs.Close()
	if fs.config.StoreFile == "" {
		return
	}
	if fs.config.IsRestore {
		if err := fs.Restore(); err != nil {
			logger.Println(err)
		}
	}
	var storeInterval time.Duration
	if fs.config.StoreInterval > 0 {
		storeInterval = fs.config.StoreInterval
	} else {
		storeInterval = fs.config.ReportInterval
	}
	storeTick := time.NewTicker(storeInterval)
	defer storeTick.Stop()
	for {
		select {
		case <-fs.ctx.Done():
			return
		case <-storeTick.C:
			if err := fs.Store(); err != nil {
				logger.Println(err)
			}
		}
	}
}
