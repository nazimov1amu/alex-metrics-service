package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
)

// Mount описывает, по какому префиксу подключается вложенный роутер.
type Mount struct {
	Pattern string
	Router  chi.Router
}

// MetricsRouter регистрирует эндпоинты API метрик.
// Необязательный audit-middleware применяется только к запросам, которые меняют метрики.
func MetricsRouter(h *Handler, audit ...func(http.Handler) http.Handler) chi.Router {
	r := chi.NewRouter()

	var updates chi.Router = r
	if len(audit) > 0 && audit[0] != nil {
		updates = r.With(audit[0])
	}

	updates.Post("/update/", h.UpdateJSON)
	r.Post("/value/", h.ValueJSON)
	updates.Post("/updates/", h.BulkUpdateJSON)

	// Plain URL API (iter1–5 compatibility)
	updates.Post("/update/{type}/{name}/{value}", h.UpdatePath)
	r.Get("/value/{type}/{name}", h.ValuePath)
	r.Get("/", h.GetBulk)

	return r
}

// HealthRouter регистрирует проверку доступности базы на GET /.
func HealthRouter(h *HealthHandler) chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.DBHealthCheck)
	return r
}

// NewGlobalRouter собирает корневой роутер: общие middleware, pprof на /debug и переданные mounts.
func NewGlobalRouter(middleware []func(http.Handler) http.Handler, mounts []Mount) chi.Router {
	r := chi.NewRouter()
	r.Use(middleware...)
	r.Mount("/debug", chimw.Profiler())
	for _, m := range mounts {
		r.Mount(m.Pattern, m.Router)
	}
	return r
}
