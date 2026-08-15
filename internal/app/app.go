package app

import (
	"log"
	"net"
	"net/http"
	"time"

	"github.com/Alexunder2003/alex-metrics-service/internal/config"
	"github.com/Alexunder2003/alex-metrics-service/internal/handler"
	"github.com/Alexunder2003/alex-metrics-service/internal/middleware"
	"github.com/Alexunder2003/alex-metrics-service/internal/model"
	"github.com/Alexunder2003/alex-metrics-service/internal/repository"
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
	store *storage.FileStorage[model.Metrics]
}

func NewApp() *App {
	cfg := config.NewServerConfig()
	logger, err := zap.NewProduction()
    if err != nil {
        log.Fatalf("failed to create logger: %v", err)
    }
    sugar := logger.Sugar()
	
	mw := []func(http.Handler) http.Handler{
		middleware.LoggingMiddleware(sugar),
		middleware.CompressingMiddleware,
	}

	store, err := storage.NewFileStorage[model.Metrics](cfg.FileStoragePath, cfg.Restore, cfg.StoreInterval == 0)
	if err != nil {
		log.Fatalf("failed to create file storage: %v", err)
	}
	metricsRepository := repository.NewMetricsRepository(store)
	metricsService := service.NewMetricsService(metricsRepository, cfg)
	metricsHandler := handler.NewMetricsHandler(metricsService, sugar)
	metricsRouter := handler.MetricsRouter(metricsHandler)

	return &App{cfg: *cfg, store: store, metricsService: metricsService, router: handler.NewGlobalRouter(mw, []handler.Mount{
		{Pattern: "/", Router: metricsRouter},
	}), logger: sugar}
}

func (a *App) Run() error {
	addr := a.cfg.Address
	if _, port, err := net.SplitHostPort(addr); err == nil {
		addr = net.JoinHostPort("", port)
	}

	ticker := time.NewTicker(time.Duration(a.cfg.StoreInterval) * time.Second)
	defer ticker.Stop()
	
	if a.cfg.StoreInterval > 0 {
		go func() {
			for range ticker.C {
				if err := a.store.Store(); err != nil {
					a.logger.Errorw("failed to store metrics", "error", err)
				}
			}	
		}()
	}

	a.logger.Infof("starting server on %s (from -a %s)", addr, a.cfg.Address)
	return http.ListenAndServe(addr, a.router)
}
