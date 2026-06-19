package logger

import (
	"context"
	"log/slog"
	"os"
)

// Logger is a shared Logger implementation based on slog.
type Logger struct {
	slog *slog.Logger
}

// New creates a new instance of Logger.
func New() *Logger {
	// ponytail: simplified logger adapter by using standard library slog instead of zerolog dependency
	l := slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
		AddSource: true,
		Level:     slog.LevelDebug,
	}))
	return &Logger{slog: l}
}

// Debug logs a debug message with optional fields.
func (l *Logger) Debug(ctx context.Context, msg string, fields ...any) {
	l.slog.DebugContext(ctx, msg, fields...)
}

// Info logs an info message with optional fields.
func (l *Logger) Info(ctx context.Context, msg string, fields ...any) {
	l.slog.InfoContext(ctx, msg, fields...)
}

// Warn logs a warning message with optional fields.
func (l *Logger) Warn(ctx context.Context, msg string, fields ...any) {
	l.slog.WarnContext(ctx, msg, fields...)
}

// Error logs an error message with optional fields.
func (l *Logger) Error(ctx context.Context, msg string, fields ...any) {
	l.slog.ErrorContext(ctx, msg, fields...)
}

// Fatal logs a fatal message with optional fields and terminates the process.
func (l *Logger) Fatal(ctx context.Context, msg string, fields ...any) {
	l.slog.ErrorContext(ctx, "FATAL: "+msg, fields...)
	os.Exit(1)
}
