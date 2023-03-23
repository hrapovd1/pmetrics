// Часть модуля handlers содержит типы и методы middleware,
// реализующий сжатие при передаче по http.
package handlers

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
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
		defer mh.logger.Println(gz.Close())

		w.Header().Set("Content-Encoding", "gzip")
		// передаём обработчику страницы переменную типа gzipWriter для вывода данных
		next.ServeHTTP(gzipWriter{ResponseWriter: w, Writer: gz}, r)
	})
}
