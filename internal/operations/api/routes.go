// Package api exposes the HTTP routes owned by the operations bounded context.
package api

import (
	"net/http"

	operationshandler "github.com/billykore/project-one/internal/operations/api/handler"
	operationsmiddleware "github.com/billykore/project-one/internal/operations/api/middleware"
	"github.com/labstack/echo/v4"
)

// RegisterRoutes mounts liveness, readiness, and authenticated metrics endpoints.
func RegisterRoutes(e *echo.Echo, healthHandler *operationshandler.HealthHandler, metricsHandler http.Handler) {
	// Liveness answers for the process only; readiness reports dependency state.
	e.GET("/healthz", healthHandler.HandleLiveness)
	e.GET("/status", healthHandler.HandleReadiness)
	RegisterMetricsRoute(e, metricsHandler)
}

// RegisterMetricsRoute mounts the authenticated Prometheus scrape handler on
// the private Compose network path. It is intentionally absent from the public
// API surface and from the Swagger contract.
func RegisterMetricsRoute(e *echo.Echo, handler http.Handler) {
	e.GET(operationsmiddleware.SelfScrapeRoute, echo.WrapHandler(handler))
}
