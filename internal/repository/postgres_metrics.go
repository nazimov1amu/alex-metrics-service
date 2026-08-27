package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/Alexunder2003/alex-metrics-service/internal/db/queries"
	"github.com/Alexunder2003/alex-metrics-service/internal/model"
	"github.com/Alexunder2003/alex-metrics-service/internal/utils"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
)

var (
	queryUpsertMetric = queries.UpsertMetric
	queryGetMetric    = queries.GetMetric
	queryGetMetrics   = queries.GetMetrics
)

const retryInterval = 2 * time.Second
const maxRetries = 3

type PGErrorClassification int

const (
    NonRetriable PGErrorClassification = iota

    Retriable
)

type PostgresErrorClassifier struct{}

func NewPostgresErrorClassifier() *PostgresErrorClassifier {
    return &PostgresErrorClassifier{}
}

func (c *PostgresErrorClassifier) Classify(err error) PGErrorClassification {
    if err == nil {
        return NonRetriable
    }

    var pgErr *pgconn.PgError
    if errors.As(err, &pgErr) {
        return СlassifyPgError(pgErr)
    }

    return NonRetriable
}

func СlassifyPgError(pgErr *pgconn.PgError) PGErrorClassification {

    switch pgErr.Code {
    case pgerrcode.ConnectionException,
        pgerrcode.ConnectionDoesNotExist,
        pgerrcode.ConnectionFailure:
        return Retriable

    case pgerrcode.TransactionRollback, 
        pgerrcode.SerializationFailure, 
        pgerrcode.DeadlockDetected:     
        return Retriable

    case pgerrcode.CannotConnectNow: 
        return Retriable
    }

    switch pgErr.Code {
    case pgerrcode.DataException,
        pgerrcode.NullValueNotAllowedDataException:
        return NonRetriable

    case pgerrcode.IntegrityConstraintViolation,
        pgerrcode.RestrictViolation,
        pgerrcode.NotNullViolation,
        pgerrcode.ForeignKeyViolation,
        pgerrcode.UniqueViolation,
        pgerrcode.CheckViolation:
        return NonRetriable

    case pgerrcode.SyntaxErrorOrAccessRuleViolation,
        pgerrcode.SyntaxError,
        pgerrcode.UndefinedColumn,
        pgerrcode.UndefinedTable,
        pgerrcode.UndefinedFunction:
        return NonRetriable
    }

    return NonRetriable
}

type PostgresMetricsRepository struct {
	db *sql.DB
	classifier *PostgresErrorClassifier
}

func NewPostgresMetricsRepository(db *sql.DB) *PostgresMetricsRepository {
	return &PostgresMetricsRepository{db: db, classifier: NewPostgresErrorClassifier()}
}

func (r *PostgresMetricsRepository) withRetry(fn func() error) error {
	err := fn()
	if err != nil {
		if r.classifier.Classify(err) == Retriable {
			return utils.LinearRetry(maxRetries, retryInterval, fn)
		}
		return err
	}
	return nil
}

func (r *PostgresMetricsRepository) Update(ctx context.Context, metric model.Metrics) error {
	return r.withRetry(func() error {
		_, err := r.db.ExecContext(ctx, queryUpsertMetric, metric.ID, metric.MType, metric.Delta, metric.Value)
		return err
	})
}

func (r *PostgresMetricsRepository) Get(ctx context.Context, id string) (model.Metrics, error) {
	var metric model.Metrics
	err := r.withRetry(func() error {
		return r.db.QueryRowContext(ctx, queryGetMetric, id).
			Scan(&metric.ID, &metric.MType, &metric.Delta, &metric.Value)
	})
	return metric, err
}

func (r *PostgresMetricsRepository) GetBulk(ctx context.Context) ([]model.Metrics, error) {
	var metrics []model.Metrics
	err := r.withRetry(func() error {
		rows, err := r.db.QueryContext(ctx, queryGetMetrics)
		if err != nil {
			return err
		}
		defer rows.Close()

		metrics = nil
		for rows.Next() {
			var metric model.Metrics
			if err := rows.Scan(&metric.ID, &metric.MType, &metric.Delta, &metric.Value); err != nil {
				return err
			}
			metrics = append(metrics, metric)
		}
		return rows.Err()
	})
	return metrics, err
}

func (r *PostgresMetricsRepository) BulkUpdate(ctx context.Context, metrics []model.Metrics) error {
	return r.withRetry(func() error {
		tx, err := r.db.BeginTx(ctx, nil)
		if err != nil {
			return err
		}
		defer tx.Rollback()

		for _, metric := range metrics {
			if _, err := tx.ExecContext(ctx, queryUpsertMetric, metric.ID, metric.MType, metric.Delta, metric.Value); err != nil {
				return err
			}
		}
		return tx.Commit()
	})
}