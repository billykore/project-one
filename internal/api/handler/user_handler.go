package handler

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/billykore/project-one/internal/api/dto"
	"github.com/billykore/project-one/internal/core/domain"
	"github.com/billykore/project-one/internal/core/ports"
	"github.com/labstack/echo/v4"
)

type UserHandler struct {
	userUseCase   ports.UserUseCase
	loginUseCase  ports.LoginUseCase
	followUseCase ports.FollowUseCase
	validator     ports.Validator
	log           ports.Logger
}

// NewUserHandler creates a new instance of UserHandler.
func NewUserHandler(
	userUseCase ports.UserUseCase,
	loginUseCase ports.LoginUseCase,
	followUseCase ports.FollowUseCase,
	validator ports.Validator,
	log ports.Logger,
) *UserHandler {
	if userUseCase == nil || loginUseCase == nil || followUseCase == nil || validator == nil || log == nil {
		panic("NewUserHandler: dependencies must not be nil")
	}
	return &UserHandler{
		userUseCase:   userUseCase,
		loginUseCase:  loginUseCase,
		followUseCase: followUseCase,
		validator:     validator,
		log:           log,
	}
}

// Me handles the GET /users/me endpoint.
//
//	@Summary		Get current user
//	@Description	Get the profile of the currently authenticated user.
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	dto.UserResponse
//	@Failure		401	{object}	dto.ErrorResponse
//	@Failure		500	{object}	dto.ErrorResponse
//	@Security		BearerAuth
//	@Router			/users/me [get]
func (h *UserHandler) Me(c echo.Context) error {
	userID, ok := c.Get("userID").(int)
	if !ok {
		return c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "Unauthorized"})
	}

	user, err := h.userUseCase.GetCurrentUser(c.Request().Context(), userID)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) || errors.Is(err, domain.ErrUnauthorized) {
			return c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "Unauthorized"})
		}
		h.log.Error(c.Request().Context(), "failed to get current user", "userID", userID, "error", err)
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "Internal Server Error"})
	}

	res := dto.UserResponse{
		Email: user.Email,
		Name:  user.FirstName + " " + user.LastName,
	}

	return c.JSON(http.StatusOK, res)
}

// HandleLogin handles the POST /users/login endpoint.
//
//	@Summary		Login
//	@Description	Authenticate a user and return access and refresh tokens via HttpOnly cookies.
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Param			LoginRequest	body		dto.LoginRequest	true	"Login credentials"
//	@Success		200				{object}	dto.LoginResponse
//	@Failure		400				{object}	dto.ErrorResponse
//	@Failure		401				{object}	dto.ErrorResponse
//	@Failure		500				{object}	dto.ErrorResponse
//	@Router			/users/login [post]
func (h *UserHandler) HandleLogin(c echo.Context) error {
	var req dto.LoginRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "Invalid request body"})
	}

	if err := h.validator.Validate(req); err != nil {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
	}

	accessToken, err := h.loginUseCase.Login(c.Request().Context(), req.Email, req.Password)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidCredentials) {
			return c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "Invalid email or password"})
		}
		h.log.Error(c.Request().Context(), "login failed", "email", req.Email, "error", err)
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "Internal server error"})
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

	return c.JSON(http.StatusOK, dto.LoginResponse{Message: "Login successful"})
}

// HandleLogout handles the POST /users/logout endpoint.
//
//	@Summary		Logout
//	@Description	Invalidate the current user's session.
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	dto.LogoutResponse
//	@Failure		401	{object}	dto.ErrorResponse
//	@Failure		500	{object}	dto.ErrorResponse
//	@Security		BearerAuth
//	@Router			/users/logout [post]
func (h *UserHandler) HandleLogout(c echo.Context) error {
	userID, ok := c.Get("userID").(int)
	if !ok {
		return c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "Unauthorized"})
	}

	if err := h.loginUseCase.Logout(c.Request().Context(), userID); err != nil {
		h.log.Error(c.Request().Context(), "logout failed", "userID", userID, "error", err)
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "Internal server error"})
	}

	return c.JSON(http.StatusOK, dto.LogoutResponse{
		Message: "Logged out successfully",
	})
}

// HandleRegister handles the POST /users/register endpoint.
//
//	@Summary		Register
//	@Description	Create a new user account.
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Param			request	body		dto.RegisterRequest	true	"User registration details"
//	@Success		201		{object}	dto.RegisterResponse
//	@Failure		400		{object}	dto.ErrorResponse
//	@Failure		500		{object}	dto.ErrorResponse
//	@Router			/users/register [post]
func (h *UserHandler) HandleRegister(c echo.Context) error {
	var req dto.RegisterRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "Invalid request body"})
	}

	if err := h.validator.Validate(req); err != nil {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
	}

	user := &domain.User{
		FirstName: strings.TrimSpace(req.FirstName),
		LastName:  strings.TrimSpace(req.LastName),
		Email:     strings.ToLower(strings.TrimSpace(req.Email)),
		Password:  req.Password,
	}

	if err := h.userUseCase.Register(c.Request().Context(), user); err != nil {
		if errors.Is(err, domain.ErrEmailAlreadyRegistered) {
			return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "Email is already registered"})
		}
		if errors.Is(err, domain.ErrValidationFailed) {
			return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		}
		h.log.Error(c.Request().Context(), "registration failed", "email", req.Email, "error", err)
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "Something went wrong"})
	}

	return c.JSON(http.StatusCreated, dto.RegisterResponse{
		Message: "User registered successfully",
	})
}

