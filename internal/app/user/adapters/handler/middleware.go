package handler

import (
	"net/http"
	"strings"

	"github.com/billykore/project-one/internal/app/user/adapters/dto"
	"github.com/billykore/project-one/internal/app/user/core/ports"
	"github.com/labstack/echo/v4"
)

// AuthMiddleware is an Echo middleware for JWT authentication.
func AuthMiddleware(tks ports.TokenService) echo.MiddlewareFunc {
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
				if strings.HasPrefix(authHeader, "Bearer ") {
					token = strings.TrimPrefix(authHeader, "Bearer ")
				}
			}

			if token == "" {
				return c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "Unauthorized"})
			}

			userID, err := tks.ValidateToken(c.Request().Context(), token)
			if err != nil {
				return c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "Unauthorized"})
			}

			// Store userID for downstream handlers
			c.Set("userID", userID)

			return next(c)
		}
	}
}
