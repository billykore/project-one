package handler

import (
	"net/http"
	"strconv"

	"github.com/billykore/project-one/internal/api/dto"
	"github.com/billykore/project-one/internal/core/domain"
	"github.com/billykore/project-one/internal/core/ports"
	"github.com/labstack/echo/v4"
)

type CommentHandler struct {
	commentUseCase ports.CommentUseCase
	validator      ports.Validator
	log            ports.Logger
}

// ponytail: nil checks removed — Go panics at method call site on nil pointer
func NewCommentHandler(commentUseCase ports.CommentUseCase, validator ports.Validator, log ports.Logger) *CommentHandler {
	return &CommentHandler{
		commentUseCase: commentUseCase,
		validator:      validator,
		log:            log,
	}
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
//	@Failure		400		{object}	dto.APIErrorResponse
//	@Failure		401		{object}	dto.APIErrorResponse
//	@Failure		404		{object}	dto.APIErrorResponse
//	@Failure		500		{object}	dto.APIErrorResponse
//	@Security		BearerAuth
//	@Router			/comments/{id} [put]
func (h *CommentHandler) EditComment(c echo.Context) error {
	username, ok := c.Get("username").(string)
	if !ok {
		return domain.ErrUnauthorized
	}

	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid comment ID")
	}

	var req dto.EditCommentRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	if err := h.validator.Validate(req); err != nil {
		return err
	}

	err = h.commentUseCase.EditComment(c.Request().Context(), id, username, req.Content)
	if err != nil {
		return err
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
//	@Failure		400	{object}	dto.APIErrorResponse
//	@Failure		401	{object}	dto.APIErrorResponse
//	@Failure		404	{object}	dto.APIErrorResponse
//	@Failure		500	{object}	dto.APIErrorResponse
//	@Security		BearerAuth
//	@Router			/comments/{id} [delete]
func (h *CommentHandler) DeleteComment(c echo.Context) error {
	username, ok := c.Get("username").(string)
	if !ok {
		return domain.ErrUnauthorized
	}

	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid comment ID")
	}

	err = h.commentUseCase.DeleteComment(c.Request().Context(), id, username)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, dto.MessageResponse{Message: "Comment deleted successfully"})
}
