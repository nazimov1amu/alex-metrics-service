package middleware

import (
	"compress/gzip"
	"net/http"
	"strings"

	"github.com/Alexunder2003/alex-metrics-service/internal/encoding"
)

type compressorResponseWriter struct {
	http.ResponseWriter
	status      int
	wroteHeader bool
	gzipWriter  *gzip.Writer
}

func (w *compressorResponseWriter) WriteHeader(status int) {
	w.status = status
}

func (w *compressorResponseWriter) writeHeader(data []byte) {
	if w.wroteHeader {
		return
	}
	w.wroteHeader = true

	if w.Header().Get("Content-Type") == "" && len(data) > 0 {
		w.Header().Set("Content-Type", http.DetectContentType(data))
	}

	if isCompressible(w.Header().Get("Content-Type")) {
		w.Header().Set("Content-Encoding", "gzip")
		w.Header().Del("Content-Length")
		w.gzipWriter = gzip.NewWriter(w.ResponseWriter)
	}

	if w.status == 0 {
		w.status = http.StatusOK
	}
	w.ResponseWriter.WriteHeader(w.status)
}

func (w *compressorResponseWriter) Write(data []byte) (int, error) {
	w.writeHeader(data)

	if w.gzipWriter != nil {
		return w.gzipWriter.Write(data)
	}
	return w.ResponseWriter.Write(data)
}

func (w *compressorResponseWriter) Close() error {
	w.writeHeader(nil)
	if w.gzipWriter != nil {
		return w.gzipWriter.Close()
	}
	return nil
}

func isCompressible(contentType string) bool {
	return strings.Contains(contentType, "text/plain") ||
		strings.Contains(contentType, "application/json")
}

func WithCompressing(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.Header.Get("Content-Encoding"), "gzip") {
			decompressed, err := encoding.Decompress(r.Body)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			_ = r.Body.Close()
			r.Body = decompressed
		}

		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
			return
		}
		crw := &compressorResponseWriter{ResponseWriter: w}
		defer crw.Close()
		next.ServeHTTP(crw, r)
	})
}

func CompressingMiddleware(next http.Handler) http.Handler {
	return WithCompressing(next)
}
