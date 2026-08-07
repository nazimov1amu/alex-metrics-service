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
	"go.uber.org/zap"
)

type App struct {
	cfg    config.Config
	router chi.Router
	logger *zap.SugaredLogger
}

func NewApp() *App {
	cfg := config.NewConfig()
	sugar, err := NewLogger()
	if err != nil {
		log.Fatalf("failed to create logger: %v", err)
	}

	mw := []func(http.Handler) http.Handler{
		LoggingMiddleware(sugar),
	}

	store := storage.NewMemStorage[model.Metrics]()
	metricsService := service.NewMetricsService(store, sugar)
	metricsHandler := handler.NewMetricsHandler(metricsService)
	metricsRouter := handler.MetricsRouter(metricsHandler)

	return &App{cfg: *cfg, router: handler.NewGlobalRouter(mw, []handler.Mount{
		{Pattern: "/", Router: metricsRouter},
	}), logger: sugar}
}

func (a *App) Run() error {
	addr := a.cfg.Address
	if _, port, err := net.SplitHostPort(addr); err == nil {
		addr = net.JoinHostPort("", port)
	}
	a.logger.Infof("starting server on %s (from -a %s)", addr, a.cfg.Address)
	return http.ListenAndServe(addr, a.router)
}
