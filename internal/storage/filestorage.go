package storage

import (
	"bufio"
	"encoding/json"
	"os"
	"time"

	"github.com/hrapovd1/pmetrics/internal/config"
	"github.com/hrapovd1/pmetrics/internal/types"
)

func newBackend(backConf config.Config, buff map[string]interface{}) fileStorage {
	fs := fileStorage{}
	if backConf.StoreFile == "" {
		return fs
	}
	var err error
	fs.config = backConf
	fs.buff = buff
	fs.file, err = os.OpenFile(backConf.StoreFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0777)
	if err != nil {
		panic(err)
	}
	fs.writer = bufio.NewWriter(fs.file)
	return fs
}

func (fs *fileStorage) Close() error {
	return fs.file.Close()
}

func (fs *fileStorage) Restore() error {
	var err error
	scan := bufio.NewScanner(fs.file)
	var data []byte
	for scan.Scan() {
		data = scan.Bytes()
	}
	metrics := make([]types.Metric, 0)
	err = json.Unmarshal(data, &metrics)
	for _, metric := range metrics {
		switch metric.MType {
		case "gauge":
			fs.buff[metric.ID] = metric.Value
		case "counter":
			fs.buff[metric.ID] = metric.Delta
		}
	}

	return err
}

func (fs *fileStorage) Store() error {
	metrics := make([]types.Metric, 0)
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

func (fs *fileStorage) Storing(donech chan struct{}) {
	defer fs.Close()
	storeTick := time.NewTicker(fs.config.StoreInterval)
	select {
	case <-donech:
		return
	case <-storeTick.C:
		fs.Store()
	}
}
