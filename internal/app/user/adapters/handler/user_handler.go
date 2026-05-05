package handler

import (
	"errors"
	"net/http"
	"strings"

	"github.com/billykore/project-one/internal/app/user/core/domain"
	"github.com/billykore/project-one/internal/app/user/core/ports"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
)

type userHandler struct {
	userSvc   ports.UserService
	loginSvc  ports.LoginService
	validator *validator.Validate
}

// NewUserHandler creates a new instance of UserHandler.
func NewUserHandler(userSvc ports.UserService, loginSvc ports.LoginService, validator *validator.Validate) *userHandler {
	return &userHandler{
		userSvc:   userSvc,
		loginSvc:  loginSvc,
		validator: validator,
	}
}

// Me handles the GET /users/me endpoint.
// @Summary      Get current user
// @Description  Get the profile of the currently authenticated user.
// @Tags         users
// @Accept       json
// @Produce      json
// @Success      200  {object}  dto.UserResponse
// @Failure      401  {object}  dto.ErrorResponse
// @Failure      500  {object}  dto.ErrorResponse
// @Security     BearerAuth
// @Router       /users/me [get]
func (h *userHandler) Me(c echo.Context) error {
	userID, ok := c.Get("userID").(int)
	if !ok {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Unauthorized"})
	}

	user, err := h.userSvc.GetCurrentUser(c.Request().Context(), userID)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) || errors.Is(err, domain.ErrUnauthorized) {
			return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Unauthorized"})
		}
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Internal Server Error"})
	}

	res := UserResponse{
		Email: user.Email,
		Name:  user.FirstName + " " + user.LastName,
	}

	return c.JSON(http.StatusOK, res)
}

// HandleLogin handles the POST /users/login endpoint.
// @Summary      Login
// @Description  Authenticate a user and return access and refresh tokens via HttpOnly cookies.
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        request body dto.LoginRequest true "Login credentials"
// @Success      200  {object}  dto.LoginResponse
// @Failure      400  {object}  dto.ErrorResponse
// @Failure      401  {object}  dto.ErrorResponse
// @Failure      500  {object}  dto.ErrorResponse
// @Router       /users/login [post]
func (h *userHandler) HandleLogin(c echo.Context) error {
	var req LoginRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid request body"})
	}

	if err := h.validator.Struct(req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
	}

	accessToken, err := h.loginSvc.Login(c.Request().Context(), req.Email, req.Password)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidCredentials) {
			return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Invalid email or password"})
		}
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Internal server error"})
	}

	// Set access token cookie
	c.SetCookie(&http.Cookie{
		Name:     "access_token",
		Value:    accessToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   false, // Set to true in production
		SameSite: http.SameSiteLaxMode,
	})

	return c.JSON(http.StatusOK, LoginResponse{Message: "Login successful"})
}

// HandleLogout handles the POST /users/logout endpoint.
// @Summary      Logout
// @Description  Invalidate the current user's session.
// @Tags         users
// @Accept       json
// @Produce      json
// @Success      200  {object}  dto.LogoutResponse
// @Failure      401  {object}  dto.ErrorResponse
// @Failure      500  {object}  dto.ErrorResponse
// @Security     BearerAuth
// @Router       /users/logout [post]
func (h *userHandler) HandleLogout(c echo.Context) error {
	userID, ok := c.Get("userID").(int)
	if !ok {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Unauthorized"})
	}

	if err := h.loginSvc.Logout(c.Request().Context(), userID); err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Internal server error"})
	}

	return c.JSON(http.StatusOK, LogoutResponse{
		Message: "Logged out successfully",
	})
}

// HandleRegister handles the POST /users/register endpoint.
// @Summary      Register
// @Description  Create a new user account.
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        request body dto.RegisterRequest true "User registration details"
// @Success      201  {object}  dto.RegisterResponse
// @Failure      400  {object}  dto.ErrorResponse
// @Failure      500  {object}  dto.ErrorResponse
// @Router       /users/register [post]
func (h *userHandler) HandleRegister(c echo.Context) error {
	var req RegisterRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid request body"})
	}

	if err := h.validator.Struct(req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
	}

	user := &domain.User{
		FirstName: strings.TrimSpace(req.FirstName),
		LastName:  strings.TrimSpace(req.LastName),
		Email:     strings.ToLower(strings.TrimSpace(req.Email)),
		Password:  req.Password,
	}

	if err := h.userSvc.Register(c.Request().Context(), user); err != nil {
		if errors.Is(err, domain.ErrEmailAlreadyRegistered) {
			return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Email is already registered"})
		}
		if errors.Is(err, domain.ErrValidationFailed) {
			return c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		}
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Something went wrong"})
	}

	return c.JSON(http.StatusCreated, RegisterResponse{
		Message: "User registered successfully",
	})
}
