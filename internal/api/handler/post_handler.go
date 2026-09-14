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

type PostHandler struct {
	postUseCase    ports.PostUseCase
	commentUseCase ports.CommentUseCase
	validator      ports.Validator
	log            ports.Logger
}

// NewPostHandler creates a new instance of PostHandler.
func NewPostHandler(postUseCase ports.PostUseCase, commentUseCase ports.CommentUseCase, validator ports.Validator, log ports.Logger) *PostHandler {
	return &PostHandler{
		postUseCase:    postUseCase,
		commentUseCase: commentUseCase,
		validator:      validator,
		log:            log,
	}
}

// CreatePost handles the POST /posts endpoint.
//
//	@Summary		Create post
//	@Description	Create a new post for the authenticated user.
//	@Tags			posts
//	@Accept			json
//	@Produce		json
//	@Param			request	body		dto.CreatePostRequest	true	"Post details"
//	@Success		201		{object}	dto.CreatePostResponse
//	@Failure		400		{object}	dto.ProblemDetail
//	@Failure		401		{object}	dto.ProblemDetail
//	@Failure		500		{object}	dto.ProblemDetail
//	@Security		BearerAuth
//	@Router			/posts [post]
func (h *PostHandler) CreatePost(c echo.Context) error {
	user, ok := c.Get("user").(*domain.User)
	if !ok {
		h.log.Error(c.Request().Context(), "CreatePost failed", "error", "User not found in context")
		return echo.ErrUnauthorized
	}

	var req dto.CreatePostRequest
	if err := c.Bind(&req); err != nil {
		h.log.Error(c.Request().Context(), "CreatePost failed", "user.Username", user.Username, "error", "Invalid request body")
		return echo.ErrBadRequest
	}

	if err := h.validator.Validate(req); err != nil {
		h.log.Error(c.Request().Context(), "CreatePost failed", "user.Username", user.Username, "validation_error", err)
		return err
	}

	post, err := h.postUseCase.CreatePost(c.Request().Context(), user, req.Title, req.Content, req.Tags)
	if err != nil {
		h.log.Error(c.Request().Context(), "CreatePost failed", "user.Username", user.Username, "error", err)
		return err
	}

	h.log.Info(c.Request().Context(), "CreatePost succeeded", "user.Username", user.Username, "post_id", post.ID)
	return c.JSON(http.StatusCreated, dto.CreatePostResponse{
		ID:      post.ID,
		Message: "Post created successfully",
	})
}

// GetPostByID handles the GET /posts/:id endpoint.
//
//	@Summary		Get post by ID
//	@Description	Retrieve a specific post by its ID.
//	@Tags			posts
//	@Produce		json
//	@Param			id	path		int	true	"Post ID"
//	@Success		200	{object}	dto.PostResponse
//	@Failure		400	{object}	dto.ProblemDetail
//	@Failure		404	{object}	dto.ProblemDetail
//	@Failure		500	{object}	dto.ProblemDetail
//	@Router			/posts/{id} [get]
func (h *PostHandler) GetPostByID(c echo.Context) error {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.log.Error(c.Request().Context(), "GetPostByID failed", "error", "Post ID must be a number")
		return domain.ErrPostIDMustBeANumber
	}

	post, err := h.postUseCase.GetPostByID(c.Request().Context(), id)
	if err != nil {
		h.log.Error(c.Request().Context(), "GetPostByID failed", "post_id", id, "error", err)
		return err
	}

	comments, err := h.commentUseCase.GetCommentsByPostID(c.Request().Context(), id)
	if err != nil {
		h.log.Error(c.Request().Context(), "GetPostByID failed", "post_id", id, "error", err)
		return err
	}

	commentsResp := make([]*dto.CommentResponse, 0)
	for _, comment := range comments {
		commentsResp = append(commentsResp, &dto.CommentResponse{
			ID:        comment.ID,
			Username:  comment.Username,
			Content:   comment.Content,
			CreatedAt: comment.CreatedAt,
		})
	}

	h.log.Info(c.Request().Context(), "GetPostByID succeeded", "post_id", id)
	return c.JSON(http.StatusOK, dto.PostResponse{
		ID:        post.ID,
		Title:     post.Title,
		Content:   post.Content,
		Tags:      post.Tags,
		Author:    post.Username,
		LikeCount: post.LikeCount,
		Comments:  commentsResp,
		CreatedAt: post.CreatedAt,
		UpdatedAt: post.UpdatedAt,
	})
}

