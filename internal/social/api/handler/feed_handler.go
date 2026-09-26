package handler

import (
	"net/http"
	"strconv"

	request "github.com/billykore/project-one/internal/platform/api/request"
	vo "github.com/billykore/project-one/internal/platform/pagination"
	platformports "github.com/billykore/project-one/internal/platform/ports"
	"github.com/billykore/project-one/internal/social/api/dto"
	socialports "github.com/billykore/project-one/internal/social/ports"
	"github.com/labstack/echo/v4"
)

type FeedHandler struct {
	feedUseCase socialports.FeedUseCase
	log         platformports.Logger
}

// NewFeedHandler creates a new instance of FeedHandler.
func NewFeedHandler(feedUseCase socialports.FeedUseCase, log platformports.Logger) *FeedHandler {
	// ponytail: nil checks removed — Go panics at method call site on nil pointer
	return &FeedHandler{
		feedUseCase: feedUseCase,
		log:         log,
	}
}

// HandleGetFeed handles the GET /feeds endpoint.
//
//	@Summary		Get feed
//	@Description	Returns paginated posts from users the authenticated user follows and their own posts.
//	@Tags			feeds
//	@Produce		json
//	@Param			cursor	query		string	false	"Pagination cursor"
//	@Param			limit	query		int		false	"Items per page (1-50, default 10)"
//	@Success		200		{object}	dto.FeedResponse
//	@Failure		400		{object}	dto.ProblemDetail
//	@Failure		401		{object}	dto.ProblemDetail
//	@Failure		500		{object}	dto.ProblemDetail
//	@Security		BearerAuth
//	@Router			/feeds [get]
func (h *FeedHandler) HandleGetFeed(c echo.Context) error {
	user, ok := request.CurrentUser(c)
	if !ok {
		h.log.Error(c.Request().Context(), "HandleGetFeed failed", "error", "User not found in context")
		return echo.ErrUnauthorized
	}

	// Parse limit.
	limit := 10
	if limitStr := c.QueryParam("limit"); limitStr != "" {
		l, err := strconv.Atoi(limitStr)
		if err != nil || l < 1 || l > 50 {
			h.log.Error(c.Request().Context(), "HandleGetFeed failed", "username", user.Username, "error", "limit must be between 1 and 50")
			return echo.NewHTTPError(http.StatusBadRequest, "Limit must be between 1 and 50")
		}
		limit = l
	}

	// Decode cursor.
	var cursor *vo.Cursor
	if cursorStr := c.QueryParam("cursor"); cursorStr != "" {
		decoded, err := vo.DecodeCursor(cursorStr)
		if err != nil {
			h.log.Error(c.Request().Context(), "HandleGetFeed failed", "username", user.Username, "error", "invalid cursor")
			return echo.NewHTTPError(http.StatusBadRequest, "Invalid cursor")
		}
		cursor = &decoded
	}

	result, err := h.feedUseCase.GetFeed(c.Request().Context(), user.ID, cursor, limit)
	if err != nil {
		h.log.Error(c.Request().Context(), "HandleGetFeed failed", "username", user.Username, "error", err)
		return err
	}

	h.log.Info(c.Request().Context(), "HandleGetFeed succeeded", "username", user.Username, "count", len(result.Posts))
	return c.JSON(http.StatusOK, dto.ToFeedResponse(result))
}
