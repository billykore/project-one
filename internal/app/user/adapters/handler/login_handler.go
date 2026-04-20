package handler

import (
	"errors"
	"net/http"
	"strings"

	"github.com/billykore/project-one/internal/app/user/adapters/dto"
	"github.com/billykore/project-one/internal/app/user/core/domain"
	"github.com/billykore/project-one/internal/app/user/core/ports"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
)

type loginHandler struct {
	svc       ports.LoginService
	validator *validator.Validate
}

// NewLoginHandler creates a new instance of LoginHandler.
func NewLoginHandler(svc ports.LoginService, validator *validator.Validate) *loginHandler {
	return &loginHandler{
		svc:       svc,
		validator: validator,
	}
}

// HandleLogin handles the POST /login endpoint.
func (h *loginHandler) HandleLogin(c echo.Context) error {
	var req dto.LoginRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid request body"})
	}

	if err := h.validator.Struct(req); err != nil {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
	}

	accessToken, refreshToken, err := h.svc.Login(c.Request().Context(), req.Email, req.Password)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidCredentials) {
			return c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "invalid email or password"})
		}
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "internal server error"})
	}

	return c.JSON(http.StatusOK, dto.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	})
}

// HandleLogout handles the POST /logout endpoint.
func (h *loginHandler) HandleLogout(c echo.Context) error {
	authHeader := c.Request().Header.Get("Authorization")
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 {
		return c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "Unauthorized"})
	}
	token := parts[1]

	if err := h.svc.Logout(c.Request().Context(), token); err != nil {
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "internal server error"})
	}

	return c.JSON(http.StatusOK, dto.LogoutResponse{
		Message: "Logged out successfully",
	})
}
