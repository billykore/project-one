package handler

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/billykore/project-one/internal/app/user/adapters/dto"
	"github.com/billykore/project-one/internal/app/user/core/domain"
	"github.com/billykore/project-one/internal/app/user/core/ports"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
)

type userHandler struct {
	svc       ports.UserService
	loginSvc  ports.LoginService
	validator *validator.Validate
}

// NewUserHandler creates a new instance of UserHandler.
func NewUserHandler(svc ports.UserService, loginSvc ports.LoginService, validator *validator.Validate) *userHandler {
	return &userHandler{
		svc:       svc,
		loginSvc:  loginSvc,
		validator: validator,
	}
}

func (h *userHandler) Me(c echo.Context) error {
	userID, ok := c.Get("userID").(int)
	if !ok {
		return c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "Unauthorized"})
	}

	user, err := h.svc.GetCurrentUser(c.Request().Context(), userID)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "Unauthorized"})
		}
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "Internal Server Error"})
	}

	res := dto.UserResponse{
		ID:    strconv.Itoa(user.ID),
		Email: user.Email,
	}

	return c.JSON(http.StatusOK, res)
}

// HandleLogin handles the POST /user/login endpoint.
func (h *userHandler) HandleLogin(c echo.Context) error {
	var req dto.LoginRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid request body"})
	}

	if err := h.validator.Struct(req); err != nil {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
	}

	accessToken, refreshToken, err := h.loginSvc.Login(c.Request().Context(), req.Email, req.Password)
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

// HandleLogout handles the POST /user/logout endpoint.
func (h *userHandler) HandleLogout(c echo.Context) error {
	authHeader := c.Request().Header.Get("Authorization")
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 {
		return c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "Unauthorized"})
	}
	token := parts[1]

	if err := h.loginSvc.Logout(c.Request().Context(), token); err != nil {
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "internal server error"})
	}

	return c.JSON(http.StatusOK, dto.LogoutResponse{
		Message: "Logged out successfully",
	})
}
