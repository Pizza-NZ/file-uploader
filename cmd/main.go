package main

import (
	"context"
	"errors"
	"flag"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/pizza-nz/file-uploader/handlers"
	"github.com/pizza-nz/file-uploader/middleware"
)

func main() {
	addr := flag.String("addr", ":2131", "address to listen")
	flag.Parse()

	mux := http.NewServeMux()
	handl := handlers.NewFileUploadHandler()
	mux.HandleFunc("POST /upload", handl.CreateFileUpload)

	server := http.Server{
		Addr:    *addr,
		Handler: middleware.RequestIDMiddleware(mux),
	}

	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		// failed
	} else {
		// graceful
	}

	os.Exit(0)
}
