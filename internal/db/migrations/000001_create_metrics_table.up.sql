CREATE TABLE IF NOT EXISTS metrics (
    id VARCHAR(255) PRIMARY KEY,
    type VARCHAR(255) NOT NULL,
    delta BIGINT,
    value DOUBLE PRECISION
);

CREATE INDEX IF NOT EXISTS idx_metrics_type ON metrics (type);
