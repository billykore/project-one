package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/billykore/project-one/internal/api/dto"
	"github.com/billykore/project-one/internal/core/domain"
	"github.com/billykore/project-one/internal/core/ports"
	"github.com/billykore/project-one/internal/core/ports/mocks"
	vo "github.com/billykore/project-one/internal/core/valueobject"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestHandleGetFeed_Unauthorized(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	feedUc := mocks.NewMockFeedUseCase(ctrl)
	logger := mocks.NewMockLogger(ctrl)

	h := NewFeedHandler(feedUc, logger)
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/feeds", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.HandleGetFeed(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestHandleGetFeed_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	feedUc := mocks.NewMockFeedUseCase(ctrl)
	logger := mocks.NewMockLogger(ctrl)

	now := time.Date(2026, 7, 1, 12, 0, 0, 0, time.UTC)

	h := NewFeedHandler(feedUc, logger)
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/feeds", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("username", "alice")

	feedUc.EXPECT().GetFeed(gomock.Any(), "alice", (*vo.Cursor)(nil), 10).
		Return(&ports.FeedResult{
			Posts: []*domain.Post{
				{ID: 1, Username: "alice", Title: "Hello", Content: "World", Tags: []string{"go"}, LikeCount: 3, CreatedAt: now, UpdatedAt: now},
			},
			HasMore: false,
		}, nil)

	err := h.HandleGetFeed(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var resp dto.FeedResponse
	assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Len(t, resp.Data, 1)
	assert.Equal(t, "Hello", resp.Data[0].Title)
	assert.Equal(t, "alice", resp.Data[0].Author)
	assert.False(t, resp.HasMore)
}

func TestHandleGetFeed_InvalidLimit(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	feedUc := mocks.NewMockFeedUseCase(ctrl)
	logger := mocks.NewMockLogger(ctrl)

	h := NewFeedHandler(feedUc, logger)
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/feeds?limit=0", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("username", "alice")

	err := h.HandleGetFeed(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestHandleGetFeed_InvalidCursor(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	feedUc := mocks.NewMockFeedUseCase(ctrl)
	logger := mocks.NewMockLogger(ctrl)

	h := NewFeedHandler(feedUc, logger)
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/feeds?cursor=!!!invalid!!!", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("username", "alice")

	err := h.HandleGetFeed(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestHandleGetFeed_WithCursor(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	feedUc := mocks.NewMockFeedUseCase(ctrl)
	logger := mocks.NewMockLogger(ctrl)

	now := time.Date(2026, 7, 1, 12, 0, 0, 0, time.UTC)
	cursorStr := vo.Cursor{CreatedAt: now, ID: 10}.Encode()

	h := NewFeedHandler(feedUc, logger)
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/feeds?cursor="+cursorStr+"&limit=5", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("username", "alice")

	expectedCursor := &vo.Cursor{CreatedAt: now, ID: 10}
	feedUc.EXPECT().GetFeed(gomock.Any(), "alice", expectedCursor, 5).
		Return(&ports.FeedResult{
			Posts:   []*domain.Post{},
			HasMore: false,
		}, nil)

	err := h.HandleGetFeed(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
}
