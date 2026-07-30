package repository

import (
	"github.com/Alexunder2003/alex-metrics-service/internal/model"
	"github.com/Alexunder2003/alex-metrics-service/internal/storage"
)


type MetricsRepository struct {
	storage *storage.MemStorage[model.Metrics]
}

func NewMetricsRepository(storage *storage.MemStorage[model.Metrics]) *MetricsRepository {
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