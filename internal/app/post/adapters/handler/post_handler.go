package handler

import (
	"fmt"
	"net/http"

	"github.com/billykore/project-one/internal/app/post/core/ports"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
)

type postHandler struct {
	postSvc   ports.PostService
	validator *validator.Validate
}

func NewPostHandler(postSvc ports.PostService, validator *validator.Validate) *postHandler {
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
// @Param        request body dto.CreatePostRequest true "Post details"
// @Success      201  {object}  dto.CreatePostResponse
// @Failure      400  {object}  dto.ErrorResponse
// @Failure      401  {object}  dto.ErrorResponse
// @Failure      500  {object}  dto.ErrorResponse
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

	if err := h.validator.Struct(req); err != nil {
		for _, err := range err.(validator.ValidationErrors) {
			if err.Field() == "Title" && err.Tag() == "required" {
				return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Title is required"})
			}
			if err.Field() == "Content" && err.Tag() == "min" {
				return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Content must be 10 characters minimum"})
			}
		}
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
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
