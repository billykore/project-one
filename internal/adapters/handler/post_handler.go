package handler

import (
	"fmt"
	"net/http"

	"github.com/billykore/project-one/internal/core/ports"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
)

type postHandler struct {
	postSvc   ports.PostService
	validator ports.Validator
}

func NewPostHandler(postSvc ports.PostService, validator ports.Validator) *postHandler {
	return &postHandler{
		postSvc:   postSvc,
		validator: validator,
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
func (h *postHandler) CreatePost(c echo.Context) error {
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

	post, err := h.postSvc.CreatePost(c.Request().Context(), userID, req.Title, req.Content, req.Tags)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Something went wrong"})
	}

	return c.JSON(http.StatusCreated, CreatePostResponse{
		Message:     "Post created successfully",
		RedirectURL: fmt.Sprintf("/posts/%d", post.ID),
	})
}
