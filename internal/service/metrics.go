package service

import (
	"context"
	"errors"

	"github.com/Alexunder2003/alex-metrics-service/internal/config"
	"github.com/Alexunder2003/alex-metrics-service/internal/model"
)

var (
	ErrInvalidMetricType   = errors.New("invalid metric type")
	ErrInvalidCounterValue = errors.New("invalid counter value")
	ErrInvalidGaugeValue   = errors.New("invalid gauge value")
	ErrMetricNotFound      = errors.New("metric not found")
)

type MetricsRepository interface {
	Update(ctx context.Context, metric model.Metrics) error
	Get(ctx context.Context, id string) (model.Metrics, error)
	GetBulk(ctx context.Context) ([]model.Metrics, error)
}

type MetricsService struct {
	repository MetricsRepository
	config     *config.ServerConfig
}

func NewMetricsService(repository MetricsRepository, config *config.ServerConfig) *MetricsService {
	return &MetricsService{repository: repository, config: config}
}

func (s *MetricsService) Update(ctx context.Context, metric *model.Metrics) error {
	switch metric.MType {
	case model.Counter:
		if metric.Delta == nil {
			return ErrInvalidCounterValue
		}
		existing, err := s.repository.Get(ctx, metric.ID)
		if err == nil && existing.Delta != nil {
			*metric.Delta += *existing.Delta
		}
	case model.Gauge:
		if metric.Value == nil {
			return ErrInvalidGaugeValue
		}
	default:
		return ErrInvalidMetricType
	}

	return s.repository.Update(ctx, *metric)
}

func (s *MetricsService) Get(ctx context.Context, id string) (model.Metrics, error) {
	got, err := s.repository.Get(ctx, id)
	if err != nil {
		return model.Metrics{}, ErrMetricNotFound
	}
	return got, nil
}

func (s *MetricsService) GetBulk(ctx context.Context) ([]model.Metrics, error) {
	return s.repository.GetBulk(ctx)
}
