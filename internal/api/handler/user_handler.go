package handler

import (
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
	postUseCase   ports.PostUseCase
	validator     ports.Validator
	log           ports.Logger
}

// NewUserHandler creates a new instance of UserHandler.
func NewUserHandler(
	userUseCase ports.UserUseCase,
	loginUseCase ports.LoginUseCase,
	followUseCase ports.FollowUseCase,
	postUseCase ports.PostUseCase,
	validator ports.Validator,
	log ports.Logger,
) *UserHandler {
	// ponytail: nil checks removed — Go panics at method call site on nil pointer
	return &UserHandler{
		userUseCase:   userUseCase,
		loginUseCase:  loginUseCase,
		followUseCase: followUseCase,
		postUseCase:   postUseCase,
		validator:     validator,
		log:           log,
	}
}

// GetUser handles the GET /users/:username endpoint.
//
//	@Summary		Get user
//	@Description	Get a user by their username.
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Param			username	path		string	true	"Username"
//	@Success		200			{object}	dto.UserResponse
//	@Failure		400			{object}	dto.ProblemDetail
//	@Failure		404			{object}	dto.ProblemDetail
//	@Failure		500			{object}	dto.ProblemDetail
//	@Router			/users/{username} [get]
func (h *UserHandler) GetUser(c echo.Context) error {
	username := c.Param("username")
	if username == "" {
		h.log.Error(c.Request().Context(), "GetUser failed", "error", "username parameter is empty")
		return echo.ErrUnauthorized
	}

	user, err := h.userUseCase.GetUser(c.Request().Context(), username)
	if err != nil {
		h.log.Error(c.Request().Context(), "GetUser failed", "username", username, "error", err)
		return err
	}

	h.log.Info(c.Request().Context(), "GetUser succeeded", "username", username)
	return c.JSON(http.StatusOK, toUserResponse(user))
}