// GetPosts handles the GET /posts endpoint.
//
//	@Summary		Get user posts
//	@Description	Retrieve all posts for the authenticated user.
//	@Tags			posts
//	@Produce		json
//	@Param			cursor	query		string	false	"Pagination cursor from the previous response"
//	@Param			limit	query		int		false	"Items per page (1-100, default 10)"
//	@Success		200		{object}	dto.PostsListResponse
//	@Failure		401		{object}	dto.ProblemDetail
//	@Failure		500		{object}	dto.ProblemDetail
//	@Security		BearerAuth
//	@Router			/posts [get]
func (h *PostHandler) GetPosts(c echo.Context) error {
	user, ok := c.Get("user").(*domain.User)
	if !ok {
		h.log.Error(c.Request().Context(), "GetPosts failed", "error", "user.Username not found in context")
		return echo.ErrUnauthorized
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
			return domain.ErrInvalidCursor
		}
		cursor = &decoded
	}

	posts, err := h.postUseCase.GetPosts(c.Request().Context(), user.Username, cursor, limit)
	if err != nil {
		h.log.Error(c.Request().Context(), "GetPosts failed", "username", user.Username, "error", err)
		return err
	}

	response := make([]dto.PostResponse, 0, len(posts.Posts))
	for _, p := range posts.Posts {
		response = append(response, dto.PostResponse{
			ID:        p.ID,
			Title:     p.Title,
			Content:   p.Content,
			Tags:      p.Tags,
			Author:    p.Username,
			CreatedAt: p.CreatedAt,
			UpdatedAt: p.UpdatedAt,
		})
	}

	h.log.Info(c.Request().Context(), "GetPosts succeeded", "username", user.Username, "count", len(response))
	nextCursor := ""
	if posts.NextCursor != nil {
		nextCursor = posts.NextCursor.Encode()
	}
	return c.JSON(http.StatusOK, dto.PostsListResponse{Data: response, NextCursor: nextCursor, HasMore: posts.HasMore})
}

// UpdatePost handles the PUT /posts/:id endpoint.
//
//	@Summary		Update post
//	@Description	Update an existing post for the authenticated user.
//	@Tags			posts
//	@Accept			json
//	@Produce		json
//	@Param			id		path		int						true	"Post ID"
//	@Param			request	body		dto.UpdatePostRequest	true	"Post details"
//	@Success		200		{object}	dto.PostResponse
//	@Failure		400		{object}	dto.ProblemDetail
//	@Failure		401		{object}	dto.ProblemDetail
//	@Failure		403		{object}	dto.ProblemDetail
//	@Failure		404		{object}	dto.ProblemDetail
//	@Failure		500		{object}	dto.ProblemDetail
//	@Security		BearerAuth
//	@Router			/posts/{id} [put]
func (h *PostHandler) UpdatePost(c echo.Context) error {
	user, ok := c.Get("user").(*domain.User)
	if !ok {
		h.log.Error(c.Request().Context(), "UpdatePost failed", "error", "user not found in context")
		return echo.ErrUnauthorized
	}

	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.log.Error(c.Request().Context(), "UpdatePost failed", "username", user.Username, "error", "Post ID must be a number")
		return domain.ErrPostIDMustBeANumber
	}

	var req dto.UpdatePostRequest
	if err := c.Bind(&req); err != nil {
		h.log.Error(c.Request().Context(), "UpdatePost failed", "username", user.Username, "post_id", id, "error", "Invalid request body")
		return echo.ErrBadRequest
	}

	post, err := h.postUseCase.UpdatePost(c.Request().Context(), user.Username, id, req.Title, req.Content)
	if err != nil {
		h.log.Error(c.Request().Context(), "UpdatePost failed", "username", user.Username, "post_id", id, "error", err)
		return err
	}

	h.log.Info(c.Request().Context(), "UpdatePost succeeded", "username", user.Username, "post_id", id)
	return c.JSON(http.StatusOK, dto.PostResponse{
		ID:        post.ID,
		Message:   "Post updated successfully",
		UpdatedAt: post.UpdatedAt,
	})
}

