package logging

import (
	"log/slog"
	"os"
)

func NewLogger(env string) *slog.Logger {
	switch env {
	case "development":
		return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		}))
	case "production":
		return slog.New(slog.NewJSONHandler(os.Stdout, nil))
	default:
		return slog.New(slog.NewTextHandler(os.Stdout, nil))
	}
}