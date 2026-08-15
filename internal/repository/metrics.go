package repository

import (
	"github.com/Alexunder2003/alex-metrics-service/internal/model"
)

type Storage interface {
	Get(key string) (model.Metrics, error)
	Update(key string, value model.Metrics) error
	GetBulk() ([]model.Metrics, error)
}

type MetricsRepository struct {
	storage Storage
}

func NewMetricsRepository(storage Storage) *MetricsRepository {
	return &MetricsRepository{storage: storage}
}

func (r *MetricsRepository) Get(key string) (model.Metrics, error) {
	return r.storage.Get(key)
}

func (r *MetricsRepository) Update(metric model.Metrics) error {
	return r.storage.Update(metric.ID, metric)
}

func (r *MetricsRepository) GetBulk() ([]model.Metrics, error) {
	return r.storage.GetBulk()
}
