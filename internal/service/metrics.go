package service

import (
	"errors"
	"strconv"

	"github.com/Alexunder2003/alex-metrics-service/internal/model"
	"github.com/Alexunder2003/alex-metrics-service/internal/repository"
	"github.com/Alexunder2003/alex-metrics-service/internal/storage"
	"go.uber.org/zap"
)

var (
	ErrInvalidMetricType = errors.New("invalid metric type")
	ErrInvalidCounterValue = errors.New("invalid counter value")
	ErrInvalidGaugeValue = errors.New("invalid gauge value")
	ErrMetricNotFound = errors.New("metric not found")
)

type MetricsRepository interface {
	Update(metric model.Metrics) error
	Get(name string) (model.Metrics, error)
	GetBulk() ([]model.Metrics, error)
}

type MetricsService struct {
	repository MetricsRepository
	logger *zap.SugaredLogger
}

func NewMetricsService(storage *storage.MemStorage[model.Metrics], logger *zap.SugaredLogger) *MetricsService {
	repository := repository.NewMetricsRepository(storage)
	return &MetricsService{repository: repository, logger: logger}
}

func (s *MetricsService) Update(input model.MetricsInput) (model.Metrics, error) {
	metric := model.Metrics{
		ID:    input.Name,
		MType: input.MType,
	}

	switch input.MType {
	case model.Counter:
		delta, err := strconv.ParseInt(input.RawValue, 10, 64)
		if err != nil {
			s.logger.Error("failed to parse counter value", zap.Error(err))
			return model.Metrics{}, ErrInvalidCounterValue
		}

		key := input.Name
		if current, err := s.repository.Get(key); err == nil {
			delta += *current.Delta
		}
		metric.Delta = &delta
	case model.Gauge:
		value, err := strconv.ParseFloat(input.RawValue, 64)
		if err != nil {
			s.logger.Error("failed to parse gauge value", zap.Error(err))
			return model.Metrics{}, ErrInvalidGaugeValue
		}
		metric.Value = &value
	default:
		s.logger.Error("failed to update metric", zap.String("name", input.Name), zap.Error(ErrInvalidMetricType))
		return model.Metrics{}, ErrInvalidMetricType
	}

	if err := s.repository.Update(metric); err != nil {
		s.logger.Error("failed to update metric", zap.String("name", input.Name), zap.Error(err))
		return model.Metrics{}, err
	}
	return metric, nil
}


func (s *MetricsService) GetBulk() ([]model.Metrics, error) {
	metrics, err := s.repository.GetBulk()
	if err != nil {
		s.logger.Error("failed to get bulk metrics", zap.Error(err))
		return nil, err
	}
	return metrics, nil
}

func (s *MetricsService) Get(name string) (model.Metrics, error) {
	metric, err := s.repository.Get(name)
	if err != nil {
		s.logger.Error("failed to get metric", zap.String("name", name), zap.Error(err))
		return model.Metrics{}, ErrMetricNotFound
	}
	return metric, nil
}