package main

import (
	"context"
	"errors"
	"flag"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/pizza-nz/file-uploader/config"
	"github.com/pizza-nz/file-uploader/handlers"
	"github.com/pizza-nz/file-uploader/logging"
	"github.com/pizza-nz/file-uploader/middleware"
)

func main() {
	configPath := flag.String("config", "/app/config.yml", "path to config file")
	flag.Parse()

	cfg, err := config.NewConfig(*configPath)
	if err != nil {
		slog.Error("Failed to load configuration", "error", err)
		os.Exit(1)
	}

	if !config.ValidateConfig(cfg) {
		slog.Error("Invalid configuration")
		os.Exit(1)
	}

	logger := logging.NewLogger(cfg.Logging.Level)
	slog.SetDefault(logger)

	mux := http.NewServeMux()
	handl := handlers.NewFileUploadHandler(cfg.File.MaxSize, cfg.File.Path)
	mux.HandleFunc("POST /upload", handl.CreateFileUpload)
	mux.HandleFunc("GET /health", handlers.HealthCheck)

	server := http.Server{
		Addr:    cfg.Server.Port,
		Handler: middleware.RequestIDMiddleware(mux),
	}

	go func() {
		slog.Info("Starting server", "addr", cfg.Server.Port)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("Server failed to start", "error", err)
			os.Exit(1)
		}
	}()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	<-ctx.Done()

	slog.Info("Shutting down server")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		slog.Error("Server shutdown failed", "error", err)
	} else {
		slog.Info("Server shutdown gracefully")
	}

	os.Exit(0)
}
