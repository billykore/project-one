package handler

import (
	"net/http"
	"strconv"

	identitydomain "github.com/billykore/project-one/internal/identity/domain"
	platformports "github.com/billykore/project-one/internal/platform/ports"
	"github.com/billykore/project-one/internal/platform/problem"
	"github.com/billykore/project-one/internal/publishing/api/dto"
	publishingports "github.com/billykore/project-one/internal/publishing/ports"
	"github.com/labstack/echo/v4"
)

// PostCommandHandler owns all state-changing post routes.
type PostCommandHandler struct {
	postUseCase    publishingports.PostCommandUseCase
	commentUseCase publishingports.CommentUseCase
	validator      platformports.Validator
	log            platformports.Logger
}

func NewPostCommandHandler(postUseCase publishingports.PostCommandUseCase, commentUseCase publishingports.CommentUseCase, validator platformports.Validator, log platformports.Logger) *PostCommandHandler {
	return &PostCommandHandler{postUseCase: postUseCase, commentUseCase: commentUseCase, validator: validator, log: log}
}

// CreatePost handles POST /posts.
//
//	@Summary	Create post
//	@Tags		posts
//	@Accept		json
//	@Produce	json
//	@Param		request		body		dto.CreatePostRequest	true	"Post details"
//	@Success	201			{object}	dto.CreatePostResponse
//	@Failure	400,401,500	{object}	dto.ProblemDetail
//	@Security	BearerAuth
//	@Router		/posts [post]
func (h *PostCommandHandler) CreatePost(c echo.Context) error {
	user, ok := c.Get("user").(*identitydomain.User)
	if !ok {
		h.log.Error(c.Request().Context(), "CreatePost failed", "error", "user not found in context")
		return echo.ErrUnauthorized
	}
	var req dto.CreatePostRequest
	if err := c.Bind(&req); err != nil {
		h.log.Error(c.Request().Context(), "CreatePost failed", "username", user.Username, "error", "invalid request body")
		return echo.ErrBadRequest
	}
	if err := h.validator.Validate(req); err != nil {
		h.log.Error(c.Request().Context(), "CreatePost failed", "username", user.Username, "validation_error", err)
		return err
	}
	post, err := h.postUseCase.CreatePost(c.Request().Context(), user, req.Title, req.Content, req.Tags)
	if err != nil {
		h.log.Error(c.Request().Context(), "CreatePost failed", "username", user.Username, "error", err)
		return err
	}
	h.log.Info(c.Request().Context(), "CreatePost succeeded", "username", user.Username, "post_id", post.ID)
	return c.JSON(http.StatusCreated, dto.CreatePostResponse{ID: post.ID, Message: "Post created successfully"})
}

// UpdatePost handles PUT /posts/:id.
//
//	@Summary	Update post
//	@Tags		posts
//	@Accept		json
//	@Produce	json
//	@Param		id					path		int						true	"Post ID"
//	@Param		request				body		dto.UpdatePostRequest	true	"Post details"
//	@Success	200					{object}	dto.PostResponse
//	@Failure	400,401,403,404,500	{object}	dto.ProblemDetail
//	@Security	BearerAuth
//	@Router		/posts/{id} [put]
func (h *PostCommandHandler) UpdatePost(c echo.Context) error {
	user, ok := c.Get("user").(*identitydomain.User)
	if !ok {
		h.log.Error(c.Request().Context(), "UpdatePost failed", "error", "user not found in context")
		return echo.ErrUnauthorized
	}
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return problem.ErrPostIDMustBeANumber
	}
	var req dto.UpdatePostRequest
	if err := c.Bind(&req); err != nil {
		return echo.ErrBadRequest
	}
	post, err := h.postUseCase.UpdatePost(c.Request().Context(), user.ID, id, req.Title, req.Content)
	if err != nil {
		h.log.Error(c.Request().Context(), "UpdatePost failed", "username", user.Username, "post_id", id, "error", err)
		return err
	}
	return c.JSON(http.StatusOK, dto.PostResponse{ID: post.ID, Message: "Post updated successfully", UpdatedAt: post.UpdatedAt})
}

