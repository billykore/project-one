package logger

import (
	"context"
	"os"

	"github.com/billykore/project-one/internal/app/login/core/ports"
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
	l.logger.Debug().Fields(fields).Msg(msg)
}

func (l *zerologLogger) Info(_ context.Context, msg string, fields ...any) {
	l.logger.Info().Fields(fields).Msg(msg)
}

func (l *zerologLogger) Warn(_ context.Context, msg string, fields ...any) {
	l.logger.Warn().Fields(fields).Msg(msg)
}

func (l *zerologLogger) Error(_ context.Context, msg string, fields ...any) {
	l.logger.Error().Fields(fields).Msg(msg)
}

func (l *zerologLogger) With(fields ...any) ports.Logger {
	return &zerologLogger{logger: l.logger.With().Fields(fields).Logger()}
}
