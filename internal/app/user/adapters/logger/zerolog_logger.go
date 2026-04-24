package logger

import (
	"context"
	"os"

	"github.com/billykore/project-one/internal/app/user/core/ports"
	"github.com/rs/zerolog"
)

type zerologLogger struct {
	logger zerolog.Logger
}

// NewZerologLogger creates a new instance of Logger.
func NewZerologLogger() ports.Logger {
	l := zerolog.New(os.Stderr).With().Timestamp().Logger()
	return &zerologLogger{logger: l}
}

func (l *zerologLogger) Debug(_ context.Context, msg string, fields ...any) {
	l.logger.Debug().Fields(toFields(fields...)).Msg(msg)
}

func (l *zerologLogger) Info(_ context.Context, msg string, fields ...any) {
	l.logger.Info().Fields(toFields(fields...)).Msg(msg)
}

func (l *zerologLogger) Warn(_ context.Context, msg string, fields ...any) {
	l.logger.Warn().Fields(toFields(fields...)).Msg(msg)
}

func (l *zerologLogger) Error(_ context.Context, msg string, fields ...any) {
	l.logger.Error().Fields(toFields(fields...)).Msg(msg)
}

func (l *zerologLogger) Fatal(_ context.Context, msg string, fields ...any) {
	l.logger.Fatal().Fields(toFields(fields...)).Msg(msg)
}

func (l *zerologLogger) With(fields ...any) ports.Logger {
	return &zerologLogger{logger: l.logger.With().Fields(toFields(fields...)).Logger()}
}

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
