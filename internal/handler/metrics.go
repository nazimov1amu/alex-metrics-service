package handler

import (
	"embed"
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"strconv"

	"github.com/Alexunder2003/alex-metrics-service/internal/model"
	"github.com/Alexunder2003/alex-metrics-service/internal/service"
	"github.com/go-chi/chi/v5"
)

//go:embed templates/*.html
var templatesFS embed.FS

var metricsTmpl = template.Must(template.ParseFS(templatesFS, "templates/*.html"))

type MetricsService interface {
	Update(input model.MetricsInput) (model.Metrics, error)
	Get(name string) (model.Metrics, error)
	GetBulk() ([]model.Metrics, error)
}

type Handler struct {
	svc MetricsService
}

func NewMetricsHandler(svc MetricsService) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	input := model.MetricsInput{
		Name:     chi.URLParam(r, "name"),
		MType:    chi.URLParam(r, "type"),
		RawValue: chi.URLParam(r, "value"),
	}

	metric, err := h.svc.Update(input)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidMetricType):
			http.Error(w, err.Error(), http.StatusBadRequest)
		case errors.Is(err, service.ErrInvalidCounterValue):
			http.Error(w, err.Error(), http.StatusBadRequest)
		case errors.Is(err, service.ErrInvalidGaugeValue):
			http.Error(w, err.Error(), http.StatusBadRequest)
		default:
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusOK)

	switch metric.MType {
	case model.Counter:
		w.Write(fmt.Appendf(nil, "%d", *metric.Delta))
	case model.Gauge:
		w.Write(fmt.Appendf(nil, "%f", *metric.Value))
	}
}


func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")

	metric, err := h.svc.Get(name)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrMetricNotFound):
			http.Error(w, err.Error(), http.StatusNotFound)
		default:
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusOK)
	switch metric.MType {
	case model.Counter:
		w.Write([]byte(strconv.FormatInt(*metric.Delta, 10)))
	case model.Gauge:
		w.Write([]byte(strconv.FormatFloat(*metric.Value, 'f', -1, 64)))
	}
}

func (h *Handler) GetBulk(w http.ResponseWriter, r *http.Request)  {
	metrics, err := h.svc.GetBulk()
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	err = metricsTmpl.Execute(w, metrics)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

}