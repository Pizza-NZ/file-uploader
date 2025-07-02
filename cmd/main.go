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

	"github.com/pizza-nz/file-uploader/handlers"
	"github.com/pizza-nz/file-uploader/logging"
	"github.com/pizza-nz/file-uploader/middleware"
)

func main() {
	addr := flag.String("addr", ":2131", "address to listen")
	env := flag.String("env", "development", "environment")
	flag.Parse()

	logger := logging.NewLogger(*env)
	slog.SetDefault(logger)

	mux := http.NewServeMux()
	handl := handlers.NewFileUploadHandler(200 << 20)
	mux.HandleFunc("POST /upload", handl.CreateFileUpload)

	server := http.Server{
		Addr:    *addr,
		Handler: middleware.RequestIDMiddleware(mux),
	}

	go func() {
		slog.Info("Starting server", "addr", *addr)
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