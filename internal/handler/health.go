package handler

import (
	"database/sql"
	"net/http"

	"go.uber.org/zap"
)

// HealthHandler отвечает на проверку доступности хранилища.
type HealthHandler struct {
	logger *zap.SugaredLogger
	db     *sql.DB
}

// NewHealthHandler создаёт обработчик проверки базы.
// Нулевой db означает, что хранилище не настроено, и проверка отвечает ошибкой.
func NewHealthHandler(db *sql.DB, logger *zap.SugaredLogger) *HealthHandler {
	return &HealthHandler{db: db, logger: logger}
}

// DBHealthCheck обрабатывает GET /ping: 200, если Ping базы успешен, иначе 500.
func (h *HealthHandler) DBHealthCheck(w http.ResponseWriter, r *http.Request) {
	if h.db == nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if err := h.db.PingContext(r.Context()); err != nil {
		h.logger.Error("failed to ping database", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
