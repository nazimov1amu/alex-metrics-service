package handler

import (
	"context"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io"
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

type MetricsService interface {
	Update(ctx context.Context, metric *model.Metrics) error
	Get(ctx context.Context, id string) (model.Metrics, error)
	GetBulk(ctx context.Context) ([]model.Metrics, error)
	BulkUpdate(ctx context.Context, metrics []model.Metrics) error
}

type Handler struct {
	svc    MetricsService
	logger *zap.SugaredLogger
}

func NewMetricsHandler(svc MetricsService, logger *zap.SugaredLogger) *Handler {
	return &Handler{svc: svc, logger: logger}
}

func (h *Handler) UpdateJSON(w http.ResponseWriter, r *http.Request) {
	var metric model.Metrics

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if err = json.Unmarshal(body, &metric); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err = h.svc.Update(r.Context(), &metric); err != nil {
		h.writeError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	writeJSON(w, http.StatusOK, metric)
}

func (h *Handler) ValueJSON(w http.ResponseWriter, r *http.Request) {
	var metric model.Metrics

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if err = json.Unmarshal(body, &metric); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	metric, err = h.svc.Get(r.Context(), metric.ID)
	if err != nil {
		h.writeError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	writeJSON(w, http.StatusOK, metric)
}

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
		_, _ = w.Write(fmt.Appendf(nil, "%d", *metric.Delta))
	case model.Gauge:
		_, _ = w.Write(fmt.Appendf(nil, "%f", *metric.Value))
	}
}

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

func (h *Handler) BulkUpdateJSON(w http.ResponseWriter, r *http.Request) {
	var metrics []model.Metrics
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()
	
	if err = json.Unmarshal(body, &metrics); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err = h.svc.BulkUpdate(r.Context(), metrics); err != nil {
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
