// Модуль usecase содержит общие для проекта методы.
package usecase

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/hrapovd1/pmetrics/internal/storage"
	"github.com/hrapovd1/pmetrics/internal/types"
)

const (
	metricType    = 2 // Позиция типа метрики в url POST запроса
	metricName    = 3 // Позиция имени метрики в url POST запроса
	metricVal     = 4 // Позиция значения метркики в url POST запроса
	getMetricType = 2 // Позиция значения метрики в url GET запроса
	getMetricName = 3 // Позиция имени метрики в url GET запроса
)

// WriteMetric сохраняет метрику в Repository при получении через
// url POST запроса.
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

// GetMetric возвращает значение метрики из Repository при запросе
// через url GET запросом.
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

// WriteJSONMetric сохраняет метрику в Repository полученную в
// JSON формате POST запроса.
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

// WriteJSONMetrics сохраняет метрики полученные в JSON формате
// POST запроса в Repository.
func WriteJSONMetrics(ctx context.Context, data *[]types.Metric, repo types.Repository) {
	repo.StoreAll(ctx, data)
}

// GetJSONMetric возвращает метрику из Repository в JSON формате
// при GET запросе
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

// GetTableMetrics возвращает все метрики в строчном виде для
// последующего отображения на html странице.
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

func GetPrivKey(fname string, logger *log.Logger) (*rsa.PrivateKey, error) {
	// read private key from file
	keyFile, err := os.Open(fname)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := keyFile.Close(); err != nil {
			logger.Println(err)
		}
	}()
	pemPrivKey := make([]byte, 4*1024)
	n, err := keyFile.Read(pemPrivKey)
	if err != nil {
		return nil, err
	}
	pemPrivKey = pemPrivKey[:n]

	// decode private key from pem format
	privKey, _ := pem.Decode(pemPrivKey)
	if privKey == nil || privKey.Type != "PRIVATE KEY" {
		return nil, errors.New("not found PRIVATE KEY in file " + fname)
	}
	// parse private key from byte slice
	rsaPrivKey, err := x509.ParsePKCS8PrivateKey(privKey.Bytes)
	if err != nil {
		return nil, err
	}
	key, ok := rsaPrivKey.(*rsa.PrivateKey)
	if !ok {
		return key, errors.New("can't convert key to *rsa.PrivateKey")
	}
	return key, nil
}

func DecryptKey(kData string, privKey *rsa.PrivateKey) ([]byte, error) {
	// Get encrypted primary data
	encSymmKey, err := base64.StdEncoding.DecodeString(kData)
	if err != nil {
		return nil, err
	}
	// Decrypt primary data
	// Decrypt symm key
	symmKey, err := rsa.DecryptPKCS1v15(rand.Reader, privKey, encSymmKey)
	if err != nil {
		return nil, err
	}
	return symmKey, nil
}

func DecryptData(data string, symm []byte) ([]byte, error) {
	encJSON, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return nil, err
	}
	// Decrypt metrics data
	chpr, err := aes.NewCipher(symm)
	if err != nil {
		return nil, err
	}
	gcmDecrypt, err := cipher.NewGCM(chpr)
	if err != nil {
		return nil, err
	}
	nonceSize := gcmDecrypt.NonceSize()
	if len(encJSON) < nonceSize {
		return nil, errors.New("len(encJSON) < nonceSize")
	}
	nonce, encDataJSON := encJSON[:nonceSize], encJSON[nonceSize:]
	return gcmDecrypt.Open(nil, nonce, encDataJSON, nil)
}

func CheckAddr(addr net.IP, trustSubNet string) (bool, error) {
	subNet := strings.Split(trustSubNet, "/")
	trustNet := net.ParseIP(subNet[0])
	mask, err := strconv.Atoi(subNet[1])
	if err != nil {
		return false, err
	}
	trustMask := net.CIDRMask(mask, 32)
	agentNet := addr.Mask(trustMask)
	return agentNet.Equal(trustNet), nil
}
