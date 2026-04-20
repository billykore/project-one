package handler

import (
	"context"
	"strings"

	"github.com/billykore/project-one/internal/app/user/adapters/dto"
	"github.com/billykore/project-one/internal/app/user/core/ports"
	"github.com/labstack/echo/v4"
)

type contextKey string

const userIDKey contextKey = "userID"

// AuthMiddleware is an Echo middleware for JWT authentication.
func AuthMiddleware(tks ports.TokenService) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				return c.JSON(403, dto.ErrorResponse{Error: "Unauthorized"})
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				return c.JSON(403, dto.ErrorResponse{Error: "Unauthorized"})
			}

			token := parts[1]
			userID, err := tks.ValidateToken(c.Request().Context(), token)
			if err != nil {
				return c.JSON(403, dto.ErrorResponse{Error: "Unauthorized"})
			}

			// Store userID in context for downstream handlers
			ctx := context.WithValue(c.Request().Context(), userIDKey, userID)
			c.SetRequest(c.Request().WithContext(ctx))

			return next(c)
		}
	}
}
