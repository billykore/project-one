package middleware

import (
	"net/http"
	"strings"

	"github.com/billykore/project-one/internal/api/dto"
	"github.com/billykore/project-one/internal/core/ports"
	"github.com/labstack/echo/v4"
)

// Authorize is an middleware to authorize requests.
func Authorize(tks ports.TokenService) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			var token string

			cookie, err := c.Cookie("access_token")
			if err == nil {
				token = cookie.Value
			}

			if token == "" {
				authHeader := c.Request().Header.Get("Authorization")
				if after, ok := strings.CutPrefix(authHeader, "Bearer "); ok {
					token = after
				}
			}

			// Query param for browser WebSocket connections — they can't set headers during upgrade.
			if token == "" {
				token = c.QueryParam("token")
			}

			if token == "" {
				return c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "Unauthorized"})
			}

			username, err := tks.ValidateToken(c.Request().Context(), token)
			if err != nil {
				return c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "Unauthorized"})
			}

			// Store username for downstream handlers
			c.Set("username", username)

			return next(c)
		}
	}
}
