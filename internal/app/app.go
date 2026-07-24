package app

import (
	"net/http"

	"github.com/Alexunder2003/alex-metrics-service/internal/config"
	"github.com/Alexunder2003/alex-metrics-service/internal/handler"
	"github.com/Alexunder2003/alex-metrics-service/internal/model"
	"github.com/Alexunder2003/alex-metrics-service/internal/service"
	"github.com/Alexunder2003/alex-metrics-service/internal/storage"
	"github.com/go-chi/chi/v5"
)

type App struct {
	cfg    config.Config
	router chi.Router
}

func NewApp() *App {
	cfg := config.NewConfig()

	store := storage.NewMemStorage[model.Metrics]()
	metricsHandler := handler.New(service.NewMetricsService(store))

	r := handler.NewRouter(
		handler.Mount{Pattern: "/", Router: handler.MetricsRouter(metricsHandler)},
	)

	return &App{cfg: *cfg, router: r}
}

func (a *App) Run() error {
	return http.ListenAndServe(a.cfg.Address, a.router)
}
