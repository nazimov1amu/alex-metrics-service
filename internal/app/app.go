package app

import (
	"log"
	"net"
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
	addr := a.cfg.Address
	if _, port, err := net.SplitHostPort(addr); err == nil {
		addr = net.JoinHostPort("", port)
	}
	log.Printf("starting server on %s (from -a %s)", addr, a.cfg.Address)
	return http.ListenAndServe(addr, a.router)
}
