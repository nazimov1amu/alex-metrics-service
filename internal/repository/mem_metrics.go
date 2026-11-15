package repository

import (
	"context"
	"errors"
	"sync"

	"github.com/Alexunder2003/alex-metrics-service/internal/model"
)

// MemMetricsRepository хранит метрики в памяти процесса.
type MemMetricsRepository struct {
	storage map[string]model.Metrics
	mutex   sync.Mutex
}

// NewMemMetricsRepository создаёт пустое хранилище в памяти.
func NewMemMetricsRepository() *MemMetricsRepository {
	return &MemMetricsRepository{storage: make(map[string]model.Metrics)}
}

// Get возвращает метрику по идентификатору или ошибку, если её нет.
func (r *MemMetricsRepository) Get(_ context.Context, id string) (model.Metrics, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	value, ok := r.storage[id]
	if !ok {
		return model.Metrics{}, errors.New("value not found")
	}
	return value, nil
}

func (r *MemMetricsRepository) upsertLocked(metric model.Metrics) {
	if metric.MType == model.Counter && metric.Delta != nil {
		if existing, ok := r.storage[metric.ID]; ok && existing.Delta != nil {
			*metric.Delta += *existing.Delta
		}
	}
	r.storage[metric.ID] = metric
}

// Update записывает gauge поверх прежнего значения, а counter прибавляет к текущему delta.
func (r *MemMetricsRepository) Update(_ context.Context, metric model.Metrics) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	r.upsertLocked(metric)
	return nil
}

// GetBulk возвращает копию всех метрик из памяти.
func (r *MemMetricsRepository) GetBulk(_ context.Context) ([]model.Metrics, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	values := make([]model.Metrics, 0, len(r.storage))
	for _, value := range r.storage {
		values = append(values, value)
	}
	return values, nil
}

// BulkUpdate сохраняет набор метрик теми же правилами, что и Update.
func (r *MemMetricsRepository) BulkUpdate(_ context.Context, metrics []model.Metrics) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	for _, metric := range metrics {
		r.upsertLocked(metric)
	}
	return nil
}
