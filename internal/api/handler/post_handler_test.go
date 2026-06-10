package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/billykore/project-one/internal/api/dto"
	"github.com/billykore/project-one/internal/core/domain"
	"github.com/billykore/project-one/internal/core/usecase/mocks"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestPostHandler_GetPostByID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPostUC := mocks.NewMockPostUseCase(ctrl)
	mockCommentUC := mocks.NewMockCommentUseCase(ctrl)
	mockValidator := mocks.NewMockValidator(ctrl)

	h := NewPostHandler(mockPostUC, mockCommentUC, mockValidator)

	postID := 1

	t.Run("success - authenticated", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/posts/1", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/posts/:id")
		c.SetParamNames("id")
		c.SetParamValues("1")
		c.Set("username", "testuser")

		expectedPost := &domain.Post{
			ID:       postID,
			Username: "authoruser",
			Title:    "Post Title",
			Content:  "Post Content",
			Tags:     []string{"tag1"},
		}
		expectedComments := []*domain.Comment{
			{
				ID:       1,
				Username: "commenter",
				Content:  "Nice post!",
			},
		}

		mockPostUC.EXPECT().GetPostByID(gomock.Any(), postID).Return(expectedPost, nil)
		mockCommentUC.EXPECT().GetCommentsByPostID(gomock.Any(), postID).Return(expectedComments, nil)

		err := h.GetPostByID(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		var resp dto.PostResponse
		err = json.Unmarshal(rec.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, postID, resp.ID)
		assert.Equal(t, "authoruser", resp.Author)
		assert.Len(t, resp.Comments, 1)
		assert.Equal(t, "commenter", resp.Comments[0].Username)
	})

	t.Run("success - unauthenticated", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/posts/1", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/posts/:id")
		c.SetParamNames("id")
		c.SetParamValues("1")

		expectedPost := &domain.Post{
			ID:       postID,
			Username: "authoruser",
			Title:    "Post Title",
			Content:  "Post Content",
		}

		mockPostUC.EXPECT().GetPostByID(gomock.Any(), postID).Return(expectedPost, nil)
		mockCommentUC.EXPECT().GetCommentsByPostID(gomock.Any(), postID).Return([]*domain.Comment{}, nil)

		err := h.GetPostByID(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		var resp dto.PostResponse
		err = json.Unmarshal(rec.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, postID, resp.ID)
		assert.Equal(t, "authoruser", resp.Author)
		assert.Empty(t, resp.Comments)
	})

	t.Run("bad request - invalid post id format", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/posts/invalid", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/posts/:id")
		c.SetParamNames("id")
		c.SetParamValues("invalid")

		err := h.GetPostByID(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("not found", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/posts/1", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/posts/:id")
		c.SetParamNames("id")
		c.SetParamValues("1")

		mockPostUC.EXPECT().GetPostByID(gomock.Any(), postID).Return(nil, domain.ErrPostNotFound)

		err := h.GetPostByID(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, rec.Code)
	})

	t.Run("internal server error - post usecase error", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/posts/1", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/posts/:id")
		c.SetParamNames("id")
		c.SetParamValues("1")

		mockPostUC.EXPECT().GetPostByID(gomock.Any(), postID).Return(nil, errors.New("usecase error"))

		err := h.GetPostByID(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	})

	t.Run("internal server error - comments error", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/posts/1", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/posts/:id")
		c.SetParamNames("id")
		c.SetParamValues("1")

		expectedPost := &domain.Post{ID: postID}
		mockPostUC.EXPECT().GetPostByID(gomock.Any(), postID).Return(expectedPost, nil)
		mockCommentUC.EXPECT().GetCommentsByPostID(gomock.Any(), postID).Return(nil, errors.New("comment error"))

		err := h.GetPostByID(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	})
}
