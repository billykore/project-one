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

			// Check for access_token cookie first
			cookie, err := c.Cookie("access_token")
			if err == nil {
				token = cookie.Value
			}

			// If no cookie, check Authorization header
			if token == "" {
				authHeader := c.Request().Header.Get("Authorization")
				if after, ok := strings.CutPrefix(authHeader, "Bearer "); ok {
					token = after
				}
			}

			// If still no token, check for "token" query parameter.
			// This is used by browser WebSocket connections which cannot
			// set custom headers during the HTTP upgrade handshake.
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
