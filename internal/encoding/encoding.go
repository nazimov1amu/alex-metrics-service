package encoding

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"strings"
)

type CompressorResponseWriter struct {
	rw   http.ResponseWriter
	grw *gzip.Writer
}

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

func Decompress(data io.ReadCloser) (io.ReadCloser, error) {
	gz, err := gzip.NewReader(data)
	if err != nil {
		return nil, err
	}
	return gz, nil
}

func NewCompressorResponseWriter(w http.ResponseWriter) *CompressorResponseWriter {
	return &CompressorResponseWriter{
		rw:     w,
		grw:    gzip.NewWriter(w),
	}
}


func (w *CompressorResponseWriter) Write(data []byte) (int, error) {
	compressed, err := Compress(data)
	if err != nil {
		return 0, err
	}

	return w.rw.Write(compressed)
}

func (w *CompressorResponseWriter) Header() http.Header {
	return w.rw.Header()
}

func (w *CompressorResponseWriter) WriteHeader(statusCode int) {
	w.rw.WriteHeader(statusCode)
}

func WithCompressing(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.Header.Get("Content-Encoding"), "gzip") {
			defer r.Body.Close()
			
			decompressed, err := Decompress(r.Body)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			r.Body = decompressed
		}

		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
			return
		}

		crw := NewCompressorResponseWriter(w)
		crw.Header().Set("Content-Encoding", "gzip")
		next.ServeHTTP(crw, r)
	})
}

func CompressingMiddleware(next http.Handler) http.Handler {
	return WithCompressing(next)
}