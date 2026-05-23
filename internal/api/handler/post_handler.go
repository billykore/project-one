package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/billykore/project-one/internal/api/dto"
	"github.com/billykore/project-one/internal/core/domain"
	"github.com/billykore/project-one/internal/core/ports"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
)

type PostHandler struct {
	postUseCase ports.PostUseCase
	validator   ports.Validator
}

func NewPostHandler(postUseCase ports.PostUseCase, validator ports.Validator) *PostHandler {
	if postUseCase == nil {
		panic("postUseCase is required")
	}
	if validator == nil {
		panic("validator is required")
	}
	return &PostHandler{
		postUseCase: postUseCase,
		validator:   validator,
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
//	@Failure		400		{object}	dto.ErrorResponse
//	@Failure		401		{object}	dto.ErrorResponse
//	@Failure		500		{object}	dto.ErrorResponse
//	@Security		BearerAuth
//	@Router			/api/v1/posts [post]
func (h *PostHandler) CreatePost(c echo.Context) error {
	username, ok := c.Get("username").(string)
	if !ok {
		return c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "Unauthorized"})
	}

	var req dto.CreatePostRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "Invalid request body"})
	}

	if err := h.validator.Validate(req); err != nil {
		validationErrs, ok := err.(validator.ValidationErrors)
		if !ok {
			return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "Invalid request body"})
		}

		for _, err := range validationErrs {
			if err.Field() == "Title" && err.Tag() == "required" {
				return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "Title is required"})
			}
			if err.Field() == "Content" && err.Tag() == "min" {
				return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "Content must be 10 characters minimum"})
			}
		}
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "Validation failed"})
	}

	post, err := h.postUseCase.CreatePost(c.Request().Context(), username, req.Title, req.Content, req.Tags)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "Something went wrong"})
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
//	@Failure		400	{object}	dto.ErrorResponse
//	@Failure		401	{object}	dto.ErrorResponse
//	@Failure		403	{object}	dto.ErrorResponse
//	@Failure		404	{object}	dto.ErrorResponse
//	@Failure		500	{object}	dto.ErrorResponse
//	@Router			/api/v1/posts/{id} [get]
func (h *PostHandler) GetPostByID(c echo.Context) error {
	username, ok := c.Get("username").(string)
	if !ok {
		return c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "Unauthorized"})
	}

	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "Post ID must be a number"})
	}

	post, err := h.postUseCase.GetPostByID(c.Request().Context(), username, id)
	if err != nil {
		if errors.Is(err, domain.ErrPostNotFound) {
			return c.JSON(http.StatusNotFound, dto.ErrorResponse{Error: "Post not found"})
		}
		if errors.Is(err, domain.ErrInvalidPost) {
			return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "Post ID must be integer and not 0"})
		}
		if errors.Is(err, domain.ErrUnauthorized) {
			return c.JSON(http.StatusForbidden, dto.ErrorResponse{Error: "You do not have permission to access this post"})
		}
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "Something went wrong"})
	}

	return c.JSON(http.StatusOK, dto.PostResponse{
		ID:        post.ID,
		Title:     post.Title,
		Content:   post.Content,
		Tags:      post.Tags,
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
//	@Failure		401		{object}	dto.ErrorResponse
//	@Failure		500		{object}	dto.ErrorResponse
//	@Security		BearerAuth
//	@Router			/api/v1/posts [get]
func (h *PostHandler) GetPosts(c echo.Context) error {
	username, ok := c.Get("username").(string)
	if !ok {
		return c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "Unauthorized"})
	}

	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	offset, _ := strconv.Atoi(c.QueryParam("offset"))

	if limit == 0 {
		limit = 10 // default limit
	}

	posts, err := h.postUseCase.GetPosts(c.Request().Context(), username, limit, offset)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "Something went wrong"})
	}

	response := make([]dto.PostResponse, 0, len(posts))
	for _, p := range posts {
		response = append(response, dto.PostResponse{
			ID:        p.ID,
			Title:     p.Title,
			Content:   p.Content,
			Tags:      p.Tags,
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
//	@Failure		400		{object}	dto.ErrorResponse
//	@Failure		401		{object}	dto.ErrorResponse
//	@Failure		403		{object}	dto.ErrorResponse
//	@Failure		404		{object}	dto.ErrorResponse
//	@Failure		500		{object}	dto.ErrorResponse
//	@Security		BearerAuth
//	@Router			/api/v1/posts/{id} [put]
func (h *PostHandler) UpdatePost(c echo.Context) error {
	username, ok := c.Get("username").(string)
	if !ok {
		return c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "Unauthorized"})
	}

	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "Post ID must be a number"})
	}

	var req dto.UpdatePostRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "Invalid request body"})
	}

	post, err := h.postUseCase.UpdatePost(c.Request().Context(), username, id, req.Title, req.Content)
	if err != nil {
		if errors.Is(err, domain.ErrPostNotFound) {
			return c.JSON(http.StatusNotFound, dto.ErrorResponse{Error: "Post not found"})
		}
		if errors.Is(err, domain.ErrInvalidPost) {
			return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "Post ID must be integer and not 0"})
		}
		if errors.Is(err, domain.ErrUnauthorized) {
			return c.JSON(http.StatusForbidden, dto.ErrorResponse{Error: "You do not have permission to update this post"})
		}
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "Something went wrong"})
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
//	@Failure		400	{object}	dto.ErrorResponse
//	@Failure		401	{object}	dto.ErrorResponse
//	@Failure		404	{object}	dto.ErrorResponse
//	@Failure		500	{object}	dto.ErrorResponse
//	@Security		BearerAuth
//	@Router			/api/v1/posts/{id} [delete]
func (h *PostHandler) DeletePost(c echo.Context) error {
	username, ok := c.Get("username").(string)
	if !ok {
		return c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "Unauthorized"})
	}

	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "Post ID must be a number"})
	}

	err = h.postUseCase.DeletePost(c.Request().Context(), username, id)
	if err != nil {
		if errors.Is(err, domain.ErrPostNotFound) {
			return c.JSON(http.StatusNotFound, dto.ErrorResponse{Error: "Post not found"})
		}
		if errors.Is(err, domain.ErrInvalidPost) {
			return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "Post ID must be integer and not 0"})
		}
		if errors.Is(err, domain.ErrUnauthorized) {
			return c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "Unauthorized"})
		}
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "Something went wrong"})
	}

	return c.JSON(http.StatusOK, dto.PostResponse{
		ID:      id,
		Message: "Post deleted successfully",
	})
}