// HandleFollow handles the POST /users/{userId}/follow endpoint.
//
//	@Summary		Follow a user
//	@Description	Allows an authenticated user to follow another user.
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Param			userId	path		int	true	"User ID to follow"
//	@Success		200		{object}	dto.FollowResponse
//	@Failure		400		{object}	dto.ErrorResponse
//	@Failure		401		{object}	dto.ErrorResponse
//	@Failure		500		{object}	dto.ErrorResponse
//	@Security		BearerAuth
//	@Router			/users/{userId}/follow [post]
func (h *UserHandler) HandleFollow(c echo.Context) error {
	followerID, ok := c.Get("userID").(int)
	if !ok {
		return c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "Unauthorized"})
	}

	followedID, err := strconv.Atoi(c.Param("userId"))
	if err != nil || followedID <= 0 {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "Invalid user ID"})
	}

	follow, err := h.followUseCase.Follow(c.Request().Context(), followerID, followedID)
	if err != nil {
		if errors.Is(err, domain.ErrCannotFollowSelf) || errors.Is(err, domain.ErrAlreadyFollowing) {
			return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		}
		if errors.Is(err, domain.ErrUserNotFound) {
			return c.JSON(http.StatusNotFound, dto.ErrorResponse{Error: "User not found"})
		}
		h.log.Error(c.Request().Context(), "follow failed", "followerID", followerID, "followedID", followedID, "error", err)
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "Something went wrong"})
	}

	return c.JSON(http.StatusOK, dto.FollowResponse{
		Message: "You are now following this user.",
		Data: dto.FollowData{
			FollowedUserID: follow.FollowedID,
			FollowedAt:     follow.CreatedAt.Format(time.RFC3339),
		},
	})
}

// GetFollowing handles the GET /users/me/following endpoint.
//
//	@Summary		Get following list
//	@Description	Get the list of users being followed by the currently authenticated user.
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Param			limit	query		int	false	"Limit for pagination"
//	@Param			offset	query		int	false	"Offset for pagination"
//	@Success		200		{array}		dto.FollowingResponse
//	@Failure		400		{object}	dto.ErrorResponse
//	@Failure		401		{object}	dto.ErrorResponse
//	@Failure		500		{object}	dto.ErrorResponse
//	@Security		BearerAuth
//	@Router			/users/me/following [get]
func (h *UserHandler) GetFollowing(c echo.Context) error {
	followerID, ok := c.Get("userID").(int)
	if !ok {
		return c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "Unauthorized"})
	}

	var req dto.GetFollowingRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "Invalid query parameters"})
	}

	if err := h.validator.Validate(req); err != nil {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
	}

	following, err := h.followUseCase.GetFollowing(c.Request().Context(), followerID, req.Limit, req.Offset)
	if err != nil {
		h.log.Error(c.Request().Context(), "get following failed", "followerID", followerID, "error", err)
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "Something went wrong"})
	}

	res := make([]dto.FollowingResponse, 0, len(following))
	for _, f := range following {
		res = append(res, toFollowingResponse(f))
	}

	return c.JSON(http.StatusOK, res)
}

// GetFollowers handles the GET /users/me/followers endpoint.
//
//	@Summary		Get followers list
//	@Description	Get the list of users following the currently authenticated user.
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Param			limit	query		int	false	"Limit for pagination"
//	@Param			offset	query		int	false	"Offset for pagination"
//	@Success		200		{array}		dto.FollowerResponse
//	@Failure		400		{object}	dto.ErrorResponse
//	@Failure		401		{object}	dto.ErrorResponse
//	@Failure		500		{object}	dto.ErrorResponse
//	@Security		BearerAuth
//	@Router			/users/me/followers [get]
func (h *UserHandler) GetFollowers(c echo.Context) error {
	followedID, ok := c.Get("userID").(int)
	if !ok {
		return c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "Unauthorized"})
	}

	var req dto.GetFollowersRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "Invalid query parameters"})
	}

	if err := h.validator.Validate(req); err != nil {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
	}

	followers, err := h.followUseCase.GetFollowers(c.Request().Context(), followedID, req.Limit, req.Offset)
	if err != nil {
		h.log.Error(c.Request().Context(), "get followers failed", "followedID", followedID, "error", err)
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "Something went wrong"})
	}

	res := make([]dto.FollowerResponse, 0, len(followers))
	for _, f := range followers {
		res = append(res, toFollowerResponse(f))
	}

	return c.JSON(http.StatusOK, res)
}

func toFollowingResponse(f domain.Following) dto.FollowingResponse {
	return dto.FollowingResponse{
		ID:         f.ID,
		Name:       f.FirstName + " " + f.LastName,
		FollowedAt: f.FollowedAt.Format(time.RFC3339),
		IsMutual:   f.IsMutual,
	}
}

func toFollowerResponse(f domain.Follower) dto.FollowerResponse {
	return dto.FollowerResponse{
		ID:         f.ID,
		Name:       f.FirstName + " " + f.LastName,
		FollowedAt: f.FollowedAt.Format(time.RFC3339),
		IsMutual:   f.IsMutual,
	}
}
