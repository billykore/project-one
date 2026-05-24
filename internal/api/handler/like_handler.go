package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/billykore/project-one/internal/api/dto"
	"github.com/billykore/project-one/internal/core/domain"
	"github.com/billykore/project-one/internal/core/ports"
	"github.com/labstack/echo/v4"
)

// LikeHandler handles HTTP requests for post like operations.
type LikeHandler struct {
	likeUseCase ports.LikeUseCase
}

// NewLikeHandler creates a new LikeHandler instance.
func NewLikeHandler(likeUseCase ports.LikeUseCase) *LikeHandler {
	if likeUseCase == nil {
		panic("likeUseCase is required")
	}
	return &LikeHandler{
		likeUseCase: likeUseCase,
	}
}

// ToggleLike handles the POST /posts/:id/likes endpoint.
//
//	@Summary		Toggle like
//	@Description	Toggle like status on a post. Like if not yet liked, unlike if already liked.
//	@Tags			likes
//	@Produce		json
//	@Param			id	path		int	true	"Post ID"
//	@Success		200	{object}	dto.LikeResponse
//	@Failure		400	{object}	dto.ErrorResponse
//	@Failure		401	{object}	dto.ErrorResponse
//	@Failure		404	{object}	dto.ErrorResponse
//	@Failure		500	{object}	dto.ErrorResponse
//	@Security		BearerAuth
//	@Router			/posts/{id}/likes [post]
func (h *LikeHandler) ToggleLike(c echo.Context) error {
	username, ok := c.Get("username").(string)
	if !ok {
		return c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "Unauthorized"})
	}

	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "Post ID must be a number"})
	}

	liked, likeCount, err := h.likeUseCase.ToggleLike(c.Request().Context(), id, username)
	if err != nil {
		if errors.Is(err, domain.ErrPostNotFound) {
			return c.JSON(http.StatusNotFound, dto.ErrorResponse{Error: "Post not found"})
		}
		if errors.Is(err, domain.ErrInvalidPost) {
			return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "Post ID must be integer and not 0"})
		}
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "Something went wrong"})
	}

	return c.JSON(http.StatusOK, dto.LikeResponse{
		Liked:     liked,
		LikeCount: likeCount,
	})
}

// GetLikeStatus handles the GET /posts/:id/likes endpoint.
//
//	@Summary		Get like status
//	@Description	Get the like count and the current user's like status for a post.
//	@Tags			likes
//	@Produce		json
//	@Param			id	path		int	true	"Post ID"
//	@Success		200	{object}	dto.LikeResponse
//	@Failure		400	{object}	dto.ErrorResponse
//	@Failure		401	{object}	dto.ErrorResponse
//	@Failure		404	{object}	dto.ErrorResponse
//	@Failure		500	{object}	dto.ErrorResponse
//	@Security		BearerAuth
//	@Router			/posts/{id}/likes [get]
func (h *LikeHandler) GetLikeStatus(c echo.Context) error {
	username, ok := c.Get("username").(string)
	if !ok {
		return c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "Unauthorized"})
	}

	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "Post ID must be a number"})
	}

	liked, likeCount, err := h.likeUseCase.GetLikeStatus(c.Request().Context(), id, username)
	if err != nil {
		if errors.Is(err, domain.ErrPostNotFound) {
			return c.JSON(http.StatusNotFound, dto.ErrorResponse{Error: "Post not found"})
		}
		if errors.Is(err, domain.ErrInvalidPost) {
			return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "Post ID must be integer and not 0"})
		}
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "Something went wrong"})
	}

	return c.JSON(http.StatusOK, dto.LikeResponse{
		Liked:     liked,
		LikeCount: likeCount,
	})
}
