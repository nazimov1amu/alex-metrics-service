package service

import (
	"encoding/json"
	"errors"
	"os"

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
	Update(metric model.Metrics) error
	Get(id string) (model.Metrics, error)
	GetBulk() ([]model.Metrics, error)
}

type MetricsService struct {
	repository MetricsRepository
	config     *config.ServerConfig
}


func NewMetricsService(repository MetricsRepository, config *config.ServerConfig) *MetricsService {
	return &MetricsService{repository: repository, config: config}
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
		return err
	}

	if s.config.StoreInterval <= 0 {
		s.Store()
	}

	return nil
}

func (s *MetricsService) Get(id string) (model.Metrics, error) {
	got, err := s.repository.Get(id)
	if err != nil {
		return model.Metrics{}, ErrMetricNotFound
	}
	return got, nil
}

func (s *MetricsService) Store() error {
	file, err := os.OpenFile(s.config.FileStoragePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	metrics, err := s.repository.GetBulk()
	if err != nil {
		return err
	}
	return encoder.Encode(metrics)
}

func (s *MetricsService) Restore() error {
	file, err := os.OpenFile(s.config.FileStoragePath, os.O_RDONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	var metrics []model.Metrics
	if err := decoder.Decode(&metrics); err != nil {
		return err
	}
	for _, metric := range metrics {
		if err := s.repository.Update(metric); err != nil {
			return err
		}
	}
	return nil
}

func (s *MetricsService) GetBulk() ([]model.Metrics, error) {
	return s.repository.GetBulk()
}
