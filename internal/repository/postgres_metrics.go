package repository

import (
	"context"
	"database/sql"

	"github.com/Alexunder2003/alex-metrics-service/internal/db/queries"
	"github.com/Alexunder2003/alex-metrics-service/internal/model"
)

var (
	queryUpsertMetric = queries.UpsertMetric
	queryGetMetric    = queries.GetMetric
	queryGetMetrics   = queries.GetMetrics
)

type PostgresMetricsRepository struct {
	db *sql.DB
}

func NewPostgresMetricsRepository(db *sql.DB) *PostgresMetricsRepository {
	return &PostgresMetricsRepository{db: db}
}

func (r *PostgresMetricsRepository) Update(ctx context.Context, metric model.Metrics) error {
	_, err := r.db.ExecContext(ctx, queryUpsertMetric, metric.ID, metric.MType, metric.Delta, metric.Value)
	return err
}

func (r *PostgresMetricsRepository) Get(ctx context.Context, id string) (model.Metrics, error) {
	row := r.db.QueryRowContext(ctx, queryGetMetric, id)

	var metric model.Metrics

	err := row.Scan(&metric.ID, &metric.MType, &metric.Delta, &metric.Value)
	return metric, err
}

func (r *PostgresMetricsRepository) GetBulk(ctx context.Context) ([]model.Metrics, error) {
	rows, err := r.db.QueryContext(ctx, queryGetMetrics)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var metrics []model.Metrics

	for rows.Next() {
		var metric model.Metrics
		err := rows.Scan(&metric.ID, &metric.MType, &metric.Delta, &metric.Value)
		if err != nil {
			return nil, err
		}
		metrics = append(metrics, metric)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return metrics, nil
}

func (r *PostgresMetricsRepository) BulkUpdate(ctx context.Context, metrics []model.Metrics) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	for _, metric := range metrics {
		_, err := tx.ExecContext(ctx, queryUpsertMetric, metric.ID, metric.MType, metric.Delta, metric.Value)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}