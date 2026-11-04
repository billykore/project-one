package ports

import "context"

// Logger is a driven port for logging.
type Logger interface {
	// Debug logs a debug message with optional fields.
	Debug(ctx context.Context, msg string, fields ...any)
	// Info logs an info message with optional fields.
	Info(ctx context.Context, msg string, fields ...any)
	// Warn logs a warning message with optional fields.
	Warn(ctx context.Context, msg string, fields ...any)
	// Error logs an error message with optional fields.
	Error(ctx context.Context, msg string, fields ...any)
	// Fatal logs a fatal message with optional fields and may terminate the process.
	Fatal(ctx context.Context, msg string, fields ...any)
}
