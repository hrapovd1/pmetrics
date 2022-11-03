package usecase

import (
	"errors"
	"fmt"

	"github.com/hrapovd1/pmetrics/internal/storage"
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
