// Package handler содержит HTTP-обработчики сервиса сбора метрик.
package handler

import (
	"context"
	"embed"
	"encoding/json"
	"errors"
	"html/template"
	"net/http"
	"strconv"

	"github.com/Alexunder2003/alex-metrics-service/internal/model"
	"github.com/Alexunder2003/alex-metrics-service/internal/service"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

//go:embed templates/*.html
var templatesFS embed.FS

var metricsTmpl = template.Must(template.ParseFS(templatesFS, "templates/*.html"))

// MetricsService описывает операции с метриками, которые вызывают HTTP-обработчики.
type MetricsService interface {
	// Update сохраняет одну метрику.
	Update(ctx context.Context, metric *model.Metrics) error
	// Get возвращает метрику по идентификатору.
	Get(ctx context.Context, id string) (model.Metrics, error)
	// GetBulk возвращает все сохранённые метрики.
	GetBulk(ctx context.Context) ([]model.Metrics, error)
	// BulkUpdate сохраняет набор метрик.
	BulkUpdate(ctx context.Context, metrics []model.Metrics) error
}

// Handler обрабатывает HTTP-запросы к API метрик.
type Handler struct {
	svc    MetricsService
	logger *zap.SugaredLogger
}

// NewMetricsHandler создаёт обработчик метрик поверх сервиса и логгера.
func NewMetricsHandler(svc MetricsService, logger *zap.SugaredLogger) *Handler {
	return &Handler{svc: svc, logger: logger}
}

// UpdateJSON обрабатывает POST /update/ и сохраняет одну метрику из JSON-тела.
func (h *Handler) UpdateJSON(w http.ResponseWriter, r *http.Request) {
	var metric model.Metrics

	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(&metric); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.svc.Update(r.Context(), &metric); err != nil {
		h.writeError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	writeJSON(w, http.StatusOK, metric)
}

// ValueJSON обрабатывает POST /value/ и возвращает метрику по идентификатору из JSON-тела.
func (h *Handler) ValueJSON(w http.ResponseWriter, r *http.Request) {
	var metric model.Metrics

	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(&metric); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	metric, err := h.svc.Get(r.Context(), metric.ID)
	if err != nil {
		h.writeError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	writeJSON(w, http.StatusOK, metric)
}

// UpdatePath обрабатывает POST /update/{type}/{name}/{value} и сохраняет метрику из пути.
func (h *Handler) UpdatePath(w http.ResponseWriter, r *http.Request) {
	metric, err := metricFromPath(
		chi.URLParam(r, "type"),
		chi.URLParam(r, "name"),
		chi.URLParam(r, "value"),
	)
	if err != nil {
		h.writeError(w, err)
		return
	}

	if err = h.svc.Update(r.Context(), &metric); err != nil {
		h.writeError(w, err)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")

	w.WriteHeader(http.StatusOK)
	switch metric.MType {
	case model.Counter:
		_, _ = w.Write([]byte(strconv.FormatInt(*metric.Delta, 10)))
	case model.Gauge:
		_, _ = w.Write([]byte(strconv.FormatFloat(*metric.Value, 'f', 6, 64)))
	}
}

// ValuePath обрабатывает GET /value/{type}/{name} и возвращает значение метрики текстом.
func (h *Handler) ValuePath(w http.ResponseWriter, r *http.Request) {
	metric, err := h.svc.Get(r.Context(), chi.URLParam(r, "name"))
	if err != nil {
		h.writeError(w, err)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	switch metric.MType {
	case model.Counter:
		_, _ = w.Write([]byte(strconv.FormatInt(*metric.Delta, 10)))
	case model.Gauge:
		_, _ = w.Write([]byte(strconv.FormatFloat(*metric.Value, 'f', -1, 64)))
	}
}

// GetBulk обрабатывает GET / и отдаёт HTML-страницу со всеми метриками.
func (h *Handler) GetBulk(w http.ResponseWriter, r *http.Request) {
	metrics, err := h.svc.GetBulk(r.Context())
	if err != nil {
		h.writeError(w, err)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	if err = metricsTmpl.Execute(w, metrics); err != nil {
		h.writeError(w, err)
	}
}

func metricFromPath(mType, name, raw string) (model.Metrics, error) {
	metric := model.Metrics{ID: name, MType: mType}
	switch mType {
	case model.Counter:
		v, err := strconv.ParseInt(raw, 10, 64)
		if err != nil {
			return model.Metrics{}, service.ErrInvalidCounterValue
		}
		metric.Delta = &v
	case model.Gauge:
		v, err := strconv.ParseFloat(raw, 64)
		if err != nil {
			return model.Metrics{}, service.ErrInvalidGaugeValue
		}
		metric.Value = &v
	default:
		return model.Metrics{}, service.ErrInvalidMetricType
	}
	return metric, nil
}

// BulkUpdateJSON обрабатывает POST /updates/ и сохраняет список метрик из JSON-тела.
func (h *Handler) BulkUpdateJSON(w http.ResponseWriter, r *http.Request) {
	var metrics []model.Metrics
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(&metrics); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.svc.BulkUpdate(r.Context(), metrics); err != nil {
		h.writeError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	writeJSON(w, http.StatusOK, metrics)
}

func (h *Handler) writeError(w http.ResponseWriter, err error) {
	h.logger.Errorw("request failed", "error", err)
	switch {
	case errors.Is(err, service.ErrInvalidMetricType):
		http.Error(w, err.Error(), http.StatusBadRequest)
	case errors.Is(err, service.ErrInvalidCounterValue):
		http.Error(w, err.Error(), http.StatusBadRequest)
	case errors.Is(err, service.ErrInvalidGaugeValue):
		http.Error(w, err.Error(), http.StatusBadRequest)
	case errors.Is(err, service.ErrMetricNotFound):
		http.Error(w, err.Error(), http.StatusNotFound)
	default:
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}
