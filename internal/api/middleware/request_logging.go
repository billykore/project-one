package middleware

import (
	"net/http"
	"time"

	"github.com/billykore/project-one/internal/core/ports"
	"github.com/labstack/echo/v4"
)

const (
	// RequestCompletedMessage is a stable event name, never user input.
	RequestCompletedMessage = "http request completed"

	FailureCategoryClient = "client_error"
	FailureCategoryServer = "server_error"
)

// RequestLogging emits one safe structured event after each request. It uses
// only the resolved route template and bounded response data; raw request
// paths, queries, headers, bodies, addresses, errors, and stack traces are
// intentionally never added.
func RequestLogging(log ports.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()
			err := next(c)

			status := c.Response().Status
			if err != nil {
				status = StatusForError(err)
			}
			if status < 100 || status > 599 {
				status = http.StatusOK
			}

			route := c.Path()
			if route == "" {
				route = UnmatchedRoute
			}

			fields := []any{
				"request_id", c.Response().Header().Get(echo.HeaderXRequestID),
				"method", c.Request().Method,
				"route", route,
				"status", status,
				"duration_ms", float64(time.Since(start).Microseconds()) / 1000,
			}

			switch {
			case status >= http.StatusInternalServerError:
				fields = appendFailure(fields, err, FailureCategoryServer)
				log.Error(c.Request().Context(), RequestCompletedMessage, fields...)
			case status >= http.StatusBadRequest:
				fields = appendFailure(fields, err, FailureCategoryClient)
				log.Warn(c.Request().Context(), RequestCompletedMessage, fields...)
			default:
				log.Info(c.Request().Context(), RequestCompletedMessage, fields...)
			}

			return err
		}
	}
}

func appendFailure(fields []any, err error, category string) []any {
	fields = append(fields, "failure_category", category)
	if err != nil {
		fields = append(fields, "error_code", LookupError(err).Code)
	}
	return fields
}
