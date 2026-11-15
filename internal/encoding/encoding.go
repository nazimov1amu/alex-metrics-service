// Package encoding сжимает и распаковывает тела HTTP-запросов алгоритмом gzip.
package encoding

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
)

// CompressorResponseWriter пишет тело ответа через gzip.
type CompressorResponseWriter struct {
	rw  http.ResponseWriter
	grw *gzip.Writer
}

// Compress возвращает data, сжатые gzip.
func Compress(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	if _, err := gz.Write(data); err != nil {
		return nil, err
	}
	if err := gz.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// Decompress открывает gzip-поток поверх data.
// Вызывающий код закрывает возвращённый ReadCloser.
func Decompress(data io.ReadCloser) (io.ReadCloser, error) {
	gz, err := gzip.NewReader(data)
	if err != nil {
		return nil, err
	}
	return gz, nil
}
