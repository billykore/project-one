package handler

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/billykore/project-one/internal/identity/api/dto"
	identitydomain "github.com/billykore/project-one/internal/identity/domain"
	identityports "github.com/billykore/project-one/internal/identity/ports"
	request "github.com/billykore/project-one/internal/platform/api/request"
	vo "github.com/billykore/project-one/internal/platform/pagination"
	platformports "github.com/billykore/project-one/internal/platform/ports"
	"github.com/billykore/project-one/internal/platform/problem"
	"github.com/labstack/echo/v4"
)

type UserHandler struct {
	userUseCase   identityports.UserUseCase
	loginUseCase  identityports.LoginUseCase
	validator     platformports.Validator
	log           platformports.Logger
	secureCookies bool
}

// NewUserHandler creates a new instance of UserHandler.
func NewUserHandler(
	userUseCase identityports.UserUseCase,
	loginUseCase identityports.LoginUseCase,
	validator platformports.Validator,
	log platformports.Logger,
	secureCookies bool,
) *UserHandler {
	// ponytail: nil checks removed — Go panics at method call site on nil pointer
	return &UserHandler{
		userUseCase:   userUseCase,
		loginUseCase:  loginUseCase,
		validator:     validator,
		log:           log,
		secureCookies: secureCookies,
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
		return echo.ErrBadRequest
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
		Secure:   h.secureCookies,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int(time.Until(accessToken.ExpiresAt).Seconds()),
	})

	c.SetCookie(&http.Cookie{
		Name:     "username",
		Value:    accessToken.Username,
		Path:     "/",
		HttpOnly: false, // ponytail: readable by JS for client-side redirect guard
		Secure:   h.secureCookies,
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
	user, ok := request.CurrentUser(c)
	if !ok {
		h.log.Error(c.Request().Context(), "HandleLogout failed", "error", "User not found in context")
		return echo.ErrUnauthorized
	}

	if err := h.loginUseCase.Logout(c.Request().Context(), user.ID); err != nil {
		h.log.Error(c.Request().Context(), "HandleLogout failed", "username", user.Username, "error", err)
		return err
	}

	// Clear auth cookies so the client cannot access protected routes after logout.
	c.SetCookie(&http.Cookie{
		Name:     "access_token",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   h.secureCookies,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	})
	c.SetCookie(&http.Cookie{
		Name:     "username",
		Value:    "",
		Path:     "/",
		HttpOnly: false,
		Secure:   h.secureCookies,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	})

	h.log.Info(c.Request().Context(), "HandleLogout succeeded", "username", user.Username)
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

	user := &identitydomain.User{
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

func toUserResponse(user *identitydomain.User) dto.UserResponse {
	return dto.UserResponse{
		Username: user.Username,
		Email:    user.Email,
		Name:     user.FirstName + " " + user.LastName,
	}
}

// SearchUsers handles the GET /users/search endpoint.
//
//	@Summary		Search users
//	@Description	Search users by username using prefix and fuzzy matching.
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Param			q		query		string	true	"Search query (min 3 characters)"
//	@Param			cursor	query		string	false	"Pagination cursor from previous response"
//	@Param			limit	query		int		false	"Max results (1-20, default 10)"
//	@Success		200		{object}	dto.SearchUsersResponse
//	@Failure		400		{object}	dto.ProblemDetail
//	@Failure		500		{object}	dto.ProblemDetail
//	@Router			/users/search [get]
func (h *UserHandler) SearchUsers(c echo.Context) error {
	var req dto.SearchUsersRequest
	if err := c.Bind(&req); err != nil {
		h.log.Error(c.Request().Context(), "SearchUsers failed", "error", "Invalid query parameters")
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid query parameters")
	}

	if limitStr := c.QueryParam("limit"); limitStr != "" {
		limit, err := strconv.Atoi(limitStr)
		if err != nil || limit < 1 || limit > 20 {
			return echo.ErrBadRequest
		}
		req.Limit = limit
	} else {
		req.Limit = 10
	}

	if err := h.validator.Validate(req); err != nil {
		h.log.Error(c.Request().Context(), "SearchUsers failed", "validation_error", err)
		return err
	}

	cursor, err := vo.DecodeCursor(req.Cursor)
	if err != nil && req.Cursor != "" {
		h.log.Error(c.Request().Context(), "SearchUsers failed", "error", "Invalid cursor")
		return problem.ErrInvalidCursor
	}
	var cursorPtr *vo.Cursor
	if req.Cursor != "" {
		cursorPtr = &cursor
	}

	results, nextCursor, hasMore, err := h.userUseCase.SearchUsers(c.Request().Context(), req.Q, cursorPtr, req.Limit)
	if err != nil {
		h.log.Error(c.Request().Context(), "SearchUsers failed", "query", req.Q, "error", err)
		return err
	}

	items := make([]dto.SearchUsersItem, 0, len(results))
	for _, r := range results {
		items = append(items, dto.SearchUsersItem{
			Username: r.Username,
			Name:     r.Name(),
		})
	}

	nextCursorStr := ""
	if nextCursor != nil {
		nextCursorStr = nextCursor.Encode()
	}

	h.log.Info(c.Request().Context(), "SearchUsers succeeded", "query", req.Q, "results", len(items))
	return c.JSON(http.StatusOK, dto.SearchUsersResponse{Data: items, NextCursor: nextCursorStr, HasMore: hasMore})
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
	user, ok := request.CurrentUser(c)
	if !ok {
		h.log.Error(c.Request().Context(), "HandleChangePassword failed", "error", "User not found in context")
		return echo.ErrUnauthorized
	}

	var req dto.ChangePasswordRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	if err := h.validator.Validate(req); err != nil {
		h.log.Error(c.Request().Context(), "HandleChangePassword failed", "username", user.Username, "validation_error", err)
		return err
	}

	err := h.userUseCase.ChangePassword(c.Request().Context(), user.ID, req.OldPassword, req.NewPassword)
	if err != nil {
		h.log.Error(c.Request().Context(), "HandleChangePassword failed", "username", user.Username, "error", err)
		return err
	}

	h.log.Info(c.Request().Context(), "HandleChangePassword succeeded", "username", user.Username)
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
	authUser, ok := request.CurrentUser(c)
	if !ok {
		h.log.Error(c.Request().Context(), "HandleUpdateProfile failed", "error", "User not found in context")
		return echo.ErrUnauthorized
	}

	var req dto.UpdateProfileRequest
	if err := c.Bind(&req); err != nil {
		h.log.Error(c.Request().Context(), "HandleUpdateProfile failed", "username", authUser.Username, "error", "Failed to bind request body")
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	if err := h.validator.Validate(req); err != nil {
		h.log.Error(c.Request().Context(), "HandleUpdateProfile failed", "username", authUser.Username, "validation_error", err)
		return err
	}

	updatedUser := &identitydomain.User{
		FirstName: strings.TrimSpace(req.FirstName),
		LastName:  strings.TrimSpace(req.LastName),
		Username:  strings.ToLower(strings.TrimSpace(req.Username)),
	}

	if err := h.userUseCase.UpdateProfile(c.Request().Context(), authUser.ID, updatedUser); err != nil {
		h.log.Error(c.Request().Context(), "HandleUpdateProfile failed", "username", authUser.Username, "error", err)
		return err
	}

	h.log.Info(c.Request().Context(), "HandleUpdateProfile succeeded", "username", authUser.Username)
	return c.JSON(http.StatusOK, dto.UpdateProfileResponse{
		Message:  "Profile updated successfully",
		Username: updatedUser.Username,
	})
}
