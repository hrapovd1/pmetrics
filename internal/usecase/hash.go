package usecase

import (
	"crypto/hmac"
	"crypto/sha256"
	"errors"
	"fmt"

	"github.com/hrapovd1/pmetrics/internal/types"
)

// Write sign data with hash function here
func SignData(data *types.Metric, key string) error {
	h := hmac.New(sha256.New, []byte(key))
	if data.MType == "counter" {
		_, err := h.Write([]byte(fmt.Sprintf("%s:%s:%d", data.ID, data.MType, *data.Delta)))
		if err != nil {
			return err
		}
		data.Hash = fmt.Sprintf("%x", h.Sum(nil))
		return nil
	} else if data.MType == "gauge" {
		_, err := h.Write([]byte(fmt.Sprintf("%s:%s:%f", data.ID, data.MType, *data.Value)))
		if err != nil {
			return err
		}
		data.Hash = fmt.Sprintf("%x", h.Sum(nil))
		return nil
	}
	return errors.New("undefined data.MType")
}

// Write check sign data hash function here
func IsSignEqual(data types.Metric, key string) bool {
	signRemote := []byte(data.Hash)
	if err := SignData(&data, key); err != nil {
		return false
	}
	signLocal := []byte(data.Hash)
	return hmac.Equal(signRemote, signLocal)
}
