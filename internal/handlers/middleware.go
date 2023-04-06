// Часть модуля handlers содержит типы и методы middleware,
// реализующий сжатие при передаче по http.
package handlers

import (
	"bytes"
	"compress/gzip"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/hrapovd1/pmetrics/internal/types"
)

// тип http ответа со сжатием
type gzipWriter struct {
	http.ResponseWriter
	Writer io.Writer
}

// Write реализует интерфейс Writer
func (w gzipWriter) Write(b []byte) (int, error) {
	// w.Writer будет отвечать за gzip-сжатие, поэтому пишем в него
	return w.Writer.Write(b)
}

// GzipMiddle промежуточный обработчик запросов для сжатия/распаковки
func (mh *MetricsHandler) GzipMiddle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// проверяем, что клиент поддерживает gzip-сжатие
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			// если gzip не поддерживается, передаём управление
			// дальше без изменений
			next.ServeHTTP(w, r)
			return
		}

		// создаём gzip.Writer поверх текущего w
		gz, err := gzip.NewWriterLevel(w, gzip.BestSpeed)
		if err != nil {
			_, err := io.WriteString(w, err.Error())
			if err != nil {
				mh.logger.Println(err)
			}
			return
		}
		defer func() {
			if err := gz.Close(); err != nil {
				mh.logger.Println(err)
			}
		}()

		w.Header().Set("Content-Encoding", "gzip")
		// передаём обработчику страницы переменную типа gzipWriter для вывода данных
		next.ServeHTTP(gzipWriter{ResponseWriter: w, Writer: gz}, r)
	})
}

func (mh *MetricsHandler) DecryptMiddle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Header.Get("Encrypt-Type"), "1") {
			next.ServeHTTP(w, r)
			return
		}
		if mh.Config.CryptoKey == "" {
			mh.logger.Print("got encrypted request, but CryptoKey wasn't provided")
			http.Error(w, "Server doesn't support encryption", http.StatusInternalServerError)
			return
		}
		key, err := getPrivKey(mh.Config.CryptoKey, mh.logger)
		if err != nil {
			mh.logger.Printf("when open key file %s, got error: %v", mh.Config.CryptoKey, err)
			http.Error(w, "Server doesn't support encryption", http.StatusInternalServerError)
			return
		}
		defer func() {
			if err := r.Body.Close(); err != nil {
				mh.logger.Println(err)
			}
		}()
		body, err := io.ReadAll(r.Body) // Get body from request
		if err != nil {
			mh.logger.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Decrypt data and send next
		encData := types.EncData{}
		err = json.Unmarshal(body, &encData)
		if err != nil {
			mh.logger.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		dataJSON, err := decryptData(encData, key)
		if err != nil {
			mh.logger.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		r.Body = io.NopCloser(bytes.NewReader(dataJSON))
		r.ContentLength = int64(len(dataJSON))

		next.ServeHTTP(w, r)
	})
}

func getPrivKey(fname string, logger *log.Logger) (*rsa.PrivateKey, error) {
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

func decryptData(dData types.EncData, privKey *rsa.PrivateKey) ([]byte, error) {
	// Get encrypted primary data
	encSymmKey, err := base64.StdEncoding.DecodeString(dData.Data0)
	if err != nil {
		return nil, err
	}
	encJSON, err := base64.StdEncoding.DecodeString(dData.Data)
	if err != nil {
		return nil, err
	}
	// Decrypt primary data
	// Decrypt symm key
	symmKey, err := rsa.DecryptPKCS1v15(rand.Reader, privKey, encSymmKey)
	if err != nil {
		return nil, err
	}
	// Decrypt metrics data
	chpr, err := aes.NewCipher(symmKey)
	if err != nil {
		return nil, err
	}
	gcmDecrypt, err := cipher.NewGCM(chpr)
	if err != nil {
		return nil, err
	}
	nonceSize := gcmDecrypt.NonceSize()
	if len(encJSON) < nonceSize {
		return nil, err
	}
	nonce, encDataJSON := encJSON[:nonceSize], encJSON[nonceSize:]
	return gcmDecrypt.Open(nil, nonce, encDataJSON, nil)
}
