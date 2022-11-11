package usecase

import (
	"errors"
	"fmt"

	"github.com/hrapovd1/pmetrics/internal/storage"
	"github.com/hrapovd1/pmetrics/internal/types"
)

const (
	metricType    = 2
	metricName    = 3
	metricVal     = 4
	getMetricType = 2
	getMetricName = 3
)

func WriteMetric(ms *storage.MemStorage, path []string) error {
	metricKey := path[metricName]
	switch path[metricType] {
	case "gauge":
		metricValue, err := storage.StrToGauge(path[metricVal])
		if err == nil {
			ms.Rewrite(metricKey, metricValue)
		}
		return err
	case "counter":
		metricValue, err := storage.StrToCounter(path[metricVal])
		if err == nil {
			ms.Append(metricKey, metricValue)
		}
		return err
	default:
		return errors.New("undefined metric type")
	}
}

func GetMetric(ms *storage.MemStorage, path []string) (string, error) {
	metricType := path[getMetricType]
	metric := path[getMetricName]
	var metricValue string
	var err error

	if metricType == "gauge" || metricType == "counter" {
		metricVal := ms.Get(metric)
		switch metricVal := metricVal.(type) {
		case int64:
			metricValue = fmt.Sprint(metricVal)
		case float64:
			metricValue = fmt.Sprint(metricVal)
		case nil:
			metricValue = ""
		}
	} else {
		err = errors.New("undefined metric type")
	}
	return metricValue, err
}

func WriteJSONMetric(ms *storage.MemStorage, data types.Metrics) error {
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

func GetJSONMetric(ms *storage.MemStorage, data *types.Metrics) error {
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
