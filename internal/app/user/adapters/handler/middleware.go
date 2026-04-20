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
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				return c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "Unauthorized"})
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				return c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "Unauthorized"})
			}

			token := parts[1]
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
