package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

type Mount struct {
	Pattern string
	Router  chi.Router
}

func MetricsRouter(h *Handler) chi.Router {
	r := chi.NewRouter()

	// JSON API (iter7)
	r.Post("/update/", h.UpdateJSON)
	r.Post("/update", h.UpdateJSON)
	r.Post("/value/", h.ValueJSON)
	r.Post("/value", h.ValueJSON)

	// Plain URL API (iter1–5 compatibility)
	r.Post("/update/{type}/{name}/{value}", h.UpdatePath)
	r.Get("/value/{type}/{name}", h.ValuePath)
	r.Get("/", h.GetBulk)

	return r
}

func NewGlobalRouter(middleware []func(http.Handler) http.Handler, mounts []Mount) chi.Router {
	r := chi.NewRouter()
	r.Use(middleware...)
	for _, m := range mounts {
		r.Mount(m.Pattern, m.Router)
	}
	return r
}
