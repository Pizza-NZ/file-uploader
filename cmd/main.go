package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/pizza-nz/file-uploader/config"
	"github.com/pizza-nz/file-uploader/handlers"
	"github.com/pizza-nz/file-uploader/instrumentation"
	"github.com/pizza-nz/file-uploader/logging"
	"github.com/pizza-nz/file-uploader/middleware"
	"github.com/pizza-nz/file-uploader/services"
	"github.com/pizza-nz/file-uploader/storage"
)

func handleStartupError(msg string, err error) {
	if err != nil {
		slog.Error(msg, "error", err)
		os.Exit(1)
	}
}

func main() {
	configPath := flag.String("config", "config.yml", "path to config file")
	flag.Parse()

	cfg, err := config.NewConfig(*configPath)
	if err != nil {
		handleStartupError("Failed to load configuration", err)
	}

	if err = config.ValidateConfig(cfg); err != nil {
		handleStartupError("Configuration validation failed", err)
	}

	baseHandler := logging.NewLogger(cfg.Environment)
	otelHandler := logging.NewOtelSlogHandler(baseHandler.Handler())

	logger := slog.New(otelHandler)
	slog.SetDefault(logger)

	shutdown, err := instrumentation.SetupOTelSDK(context.Background())
	if err != nil {
		handleStartupError("Failed to set up OpenTelemetry SDK", err)
	}
	defer shutdown(context.Background())

	metricsMiddleware := middleware.NewMetricsMiddleware()

	var fileStorage storage.FileStorage
	switch cfg.StorageType {
	case "s3":
		var err error
		fileStorage, err = storage.NewS3Storage(context.Background(), cfg.AWS)
		if err != nil {
			handleStartupError("Failed to create S3 storage", err)
		}
	case "mock":
		fileStorage = storage.NewMockFileStorage()
	default:
		handleStartupError("Invalid storage type", fmt.Errorf("storage type '%s' is not supported", cfg.StorageType))
	}

	fileUploadService := services.NewFileUploadService(fileStorage, cfg.File.AllowedTypes)

	mux := http.NewServeMux()
	handl := handlers.NewFileUploadHandler(cfg.File.MaxSize, fileUploadService)
	mux.HandleFunc("POST /upload", handl.CreateFileUpload)
	mux.HandleFunc("GET /health", handlers.HealthCheck)

	finalHandler := middleware.RequestIDMiddleware(
		metricsMiddleware.Handler(
			middleware.OpenTelemetryMiddleware(mux),
		),
	)

	server := http.Server{
		Addr:    cfg.Server.Port,
		Handler: finalHandler,
	}

	go func() {
		slog.Info("Starting server", "addr", cfg.Server.Port)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			handleStartupError("Server failed to start", err)
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
