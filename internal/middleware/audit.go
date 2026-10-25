package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"time"

	"github.com/Alexunder2003/alex-metrics-service/internal/model"
	"github.com/go-chi/chi/v5"
)

type auditResponseWriter struct {
	http.ResponseWriter
	status int
}

func (w *auditResponseWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

func (w *auditResponseWriter) Write(data []byte) (int, error) {
	return w.ResponseWriter.Write(data)
}

func AuditMiddleware(channels []chan *model.Audit) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			r.Body = io.NopCloser(bytes.NewReader(body))

			sw := &auditResponseWriter{ResponseWriter: w}
			next.ServeHTTP(sw, r)

			if sw.status != http.StatusOK {
				return
			}

			event := &model.Audit{
				Timestamp: time.Now().Unix(),
				Metrics:   metricNames(r, body),
				IPAddress: clientIP(r),
			}
			for _, ch := range channels {
				go func(ch chan *model.Audit) {
					ch <- event
				}(ch)
			}
		})
	}
}

func metricNames(r *http.Request, body []byte) []string {
	if name := chi.URLParam(r, "name"); name != "" {
		return []string{name}
	}

	if r.URL.Path == "/updates/" {
		var metrics []model.Metrics
		if err := json.Unmarshal(body, &metrics); err == nil {
			names := make([]string, 0, len(metrics))
			for _, metric := range metrics {
				names = append(names, metric.ID)
			}
			return names
		}
	}

	var metric model.Metrics
	if err := json.Unmarshal(body, &metric); err == nil && metric.ID != "" {
		return []string{metric.ID}
	}
	return []string{}
}

func clientIP(r *http.Request) string {
	if ip := r.Header.Get("X-Real-IP"); ip != "" {
		return ip
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