// GetUserPosts handles the GET /users/:username/posts endpoint.
//
//	@Summary		Get user posts by username
//	@Description	Retrieve all posts for a specific user by username.
//	@Tags			posts
//	@Produce		json
//	@Param			username	path		string	true	"Username"
//	@Param			limit		query		int		false	"Limit"
//	@Param			offset		query		int		false	"Offset"
//	@Success		200			{array}		dto.PostResponse
//	@Failure		400			{object}	dto.ErrorResponse
//	@Failure		500			{object}	dto.ErrorResponse
//	@Router			/api/v1/users/{username}/posts [get]
func (h *PostHandler) GetUserPosts(c echo.Context) error {
	username := c.Param("username")

	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	offset, _ := strconv.Atoi(c.QueryParam("offset"))

	if limit == 0 {
		limit = 10 // default limit
	}

	posts, err := h.postUseCase.GetPosts(c.Request().Context(), username, limit, offset)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) || errors.Is(err, domain.ErrValidationFailed) {
			return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "User not found"})
		}
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "Something went wrong"})
	}

	response := make([]dto.PostResponse, 0, len(posts))
	for _, p := range posts {
		response = append(response, dto.PostResponse{
			ID:        p.ID,
			Title:     p.Title,
			Content:   p.Content,
			Tags:      p.Tags,
			Author:    username,
			CreatedAt: p.CreatedAt,
			UpdatedAt: p.UpdatedAt,
		})
	}

	return c.JSON(http.StatusOK, response)
}