// DeletePost handles the DELETE /posts/:id endpoint.
//
//	@Summary		Delete post
//	@Description	Soft delete a post for the authenticated user.
//	@Tags			posts
//	@Param			id	path		int	true	"Post ID"
//	@Success		200	{object}	map[string]interface{}
//	@Failure		400	{object}	dto.ProblemDetail
//	@Failure		401	{object}	dto.ProblemDetail
//	@Failure		404	{object}	dto.ProblemDetail
//	@Failure		500	{object}	dto.ProblemDetail
//	@Security		BearerAuth
//	@Router			/posts/{id} [delete]
func (h *PostHandler) DeletePost(c echo.Context) error {
	user, ok := c.Get("user").(*domain.User)
	if !ok {
		h.log.Error(c.Request().Context(), "DeletePost failed", "error", "user not found in context")
		return echo.ErrUnauthorized
	}

	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.log.Error(c.Request().Context(), "DeletePost failed", "username", user.Username, "error", "Post ID must be a number")
		return domain.ErrPostIDMustBeANumber
	}

	err = h.postUseCase.DeletePost(c.Request().Context(), user.Username, id)
	if err != nil {
		h.log.Error(c.Request().Context(), "DeletePost failed", "username", user.Username, "post_id", id, "error", err)
		return err
	}

	h.log.Info(c.Request().Context(), "DeletePost succeeded", "username", user.Username, "post_id", id)
	return c.JSON(http.StatusOK, dto.PostResponse{
		ID:      id,
		Message: "Post deleted successfully",
	})
}

// CreateComment handles the POST /posts/:id/comments endpoint.
//
//	@Summary		Create comment
//	@Description	Create a new comment on a post.
//	@Tags			posts,comments
//	@Accept			json
//	@Produce		json
//	@Param			id		path	int							true	"Post ID"
//	@Param			request	body	dto.CreateCommentRequest	true	"Comment details"
//	@Success		201		"Created"
//	@Failure		400		{object}	dto.ProblemDetail
//	@Failure		401		{object}	dto.ProblemDetail
//	@Failure		404		{object}	dto.ProblemDetail
//	@Failure		500		{object}	dto.ProblemDetail
//	@Security		BearerAuth
//	@Router			/posts/{id}/comments [post]
func (h *PostHandler) CreateComment(c echo.Context) error {
	user, ok := c.Get("user").(*domain.User)
	if !ok {
		h.log.Error(c.Request().Context(), "CreateComment failed", "error", "user not found in context")
		return echo.ErrUnauthorized
	}

	var req dto.CreateCommentRequest
	if err := c.Bind(&req); err != nil {
		h.log.Error(c.Request().Context(), "CreateComment failed", "username", user.Username, "error", "Invalid request body")
		return echo.ErrBadRequest
	}

	if err := h.validator.Validate(req); err != nil {
		h.log.Error(c.Request().Context(), "CreateComment failed", "username", user.Username, "validation_error", err)
		return err
	}

	err := h.commentUseCase.AddComment(c.Request().Context(), req.ID, user.Username, req.Content)
	if err != nil {
		h.log.Error(c.Request().Context(), "CreateComment failed", "username", user.Username, "post_id", req.ID, "error", err)
		return err
	}

	h.log.Info(c.Request().Context(), "CreateComment succeeded", "username", user.Username, "post_id", req.ID)
	return c.NoContent(http.StatusCreated)
}

