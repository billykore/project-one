package logger

import (
	"context"
	"os"

	"github.com/rs/zerolog"
)

// Logger is a shared logger implementation based on zerolog.
type Logger struct {
	logger zerolog.Logger
}

// New creates a new instance of Logger.
func New() *Logger {
	l := zerolog.New(os.Stderr).With().Timestamp().Logger()
	return &Logger{logger: l}
}

// Debug logs a debug message with optional fields.
func (l *Logger) Debug(_ context.Context, msg string, fields ...any) {
	l.logger.Debug().Fields(toFields(fields...)).Msg(msg)
}

// Info logs an info message with optional fields.
func (l *Logger) Info(_ context.Context, msg string, fields ...any) {
	l.logger.Info().Fields(toFields(fields...)).Msg(msg)
}

// Warn logs a warning message with optional fields.
func (l *Logger) Warn(_ context.Context, msg string, fields ...any) {
	l.logger.Warn().Fields(toFields(fields...)).Msg(msg)
}

// Error logs an error message with optional fields.
func (l *Logger) Error(_ context.Context, msg string, fields ...any) {
	l.logger.Error().Fields(toFields(fields...)).Msg(msg)
}

// Fatal logs a fatal message with optional fields.
func (l *Logger) Fatal(_ context.Context, msg string, fields ...any) {
	l.logger.Fatal().Fields(toFields(fields...)).Msg(msg)
}

// toFields converts a slice of any to a map of string to any.
func toFields(args ...any) map[string]any {
	fields := make(map[string]any)
	for i := 0; i < len(args); i += 2 {
		if i+1 < len(args) {
			if key, ok := args[i].(string); ok {
				fields[key] = args[i+1]
			}
		}
	}
	return fields
}
