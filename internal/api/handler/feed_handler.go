package handler

import (
	"net/http"
	"strconv"

	"github.com/billykore/project-one/internal/api/dto"
	"github.com/billykore/project-one/internal/core/domain"
	"github.com/billykore/project-one/internal/core/ports"
	vo "github.com/billykore/project-one/internal/core/valueobject"
	"github.com/labstack/echo/v4"
)

type FeedHandler struct {
	feedUseCase ports.FeedUseCase
	log         ports.Logger
}

// NewFeedHandler creates a new instance of FeedHandler.
func NewFeedHandler(feedUseCase ports.FeedUseCase, log ports.Logger) *FeedHandler {
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
	username, ok := c.Get("username").(string)
	if !ok {
		return domain.ErrUnauthorized
	}

	// Parse limit.
	limit := 10
	if limitStr := c.QueryParam("limit"); limitStr != "" {
		l, err := strconv.Atoi(limitStr)
		if err != nil || l < 1 || l > 50 {
			return echo.NewHTTPError(http.StatusBadRequest, "limit must be between 1 and 50")
		}
		limit = l
	}

	// Decode cursor.
	var cursor *vo.Cursor
	if cursorStr := c.QueryParam("cursor"); cursorStr != "" {
		decoded, err := vo.DecodeCursor(cursorStr)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid cursor")
		}
		cursor = &decoded
	}

	result, err := h.feedUseCase.GetFeed(c.Request().Context(), username, cursor, limit)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, dto.ToFeedResponse(result))
}
