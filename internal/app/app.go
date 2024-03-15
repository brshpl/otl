package app

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/brshpl/otl/config"
	"github.com/brshpl/otl/pkg/random"
	"github.com/gin-gonic/gin"

	v1 "github.com/brshpl/otl/internal/controller/http/v1"
	"github.com/brshpl/otl/internal/usecase"
	"github.com/brshpl/otl/internal/usecase/repo"
	"github.com/brshpl/otl/pkg/httpserver"
	"github.com/brshpl/otl/pkg/logger"
	"github.com/brshpl/otl/pkg/postgres"
)

// Run creates objects via constructors.
func Run(cfg *config.Config) {
	l := logger.New(cfg.Log.Level)

	// Repository
	pg, err := postgres.New(cfg.PG.URL,
		postgres.MaxPoolSize(cfg.PG.PoolMax),
		postgres.ConnAttempts(10),
		postgres.ConnTimeout(time.Second))
	if err != nil {
		l.Fatal(fmt.Errorf("app - Run - postgres.NewPostgres: %w", err))
	}
	defer pg.Close()

	// Use case
	oneTimeLinkUseCase := usecase.New(repo.NewPostgres(pg), random.GenerateString, cfg.OTL.SymbolsCount)

	// HTTP Server
	handler := gin.New()
	v1.NewRouter(handler, l, oneTimeLinkUseCase)
	httpServer := httpserver.New(handler,
		httpserver.Port(cfg.HTTP.Port),
		httpserver.WriteTimeout(5*time.Second),
		httpserver.ReadTimeout(5*time.Second),
		httpserver.ShutdownTimeout(5*time.Second),
	)

	// Waiting signal
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	select {
	case s := <-interrupt:
		l.Info("app - Run - signal: " + s.String())
	case err = <-httpServer.Notify():
		l.Error(fmt.Errorf("app - Run - httpServer.Notify: %w", err))
	}

	// Shutdown
	err = httpServer.Shutdown()
	if err != nil {
		l.Error(fmt.Errorf("app - Run - httpServer.Shutdown: %w", err))
	}
}
