package middleware

import (
	"strings"

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
				return echo.ErrUnauthorized
			}

			user, err := tks.ValidateToken(c.Request().Context(), token)
			if err != nil {
				return echo.ErrUnauthorized
			}
			if user == nil {
				return echo.ErrUnauthorized
			}

			// Store user for downstream handlers
			c.Set("user", user)
			c.Set("username", user.Username)

			return next(c)
		}
	}
}

// OptionalAuthorize attaches a user when a valid session is present and treats
// missing or invalid credentials as anonymous access.
func OptionalAuthorize(tks ports.TokenService) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			token := requestToken(c)
			if token == "" {
				return next(c)
			}
			user, err := tks.ValidateToken(c.Request().Context(), token)
			if err == nil && user != nil {
				c.Set("user", user)
				c.Set("username", user.Username)
			}
			return next(c)
		}
	}
}

func requestToken(c echo.Context) string {
	if cookie, err := c.Cookie("access_token"); err == nil && cookie.Value != "" {
		return cookie.Value
	}
	if after, ok := strings.CutPrefix(c.Request().Header.Get("Authorization"), "Bearer "); ok {
		return after
	}
	return c.QueryParam("token")
}
