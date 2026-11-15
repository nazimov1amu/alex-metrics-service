// Package queries хранит SQL-запросы метрик, встроенные в бинарник.
package queries

import _ "embed"

// UpsertMetric — SQL-запрос вставки или обновления одной метрики.
//
//go:embed metrics/upsert_metric.sql
var UpsertMetric string

// GetMetric — SQL-запрос чтения одной метрики по идентификатору.
//
//go:embed metrics/get_metric.sql
var GetMetric string

// GetMetrics — SQL-запрос чтения всех метрик.
//
//go:embed metrics/get_metrics.sql
var GetMetrics string
