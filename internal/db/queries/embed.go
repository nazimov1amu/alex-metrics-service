package queries

import _ "embed"

//go:embed metrics/upsert_metric.sql
var UpsertMetric string

//go:embed metrics/get_metric.sql
var GetMetric string

//go:embed metrics/get_metrics.sql
var GetMetrics string
