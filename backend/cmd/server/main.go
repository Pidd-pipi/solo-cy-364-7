package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ld/storeinventory/internal/config"
	"github.com/ld/storeinventory/internal/router"
	"github.com/ld/storeinventory/internal/util"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("load config failed", "error", err)
		os.Exit(1)
	}
	logger := util.NewLogger(cfg.AppEnv)
	util.SetLogger(logger)

	r, err := router.NewRouter(cfg, logger)
	if err != nil {
		logger.Error("init router failed", "error", err)
		os.Exit(1)
	}

	srv := &http.Server{
		Addr:    ":" + cfg.ServerPort,
		Handler: r,
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		logger.Info("store inventory server started", "port", cfg.ServerPort)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("server run failed", "error", err)
			stop()
		}
	}()

	<-ctx.Done()
	logger.Info("shutdown signal received, draining connections")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("server shutdown failed", "error", err)
		os.Exit(1)
	}
	logger.Info("server stopped gracefully")
}
