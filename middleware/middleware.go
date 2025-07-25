package middleware

import (
	"log/slog"
	"net/http"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// RequestIDMiddleware is a middleware that generates a unique request ID for each incoming HTTP request.
// It adds the request ID to the response header and logs the request details.
func RequestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := uuid.New().String()

		w.Header().Set("X-Request-ID", requestID)
		slog.Info("Received request", "requestID", requestID, "method", r.Method, "url", r.URL.String())

		next.ServeHTTP(w, r)
	})
}

func OpenTelemetryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Start a new span for the request
		ctx, span := otel.Tracer("middleware").Start(r.Context(), "http.method "+r.Method, trace.WithAttributes(attribute.String("http.url", r.URL.String())))
		defer span.End()

		span.SetAttributes(
			attribute.String("request.id", r.Header.Get("X-Request-ID")),
			attribute.String("user.agent", r.UserAgent()),
		)

		// Set the request context with the span
		r = r.WithContext(ctx)

		// Call the next handler
		next.ServeHTTP(w, r)

		// Log the request completion
		slog.Info("Request completed", "requestID", r.Header.Get("X-Request-ID"), "method", r.Method, "url", r.URL.String())
	})
}
