// Package middleware содержит HTTP-middleware логирования, gzip, подписи и аудита.
package middleware

import (
	"net/http"
	"time"

	"go.uber.org/zap"
)

// NewLogger создаёт production-логгер zap в sugared-режиме.
func NewLogger() (*zap.SugaredLogger, error) {
	logger, err := zap.NewProduction()
	if err != nil {
		return nil, err
	}
	sugar := logger.Sugar()
	return sugar, nil
}

type (
	// RequestData хранит метод, URI и длительность обработанного запроса.
	RequestData struct {
		method  string
		uri     string
		latency time.Duration
	}

	// ResponseData хранит HTTP-статус и размер тела ответа.
	ResponseData struct {
		status int
		size   int
	}

	// LoggerResponseWriter запоминает статус и размер ответа, не меняя запись клиенту.
	LoggerResponseWriter struct {
		http.ResponseWriter
		*ResponseData
	}
)

// WriteHeader сохраняет статус и передаёт его исходному ResponseWriter.
func (w *LoggerResponseWriter) WriteHeader(status int) {
	w.ResponseData.status = status
	w.ResponseWriter.WriteHeader(status)
}

// Write дописывает размер тела к ResponseData и передаёт байты исходному ResponseWriter.
func (w *LoggerResponseWriter) Write(b []byte) (int, error) {
	size, err := w.ResponseWriter.Write(b)
	w.ResponseData.size += size
	return size, err
}

// Header возвращает заголовки исходного ResponseWriter.
func (w *LoggerResponseWriter) Header() http.Header {
	return w.ResponseWriter.Header()
}

// WithLogging пишет в лог метод, URL, длительность, статус и размер ответа.
func WithLogging(handler http.Handler, logger *zap.SugaredLogger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()
		loggerResponseWriter := &LoggerResponseWriter{
			ResponseWriter: w,
			ResponseData:   &ResponseData{},
		}
		handler.ServeHTTP(loggerResponseWriter, r)
		latency := time.Since(startTime)
		logger.Infow("request", "method", r.Method, "url", r.URL.String(), "latency", latency, "status", loggerResponseWriter.ResponseData.status, "size", loggerResponseWriter.ResponseData.size)
	})
}

// LoggingMiddleware возвращает middleware логирования запроса и ответа.
func LoggingMiddleware(logger *zap.SugaredLogger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return WithLogging(next, logger)
	}
}
