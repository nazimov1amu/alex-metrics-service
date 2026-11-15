// Package repository хранит метрики и события аудита в памяти, файле или PostgreSQL.
package repository

import (
	"context"
	"database/sql"

	"github.com/Alexunder2003/alex-metrics-service/internal/config"
	"github.com/Alexunder2003/alex-metrics-service/internal/model"
	_ "github.com/jackc/pgx/v5/stdlib"
)

// MetricsRepository читает и сохраняет метрики независимо от выбранного хранилища.
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

// NewMetricsRepository выбирает хранилище по конфигурации сервера.
// При непустом DatabaseDSN создаётся PostgreSQL, при FileStoragePath — файл, иначе память.
func NewMetricsRepository(cfg *config.ServerConfig) (MetricsRepository, error) {
	if cfg.DatabaseDSN != "" {
		db, err := sql.Open("pgx", cfg.DatabaseDSN)
		if err != nil {
			return nil, err
		}
		return NewPostgresMetricsRepository(db), nil
	}

	if cfg.FileStoragePath != "" {
		return NewFileMetricsRepository(cfg.FileStoragePath, cfg.Restore, cfg.StoreInterval == 0)
	}

	return NewMemMetricsRepository(), nil
}
