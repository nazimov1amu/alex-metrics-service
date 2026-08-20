INSERT INTO metrics (id, type, delta, value) 
VALUES ($1, $2, $3, $4) 
ON CONFLICT (id) DO UPDATE SET delta = $3, value = $4;