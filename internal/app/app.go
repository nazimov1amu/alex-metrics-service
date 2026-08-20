package app

import (
	"database/sql"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/Alexunder2003/alex-metrics-service/internal/config"
	"github.com/Alexunder2003/alex-metrics-service/internal/handler"
	"github.com/Alexunder2003/alex-metrics-service/internal/middleware"
	"github.com/Alexunder2003/alex-metrics-service/internal/repository"
	"github.com/Alexunder2003/alex-metrics-service/internal/service"
	"github.com/go-chi/chi/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"
)

type App struct {
	cfg            config.ServerConfig
	metricsService *service.MetricsService
	router         chi.Router
	logger         *zap.SugaredLogger
	metricsRepo    repository.MetricsRepository
	db             *sql.DB
}

func NewApp() *App {
	cfg := config.NewServerConfig()

	db, err := sql.Open("pgx", cfg.DatabaseDSN)
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}

	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("failed to create logger: %v", err)
	}
	sugar := logger.Sugar()

	mw := []func(http.Handler) http.Handler{
		middleware.LoggingMiddleware(sugar),
		middleware.CompressingMiddleware,
	}

	metricsRepo, err := repository.NewMetricsRepository(cfg)
	if err != nil {
		log.Fatalf("failed to create metrics repository: %v", err)
	}
	
	metricsService := service.NewMetricsService(metricsRepo, cfg)
	metricsHandler := handler.NewMetricsHandler(metricsService, sugar)
	metricsRouter := handler.MetricsRouter(metricsHandler)

	healthHandler := handler.NewHealthHandler(db, sugar)
	healthRouter := handler.HealthRouter(healthHandler)

	return &App{
		cfg:            *cfg,
		metricsRepo:    metricsRepo,
		db:             db,
		metricsService: metricsService,
		router: handler.NewGlobalRouter(mw, []handler.Mount{
			{Pattern: "/", Router: metricsRouter},
			{Pattern: "/ping", Router: healthRouter},
		}),
		logger: sugar,
	}
}

func (a *App) Run() error {
	addr := a.cfg.Address
	if _, port, err := net.SplitHostPort(addr); err == nil {
		addr = net.JoinHostPort("", port)
	}

	if fileRepo, ok := a.metricsRepo.(*repository.FileMetricsRepository); ok && a.cfg.StoreInterval > 0 {
		ticker := time.NewTicker(time.Duration(a.cfg.StoreInterval) * time.Second)
		go func() {
			defer ticker.Stop()
			for range ticker.C {
				if err := fileRepo.Store(); err != nil {
					a.logger.Errorw("failed to store metrics", "error", err)
				}
			}
		}()
	}

	a.logger.Infof("starting server on %s (from -a %s)", addr, a.cfg.Address)
	return http.ListenAndServe(addr, a.router)
}
