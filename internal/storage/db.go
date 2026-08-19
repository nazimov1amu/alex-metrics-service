package storage

import (
	"context"
	"database/sql"
	"log"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type DBStorage struct {
	db *sql.DB
}

func NewDBStorage(dbDSN string) *DBStorage {
	db, err := sql.Open("pgx", dbDSN)
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	return &DBStorage{db: db}
}

func (s *DBStorage) Ping() error {
	return s.db.Ping()
}

func (s *DBStorage) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return s.db.QueryContext(ctx, query, args...)
}

func (s *DBStorage) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return s.db.ExecContext(ctx, query, args...)
}

func (s *DBStorage) Close() error {
	return s.db.Close()
}
