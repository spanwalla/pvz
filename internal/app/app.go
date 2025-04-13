package app

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	trmpgx "github.com/avito-tech/go-transaction-manager/drivers/pgxv5/v2"
	"github.com/avito-tech/go-transaction-manager/trm/v2/manager"
	"github.com/jonboulle/clockwork"
	"github.com/labstack/echo/v4"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"

	"github.com/spanwalla/pvz/config"
	grpccontroller "github.com/spanwalla/pvz/internal/controller/grpc"
	httpcontroller "github.com/spanwalla/pvz/internal/controller/http"
	"github.com/spanwalla/pvz/internal/metrics"
	"github.com/spanwalla/pvz/internal/repository"
	"github.com/spanwalla/pvz/internal/service"
	"github.com/spanwalla/pvz/pkg/grpcserver"
	"github.com/spanwalla/pvz/pkg/hasher"
	"github.com/spanwalla/pvz/pkg/httpserver"
	"github.com/spanwalla/pvz/pkg/postgres"
	"github.com/spanwalla/pvz/pkg/validator"
)

// Run creates objects via constructors
func Run() {
	// Config
	configPath, ok := os.LookupEnv("CONFIG_PATH")
	if !ok || len(configPath) == 0 {
		panic("app - os.LookupEnv: CONFIG_PATH is empty")
	}

	cfg, err := config.New(configPath)
	if err != nil {
		panic(fmt.Errorf("app - config.New: %w", err))
	}

	// Logger
	initLogger(cfg.Log.Level)
	log.Info("Config read")

	// Postgres
	log.Info("Connecting to postgres...")
	pg, err := postgres.New(cfg.PG.URL, postgres.MaxPoolSize(cfg.PG.PoolMax))
	if err != nil {
		panic(fmt.Errorf("app - Run - postgres.New: %w", err))
	}
	defer pg.Close()

	// Services and dependencies
	log.Info("Initializing services and dependencies...")
	services := service.New(service.Dependencies{
		Repos:          repository.New(pg),
		Counters:       metrics.New(),
		Transaction:    manager.Must(trmpgx.NewDefaultFactory(pg.Pool)),
		PasswordHasher: hasher.NewBcrypt(),
		Clock:          clockwork.NewRealClock(),
		SecretKey:      cfg.Auth.JWTSecretKey,
		TokenTTL:       cfg.Auth.TokenTTL,
	})

	// Echo handler
	log.Info("Initializing handlers and routes...")
	handler := echo.New()
	handler.Validator = validator.NewCustomValidator()
	httpcontroller.ConfigureRouter(handler, services)

	// gRPC Server
	log.Infof("Starting gRPC server...")
	log.Debugf("Server port: %s", cfg.GRPC.Port)
	grpcHandler := grpc.NewServer()
	grpccontroller.ConfigureHandler(grpcHandler, services)
	grpcServer, err := grpcserver.New(grpcHandler, grpcserver.WithPort(cfg.GRPC.Port))
	if err != nil {
		panic(fmt.Errorf("app - Run - grpcserver.New: %w", err))
	}

	// HTTP Server
	log.Info("Starting HTTP server...")
	log.Debugf("Server port: %s", cfg.HTTP.Port)
	httpServer := httpserver.New(handler, httpserver.Port(cfg.HTTP.Port))

	// Prometheus server
	log.Infof("Starting metrics server...")
	log.Debugf("Server port: %s", cfg.Prometheus.Port)
	metricsHandler := echo.New()
	metrics.ConfigureHandler(metricsHandler)
	metricsServer := httpserver.New(metricsHandler, httpserver.Port(cfg.Prometheus.Port))

	log.Info("Configuring graceful shutdown...")
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	select {
	case s := <-interrupt:
		log.Info("app - Run - signal: " + s.String())
	case err = <-httpServer.Notify():
		log.Errorf("app - Run - httpServer.Notify: %v", err)
	case err = <-metricsServer.Notify():
		log.Errorf("app - Run - metricsServer.Notify: %v", err)
	case err = <-grpcServer.Notify():
		log.Errorf("app - Run - grpcServer.Notify: %v", err)
	}

	// Graceful shutdown
	log.Info("Shutting down...")

	err = httpServer.Shutdown()
	if err != nil {
		log.Errorf("app - Run - httpServer.Shutdown: %v", err)
	}

	err = metricsServer.Shutdown()
	if err != nil {
		log.Errorf("app - Run - metricsServer.Shutdown: %v", err)
	}

	grpcServer.Shutdown()
}
