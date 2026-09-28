package handler

import (
	"net/http"
	"strconv"
	"time"

	request "github.com/billykore/project-one/internal/platform/api/request"
	vo "github.com/billykore/project-one/internal/platform/pagination"
	platformports "github.com/billykore/project-one/internal/platform/ports"
	"github.com/billykore/project-one/internal/platform/problem"
	"github.com/billykore/project-one/internal/social/api/dto"
	"github.com/billykore/project-one/internal/social/domain"
	socialports "github.com/billykore/project-one/internal/social/ports"
	"github.com/labstack/echo/v4"
)

// FollowHandler serves the HTTP API owned by the social context.
type FollowHandler struct {
	followUseCase socialports.FollowUseCase
	validator     platformports.Validator
	log           platformports.Logger
}

// NewFollowHandler creates a follow API handler.
func NewFollowHandler(followUseCase socialports.FollowUseCase, validator platformports.Validator, log platformports.Logger) *FollowHandler {
	return &FollowHandler{followUseCase: followUseCase, validator: validator, log: log}
}

// HandleFollow handles the POST /users/{username}/followers endpoint.
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
func (h *FollowHandler) HandleFollow(c echo.Context) error {
	user, ok := request.CurrentUser(c)
	if !ok {
		h.log.Error(c.Request().Context(), "HandleFollow failed", "error", "User not found in context")
		return echo.ErrUnauthorized
	}

	followedUsername := c.Param("username")
	if followedUsername == "" {
		h.log.Error(c.Request().Context(), "HandleFollow failed", "error", "Username parameter is empty")
		return echo.ErrBadRequest
	}

	follow, err := h.followUseCase.Follow(c.Request().Context(), user, followedUsername)
	if err != nil {
		h.log.Error(c.Request().Context(), "HandleFollow failed", "follower", user.Username, "followed", followedUsername, "error", err)
		return err
	}

	h.log.Info(c.Request().Context(), "HandleFollow succeeded", "follower", user.Username, "followed", followedUsername)
	return c.JSON(http.StatusOK, dto.FollowResponse{
		Message: "You are now following this user.",
		Data: dto.FollowData{
			FollowedUsername: follow.FollowedUsername,
			FollowedAt:       follow.CreatedAt.Format(time.RFC3339),
		},
	})
}

// HandleUnfollow handles the DELETE /users/{username}/followers endpoint.
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
func (h *FollowHandler) HandleUnfollow(c echo.Context) error {
	user, ok := request.CurrentUser(c)
	if !ok {
		h.log.Error(c.Request().Context(), "HandleUnfollow failed", "error", "User not found in context")
		return echo.ErrUnauthorized
	}

	followedUsername := c.Param("username")
	if followedUsername == "" {
		h.log.Error(c.Request().Context(), "HandleUnfollow failed", "error", "Username parameter is empty")
		return echo.ErrBadRequest
	}

	if err := h.followUseCase.Unfollow(c.Request().Context(), user, followedUsername); err != nil {
		h.log.Error(c.Request().Context(), "HandleUnfollow failed", "follower", user.Username, "followed", followedUsername, "error", err)
		return err
	}

	h.log.Info(c.Request().Context(), "HandleUnfollow succeeded", "follower", user.Username, "followed", followedUsername)
	return c.JSON(http.StatusOK, dto.UnfollowResponse{Message: "Successfully unfollowed this user."})
}

