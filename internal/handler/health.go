package handler

import (
	"database/sql"
	"net/http"

	"go.uber.org/zap"
)

type HealthHandler struct {
	logger *zap.SugaredLogger
	db *sql.DB
}

func NewHealthHandler(db *sql.DB, logger *zap.SugaredLogger) *HealthHandler {
	return &HealthHandler{db: db, logger: logger}
}

func (h *HealthHandler) DBHealthCheck(w http.ResponseWriter, r *http.Request) {
	if err := h.db.PingContext(r.Context()); err != nil {
		h.logger.Error("failed to ping database", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}