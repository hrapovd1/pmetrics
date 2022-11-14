package storage

import (
	"bufio"
	"encoding/json"
	"os"

	"github.com/hrapovd1/pmetrics/internal/config"
	"github.com/hrapovd1/pmetrics/internal/types"
	"github.com/hrapovd1/pmetrics/internal/usecase"
)

func newBackend(backConf config.Config, buff *map[string]interface{}) fileStorage {
	if backConf.StoreFile == "" {
		return fileStorage{}
	}
	var err error
	fs := fileStorage{}
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
	ms := NewMemStorage(fs.config, WithBuffer(*fs.buff))
	var err error
	scan := bufio.NewScanner(fs.file)
	var data []byte
	for scan.Scan() {
		data = scan.Bytes()
	}
	metrics := make([]types.Metrics, 0)
	err = json.Unmarshal(data, &metrics)
	for _, metric := range metrics {
		err = usecase.WriteJSONMetric(&ms, metric)
	}

	return err
}

func (fs *fileStorage) Store() error {
	metrics := make([]types.Metrics, 0)
	for k, v := range *fs.buff {
		metric := types.Metrics{ID: k}
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
