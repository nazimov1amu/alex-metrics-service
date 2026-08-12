package service

import (
	"encoding/json"
	"errors"
	"os"

	"github.com/Alexunder2003/alex-metrics-service/internal/config"
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
	config     *config.ServerConfig
}

func NewMetricsService(storage *storage.MemStorage[model.Metrics], logger *zap.SugaredLogger, config *config.ServerConfig) *MetricsService {
	repository := repository.NewMetricsRepository(storage)
	return &MetricsService{repository: repository, logger: logger, config: config}
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

func (s *MetricsService) Store() error {
	file, err := os.OpenFile(s.config.FileStoragePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		s.logger.Errorw("failed to open file", "path", s.config.FileStoragePath, "error", err)
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	metrics, err := s.repository.GetBulk()
	if err != nil {
		s.logger.Errorw("failed to get metrics", "error", err)
		return err
	}
	return encoder.Encode(metrics)
}

func (s *MetricsService) Restore() error {
	file, err := os.OpenFile(s.config.FileStoragePath, os.O_RDONLY, 0644)
	if err != nil {
		s.logger.Errorw("failed to open file", "path", s.config.FileStoragePath, "error", err)
		return err
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	var metrics []model.Metrics
	if err := decoder.Decode(&metrics); err != nil {
		s.logger.Errorw("failed to decode metrics", "error", err)
		return err
	}
	for _, metric := range metrics {
		if err := s.repository.Update(metric); err != nil {
			s.logger.Errorw("failed to update metric", "metric", metric, "error", err)
			return err
		}
	}
	return nil
}

func (s *MetricsService) GetBulk() ([]model.Metrics, error) {
	return s.repository.GetBulk()
}
