INSERT INTO metrics (id, type, delta, value)
VALUES ($1, $2, $3, $4)
ON CONFLICT (id) DO UPDATE SET
    type  = EXCLUDED.type,
    delta = CASE
        WHEN EXCLUDED.type = 'counter'
            THEN COALESCE(metrics.delta, 0) + COALESCE(EXCLUDED.delta, 0)
        ELSE EXCLUDED.delta
    END,
    value = EXCLUDED.value;