package handler

import (
	"net/http"
	"strconv"

	vo "github.com/billykore/project-one/internal/platform/pagination"
	platformports "github.com/billykore/project-one/internal/platform/ports"
	"github.com/billykore/project-one/internal/platform/problem"
	"github.com/billykore/project-one/internal/publishing/api/dto"
	publishingports "github.com/billykore/project-one/internal/publishing/ports"
	"github.com/labstack/echo/v4"
)

// UserPostsHandler serves profile-post reads owned by the publishing context.
type UserPostsHandler struct {
	postUseCase publishingports.PostQueryUseCase
	userLookup  publishingports.UserLookup
	log         platformports.Logger
}

// NewUserPostsHandler creates a profile-posts API handler.
func NewUserPostsHandler(postUseCase publishingports.PostQueryUseCase, userLookup publishingports.UserLookup, log platformports.Logger) *UserPostsHandler {
	return &UserPostsHandler{postUseCase: postUseCase, userLookup: userLookup, log: log}
}

// GetUserPosts handles the GET /users/:username/posts endpoint.
//
//	@Summary		Get user posts by username
//	@Description	Retrieve all posts for a specific user by username.
//	@Tags			users
//	@Produce		json
//	@Param			username	path		string	true	"Username"
//	@Param			cursor		query		string	false	"Pagination cursor from the previous response"
//	@Param			limit		query		int		false	"Items per page (1-100, default 10)"
//	@Success		200			{object}	dto.PostsListResponse
//	@Failure		400			{object}	dto.ProblemDetail
//	@Failure		500			{object}	dto.ProblemDetail
//	@Router			/users/{username}/posts [get]
func (h *UserPostsHandler) GetUserPosts(c echo.Context) error {
	username := c.Param("username")
	if username == "" {
		h.log.Error(c.Request().Context(), "GetUserPosts failed", "error", "Username parameter is empty")
		return echo.ErrBadRequest
	}

	limit := 10
	if limitStr := c.QueryParam("limit"); limitStr != "" {
		parsed, err := strconv.Atoi(limitStr)
		if err != nil || parsed < 1 || parsed > 100 {
			return echo.ErrBadRequest
		}
		limit = parsed
	}
	var cursor *vo.Cursor
	if cursorStr := c.QueryParam("cursor"); cursorStr != "" {
		decoded, err := vo.DecodeCursor(cursorStr)
		if err != nil {
			return problem.ErrInvalidCursor
		}
		cursor = &decoded
	}

	profile, err := h.userLookup.GetUserByUsername(c.Request().Context(), username)
	if err != nil {
		h.log.Error(c.Request().Context(), "GetUserPosts failed", "username", username, "error", err)
		return err
	}
	posts, nextCursor, hasMore, err := h.postUseCase.GetPosts(c.Request().Context(), profile.ID, cursor, limit)
	if err != nil {
		h.log.Error(c.Request().Context(), "GetUserPosts failed", "username", username, "error", err)
		return err
	}

	response := make([]dto.PostResponse, 0, len(posts))
	for _, p := range posts {
		response = append(response, dto.PostResponse{ID: p.ID, Title: p.Title, Content: p.Content, Tags: p.Tags, Author: p.Username, CreatedAt: p.CreatedAt, UpdatedAt: p.UpdatedAt})
	}
	nextCursorValue := ""
	if nextCursor != nil {
		nextCursorValue = nextCursor.Encode()
	}
	h.log.Info(c.Request().Context(), "GetUserPosts succeeded", "username", username, "count", len(response))
	return c.JSON(http.StatusOK, dto.PostsListResponse{Data: response, NextCursor: nextCursorValue, HasMore: hasMore})
}
