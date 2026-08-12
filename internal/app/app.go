package app

import (
	"log"
	"net"
	"net/http"
	"time"

	"github.com/Alexunder2003/alex-metrics-service/internal/config"
	"github.com/Alexunder2003/alex-metrics-service/internal/encoding"
	"github.com/Alexunder2003/alex-metrics-service/internal/handler"
	"github.com/Alexunder2003/alex-metrics-service/internal/logger"
	"github.com/Alexunder2003/alex-metrics-service/internal/model"
	"github.com/Alexunder2003/alex-metrics-service/internal/service"
	"github.com/Alexunder2003/alex-metrics-service/internal/storage"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

type App struct {
	cfg    config.ServerConfig
	metricsService *service.MetricsService
	router chi.Router
	logger *zap.SugaredLogger
}

func NewApp() *App {
	cfg := config.NewServerConfig()
	sugar, err := logger.NewLogger()
	if err != nil {
		log.Fatalf("failed to create logger: %v", err)
	}
	
	mw := []func(http.Handler) http.Handler{
		logger.LoggingMiddleware(sugar),
		encoding.CompressingMiddleware,
	}

	store := storage.NewMemStorage[model.Metrics]()
	metricsService := service.NewMetricsService(store, sugar, cfg)
	metricsHandler := handler.NewMetricsHandler(metricsService)
	metricsRouter := handler.MetricsRouter(metricsHandler)

	return &App{cfg: *cfg, metricsService: metricsService, router: handler.NewGlobalRouter(mw, []handler.Mount{
		{Pattern: "/", Router: metricsRouter},
	}), logger: sugar}
}

func (a *App) Run() error {
	addr := a.cfg.Address
	if _, port, err := net.SplitHostPort(addr); err == nil {
		addr = net.JoinHostPort("", port)
	}

	if a.cfg.Restore {
		if err := a.metricsService.Restore(); err != nil {
			a.logger.Errorw("failed to restore metrics", "error", err)
		}
	}

	go func() {
		for {
			time.Sleep(time.Duration(a.cfg.StoreInterval) * time.Second)
			if err := a.metricsService.Store(); err != nil {
				a.logger.Errorw("failed to store metrics", "error", err)
			}
		}
	}()

	a.logger.Infof("starting server on %s (from -a %s)", addr, a.cfg.Address)
	return http.ListenAndServe(addr, a.router)
}