// LikePost handles the POST /posts/:id/likes endpoint.
//
//	@Summary		Like post
//	@Description	Like a post idempotently.
//	@Tags			posts,likes
//	@Produce		json
//	@Param			id	path		int	true	"Post ID"
//	@Success		200	{object}	dto.LikeResponse
//	@Failure		400	{object}	dto.ProblemDetail
//	@Failure		401	{object}	dto.ProblemDetail
//	@Failure		404	{object}	dto.ProblemDetail
//	@Failure		500	{object}	dto.ProblemDetail
//	@Security		BearerAuth
//	@Router			/posts/{id}/likes [post]
func (h *PostHandler) LikePost(c echo.Context) error {
	user, ok := c.Get("user").(*domain.User)
	if !ok {
		h.log.Error(c.Request().Context(), "LikePost failed", "error", "user not found in context")
		return echo.ErrUnauthorized
	}

	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.log.Error(c.Request().Context(), "LikePost failed", "username", user.Username, "error", "Post ID must be a number")
		return domain.ErrPostIDMustBeANumber
	}

	likeCount, err := h.postUseCase.LikePost(c.Request().Context(), id, user.Username)
	if err != nil {
		h.log.Error(c.Request().Context(), "LikePost failed", "username", user.Username, "post_id", id, "error", err)
		return err
	}

	h.log.Info(c.Request().Context(), "LikePost succeeded", "username", user.Username, "post_id", id, "like_count", likeCount)
	return c.JSON(http.StatusOK, dto.LikeResponse{
		Liked:     true,
		LikeCount: likeCount,
	})
}

// UnlikePost handles the DELETE /posts/:id/likes endpoint.
//
//	@Summary		Unlike post
//	@Description	Unlike a post idempotently.
//	@Tags			posts,likes
//	@Produce		json
//	@Param			id	path		int	true	"Post ID"
//	@Success		200	{object}	dto.LikeResponse
//	@Failure		400	{object}	dto.ProblemDetail
//	@Failure		401	{object}	dto.ProblemDetail
//	@Failure		404	{object}	dto.ProblemDetail
//	@Failure		500	{object}	dto.ProblemDetail
//	@Security		BearerAuth
//	@Router			/posts/{id}/likes [delete]
func (h *PostHandler) UnlikePost(c echo.Context) error {
	user, ok := c.Get("user").(*domain.User)
	if !ok {
		h.log.Error(c.Request().Context(), "UnlikePost failed", "error", "user not found in context")
		return echo.ErrUnauthorized
	}

	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.log.Error(c.Request().Context(), "UnlikePost failed", "username", user.Username, "error", "Post ID must be a number")
		return domain.ErrPostIDMustBeANumber
	}

	likeCount, err := h.postUseCase.UnlikePost(c.Request().Context(), id, user.Username)
	if err != nil {
		h.log.Error(c.Request().Context(), "UnlikePost failed", "username", user.Username, "post_id", id, "error", err)
		return err
	}

	h.log.Info(c.Request().Context(), "UnlikePost succeeded", "username", user.Username, "post_id", id, "like_count", likeCount)
	return c.JSON(http.StatusOK, dto.LikeResponse{
		Liked:     false,
		LikeCount: likeCount,
	})
}

// GetLikeStatus handles the GET /posts/:id/likes endpoint.
//
//	@Summary		Get like status
//	@Description	Get the like count and the current user's like status for a post.
//	@Tags			posts,likes
//	@Produce		json
//	@Param			id	path		int	true	"Post ID"
//	@Success		200	{object}	dto.LikeResponse
//	@Failure		400	{object}	dto.ProblemDetail
//	@Failure		401	{object}	dto.ProblemDetail
//	@Failure		404	{object}	dto.ProblemDetail
//	@Failure		500	{object}	dto.ProblemDetail
//	@Security		BearerAuth
//	@Router			/posts/{id}/likes [get]
func (h *PostHandler) GetLikeStatus(c echo.Context) error {
	user, ok := c.Get("user").(*domain.User)
	if !ok {
		h.log.Error(c.Request().Context(), "GetLikeStatus failed", "error", "user not found in context")
		return echo.ErrUnauthorized
	}

	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.log.Error(c.Request().Context(), "GetLikeStatus failed", "username", user.Username, "error", "Post ID must be a number")
		return domain.ErrPostIDMustBeANumber
	}

	liked, likeCount, err := h.postUseCase.GetLikeStatus(c.Request().Context(), id, user.Username)
	if err != nil {
		h.log.Error(c.Request().Context(), "GetLikeStatus failed", "username", user.Username, "post_id", id, "error", err)
		return err
	}

	h.log.Info(c.Request().Context(), "GetLikeStatus succeeded", "username", user.Username, "post_id", id, "liked", liked, "like_count", likeCount)
	return c.JSON(http.StatusOK, dto.LikeResponse{
		Liked:     liked,
		LikeCount: likeCount,
	})
}
