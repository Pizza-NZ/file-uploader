package logging

import (
	"context"
	"log/slog"

	"go.opentelemetry.io/otel/trace"
)

type OtelSlogHandler struct {
	slog.Handler
	next slog.Handler
}

func NewOtelSlogHandler(next slog.Handler) *OtelSlogHandler {
	return &OtelSlogHandler{next: next}
}

func (h *OtelSlogHandler) Handle(ctx context.Context, r slog.Record) error {
	span := trace.SpanFromContext(ctx)
	if span.SpanContext().IsValid() {
		r.AddAttrs(
			slog.String("trace_id", span.SpanContext().TraceID().String()),
			slog.String("span_id", span.SpanContext().SpanID().String()),
		)
	}

	return h.next.Handle(ctx, r)
}

func (h *OtelSlogHandler) WithAttributes(attrs []slog.Attr) slog.Handler {
	return &OtelSlogHandler{next: h.next.WithAttrs(attrs)}
}

func (h *OtelSlogHandler) WithGroup(name string) slog.Handler {
	return &OtelSlogHandler{next: h.next.WithGroup(name)}
}
func (h *OtelSlogHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.next.Enabled(ctx, level)
}
