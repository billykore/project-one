package middleware

import (
	"strings"

	"github.com/billykore/project-one/internal/platform/problem"
	"github.com/labstack/echo/v4"
)

// OperatorOnly permits only usernames in the configured operator allowlist.
func OperatorOnly(usernames []string) echo.MiddlewareFunc {
	allowed := make(map[string]struct{}, len(usernames))
	for _, username := range usernames {
		if normalized := strings.TrimSpace(username); normalized != "" {
			allowed[normalized] = struct{}{}
		}
	}
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			username, ok := c.Get("username").(string)
			if !ok {
				return echo.ErrForbidden
			}
			if _, ok := allowed[username]; !ok {
				return problem.ErrOperatorOnly
			}
			return next(c)
		}
	}
}
