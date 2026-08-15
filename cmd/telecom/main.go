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

	"github.com/local/telecom/internal/api"
	"github.com/local/telecom/internal/backup"
	"github.com/local/telecom/internal/config"
	"github.com/local/telecom/internal/database"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("invalid configuration", "error", err)
		os.Exit(1)
	}
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: cfg.LogLevel}))
	slog.SetDefault(logger)
	if applied, applyErr := backup.ApplyPending(cfg.DataDir); applyErr != nil {
		logger.Error("pending restore failed", "error", applyErr)
		os.Exit(1)
	} else if applied {
		logger.Info("pending restore applied")
	}

	db, err := database.Open(cfg.DatabasePath)
	if err != nil {
		logger.Error("database initialization failed", "error", err)
		os.Exit(1)
	}
	defer db.Close()
	if err := database.Migrate(db); err != nil {
		logger.Error("database migration failed", "error", err)
		os.Exit(1)
	}

	restoreRequested := make(chan struct{}, 1)
	server := &http.Server{
		Addr: cfg.HTTPAddress(),
		Handler: api.NewRouter(db, logger, cfg.ScanWorkers, cfg.DataDir, func() {
			select {
			case restoreRequested <- struct{}{}:
			default:
			}
		}),
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		// SSE de progresso permanece aberto durante todo o scan; o encerramento
		// gracioso e os timeouts internos do scanner limitam esses streams.
		WriteTimeout: 0,
		IdleTimeout:  60 * time.Second,
	}
	go func() {
		logger.Info("http server started", "address", server.Addr, "data_dir", cfg.DataDir)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("http server failed", "error", err)
			os.Exit(1)
		}
	}()

	signalContext, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	restore := false
	select {
	case <-signalContext.Done():
	case <-restoreRequested:
		restore = true
	}
	logger.Info("shutdown requested")
	shutdownContext, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownContext); err != nil {
		logger.Error("graceful shutdown failed", "error", err)
		os.Exit(1)
	}
	if restore {
		_, _ = db.Exec("PRAGMA wal_checkpoint(TRUNCATE)")
		_ = db.Close()
		applied, restoreErr := backup.ApplyPending(cfg.DataDir)
		if restoreErr != nil {
			logger.Error("restore failed", "error", restoreErr)
			os.Exit(1)
		}
		logger.Info("restore applied", "applied", applied, "restart_required", true)
	}
	logger.Info("shutdown complete")
}
