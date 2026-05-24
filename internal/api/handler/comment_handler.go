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

type CommentHandler struct {
	commentUseCase ports.CommentUseCase
	validator      ports.Validator
	log            ports.Logger
}

func NewCommentHandler(commentUseCase ports.CommentUseCase, validator ports.Validator, log ports.Logger) *CommentHandler {
	if commentUseCase == nil {
		panic("commentUseCase is required")
	}
	if validator == nil {
		panic("validator is required")
	}
	if log == nil {
		panic("log is required")
	}
	return &CommentHandler{
		commentUseCase: commentUseCase,
		validator:      validator,
		log:            log,
	}
}

// CreateComment handles the POST /posts/:id/comments endpoint.
//
//	@Summary		Create comment
//	@Description	Create a new comment on a post.
//	@Tags			comments
//	@Accept			json
//	@Produce		json
//	@Param			id		path	int							true	"Post ID"
//	@Param			request	body	dto.CreateCommentRequest	true	"Comment details"
//	@Success		201		"Created"
//	@Failure		400		{object}	dto.ErrorResponse
//	@Failure		401		{object}	dto.ErrorResponse
//	@Failure		404		{object}	dto.ErrorResponse
//	@Failure		500		{object}	dto.ErrorResponse
//	@Security		BearerAuth
//	@Router			/posts/{id}/comments [post]
func (h *CommentHandler) CreateComment(c echo.Context) error {
	username, ok := c.Get("username").(string)
	if !ok {
		return c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "Unauthorized"})
	}

	var req dto.CreateCommentRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "Invalid request body"})
	}

	if err := h.validator.Validate(req); err != nil {
		validationErrs, ok := err.(validator.ValidationErrors)
		if !ok {
			return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "Invalid request body"})
		}

		for _, err := range validationErrs {
			if err.Field() == "Content" && (err.Tag() == "required" || err.Tag() == "min") {
				return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "Comment must be at least 1 character"})
			}
		}
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "Validation failed"})
	}

	err := h.commentUseCase.AddComment(c.Request().Context(), req.ID, username, req.Content)
	if err != nil {
		if errors.Is(err, domain.ErrPostNotFound) {
			return c.JSON(http.StatusNotFound, dto.ErrorResponse{Error: "Post not found"})
		}
		if errors.Is(err, domain.ErrValidationFailed) {
			return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "Comment must be at least 1 character"})
		}
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "Something went wrong"})
	}

	return c.NoContent(http.StatusCreated)
}

// EditComment handles the PUT /comments/:id endpoint.
//
//	@Summary		Edit comment
//	@Description	Edit an existing comment by the author.
//	@Tags			comments
//	@Accept			json
//	@Produce		json
//	@Param			id		path		int						true	"Comment ID"
//	@Param			request	body		dto.EditCommentRequest	true	"Comment update content"
//	@Success		200		{object}	dto.MessageResponse
//	@Failure		400		{object}	dto.ErrorResponse
//	@Failure		401		{object}	dto.ErrorResponse
//	@Failure		404		{object}	dto.ErrorResponse
//	@Failure		500		{object}	dto.ErrorResponse
//	@Security		BearerAuth
//	@Router			/comments/{id} [put]
func (h *CommentHandler) EditComment(c echo.Context) error {
	username, ok := c.Get("username").(string)
	if !ok {
		return c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "Unauthorized"})
	}

	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "Invalid request body"})
	}

	var req dto.EditCommentRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "Invalid request body"})
	}

	if err := h.validator.Validate(req); err != nil {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "Invalid request body"})
	}

	err = h.commentUseCase.EditComment(c.Request().Context(), id, username, req.Content)
	if err != nil {
		if errors.Is(err, domain.ErrCommentNotFound) {
			return c.JSON(http.StatusNotFound, dto.ErrorResponse{Error: "Comment not found"})
		}
		if errors.Is(err, domain.ErrUnauthorized) {
			return c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "Unauthorized"})
		}
		if errors.Is(err, domain.ErrValidationFailed) {
			return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "Invalid request body"})
		}
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "Something went wrong"})
	}

	return c.JSON(http.StatusOK, dto.MessageResponse{Message: "Comment updated succesfully"})
}

// DeleteComment handles the DELETE /comments/:id endpoint.
//
//	@Summary		Delete comment
//	@Description	Delete an existing comment by the author.
//	@Tags			comments
//	@Param			id	path		int	true	"Comment ID"
//	@Success		200	{object}	dto.MessageResponse
//	@Failure		400	{object}	dto.ErrorResponse
//	@Failure		401	{object}	dto.ErrorResponse
//	@Failure		404	{object}	dto.ErrorResponse
//	@Failure		500	{object}	dto.ErrorResponse
//	@Security		BearerAuth
//	@Router			/comments/{id} [delete]
func (h *CommentHandler) DeleteComment(c echo.Context) error {
	username, ok := c.Get("username").(string)
	if !ok {
		return c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "Unauthorized"})
	}

	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "Invalid comment ID"})
	}

	err = h.commentUseCase.DeleteComment(c.Request().Context(), id, username)
	if err != nil {
		if errors.Is(err, domain.ErrCommentNotFound) {
			return c.JSON(http.StatusNotFound, dto.ErrorResponse{Error: "Comment not found"})
		}
		if errors.Is(err, domain.ErrUnauthorized) {
			return c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "Unauthorized"})
		}
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "Something went wrong"})
	}

	return c.JSON(http.StatusOK, dto.MessageResponse{Message: "Comment deleted successfully"})
}
