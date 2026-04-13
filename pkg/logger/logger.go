package logger

import (
	"log/slog"
	"os"
)

// New creates a new slog.Logger configured for the application.
func New() *slog.Logger {
	return slog.New(slog.NewJSONHandler(os.Stderr, nil))
}