// GetFollowing handles the GET /users/:username/following endpoint.
//
//	@Summary		Get following list
//	@Description	Get the list of users being followed by the currently authenticated user.
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Param			cursor	query		string	false	"Pagination cursor from the previous response"
//	@Param			limit	query		int		false	"Items per page (1-100, default 10)"
//	@Success		200		{object}	dto.FollowingListResponse
//	@Failure		400		{object}	dto.ProblemDetail
//	@Failure		401		{object}	dto.ProblemDetail
//	@Failure		500		{object}	dto.ProblemDetail
//	@Security		BearerAuth
//	@Router			/users/{username}/following [get]
func (h *FollowHandler) GetFollowing(c echo.Context) error {
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
	if limitStr := c.QueryParam("limit"); limitStr != "" {
		limit, err := strconv.Atoi(limitStr)
		if err != nil || limit < 1 || limit > 100 {
			return echo.ErrBadRequest
		}
		req.Limit = limit
	} else {
		req.Limit = 10
	}
	if err := h.validator.Validate(req); err != nil {
		h.log.Error(c.Request().Context(), "GetFollowing failed", "validation_error", err)
		return err
	}

	var cursor *vo.Cursor
	if req.Cursor != "" {
		decoded, err := vo.DecodeCursor(req.Cursor)
		if err != nil {
			return problem.ErrInvalidCursor
		}
		cursor = &decoded
	}
	following, err := h.followUseCase.GetFollowing(c.Request().Context(), followerUsername, cursor, req.Limit)
	if err != nil {
		h.log.Error(c.Request().Context(), "GetFollowing failed", "follower", followerUsername, "error", err)
		return err
	}

	res := make([]dto.FollowingResponse, 0, len(following.Data))
	for _, f := range following.Data {
		res = append(res, toFollowingResponse(f))
	}
	nextCursor := ""
	if following.NextCursor != nil {
		nextCursor = following.NextCursor.Encode()
	}
	h.log.Info(c.Request().Context(), "GetFollowing succeeded", "follower", followerUsername, "count", len(res))
	return c.JSON(http.StatusOK, dto.FollowingListResponse{Data: res, NextCursor: nextCursor, HasMore: following.HasMore})
}

// GetFollowers handles the GET /users/:username/followers endpoint.
//
//	@Summary		Get followers list
//	@Description	Get the list of users following the currently authenticated user.
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Param			cursor	query		string	false	"Pagination cursor from the previous response"
//	@Param			limit	query		int		false	"Items per page (1-100, default 10)"
//	@Success		200		{object}	dto.FollowersListResponse
//	@Failure		400		{object}	dto.ProblemDetail
//	@Failure		401		{object}	dto.ProblemDetail
//	@Failure		500		{object}	dto.ProblemDetail
//	@Security		BearerAuth
//	@Router			/users/{username}/followers [get]
func (h *FollowHandler) GetFollowers(c echo.Context) error {
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
	if limitStr := c.QueryParam("limit"); limitStr != "" {
		limit, err := strconv.Atoi(limitStr)
		if err != nil || limit < 1 || limit > 100 {
			return echo.ErrBadRequest
		}
		req.Limit = limit
	} else {
		req.Limit = 10
	}
	if err := h.validator.Validate(req); err != nil {
		h.log.Error(c.Request().Context(), "GetFollowers failed", "validation_error", err)
		return err
	}

	var cursor *vo.Cursor
	if req.Cursor != "" {
		decoded, err := vo.DecodeCursor(req.Cursor)
		if err != nil {
			return problem.ErrInvalidCursor
		}
		cursor = &decoded
	}
	followers, err := h.followUseCase.GetFollowers(c.Request().Context(), followedUsername, cursor, req.Limit)
	if err != nil {
		h.log.Error(c.Request().Context(), "GetFollowers failed", "followed", followedUsername, "error", err)
		return err
	}

	res := make([]dto.FollowerResponse, 0, len(followers.Data))
	for _, f := range followers.Data {
		res = append(res, toFollowerResponse(f))
	}
	nextCursor := ""
	if followers.NextCursor != nil {
		nextCursor = followers.NextCursor.Encode()
	}
	h.log.Info(c.Request().Context(), "GetFollowers succeeded", "followed", followedUsername, "count", len(res))
	return c.JSON(http.StatusOK, dto.FollowersListResponse{Data: res, NextCursor: nextCursor, HasMore: followers.HasMore})
}

func toFollowingResponse(f domain.Following) dto.FollowingResponse {
	return dto.FollowingResponse{Username: f.Username, Name: f.FirstName + " " + f.LastName, FollowedAt: f.FollowedAt.Format(time.RFC3339), IsMutual: f.IsMutual}
}

func toFollowerResponse(f domain.Follower) dto.FollowerResponse {
	return dto.FollowerResponse{Username: f.Username, Name: f.FirstName + " " + f.LastName, FollowedAt: f.FollowedAt.Format(time.RFC3339), IsMutual: f.IsMutual}
}
