package repository

import (
	"context"
	"database/sql"

	"github.com/Alexunder2003/alex-metrics-service/internal/config"
	"github.com/Alexunder2003/alex-metrics-service/internal/model"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type MetricsRepository interface {
	Update(ctx context.Context, metric model.Metrics) error
	Get(ctx context.Context, id string) (model.Metrics, error)
	GetBulk(ctx context.Context) ([]model.Metrics, error)
	BulkUpdate(ctx context.Context, metrics []model.Metrics) error
}

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
