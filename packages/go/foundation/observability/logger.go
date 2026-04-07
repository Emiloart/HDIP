package observability

import (
	"log/slog"
	"os"
	"strings"
)

func NewJSONLogger(service string, rawLevel string) (*slog.Logger, error) {
	var level slog.Level
	switch strings.ToUpper(strings.TrimSpace(rawLevel)) {
	case "DEBUG":
		level = slog.LevelDebug
	case "", "INFO":
		level = slog.LevelInfo
	case "WARN":
		level = slog.LevelWarn
	case "ERROR":
		level = slog.LevelError
	default:
		return nil, ErrInvalidLogLevel
	}

	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level})
	return slog.New(handler).With("service", service), nil
}
