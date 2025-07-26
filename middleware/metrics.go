package middleware

import (
	"net/http"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

type MetricsMiddleware struct {
	requestCounter  metric.Int64Counter
	requestDuration metric.Int64Histogram
}

func NewMetricsMiddleware() *MetricsMiddleware {
	meter := otel.GetMeterProvider().Meter("file-uploader/http-middleware")

	requestCounter, _ := meter.Int64Counter(
		"http.server.requests.total",
		metric.WithDescription("Total number of HTTP requests."),
		metric.WithUnit("{request}"),
	)

	requestDuration, _ := meter.Int64Histogram(
		"http.server.requests.duration",
		metric.WithDescription("The duration of HTTP requests."),
		metric.WithUnit("ms"),
	)

	return &MetricsMiddleware{
		requestCounter:  requestCounter,
		requestDuration: requestDuration,
	}
}

func (m *MetricsMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()

		rw := &responseWriter{ResponseWriter: w}

		next.ServeHTTP(rw.ResponseWriter, r)

		duration := time.Since(startTime).Milliseconds()

		attrs := []attribute.KeyValue{
			attribute.String("http.method", r.Method),
			attribute.String("http.route", r.URL.Path),
			attribute.Int("http.status_code", rw.statusCode),
		}

		m.requestCounter.Add(r.Context(), 1, metric.WithAttributes(attrs...))
		m.requestDuration.Record(r.Context(), duration, metric.WithAttributes(attrs...))
	})
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(statusCode int) {
	rw.statusCode = statusCode
	rw.ResponseWriter.WriteHeader(statusCode)
}
