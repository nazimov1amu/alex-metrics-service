package service

import (
	"errors"
	"strconv"

	"github.com/Alexunder2003/alex-metrics-service/internal/model"
	"github.com/Alexunder2003/alex-metrics-service/internal/repository"
	"github.com/Alexunder2003/alex-metrics-service/internal/storage"
)

var (
	ErrInvalidMetricType = errors.New("invalid metric type")
	ErrInvalidCounterValue = errors.New("invalid counter value")
	ErrInvalidGaugeValue = errors.New("invalid gauge value")
	ErrMetricNotFound = errors.New("metric not found")
)

type MetricsService struct {
	repository *repository.MetricsRepository
}

func NewMetricsService(storage *storage.MemStorage[model.Metrics]) *MetricsService {
	repository := repository.NewMetricsRepository(storage)
	return &MetricsService{repository: repository}
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
			return model.Metrics{}, ErrInvalidGaugeValue
		}
		metric.Value = &value
	default:
		return model.Metrics{}, ErrInvalidMetricType
	}

	if err := s.repository.Update(metric); err != nil {
		return model.Metrics{}, err
	}
	return metric, nil
}


func (s *MetricsService) GetBulk() ([]model.Metrics, error) {
	metrics, err := s.repository.GetBulk()
	if err != nil {
		return nil, err
	}
	return metrics, nil
}

func (s *MetricsService) Get(name string) (model.Metrics, error) {
	metric, err := s.repository.Get(name)
	if err != nil {
		return model.Metrics{}, ErrMetricNotFound
	}
	return metric, nil
}