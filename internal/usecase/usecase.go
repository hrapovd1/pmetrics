package usecase

import (
	"context"
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

func WriteMetric(ctx context.Context, path []string, repo types.Repository) error {
	metricKey := path[metricName]
	switch path[metricType] {
	case "gauge":
		metricValue, err := storage.StrToFloat64(path[metricVal])
		if err == nil {
			repo.Rewrite(ctx, metricKey, metricValue)
		}
		return err
	case "counter":
		metricValue, err := storage.StrToInt64(path[metricVal])
		if err == nil {
			repo.Append(ctx, metricKey, metricValue)
		}
		return err
	default:
		return errors.New("undefined metric type")
	}
}

func GetMetric(ctx context.Context, repo types.Repository, path []string) (string, error) {
	metricType := path[getMetricType]
	metric := path[getMetricName]
	var metricValue string
	var err error

	if metricType == "gauge" || metricType == "counter" {
		metricVal := repo.Get(ctx, metric)
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

func WriteJSONMetric(ctx context.Context, data types.Metric, repo types.Repository) error {
	switch data.MType {
	case "gauge":
		repo.Rewrite(ctx, data.ID, *data.Value)
		return nil
	case "counter":
		repo.Append(ctx, data.ID, *data.Delta)
		return nil
	default:
		return errors.New("undefined metric type")
	}
}

func WriteJSONMetrics(ctx context.Context, data *[]types.Metric, repo types.Repository) {
	repo.StoreAll(ctx, data)
}

func GetJSONMetric(ctx context.Context, repo types.Repository, data *types.Metric) error {
	var err error

	switch data.MType {
	case "gauge":
		val := repo.Get(ctx, data.ID)
		if val == nil {
			return errors.New("not found")
		}
		value := val.(float64)
		data.Value = &value
		err = nil
	case "counter":
		val := repo.Get(ctx, data.ID)
		if val == nil {
			return errors.New("not found")
		}
		value := val.(int64)
		data.Delta = &value
		err = nil
	default:
		err = errors.New("undefined metric type")
	}
	return err
}

func GetTableMetrics(ctx context.Context, repo types.Repository) map[string]string {
	outTable := make(map[string]string)

	for k, v := range repo.GetAll(ctx) {
		switch value := v.(type) {
		case int64:
			outTable[k] = fmt.Sprint(value)
		case float64:
			outTable[k] = fmt.Sprint(value)
		}
	}
	return outTable
}
