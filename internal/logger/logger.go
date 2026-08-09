package logger

import (
	"net/http"
	"time"

	"go.uber.org/zap"
)


func NewLogger() (*zap.SugaredLogger, error) {
    logger, err := zap.NewProduction()
    if err != nil {
        return nil, err
    }
    sugar := logger.Sugar()
    return sugar, nil
}

type (
    RequestData struct {
        method string
        uri string
        latency time.Duration
    }

    ResponseData struct {
        status int
        size int
    }

    LoggerResponseWriter struct {
        http.ResponseWriter
        *ResponseData
    }
)

func (w *LoggerResponseWriter) WriteHeader(status int) {
    w.ResponseData.status = status
    w.ResponseWriter.WriteHeader(status)
}

func (w *LoggerResponseWriter) Write(b []byte) (int, error) {
    size, err := w.ResponseWriter.Write(b)
    w.ResponseData.size += size
    return size, err
}

func (w *LoggerResponseWriter) Header() http.Header {
    return w.ResponseWriter.Header()
}

func WithLogging(handler http.Handler, logger *zap.SugaredLogger) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        startTime := time.Now()
        loggerResponseWriter := &LoggerResponseWriter{
            ResponseWriter: w,
            ResponseData: &ResponseData{},
        }
        handler.ServeHTTP(loggerResponseWriter, r)
        latency := time.Since(startTime)
        logger.Infow("request", "method", r.Method, "url", r.URL.String(), "latency", latency, "status", loggerResponseWriter.ResponseData.status, "size", loggerResponseWriter.ResponseData.size)
    })
}

func LoggingMiddleware(logger *zap.SugaredLogger) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return WithLogging(next, logger)
    }
}