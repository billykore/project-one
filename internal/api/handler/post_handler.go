package handler

import (
	"net/http"
	"strconv"

	"github.com/billykore/project-one/internal/api/dto"
	"github.com/billykore/project-one/internal/core/domain"
	"github.com/billykore/project-one/internal/core/ports"
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
	username, ok := c.Get("username").(string)
	if !ok {
		return domain.ErrUnauthorized
	}

	var req dto.CreatePostRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	if err := h.validator.Validate(req); err != nil {
		return err
	}

	post, err := h.postUseCase.CreatePost(c.Request().Context(), username, req.Title, req.Content, req.Tags)
	if err != nil {
		return err
	}

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
		return echo.NewHTTPError(http.StatusBadRequest, "Post ID must be a number")
	}

	post, err := h.postUseCase.GetPostByID(c.Request().Context(), id)
	if err != nil {
		return err
	}

	comments, err := h.commentUseCase.GetCommentsByPostID(c.Request().Context(), id)
	if err != nil {
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
//	@Param			limit	query		int	false	"Limit"
//	@Param			offset	query		int	false	"Offset"
//	@Success		200		{array}		dto.PostResponse
//	@Failure		401		{object}	dto.ProblemDetail
//	@Failure		500		{object}	dto.ProblemDetail
//	@Security		BearerAuth
//	@Router			/posts [get]
func (h *PostHandler) GetPosts(c echo.Context) error {
	username, ok := c.Get("username").(string)
	if !ok {
		return domain.ErrUnauthorized
	}

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
			Author:    p.Username,
			CreatedAt: p.CreatedAt,
			UpdatedAt: p.UpdatedAt,
		})
	}

	return c.JSON(http.StatusOK, response)
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
	username, ok := c.Get("username").(string)
	if !ok {
		return domain.ErrUnauthorized
	}

	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Post ID must be a number")
	}

	var req dto.UpdatePostRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	post, err := h.postUseCase.UpdatePost(c.Request().Context(), username, id, req.Title, req.Content)
	if err != nil {
		return err
	}

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
	username, ok := c.Get("username").(string)
	if !ok {
		return domain.ErrUnauthorized
	}

	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Post ID must be a number")
	}

	err = h.postUseCase.DeletePost(c.Request().Context(), username, id)
	if err != nil {
		return err
	}

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
	username, ok := c.Get("username").(string)
	if !ok {
		return domain.ErrUnauthorized
	}

	var req dto.CreateCommentRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	if err := h.validator.Validate(req); err != nil {
		return err
	}

	err := h.commentUseCase.AddComment(c.Request().Context(), req.ID, username, req.Content)
	if err != nil {
		return err
	}

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
	username, ok := c.Get("username").(string)
	if !ok {
		return domain.ErrUnauthorized
	}

	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Post ID must be a number")
	}

	likeCount, err := h.postUseCase.LikePost(c.Request().Context(), id, username)
	if err != nil {
		return err
	}

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
	username, ok := c.Get("username").(string)
	if !ok {
		return domain.ErrUnauthorized
	}

	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Post ID must be a number")
	}

	likeCount, err := h.postUseCase.UnlikePost(c.Request().Context(), id, username)
	if err != nil {
		return err
	}

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
	username, ok := c.Get("username").(string)
	if !ok {
		return domain.ErrUnauthorized
	}

	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Post ID must be a number")
	}

	liked, likeCount, err := h.postUseCase.GetLikeStatus(c.Request().Context(), id, username)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, dto.LikeResponse{
		Liked:     liked,
		LikeCount: likeCount,
	})
}
