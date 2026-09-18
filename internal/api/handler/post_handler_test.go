package handler

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/billykore/project-one/internal/core/domain"
	"github.com/billykore/project-one/internal/core/ports"
	"github.com/billykore/project-one/internal/core/ports/mocks"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type discardPostLogger struct{}

func (discardPostLogger) Debug(context.Context, string, ...any) {}
func (discardPostLogger) Info(context.Context, string, ...any)  {}
func (discardPostLogger) Warn(context.Context, string, ...any)  {}
func (discardPostLogger) Error(context.Context, string, ...any) {}
func (discardPostLogger) Fatal(context.Context, string, ...any) {}

type validPostRequest struct{}

func (validPostRequest) Validate(any) error { return nil }

var _ ports.Logger = discardPostLogger{}
var _ ports.Validator = validPostRequest{}

func postContext(method, path, body string, user *domain.User) (echo.Context, *httptest.ResponseRecorder) {
	e := echo.New()
	recorder := httptest.NewRecorder()
	ctx := e.NewContext(httptest.NewRequest(method, path, bytes.NewBufferString(body)), recorder)
	ctx.Request().Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	ctx.Set("user", user)
	return ctx, recorder
}

func TestPostCommandHandlerRoutes(t *testing.T) {
	ctrl := gomock.NewController(t)
	commands := mocks.NewMockPostCommandUseCase(ctrl)
	comments := mocks.NewMockCommentUseCase(ctrl)
	handler := NewPostCommandHandler(commands, comments, validPostRequest{}, discardPostLogger{})
	user := &domain.User{ID: 1, Username: "author"}

	t.Run("create", func(t *testing.T) {
		ctx, recorder := postContext(http.MethodPost, "/posts", `{"title":"title","content":"body"}`, user)
		commands.EXPECT().CreatePost(gomock.Any(), user, "title", "body", gomock.Any()).Return(&domain.Post{ID: 1}, nil)
		assert.NoError(t, handler.CreatePost(ctx))
		assert.Equal(t, http.StatusCreated, recorder.Code)
	})
	t.Run("update", func(t *testing.T) {
		ctx, recorder := postContext(http.MethodPut, "/posts/1", `{"title":"title"}`, user)
		ctx.SetParamNames("id")
		ctx.SetParamValues("1")
		commands.EXPECT().UpdatePost(gomock.Any(), "author", 1, "title", "").Return(&domain.Post{ID: 1}, nil)
		assert.NoError(t, handler.UpdatePost(ctx))
		assert.Equal(t, http.StatusOK, recorder.Code)
	})
	t.Run("delete", func(t *testing.T) {
		ctx, recorder := postContext(http.MethodDelete, "/posts/1", "", user)
		ctx.SetParamNames("id")
		ctx.SetParamValues("1")
		commands.EXPECT().DeletePost(gomock.Any(), "author", 1).Return(nil)
		assert.NoError(t, handler.DeletePost(ctx))
		assert.Equal(t, http.StatusOK, recorder.Code)
	})
	t.Run("comment", func(t *testing.T) {
		ctx, recorder := postContext(http.MethodPost, "/posts/1/comments", `{"id":1,"content":"comment"}`, user)
		comments.EXPECT().AddComment(gomock.Any(), 1, "author", "comment").Return(nil)
		assert.NoError(t, handler.CreateComment(ctx))
		assert.Equal(t, http.StatusCreated, recorder.Code)
	})
	t.Run("like and unlike", func(t *testing.T) {
		ctx, recorder := postContext(http.MethodPost, "/posts/1/likes", "", user)
		ctx.SetParamNames("id")
		ctx.SetParamValues("1")
		commands.EXPECT().LikePost(gomock.Any(), 1, "author").Return(2, nil)
		assert.NoError(t, handler.LikePost(ctx))
		assert.Equal(t, http.StatusOK, recorder.Code)

		ctx, recorder = postContext(http.MethodDelete, "/posts/1/likes", "", user)
		ctx.SetParamNames("id")
		ctx.SetParamValues("1")
		commands.EXPECT().UnlikePost(gomock.Any(), 1, "author").Return(1, nil)
		assert.NoError(t, handler.UnlikePost(ctx))
		assert.Equal(t, http.StatusOK, recorder.Code)
	})
}

func TestPostQueryHandlerRoutes(t *testing.T) {
	ctrl := gomock.NewController(t)
	queries := mocks.NewMockPostQueryUseCase(ctrl)
	comments := mocks.NewMockCommentUseCase(ctrl)
	handler := NewPostQueryHandler(queries, comments, discardPostLogger{})
	user := &domain.User{Username: "reader"}
	post := &domain.Post{ID: 1, Username: "author", Title: "title", CreatedAt: time.Now()}

	t.Run("detail", func(t *testing.T) {
		ctx, recorder := postContext(http.MethodGet, "/posts/1", "", user)
		ctx.SetParamNames("id")
		ctx.SetParamValues("1")
		queries.EXPECT().GetPostByID(gomock.Any(), 1).Return(post, nil)
		comments.EXPECT().GetCommentsByPostID(gomock.Any(), 1).Return(nil, nil)
		assert.NoError(t, handler.GetPostByID(ctx))
		assert.Equal(t, http.StatusOK, recorder.Code)
	})
	t.Run("list", func(t *testing.T) {
		ctx, recorder := postContext(http.MethodGet, "/posts", "", user)
		queries.EXPECT().GetPosts(gomock.Any(), "reader", nil, 10).Return([]*domain.Post{post}, nil, false, nil)
		assert.NoError(t, handler.GetPosts(ctx))
		assert.Equal(t, http.StatusOK, recorder.Code)
	})
	t.Run("like status", func(t *testing.T) {
		ctx, recorder := postContext(http.MethodGet, "/posts/1/likes", "", user)
		ctx.SetParamNames("id")
		ctx.SetParamValues("1")
		queries.EXPECT().GetLikeStatus(gomock.Any(), 1, "reader").Return(true, 2, nil)
		assert.NoError(t, handler.GetLikeStatus(ctx))
		assert.Equal(t, http.StatusOK, recorder.Code)
	})
}