// HandleLogin handles the POST /auth/login endpoint.
//
//	@Summary		Login
//	@Description	Authenticate a user and return access and refresh tokens via HttpOnly cookies.
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			LoginRequest	body		dto.LoginRequest	true	"Login credentials"
//	@Success		200				{object}	dto.LoginResponse
//	@Failure		400				{object}	dto.ProblemDetail
//	@Failure		401				{object}	dto.ProblemDetail
//	@Failure		500				{object}	dto.ProblemDetail
//	@Router			/auth/login [post]
func (h *UserHandler) HandleLogin(c echo.Context) error {
	var req dto.LoginRequest
	if err := c.Bind(&req); err != nil {
		h.log.Error(c.Request().Context(), "HandleLogin failed", "error", "Invalid request body")
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	if err := h.validator.Validate(req); err != nil {
		h.log.Error(c.Request().Context(), "HandleLogin failed", "validation_error", err)
		return err
	}

	accessToken, err := h.loginUseCase.Login(c.Request().Context(), req.Email, req.Password)
	if err != nil {
		h.log.Error(c.Request().Context(), "HandleLogin failed", "email", req.Email, "error", err)
		return err
	}

	// Set access token cookie
	c.SetCookie(&http.Cookie{
		Name:     "access_token",
		Value:    accessToken.Token,
		Path:     "/",
		HttpOnly: true,
		Secure:   false, // Set to true in production
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int(time.Until(accessToken.ExpiresAt).Seconds()),
	})

	c.SetCookie(&http.Cookie{
		Name:     "username",
		Value:    accessToken.Username,
		Path:     "/",
		HttpOnly: false, // ponytail: readable by JS for client-side redirect guard
		Secure:   false, // Set to true in production
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int(time.Until(accessToken.ExpiresAt).Seconds()),
	})

	h.log.Info(c.Request().Context(), "HandleLogin succeeded", "email", req.Email)
	return c.JSON(http.StatusOK, dto.LoginResponse{
		Message:  "Login successful",
		Username: accessToken.Username,
	})
}

// HandleLogout handles the POST /auth/logout endpoint.
//
//	@Summary		Logout
//	@Description	Invalidate the current user's session.
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	dto.LogoutResponse
//	@Failure		401	{object}	dto.ProblemDetail
//	@Failure		500	{object}	dto.ProblemDetail
//	@Security		BearerAuth
//	@Router			/auth/logout [post]
func (h *UserHandler) HandleLogout(c echo.Context) error {
	username, ok := c.Get("username").(string)
	if !ok {
		h.log.Error(c.Request().Context(), "HandleLogout failed", "error", "Username not found in context")
		return echo.ErrUnauthorized
	}

	if err := h.loginUseCase.Logout(c.Request().Context(), username); err != nil {
		h.log.Error(c.Request().Context(), "HandleLogout failed", "username", username, "error", err)
		return err
	}

	// Clear auth cookies so the client cannot access protected routes after logout.
	c.SetCookie(&http.Cookie{
		Name:     "access_token",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	})
	c.SetCookie(&http.Cookie{
		Name:     "username",
		Value:    "",
		Path:     "/",
		HttpOnly: false,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	})

	h.log.Info(c.Request().Context(), "HandleLogout succeeded", "username", username)
	return c.JSON(http.StatusOK, dto.LogoutResponse{
		Message: "Logged out successfully",
	})
}

// HandleRegister handles the POST /auth/register endpoint.
//
//	@Summary		Register
//	@Description	Create a new user account.
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			request	body		dto.RegisterRequest	true	"User registration details"
//	@Success		201		{object}	dto.RegisterResponse
//	@Failure		400		{object}	dto.ProblemDetail
//	@Failure		500		{object}	dto.ProblemDetail
//	@Router			/auth/register [post]
func (h *UserHandler) HandleRegister(c echo.Context) error {
	var req dto.RegisterRequest
	if err := c.Bind(&req); err != nil {
		h.log.Error(c.Request().Context(), "HandleRegister failed", "error", "Invalid request body")
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	if err := h.validator.Validate(req); err != nil {
		h.log.Error(c.Request().Context(), "HandleRegister failed", "validation_error", err)
		return err
	}

	user := &domain.User{
		FirstName: strings.TrimSpace(req.FirstName),
		LastName:  strings.TrimSpace(req.LastName),
		Username:  strings.ToLower(strings.TrimSpace(req.Username)),
		Email:     strings.ToLower(strings.TrimSpace(req.Email)),
		Password:  req.Password,
	}

	if err := h.userUseCase.Register(c.Request().Context(), user); err != nil {
		h.log.Error(c.Request().Context(), "HandleRegister failed", "email", req.Email, "error", err)
		return err
	}

	h.log.Info(c.Request().Context(), "HandleRegister succeeded", "email", req.Email)
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
//	@Failure		400			{object}	dto.ProblemDetail
//	@Failure		401			{object}	dto.ProblemDetail
//	@Failure		404			{object}	dto.ProblemDetail
//	@Failure		500			{object}	dto.ProblemDetail
//	@Security		BearerAuth
//	@Router			/users/{username}/followers [post]
func (h *UserHandler) HandleFollow(c echo.Context) error {
	followerUsername, ok := c.Get("username").(string)
	if !ok {
		h.log.Error(c.Request().Context(), "HandleFollow failed", "error", "Username not found in context")
		return echo.ErrUnauthorized
	}

	followedUsername := c.Param("username")
	if followedUsername == "" {
		h.log.Error(c.Request().Context(), "HandleFollow failed", "error", "Username parameter is empty")
		return echo.ErrBadRequest
	}

	follow, err := h.followUseCase.Follow(c.Request().Context(), followerUsername, followedUsername)
	if err != nil {
		h.log.Error(c.Request().Context(), "HandleFollow failed", "follower", followerUsername, "followed", followedUsername, "error", err)
		return err
	}

	h.log.Info(c.Request().Context(), "HandleFollow succeeded", "follower", followerUsername, "followed", followedUsername)
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
//	@Failure		400			{object}	dto.ProblemDetail
//	@Failure		401			{object}	dto.ProblemDetail
//	@Failure		500			{object}	dto.ProblemDetail
//	@Security		BearerAuth
//	@Router			/users/{username}/followers [delete]
func (h *UserHandler) HandleUnfollow(c echo.Context) error {
	followerUsername, ok := c.Get("username").(string)
	if !ok {
		h.log.Error(c.Request().Context(), "HandleUnfollow failed", "error", "Username not found in context")
		return echo.ErrUnauthorized
	}

	followedUsername := c.Param("username")
	if followedUsername == "" {
		h.log.Error(c.Request().Context(), "HandleUnfollow failed", "error", "Username parameter is empty")
		return echo.ErrBadRequest
	}

	err := h.followUseCase.Unfollow(c.Request().Context(), followerUsername, followedUsername)
	if err != nil {
		h.log.Error(c.Request().Context(), "HandleUnfollow failed", "follower", followerUsername, "followed", followedUsername, "error", err)
		return err
	}

	h.log.Info(c.Request().Context(), "HandleUnfollow succeeded", "follower", followerUsername, "followed", followedUsername)
	return c.JSON(http.StatusOK, dto.UnfollowResponse{
		Message: "Successfully unfollowed this user.",
	})
}

// GetFollowing handles the GET /users/:username/following endpoint.
//
//	@Summary		Get following list
//	@Description	Get the list of users being followed by the currently authenticated user.
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Param			limit	query		int	false	"Limit for pagination"
//	@Param			offset	query		int	false	"Offset for pagination"
//	@Success		200		{array}		dto.FollowingResponse
//	@Failure		400		{object}	dto.ProblemDetail
//	@Failure		401		{object}	dto.ProblemDetail
//	@Failure		500		{object}	dto.ProblemDetail
//	@Security		BearerAuth
//	@Router			/users/{username}/following [get]
func (h *UserHandler) GetFollowing(c echo.Context) error {
	followerUsername := c.Param("username")
	if followerUsername == "" {
		h.log.Error(c.Request().Context(), "GetFollowing failed", "error", "Username parameter is empty")
		return echo.ErrBadRequest
	}

	var req dto.GetFollowingRequest
	if err := c.Bind(&req); err != nil {
		h.log.Error(c.Request().Context(), "GetFollowing failed", "error", "Invalid query parameters")
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid query parameters")
	}

	if err := h.validator.Validate(req); err != nil {
		h.log.Error(c.Request().Context(), "GetFollowing failed", "validation_error", err)
		return err
	}

	following, err := h.followUseCase.GetFollowing(c.Request().Context(), followerUsername, req.Limit, req.Offset)
	if err != nil {
		h.log.Error(c.Request().Context(), "GetFollowing failed", "follower", followerUsername, "error", err)
		return err
	}

	res := make([]dto.FollowingResponse, 0, len(following))
	for _, f := range following {
		res = append(res, toFollowingResponse(f))
	}

	h.log.Info(c.Request().Context(), "GetFollowing succeeded", "follower", followerUsername, "count", len(res))
	return c.JSON(http.StatusOK, res)
}

// GetFollowers handles the GET /users/:username/followers endpoint.
//
//	@Summary		Get followers list
//	@Description	Get the list of users following the currently authenticated user.
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Param			limit	query		int	false	"Limit for pagination"
//	@Param			offset	query		int	false	"Offset for pagination"
//	@Success		200		{array}		dto.FollowerResponse
//	@Failure		400		{object}	dto.ProblemDetail
//	@Failure		401		{object}	dto.ProblemDetail
//	@Failure		500		{object}	dto.ProblemDetail
//	@Security		BearerAuth
//	@Router			/users/{username}/followers [get]
func (h *UserHandler) GetFollowers(c echo.Context) error {
	followedUsername := c.Param("username")
	if followedUsername == "" {
		h.log.Error(c.Request().Context(), "GetFollowers failed", "error", "Username parameter is empty")
		return echo.ErrBadRequest
	}

	var req dto.GetFollowersRequest
	if err := c.Bind(&req); err != nil {
		h.log.Error(c.Request().Context(), "GetFollowers failed", "error", "Invalid query parameters")
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid query parameters")
	}

	if err := h.validator.Validate(req); err != nil {
		h.log.Error(c.Request().Context(), "GetFollowers failed", "validation_error", err)
		return err
	}

	followers, err := h.followUseCase.GetFollowers(c.Request().Context(), followedUsername, req.Limit, req.Offset)
	if err != nil {
		h.log.Error(c.Request().Context(), "GetFollowers failed", "followed", followedUsername, "error", err)
		return err
	}

	res := make([]dto.FollowerResponse, 0, len(followers))
	for _, f := range followers {
		res = append(res, toFollowerResponse(f))
	}

	h.log.Info(c.Request().Context(), "GetFollowers succeeded", "followed", followedUsername, "count", len(res))
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

// GetUserPosts handles the GET /users/:username/posts endpoint.
//
//	@Summary		Get user posts by username
//	@Description	Retrieve all posts for a specific user by username.
//	@Tags			users
//	@Produce		json
//	@Param			username	path		string	true	"Username"
//	@Param			limit		query		int		false	"Limit"
//	@Param			offset		query		int		false	"Offset"
//	@Success		200			{array}		dto.PostResponse
//	@Failure		400			{object}	dto.ProblemDetail
//	@Failure		500			{object}	dto.ProblemDetail
//	@Router			/users/{username}/posts [get]
func (h *UserHandler) GetUserPosts(c echo.Context) error {
	username := c.Param("username")

	if username == "" {
		h.log.Error(c.Request().Context(), "GetUserPosts failed", "error", "Username parameter is empty")
		return echo.ErrBadRequest
	}

	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	offset, _ := strconv.Atoi(c.QueryParam("offset"))

	if limit == 0 {
		limit = 10 // default limit
	}

	posts, err := h.postUseCase.GetPosts(c.Request().Context(), username, limit, offset)
	if err != nil {
		h.log.Error(c.Request().Context(), "GetUserPosts failed", "username", username, "error", err)
		return err
	}

	response := make([]dto.PostResponse, 0, len(posts))
	for _, p := range posts {
		response = append(response, dto.PostResponse{
			ID:        p.ID,
			Title:     p.Title,
			Content:   p.Content,
			Tags:      p.Tags,
			Author:    username,
			CreatedAt: p.CreatedAt,
			UpdatedAt: p.UpdatedAt,
		})
	}

	h.log.Info(c.Request().Context(), "GetUserPosts succeeded", "username", username, "count", len(response))
	return c.JSON(http.StatusOK, response)
}

// HandleChangePassword handles the PUT /users/password endpoint.
//
//	@Summary		Change password
//	@Description	Verify old password and update to new password.
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Param			request	body		dto.ChangePasswordRequest	true	"Change password request details"
//	@Success		200		{object}	dto.MessageResponse
//	@Failure		400		{object}	dto.ProblemDetail
//	@Failure		401		{object}	dto.ProblemDetail
//	@Failure		500		{object}	dto.ProblemDetail
//	@Security		BearerAuth
//	@Router			/users/password [put]
func (h *UserHandler) HandleChangePassword(c echo.Context) error {
	username, ok := c.Get("username").(string)
	if !ok {
		h.log.Error(c.Request().Context(), "HandleChangePassword failed", "error", "Username not found in context")
		return echo.ErrUnauthorized
	}

	var req dto.ChangePasswordRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	if err := h.validator.Validate(req); err != nil {
		h.log.Error(c.Request().Context(), "HandleChangePassword failed", "username", username, "validation_error", err)
		return err
	}

	err := h.userUseCase.ChangePassword(c.Request().Context(), username, req.OldPassword, req.NewPassword)
	if err != nil {
		h.log.Error(c.Request().Context(), "HandleChangePassword failed", "username", username, "error", err)
		return err
	}

	h.log.Info(c.Request().Context(), "HandleChangePassword succeeded", "username", username)
	return c.JSON(http.StatusOK, dto.MessageResponse{Message: "Password updated successfully"})
}

// HandleUpdateProfile handles the PUT /users/profile endpoint.
//
//	@Summary		Update user profile
//	@Description	Update the authenticated user's first name, last name, and username.
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Param			request	body		dto.UpdateProfileRequest	true	"Updated profile fields"
//	@Success		200		{object}	dto.UpdateProfileResponse
//	@Failure		400		{object}	dto.ProblemDetail
//	@Failure		401		{object}	dto.ProblemDetail
//	@Failure		500		{object}	dto.ProblemDetail
//	@Security		BearerAuth
//	@Router			/users/profile [put]
func (h *UserHandler) HandleUpdateProfile(c echo.Context) error {
	username, ok := c.Get("username").(string)
	if !ok {
		h.log.Error(c.Request().Context(), "HandleUpdateProfile failed", "error", "Username not found in context")
		return echo.ErrUnauthorized
	}

	var req dto.UpdateProfileRequest
	if err := c.Bind(&req); err != nil {
		h.log.Error(c.Request().Context(), "HandleUpdateProfile failed", "username", username, "error", "Failed to bind request body")
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	if err := h.validator.Validate(req); err != nil {
		h.log.Error(c.Request().Context(), "HandleUpdateProfile failed", "username", username, "validation_error", err)
		return err
	}

	user := &domain.User{
		FirstName: strings.TrimSpace(req.FirstName),
		LastName:  strings.TrimSpace(req.LastName),
		Username:  strings.ToLower(strings.TrimSpace(req.Username)),
	}

	if err := h.userUseCase.UpdateProfile(c.Request().Context(), username, user); err != nil {
		h.log.Error(c.Request().Context(), "HandleUpdateProfile failed", "username", username, "error", err)
		return err
	}

	h.log.Info(c.Request().Context(), "HandleUpdateProfile succeeded", "username", username)
	return c.JSON(http.StatusOK, dto.UpdateProfileResponse{
		Message:  "Profile updated successfully",
		Username: user.Username,
	})
}
