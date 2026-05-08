package handler

import (
	"errors"
	"net/http"
	"strconv"

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
// @Summary      Create post
// @Description  Create a new post for the authenticated user.
// @Tags         posts
// @Accept       json
// @Produce      json
// @Param        request body CreatePostRequest true "Post details"
// @Success      201  {object}  CreatePostResponse
// @Failure      400  {object}  ErrorResponse
// @Failure      401  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Security     BearerAuth
// @Router       /posts [post]
func (h *PostHandler) CreatePost(c echo.Context) error {
	userID, ok := c.Get("userID").(int)
	if !ok {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Unauthorized"})
	}

	var req CreatePostRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid request body"})
	}

	if err := h.validator.Validate(req); err != nil {
		validationErrs, ok := err.(validator.ValidationErrors)
		if !ok {
			return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid request body"})
		}

		for _, err := range validationErrs {
			if err.Field() == "Title" && err.Tag() == "required" {
				return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Title is required"})
			}
			if err.Field() == "Content" && err.Tag() == "min" {
				return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Content must be 10 characters minimum"})
			}
		}
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Validation failed"})
	}

	post, err := h.postUseCase.CreatePost(c.Request().Context(), userID, req.Title, req.Content, req.Tags)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Something went wrong"})
	}

	return c.JSON(http.StatusCreated, CreatePostResponse{
		ID:      post.ID,
		Message: "Post created successfully",
	})
}

// GetPostByID handles the GET /posts/:id endpoint.
// @Summary      Get post by ID
// @Description  Retrieve a specific post by its ID.
// @Tags         posts
// @Produce      json
// @Param        id   path      int  true  "Post ID"
// @Success      200  {object}  PostResponse
// @Failure      400  {object}  ErrorResponse
// @Failure      401  {object}  ErrorResponse
// @Failure      403  {object}  ErrorResponse
// @Failure      404  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /posts/{id} [get]
func (h *PostHandler) GetPostByID(c echo.Context) error {
	userID, ok := c.Get("userID").(int)
	if !ok {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Unauthorized"})
	}

	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Post ID must be a number"})
	}

	post, err := h.postUseCase.GetPostByID(c.Request().Context(), userID, id)
	if err != nil {
		if errors.Is(err, domain.ErrPostNotFound) {
			return c.JSON(http.StatusNotFound, ErrorResponse{Error: "Post not found"})
		}
		if errors.Is(err, domain.ErrInvalidPost) {
			return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Post ID must be integer and not 0"})
		}
		if errors.Is(err, domain.ErrUnauthorized) {
			return c.JSON(http.StatusForbidden, ErrorResponse{Error: "You do not have permission to access this post"})
		}
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Something went wrong"})
	}

	return c.JSON(http.StatusOK, PostResponse{
		ID:        post.ID,
		Message:   post.Title,
		Content:   post.Content,
		Tags:      post.Tags,
		CreatedAt: post.CreatedAt,
		UpdatedAt: post.UpdatedAt,
	})
}

// UpdatePost handles the PUT /posts/:id endpoint.
// @Summary      Update post
// @Description  Update an existing post for the authenticated user.
// @Tags         posts
// @Accept       json
// @Produce      json
// @Param        id       path      int                true  "Post ID"
// @Param        request  body      UpdatePostRequest  true  "Post details"
// @Success      200      {object}  PostResponse
// @Failure      400      {object}  ErrorResponse
// @Failure      401      {object}  ErrorResponse
// @Failure      403      {object}  ErrorResponse
// @Failure      404      {object}  ErrorResponse
// @Failure      500      {object}  ErrorResponse
// @Security     BearerAuth
// @Router       /posts/{id} [put]
func (h *PostHandler) UpdatePost(c echo.Context) error {
	userID, ok := c.Get("userID").(int)
	if !ok {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Unauthorized"})
	}

	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Post ID must be a number"})
	}

	var req UpdatePostRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid request body"})
	}

	post, err := h.postUseCase.UpdatePost(c.Request().Context(), userID, id, req.Title, req.Content)
	if err != nil {
		if errors.Is(err, domain.ErrPostNotFound) {
			return c.JSON(http.StatusNotFound, ErrorResponse{Error: "Post not found"})
		}
		if errors.Is(err, domain.ErrInvalidPost) {
			return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Post ID must be integer and not 0"})
		}
		if errors.Is(err, domain.ErrUnauthorized) {
			return c.JSON(http.StatusForbidden, ErrorResponse{Error: "You do not have permission to update this post"})
		}
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Something went wrong"})
	}

	return c.JSON(http.StatusOK, PostResponse{
		ID:        post.ID,
		Message:   "Post updated successfully",
		Content:   post.Content,
		Tags:      post.Tags,
		CreatedAt: post.CreatedAt,
		UpdatedAt: post.UpdatedAt,
	})
}

// DeletePost handles the DELETE /posts/:id endpoint.
// @Summary      Delete post
// @Description  Soft delete a post for the authenticated user.
// @Tags         posts
// @Param        id   path      int  true  "Post ID"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  ErrorResponse
// @Failure      401  {object}  ErrorResponse
// @Failure      404  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Security     BearerAuth
// @Router       /posts/{id} [delete]
func (h *PostHandler) DeletePost(c echo.Context) error {
	userID, ok := c.Get("userID").(int)
	if !ok {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Unauthorized"})
	}

	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Post ID must be a number"})
	}

	err = h.postUseCase.DeletePost(c.Request().Context(), userID, id)
	if err != nil {
		if errors.Is(err, domain.ErrPostNotFound) {
			return c.JSON(http.StatusNotFound, ErrorResponse{Error: "Post not found"})
		}
		if errors.Is(err, domain.ErrInvalidPost) {
			return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Post ID must be integer and not 0"})
		}
		if errors.Is(err, domain.ErrUnauthorized) {
			return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Unauthorized"})
		}
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Something went wrong"})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"id":      id,
		"message": "Post deleted successfully",
	})
}
