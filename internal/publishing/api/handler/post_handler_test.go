package handler

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	identitydomain "github.com/billykore/project-one/internal/identity/domain"
	platformports "github.com/billykore/project-one/internal/platform/ports"
	publishingdomain "github.com/billykore/project-one/internal/publishing/domain"
	"github.com/billykore/project-one/internal/testkit/mocks"
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

var _ platformports.Logger = discardPostLogger{}
var _ platformports.Validator = validPostRequest{}

func postContext(method, path, body string, user *identitydomain.User) (echo.Context, *httptest.ResponseRecorder) {
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
	user := &identitydomain.User{ID: 1, Username: "author"}

	t.Run("create", func(t *testing.T) {
		ctx, recorder := postContext(http.MethodPost, "/posts", `{"title":"title","content":"body"}`, user)
		commands.EXPECT().CreatePost(gomock.Any(), user, "title", "body", gomock.Any()).Return(&publishingdomain.Post{ID: 1}, nil)
		assert.NoError(t, handler.CreatePost(ctx))
		assert.Equal(t, http.StatusCreated, recorder.Code)
	})
	t.Run("update", func(t *testing.T) {
		ctx, recorder := postContext(http.MethodPut, "/posts/1", `{"title":"title"}`, user)
		ctx.SetParamNames("id")
		ctx.SetParamValues("1")
		commands.EXPECT().UpdatePost(gomock.Any(), 1, 1, "title", "").Return(&publishingdomain.Post{ID: 1}, nil)
		assert.NoError(t, handler.UpdatePost(ctx))
		assert.Equal(t, http.StatusOK, recorder.Code)
	})
	t.Run("delete", func(t *testing.T) {
		ctx, recorder := postContext(http.MethodDelete, "/posts/1", "", user)
		ctx.SetParamNames("id")
		ctx.SetParamValues("1")
		commands.EXPECT().DeletePost(gomock.Any(), 1, 1).Return(nil)
		assert.NoError(t, handler.DeletePost(ctx))
		assert.Equal(t, http.StatusOK, recorder.Code)
	})
	t.Run("comment", func(t *testing.T) {
		ctx, recorder := postContext(http.MethodPost, "/posts/1/comments", `{"id":1,"content":"comment"}`, user)
		comments.EXPECT().AddComment(gomock.Any(), 1, user, "comment").Return(nil)
		assert.NoError(t, handler.CreateComment(ctx))
		assert.Equal(t, http.StatusCreated, recorder.Code)
	})
	t.Run("like and unlike", func(t *testing.T) {
		ctx, recorder := postContext(http.MethodPost, "/posts/1/likes", "", user)
		ctx.SetParamNames("id")
		ctx.SetParamValues("1")
		commands.EXPECT().LikePost(gomock.Any(), 1, user).Return(2, nil)
		assert.NoError(t, handler.LikePost(ctx))
		assert.Equal(t, http.StatusOK, recorder.Code)

		ctx, recorder = postContext(http.MethodDelete, "/posts/1/likes", "", user)
		ctx.SetParamNames("id")
		ctx.SetParamValues("1")
		commands.EXPECT().UnlikePost(gomock.Any(), 1, user).Return(1, nil)
		assert.NoError(t, handler.UnlikePost(ctx))
		assert.Equal(t, http.StatusOK, recorder.Code)
	})
}

func TestPostQueryHandlerRoutes(t *testing.T) {
	ctrl := gomock.NewController(t)
	queries := mocks.NewMockPostQueryUseCase(ctrl)
	comments := mocks.NewMockCommentUseCase(ctrl)
	handler := NewPostQueryHandler(queries, comments, discardPostLogger{})
	user := &identitydomain.User{ID: 1, Username: "reader"}
	post := &publishingdomain.Post{ID: 1, Username: "author", Title: "title", CreatedAt: time.Now()}

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
		queries.EXPECT().GetPosts(gomock.Any(), 1, nil, 10).Return([]*publishingdomain.Post{post}, nil, false, nil)
		assert.NoError(t, handler.GetPosts(ctx))
		assert.Equal(t, http.StatusOK, recorder.Code)
	})
	t.Run("like status", func(t *testing.T) {
		ctx, recorder := postContext(http.MethodGet, "/posts/1/likes", "", user)
		ctx.SetParamNames("id")
		ctx.SetParamValues("1")
		queries.EXPECT().GetLikeStatus(gomock.Any(), 1, user.ID).Return(true, 2, nil)
		assert.NoError(t, handler.GetLikeStatus(ctx))
		assert.Equal(t, http.StatusOK, recorder.Code)
	})
}
