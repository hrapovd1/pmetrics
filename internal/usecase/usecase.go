package usecase

import (
	"errors"
	"fmt"

	"github.com/hrapovd1/pmetrics/internal/storage"
	"github.com/hrapovd1/pmetrics/internal/types"
)

func WriteMetric(ms *storage.MemStorage, data types.Metrics) error {
	switch data.MType {
	case "gauge":
		metricValue := storage.ToGauge(*data.Value)
		ms.Rewrite(data.ID, metricValue)
		return nil
	case "counter":
		metricValue := storage.ToCounter(*data.Delta)
		ms.Append(data.ID, metricValue)
		return nil
	default:
		return errors.New("undefined metric type")
	}
}

func GetMetric(ms *storage.MemStorage, data *types.Metrics) error {
	var err error

	switch data.MType {
	case "gauge":
		val := ms.Get(data.ID).(float64)
		data.Value = &val
		err = nil
	case "counter":
		val := ms.Get(data.ID).(int64)
		data.Delta = &val
		err = nil
	default:
		err = errors.New("undefined metric type")
	}
	return err
}

func GetTableMetrics(ms *storage.MemStorage) map[string]string {
	outTable := make(map[string]string)

	for k, v := range ms.GetAll() {
		switch value := v.(type) {
		case int64:
			outTable[k] = fmt.Sprint(value)
		case float64:
			outTable[k] = fmt.Sprint(value)
		}
	}
	return outTable
}
