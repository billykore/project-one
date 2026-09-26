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

// PostQueryHandler owns all read-only post routes.
type PostQueryHandler struct {
	postUseCase    ports.PostQueryUseCase
	commentUseCase ports.CommentUseCase
	log            ports.Logger
}

func NewPostQueryHandler(postUseCase ports.PostQueryUseCase, commentUseCase ports.CommentUseCase, log ports.Logger) *PostQueryHandler {
	return &PostQueryHandler{postUseCase: postUseCase, commentUseCase: commentUseCase, log: log}
}

// GetPostByID handles GET /posts/:id.
//
//	@Summary	Get post by ID
//	@Tags		posts
//	@Produce	json
//	@Param		id			path		int	true	"Post ID"
//	@Success	200			{object}	dto.PostResponse
//	@Failure	400,404,500	{object}	dto.ProblemDetail
//	@Router		/posts/{id} [get]
func (h *PostQueryHandler) GetPostByID(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return domain.ErrPostIDMustBeANumber
	}
	post, err := h.postUseCase.GetPostByID(c.Request().Context(), id)
	if err != nil {
		return err
	}
	comments, err := h.commentUseCase.GetCommentsByPostID(c.Request().Context(), id)
	if err != nil {
		return err
	}
	response := make([]*dto.CommentResponse, 0, len(comments))
	for _, comment := range comments {
		response = append(response, &dto.CommentResponse{ID: comment.ID, Username: comment.Username, Content: comment.Content, CreatedAt: comment.CreatedAt})
	}
	return c.JSON(http.StatusOK, dto.PostResponse{ID: post.ID, Title: post.Title, Content: post.Content, Tags: post.Tags, Author: post.Username, LikeCount: post.LikeCount, Comments: response, CreatedAt: post.CreatedAt, UpdatedAt: post.UpdatedAt})
}

// GetPosts handles GET /posts.
//
//	@Summary	Get user posts
//	@Tags		posts
//	@Produce	json
//	@Param		cursor	query		string	false	"Pagination cursor from the previous response"
//	@Param		limit	query		int		false	"Items per page (1-100, default 10)"
//	@Success	200		{object}	dto.PostsListResponse
//	@Failure	401,500	{object}	dto.ProblemDetail
//	@Security	BearerAuth
//	@Router		/posts [get]
func (h *PostQueryHandler) GetPosts(c echo.Context) error {
	user, ok := c.Get("user").(*domain.User)
	if !ok {
		return echo.ErrUnauthorized
	}
	limit := 10
	if value := c.QueryParam("limit"); value != "" {
		parsed, err := strconv.Atoi(value)
		if err != nil || parsed < 1 || parsed > 100 {
			return echo.ErrBadRequest
		}
		limit = parsed
	}
	var cursor *vo.Cursor
	if value := c.QueryParam("cursor"); value != "" {
		decoded, err := vo.DecodeCursor(value)
		if err != nil {
			return domain.ErrInvalidCursor
		}
		cursor = &decoded
	}
	posts, nextCursor, hasMore, err := h.postUseCase.GetPosts(c.Request().Context(), user.ID, cursor, limit)
	if err != nil {
		return err
	}
	response := make([]dto.PostResponse, 0, len(posts))
	for _, post := range posts {
		response = append(response, dto.PostResponse{ID: post.ID, Title: post.Title, Content: post.Content, Tags: post.Tags, Author: post.Username, CreatedAt: post.CreatedAt, UpdatedAt: post.UpdatedAt})
	}
	nextCursorValue := ""
	if nextCursor != nil {
		nextCursorValue = nextCursor.Encode()
	}
	return c.JSON(http.StatusOK, dto.PostsListResponse{Data: response, NextCursor: nextCursorValue, HasMore: hasMore})
}

// GetLikeStatus handles GET /posts/:id/likes.
//
//	@Summary	Get like status
//	@Tags		posts,likes
//	@Produce	json
//	@Param		id				path		int	true	"Post ID"
//	@Success	200				{object}	dto.LikeResponse
//	@Failure	400,401,404,500	{object}	dto.ProblemDetail
//	@Security	BearerAuth
//	@Router		/posts/{id}/likes [get]
func (h *PostQueryHandler) GetLikeStatus(c echo.Context) error {
	user, ok := c.Get("user").(*domain.User)
	if !ok {
		return echo.ErrUnauthorized
	}
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return domain.ErrPostIDMustBeANumber
	}
	liked, count, err := h.postUseCase.GetLikeStatus(c.Request().Context(), id, user.Username)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, dto.LikeResponse{Liked: liked, LikeCount: count})
}