// DeletePost handles DELETE /posts/:id.
//
//	@Summary	Delete post
//	@Tags		posts
//	@Param		id				path		int	true	"Post ID"
//	@Success	200				{object}	dto.PostResponse
//	@Failure	400,401,404,500	{object}	dto.ProblemDetail
//	@Security	BearerAuth
//	@Router		/posts/{id} [delete]
func (h *PostCommandHandler) DeletePost(c echo.Context) error {
	user, ok := c.Get("user").(*identitydomain.User)
	if !ok {
		return echo.ErrUnauthorized
	}
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return problem.ErrPostIDMustBeANumber
	}
	if err := h.postUseCase.DeletePost(c.Request().Context(), user.ID, id); err != nil {
		h.log.Error(c.Request().Context(), "DeletePost failed", "username", user.Username, "post_id", id, "error", err)
		return err
	}
	return c.JSON(http.StatusOK, dto.PostResponse{ID: id, Message: "Post deleted successfully"})
}

// CreateComment handles POST /posts/:id/comments.
//
//	@Summary	Create comment
//	@Tags		posts,comments
//	@Accept		json
//	@Produce	json
//	@Param		id		path	int							true	"Post ID"
//	@Param		request	body	dto.CreateCommentRequest	true	"Comment details"
//	@Success	201
//	@Failure	400,401,404,500	{object}	dto.ProblemDetail
//	@Security	BearerAuth
//	@Router		/posts/{id}/comments [post]
func (h *PostCommandHandler) CreateComment(c echo.Context) error {
	user, ok := c.Get("user").(*identitydomain.User)
	if !ok {
		return echo.ErrUnauthorized
	}
	var req dto.CreateCommentRequest
	if err := c.Bind(&req); err != nil {
		return echo.ErrBadRequest
	}
	if err := h.validator.Validate(req); err != nil {
		return err
	}
	if err := h.commentUseCase.AddComment(c.Request().Context(), req.ID, user, req.Content); err != nil {
		h.log.Error(c.Request().Context(), "CreateComment failed", "username", user.Username, "post_id", req.ID, "error", err)
		return err
	}
	return c.NoContent(http.StatusCreated)
}

// LikePost handles POST /posts/:id/likes.
//
//	@Summary	Like post
//	@Tags		posts,likes
//	@Produce	json
//	@Param		id				path		int	true	"Post ID"
//	@Success	200				{object}	dto.LikeResponse
//	@Failure	400,401,404,500	{object}	dto.ProblemDetail
//	@Security	BearerAuth
//	@Router		/posts/{id}/likes [post]
func (h *PostCommandHandler) LikePost(c echo.Context) error {
	user, ok := c.Get("user").(*identitydomain.User)
	if !ok {
		return echo.ErrUnauthorized
	}
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return problem.ErrPostIDMustBeANumber
	}
	count, err := h.postUseCase.LikePost(c.Request().Context(), id, user)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, dto.LikeResponse{Liked: true, LikeCount: count})
}

// UnlikePost handles DELETE /posts/:id/likes.
//
//	@Summary	Unlike post
//	@Tags		posts,likes
//	@Produce	json
//	@Param		id				path		int	true	"Post ID"
//	@Success	200				{object}	dto.LikeResponse
//	@Failure	400,401,404,500	{object}	dto.ProblemDetail
//	@Security	BearerAuth
//	@Router		/posts/{id}/likes [delete]
func (h *PostCommandHandler) UnlikePost(c echo.Context) error {
	user, ok := c.Get("user").(*identitydomain.User)
	if !ok {
		return echo.ErrUnauthorized
	}
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return problem.ErrPostIDMustBeANumber
	}
	count, err := h.postUseCase.UnlikePost(c.Request().Context(), id, user)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, dto.LikeResponse{Liked: false, LikeCount: count})
}
