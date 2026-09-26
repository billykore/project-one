package handler

import (
	"net/http"
	"strconv"

	request "github.com/billykore/project-one/internal/platform/api/request"
	platformports "github.com/billykore/project-one/internal/platform/ports"
	"github.com/billykore/project-one/internal/publishing/api/dto"
	publishingports "github.com/billykore/project-one/internal/publishing/ports"
	"github.com/labstack/echo/v4"
)

type CommentHandler struct {
	commentUseCase publishingports.CommentUseCase
	validator      platformports.Validator
	log            platformports.Logger
}

// ponytail: nil checks removed — Go panics at method call site on nil pointer
func NewCommentHandler(commentUseCase publishingports.CommentUseCase, validator platformports.Validator, log platformports.Logger) *CommentHandler {
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
//	@Failure		400		{object}	dto.ProblemDetail
//	@Failure		401		{object}	dto.ProblemDetail
//	@Failure		404		{object}	dto.ProblemDetail
//	@Failure		500		{object}	dto.ProblemDetail
//	@Security		BearerAuth
//	@Router			/comments/{id} [put]
func (h *CommentHandler) EditComment(c echo.Context) error {
	user, ok := request.CurrentUser(c)
	if !ok {
		h.log.Error(c.Request().Context(), "EditComment failed", "error", "User not found in context")
		return echo.ErrUnauthorized
	}

	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.log.Error(c.Request().Context(), "EditComment failed", "username", user.Username, "error", "Invalid comment ID")
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid comment ID")
	}

	var req dto.EditCommentRequest
	if err := c.Bind(&req); err != nil {
		h.log.Error(c.Request().Context(), "EditComment failed", "username", user.Username, "comment_id", id, "error", "Invalid request body")
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	if err := h.validator.Validate(req); err != nil {
		h.log.Error(c.Request().Context(), "EditComment failed", "username", user.Username, "comment_id", id, "validation_error", err)
		return err
	}

	err = h.commentUseCase.EditComment(c.Request().Context(), id, user.ID, req.Content)
	if err != nil {
		h.log.Error(c.Request().Context(), "EditComment failed", "username", user.Username, "comment_id", id, "error", err)
		return err
	}

	h.log.Info(c.Request().Context(), "EditComment succeeded", "username", user.Username, "comment_id", id)
	return c.JSON(http.StatusOK, dto.MessageResponse{Message: "Comment updated succesfully"})
}

// DeleteComment handles the DELETE /comments/:id endpoint.
//
//	@Summary		Delete comment
//	@Description	Delete an existing comment by the author.
//	@Tags			comments
//	@Param			id	path		int	true	"Comment ID"
//	@Success		200	{object}	dto.MessageResponse
//	@Failure		400	{object}	dto.ProblemDetail
//	@Failure		401	{object}	dto.ProblemDetail
//	@Failure		404	{object}	dto.ProblemDetail
//	@Failure		500	{object}	dto.ProblemDetail
//	@Security		BearerAuth
//	@Router			/comments/{id} [delete]
func (h *CommentHandler) DeleteComment(c echo.Context) error {
	user, ok := request.CurrentUser(c)
	if !ok {
		h.log.Error(c.Request().Context(), "DeleteComment failed", "error", "User not found in context")
		return echo.ErrUnauthorized
	}

	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		h.log.Error(c.Request().Context(), "DeleteComment failed", "username", user.Username, "error", "Invalid comment ID")
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid comment ID")
	}

	err = h.commentUseCase.DeleteComment(c.Request().Context(), id, user.ID)
	if err != nil {
		h.log.Error(c.Request().Context(), "DeleteComment failed", "username", user.Username, "comment_id", id, "error", err)
		return err
	}

	h.log.Info(c.Request().Context(), "DeleteComment succeeded", "username", user.Username, "comment_id", id)
	return c.JSON(http.StatusOK, dto.MessageResponse{Message: "Comment deleted successfully"})
}
