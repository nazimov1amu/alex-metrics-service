// Package service содержит бизнес-логику хранения метрик, аудита и агента сбора.
package service

import (
	"context"
	"errors"

	"github.com/Alexunder2003/alex-metrics-service/internal/config"
	"github.com/Alexunder2003/alex-metrics-service/internal/model"
)

var (
	// ErrInvalidMetricType возвращается, когда тип метрики не counter и не gauge.
	ErrInvalidMetricType = errors.New("invalid metric type")
	// ErrInvalidCounterValue возвращается, когда у counter нет целочисленного delta.
	ErrInvalidCounterValue = errors.New("invalid counter value")
	// ErrInvalidGaugeValue возвращается, когда у gauge нет числового value.
	ErrInvalidGaugeValue = errors.New("invalid gauge value")
	// ErrMetricNotFound возвращается, когда метрика с указанным именем не сохранена.
	ErrMetricNotFound = errors.New("metric not found")
)

// MetricsRepository хранит и читает метрики.
type MetricsRepository interface {
	// Update сохраняет одну метрику.
	Update(ctx context.Context, metric model.Metrics) error
	// Get возвращает метрику по идентификатору.
	Get(ctx context.Context, id string) (model.Metrics, error)
	// GetBulk возвращает все сохранённые метрики.
	GetBulk(ctx context.Context) ([]model.Metrics, error)
	// BulkUpdate сохраняет набор метрик.
	BulkUpdate(ctx context.Context, metrics []model.Metrics) error
}

// MetricsService проверяет метрики и передаёт их в хранилище.
type MetricsService struct {
	repository MetricsRepository
	config     *config.ServerConfig
}

// NewMetricsService создаёт сервис метрик поверх хранилища и конфигурации сервера.
func NewMetricsService(repository MetricsRepository, config *config.ServerConfig) *MetricsService {
	return &MetricsService{repository: repository, config: config}
}

func (s *MetricsService) validate(metric *model.Metrics) error {
	switch metric.MType {
	case model.Counter:
		if metric.Delta == nil {
			return ErrInvalidCounterValue
		}
	case model.Gauge:
		if metric.Value == nil {
			return ErrInvalidGaugeValue
		}
	default:
		return ErrInvalidMetricType
	}
	return nil
}

// Update проверяет метрику и сохраняет её.
// Для counter хранилище прибавляет delta к текущему значению, для gauge заменяет value.
func (s *MetricsService) Update(ctx context.Context, metric *model.Metrics) error {
	if err := s.validate(metric); err != nil {
		return err
	}
	return s.repository.Update(ctx, *metric)
}

// Get возвращает метрику по идентификатору или ErrMetricNotFound.
func (s *MetricsService) Get(ctx context.Context, id string) (model.Metrics, error) {
	got, err := s.repository.Get(ctx, id)
	if err != nil {
		return model.Metrics{}, ErrMetricNotFound
	}
	return got, nil
}

// GetBulk возвращает все сохранённые метрики.
func (s *MetricsService) GetBulk(ctx context.Context) ([]model.Metrics, error) {
	return s.repository.GetBulk(ctx)
}

// BulkUpdate проверяет каждую метрику и сохраняет набор одной операцией хранилища.
func (s *MetricsService) BulkUpdate(ctx context.Context, metrics []model.Metrics) error {
	for i := range metrics {
		if err := s.validate(&metrics[i]); err != nil {
			return err
		}
	}
	return s.repository.BulkUpdate(ctx, metrics)
}
