package service

import (
	"errors"

	"github.com/Alexunder2003/alex-metrics-service/internal/model"
	"github.com/Alexunder2003/alex-metrics-service/internal/repository"
	"github.com/Alexunder2003/alex-metrics-service/internal/storage"
	"go.uber.org/zap"
)

var (
	ErrInvalidMetricType   = errors.New("invalid metric type")
	ErrInvalidCounterValue = errors.New("invalid counter value")
	ErrInvalidGaugeValue   = errors.New("invalid gauge value")
	ErrMetricNotFound      = errors.New("metric not found")
)

type MetricsRepository interface {
	Update(metric model.Metrics) error
	Get(id string) (model.Metrics, error)
	GetBulk() ([]model.Metrics, error)
}

type MetricsService struct {
	repository MetricsRepository
	logger     *zap.SugaredLogger
}

func NewMetricsService(storage *storage.MemStorage[model.Metrics], logger *zap.SugaredLogger) *MetricsService {
	repository := repository.NewMetricsRepository(storage)
	return &MetricsService{repository: repository, logger: logger}
}

func (s *MetricsService) Update(metric *model.Metrics) error {
	switch metric.MType {
	case model.Counter:
		if metric.Delta == nil {
			return ErrInvalidCounterValue
		}
		existing, err := s.repository.Get(metric.ID)
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

	if err := s.repository.Update(*metric); err != nil {
		s.logger.Errorw("failed to update metric", "name", metric.ID, "error", err)
		return err
	}
	return nil
}

func (s *MetricsService) Get(id string) (model.Metrics, error) {
	got, err := s.repository.Get(id)
	if err != nil {
		s.logger.Errorw("failed to get metric", "name", id, "error", err)
		return model.Metrics{}, ErrMetricNotFound
	}
	return got, nil
}

func (s *MetricsService) GetBulk() ([]model.Metrics, error) {
	return s.repository.GetBulk()
}
