package handler

import (
	"github.com/go-chi/chi/v5"
)

type Mount struct {
	Pattern string
	Router  chi.Router
}

func MetricsRouter(h *Handler) chi.Router {
	r := chi.NewRouter()

	r.Post("/update/{type}/{name}/{value}", h.Update)
	r.Get("/value/{type}/{name}", h.Get)
	r.Get("/", h.GetBulk)

	return r
}

func NewRouter(mounts ...Mount) chi.Router {
	r := chi.NewRouter()
	for _, m := range mounts {
		r.Mount(m.Pattern, m.Router)
	}
	return r
}
