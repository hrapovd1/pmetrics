// Часть модуля handlers содержит типы и методы middleware,
// реализующий сжатие при передаче по http.
package handlers

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"strings"

	"github.com/hrapovd1/pmetrics/internal/types"
	"github.com/hrapovd1/pmetrics/internal/usecase"
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
		key, err := usecase.GetPrivKey(mh.Config.CryptoKey, mh.logger)
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
		symmKey, err := usecase.DecryptKey(encData.Data0, key)
		if err != nil {
			mh.logger.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		dataJSON, err := usecase.DecryptData(encData.Data, symmKey)
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

func (mh *MetricsHandler) CheckAgentNetMiddle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case mh.Config.TrustedSubnet == "":
			next.ServeHTTP(w, r)
			return
		case r.Header.Get("X-Real-IP") != "":
			agentAddr := net.ParseIP(r.Header.Get("X-Real-IP"))
			if agentAddr == nil {
				break
			}
			allow, err := usecase.CheckAddr(agentAddr, mh.Config.TrustedSubnet)
			if err != nil {
				mh.logger.Printf("got error when check agent address: %v\n", err)
				break
			}
			if !allow {
				mh.logger.Printf("Try to connect from untusted address: %v\n", agentAddr.String())
				break
			}
			next.ServeHTTP(w, r)
			return
		default:
			mh.logger.Println("Try to connect without X-Real-IP header")
		}
		http.Error(w, "Unknown agent forbidden", http.StatusForbidden)
	})
}
