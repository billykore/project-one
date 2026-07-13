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
//	@Failure		400			{object}	dto.APIErrorResponse
//	@Failure		404			{object}	dto.APIErrorResponse
//	@Failure		500			{object}	dto.APIErrorResponse
//	@Router			/users/{username} [get]
func (h *UserHandler) GetUser(c echo.Context) error {
	username := c.Param("username")
	if username == "" {
		return domain.ErrValidationFailed
	}

	user, err := h.userUseCase.GetUser(c.Request().Context(), username)
	if err != nil {
		return err
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
//	@Failure		400				{object}	dto.APIErrorResponse
//	@Failure		401				{object}	dto.APIErrorResponse
//	@Failure		500				{object}	dto.APIErrorResponse
//	@Router			/auth/login [post]
func (h *UserHandler) HandleLogin(c echo.Context) error {
	var req dto.LoginRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	if err := h.validator.Validate(req); err != nil {
		return err
	}

	accessToken, err := h.loginUseCase.Login(c.Request().Context(), req.Email, req.Password)
	if err != nil {
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

	return c.JSON(http.StatusOK, dto.LoginResponse{
		Message:  "Login successful",
		Username: accessToken.Username,
	})
}

// HandleLogout handles the POST /users/logout endpoint.
//
//	@Summary		Logout
//	@Description	Invalidate the current user's session.
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	dto.LogoutResponse
//	@Failure		401	{object}	dto.APIErrorResponse
//	@Failure		500	{object}	dto.APIErrorResponse
//	@Security		BearerAuth
//	@Router			/auth/logout [post]
func (h *UserHandler) HandleLogout(c echo.Context) error {
	username, ok := c.Get("username").(string)
	if !ok {
		return domain.ErrUnauthorized
	}

	if err := h.loginUseCase.Logout(c.Request().Context(), username); err != nil {
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
//	@Failure		400		{object}	dto.APIErrorResponse
//	@Failure		500		{object}	dto.APIErrorResponse
//	@Router			/auth/register [post]
func (h *UserHandler) HandleRegister(c echo.Context) error {
	var req dto.RegisterRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	if err := h.validator.Validate(req); err != nil {
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
		return err
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
//	@Failure		400			{object}	dto.APIErrorResponse
//	@Failure		401			{object}	dto.APIErrorResponse
//	@Failure		404			{object}	dto.APIErrorResponse
//	@Failure		500			{object}	dto.APIErrorResponse
//	@Security		BearerAuth
//	@Router			/users/{username}/followers [post]
func (h *UserHandler) HandleFollow(c echo.Context) error {
	followerUsername, ok := c.Get("username").(string)
	if !ok {
		return domain.ErrUnauthorized
	}

	followedUsername := c.Param("username")
	if followedUsername == "" {
		return domain.ErrValidationFailed
	}

	follow, err := h.followUseCase.Follow(c.Request().Context(), followerUsername, followedUsername)
	if err != nil {
		return err
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
//	@Failure		400			{object}	dto.APIErrorResponse
//	@Failure		401			{object}	dto.APIErrorResponse
//	@Failure		500			{object}	dto.APIErrorResponse
//	@Security		BearerAuth
//	@Router			/users/{username}/followers [delete]
func (h *UserHandler) HandleUnfollow(c echo.Context) error {
	followerUsername, ok := c.Get("username").(string)
	if !ok {
		return domain.ErrUnauthorized
	}

	followedUsername := c.Param("username")
	if followedUsername == "" {
		return domain.ErrValidationFailed
	}

	err := h.followUseCase.Unfollow(c.Request().Context(), followerUsername, followedUsername)
	if err != nil {
		return err
	}

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
//	@Failure		400		{object}	dto.APIErrorResponse
//	@Failure		401		{object}	dto.APIErrorResponse
//	@Failure		500		{object}	dto.APIErrorResponse
//	@Security		BearerAuth
//	@Router			/users/{username}/following [get]
func (h *UserHandler) GetFollowing(c echo.Context) error {
	followerUsername := c.Param("username")

	var req dto.GetFollowingRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid query parameters")
	}

	if err := h.validator.Validate(req); err != nil {
		return err
	}

	following, err := h.followUseCase.GetFollowing(c.Request().Context(), followerUsername, req.Limit, req.Offset)
	if err != nil {
		return err
	}

	res := make([]dto.FollowingResponse, 0, len(following))
	for _, f := range following {
		res = append(res, toFollowingResponse(f))
	}

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
//	@Failure		400		{object}	dto.APIErrorResponse
//	@Failure		401		{object}	dto.APIErrorResponse
//	@Failure		500		{object}	dto.APIErrorResponse
//	@Security		BearerAuth
//	@Router			/users/{username}/followers [get]
func (h *UserHandler) GetFollowers(c echo.Context) error {
	followedUsername := c.Param("username")

	var req dto.GetFollowersRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid query parameters")
	}

	if err := h.validator.Validate(req); err != nil {
		return err
	}

	followers, err := h.followUseCase.GetFollowers(c.Request().Context(), followedUsername, req.Limit, req.Offset)
	if err != nil {
		return err
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
//	@Failure		400			{object}	dto.APIErrorResponse
//	@Failure		500			{object}	dto.APIErrorResponse
//	@Router			/users/{username}/posts [get]
func (h *UserHandler) GetUserPosts(c echo.Context) error {
	username := c.Param("username")

	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	offset, _ := strconv.Atoi(c.QueryParam("offset"))

	if limit == 0 {
		limit = 10 // default limit
	}

	posts, err := h.postUseCase.GetPosts(c.Request().Context(), username, limit, offset)
	if err != nil {
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
//	@Failure		400		{object}	dto.APIErrorResponse
//	@Failure		401		{object}	dto.APIErrorResponse
//	@Failure		500		{object}	dto.APIErrorResponse
//	@Security		BearerAuth
//	@Router			/users/password [put]
func (h *UserHandler) HandleChangePassword(c echo.Context) error {
	username, ok := c.Get("username").(string)
	if !ok {
		return domain.ErrUnauthorized
	}

	var req dto.ChangePasswordRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	if err := h.validator.Validate(req); err != nil {
		return err
	}

	err := h.userUseCase.ChangePassword(c.Request().Context(), username, req.OldPassword, req.NewPassword)
	if err != nil {
		return err
	}

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
//	@Failure		400		{object}	dto.APIErrorResponse
//	@Failure		401		{object}	dto.APIErrorResponse
//	@Failure		500		{object}	dto.APIErrorResponse
//	@Security		BearerAuth
//	@Router			/users/profile [put]
func (h *UserHandler) HandleUpdateProfile(c echo.Context) error {
	username, ok := c.Get("username").(string)
	if !ok {
		return domain.ErrUnauthorized
	}

	var req dto.UpdateProfileRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	if err := h.validator.Validate(req); err != nil {
		return err
	}

	user := &domain.User{
		FirstName: strings.TrimSpace(req.FirstName),
		LastName:  strings.TrimSpace(req.LastName),
		Username:  strings.ToLower(strings.TrimSpace(req.Username)),
	}

	if err := h.userUseCase.UpdateProfile(c.Request().Context(), username, user); err != nil {
		return err
	}

	return c.JSON(http.StatusOK, dto.UpdateProfileResponse{
		Message:  "Profile updated successfully",
		Username: user.Username,
	})
}
