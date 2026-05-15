package handler

import (
	"errors"
	"fmt"
	"net/http"
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
	username, ok := c.Get("username").(string)
	if !ok {
		return c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "Unauthorized"})
	}

	user, err := h.userUseCase.GetCurrentUser(c.Request().Context(), username)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) || errors.Is(err, domain.ErrUnauthorized) {
			return c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "Unauthorized"})
		}
		h.log.Error(c.Request().Context(), "failed to get current user", "username", username, "error", err)
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "Internal server error"})
	}

	return c.JSON(http.StatusOK, toUserResponse(user))
}

// GetProfile handles the GET /users/:username endpoint.
//
//	@Summary		Get user profile
//	@Description	Get the profile of a user by their username.
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Param			username	path		string	true	"Username"
//	@Success		200			{object}	dto.UserResponse
//	@Failure		400			{object}	dto.ErrorResponse
//	@Failure		404			{object}	dto.ErrorResponse
//	@Failure		500			{object}	dto.ErrorResponse
//	@Router			/users/{username} [get]
func (h *UserHandler) GetProfile(c echo.Context) error {
	username := c.Param("username")
	if username == "" {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "Invalid username"})
	}

	user, err := h.userUseCase.GetUserProfile(c.Request().Context(), username)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return c.JSON(http.StatusNotFound, dto.ErrorResponse{Error: fmt.Sprintf("User %s not found", username)})
		}
		h.log.Error(c.Request().Context(), "failed to get user profile", "username", username, "error", err)
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "Internal server error"})
	}

	return c.JSON(http.StatusOK, toUserResponse(user))
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
	username, ok := c.Get("username").(string)
	if !ok {
		return c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "Unauthorized"})
	}

	if err := h.loginUseCase.Logout(c.Request().Context(), username); err != nil {
		h.log.Error(c.Request().Context(), "logout failed", "username", username, "error", err)
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
		Username:  strings.ToLower(strings.TrimSpace(req.Username)),
		Email:     strings.ToLower(strings.TrimSpace(req.Email)),
		Password:  req.Password,
	}

	if err := h.userUseCase.Register(c.Request().Context(), user); err != nil {
		if errors.Is(err, domain.ErrEmailAlreadyRegistered) {
			return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "Email is already registered"})
		}
		if errors.Is(err, domain.ErrUsernameAlreadyTaken) {
			return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "Username is already taken"})
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

// HandleFollow handles the POST /users/{username}/follow endpoint.
//
//	@Summary		Follow a user
//	@Description	Allows an authenticated user to follow another user.
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Param			username	path		string	true	"Username to follow"
//	@Success		200			{object}	dto.FollowResponse
//	@Failure		400			{object}	dto.ErrorResponse
//	@Failure		401			{object}	dto.ErrorResponse
//	@Failure		500			{object}	dto.ErrorResponse
//	@Security		BearerAuth
//	@Router			/users/{username}/follow [post]
func (h *UserHandler) HandleFollow(c echo.Context) error {
	followerUsername, ok := c.Get("username").(string)
	if !ok {
		return c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "Unauthorized"})
	}

	followedUsername := c.Param("username")
	if followedUsername == "" {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "Invalid username"})
	}

	follow, err := h.followUseCase.Follow(c.Request().Context(), followerUsername, followedUsername)
	if err != nil {
		if errors.Is(err, domain.ErrCannotFollowSelf) {
			return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: domain.ErrCannotFollowSelf.Error()})
		}
		if errors.Is(err, domain.ErrAlreadyFollowing) {
			return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: domain.ErrAlreadyFollowing.Error()})
		}
		if errors.Is(err, domain.ErrUserNotFound) {
			return c.JSON(http.StatusNotFound, dto.ErrorResponse{Error: "User not found"})
		}
		h.log.Error(c.Request().Context(), "follow failed", "followerUsername", followerUsername, "followedUsername", followedUsername, "error", err)
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "Something went wrong"})
	}

	return c.JSON(http.StatusOK, dto.FollowResponse{
		Message: "You are now following this user.",
		Data: dto.FollowData{
			FollowedUsername: follow.FollowedUsername,
			FollowedAt:       follow.CreatedAt.Format(time.RFC3339),
		},
	})
}

// HandleUnfollow handles the DELETE /users/{username}/follow endpoint.
//
//	@Summary		Unfollow a user
//	@Description	Allows an authenticated user to unfollow another user.
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Param			username	path		string	true	"Username to unfollow"
//	@Success		200			{object}	dto.UnfollowResponse
//	@Failure		400			{object}	dto.ErrorResponse
//	@Failure		401			{object}	dto.ErrorResponse
//	@Failure		500			{object}	dto.ErrorResponse
//	@Security		BearerAuth
//	@Router			/users/{username}/follow [delete]
func (h *UserHandler) HandleUnfollow(c echo.Context) error {
	followerUsername, ok := c.Get("username").(string)
	if !ok {
		return c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "Unauthorized"})
	}

	followedUsername := c.Param("username")
	if followedUsername == "" {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "Invalid username"})
	}

	err := h.followUseCase.Unfollow(c.Request().Context(), followerUsername, followedUsername)
	if err != nil {
		if errors.Is(err, domain.ErrCannotUnfollowSelf) {
			return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: domain.ErrCannotUnfollowSelf.Error()})
		}
		if errors.Is(err, domain.ErrNotFollowing) {
			return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: domain.ErrNotFollowing.Error()})
		}
		h.log.Error(c.Request().Context(), "unfollow failed", "followerUsername", followerUsername, "followedUsername", followedUsername, "error", err)
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "Something went wrong"})
	}

	return c.JSON(http.StatusOK, dto.UnfollowResponse{
		Message: "Successfully unfollowed this user.",
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
	followerUsername, ok := c.Get("username").(string)
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

	following, err := h.followUseCase.GetFollowing(c.Request().Context(), followerUsername, req.Limit, req.Offset)
	if err != nil {
		h.log.Error(c.Request().Context(), "get following failed", "followerUsername", followerUsername, "error", err)
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
	followedUsername, ok := c.Get("username").(string)
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

	followers, err := h.followUseCase.GetFollowers(c.Request().Context(), followedUsername, req.Limit, req.Offset)
	if err != nil {
		h.log.Error(c.Request().Context(), "get followers failed", "followedUsername", followedUsername, "error", err)
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
		Username:   f.Username,
		Name:       f.FirstName + " " + f.LastName,
		FollowedAt: f.FollowedAt.Format(time.RFC3339),
		IsMutual:   f.IsMutual,
	}
}

func toFollowerResponse(f domain.Follower) dto.FollowerResponse {
	return dto.FollowerResponse{
		Username:   f.Username,
		Name:       f.FirstName + " " + f.LastName,
		FollowedAt: f.FollowedAt.Format(time.RFC3339),
		IsMutual:   f.IsMutual,
	}
}

func toUserResponse(user *domain.User) dto.UserResponse {
	return dto.UserResponse{
		Username: user.Username,
		Email:    user.Email,
		Name:     user.FirstName + " " + user.LastName,
	}
}
