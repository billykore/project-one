package middleware

import (
	"fmt"
	"net/http"
	"time"

	operationsports "github.com/billykore/project-one/internal/operations/ports"
	platformmiddleware "github.com/billykore/project-one/internal/platform/api/middleware"
	"github.com/labstack/echo/v4"
)

// UnmatchedRoute is the single route label used when a request resolved to no
// route template. Raw request paths and query strings are never label values.
const UnmatchedRoute = "unmatched"

// SelfScrapeRoute is excluded from its own observations.
const SelfScrapeRoute = "/metrics"

// Metrics observes completed requests through the metrics port using only the
// HTTP method, the resolved route template, and the response status class.
func Metrics(recorder operationsports.HTTPMetrics) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()
			err := next(c)

			route := c.Path()
			if route == SelfScrapeRoute {
				// ponytail: a scrape must not observe itself, otherwise every
				// interval would inflate the request rate it is measuring.
				return err
			}
			if route == "" {
				route = UnmatchedRoute
			}

			status := c.Response().Status
			if err != nil {
				// Echo runs the HTTPErrorHandler after this middleware returns,
				// so resolve the status the client will actually receive.
				status = platformmiddleware.StatusForError(err)
			}

			method := c.Request().Method
			ctx := c.Request().Context()
			recorder.ObserveRequest(ctx, method, route, statusClass(status))
			recorder.ObserveRequestDuration(ctx, method, route, time.Since(start).Seconds())

			return err
		}
	}
}

// statusClass reduces a status code to its bounded class, e.g. 404 to "4xx".
func statusClass(status int) string {
	if status < 100 || status > 599 {
		// Echo defaults to 200 when a handler completes without writing.
		status = http.StatusOK
	}
	return fmt.Sprintf("%dxx", status/100)
}
