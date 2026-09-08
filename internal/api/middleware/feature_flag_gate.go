package middleware

import (
	"context"

	"github.com/billykore/project-one/internal/core/domain"
	"github.com/labstack/echo/v4"
)

// FeatureFlagGate rejects guarded requests when the current decision is disabled.
func FeatureFlagGate(evaluator interface {
	Evaluate(ctx context.Context, key string, username string) domain.FeatureFlagDecision
}, key string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			username, _ := c.Get("username").(string)
			decision := evaluator.Evaluate(c.Request().Context(), key, username)
			if !decision.Enabled {
				return echo.ErrNotFound
			}
			return next(c)
		}
	}
}
