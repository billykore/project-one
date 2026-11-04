package repository

import (
	"context"
	"errors"
	"testing"
	"time"

	vo "github.com/billykore/project-one/internal/pkg/pagination"
	publishingdomain "github.com/billykore/project-one/internal/publishing/domain"
	"github.com/billykore/project-one/internal/social/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakePublishingPosts struct {
	posts []*publishingdomain.Post
	err   error

	gotUserIDs []int
	gotCursor  *vo.Cursor
	gotLimit   int
}

func (f *fakePublishingPosts) GetFeed(_ context.Context, userIDs []int, cursor *vo.Cursor, limit int) ([]*publishingdomain.Post, error) {
	f.gotUserIDs, f.gotCursor, f.gotLimit = userIDs, cursor, limit
	return f.posts, f.err
}

func TestNewTimelineReader_ProjectsPublishingPosts(t *testing.T) {
	created := time.Date(2026, 7, 1, 12, 0, 0, 0, time.UTC)
	updated := created.Add(time.Hour)
	cursor := &vo.Cursor{CreatedAt: created, ID: 9}
	posts := &fakePublishingPosts{posts: []*publishingdomain.Post{{
		ID:        5,
		UserID:    42,
		Username:  "alice",
		Title:     "Title",
		Content:   "Content",
		Tags:      []string{"go"},
		LikeCount: 3,
		CreatedAt: created,
		UpdatedAt: updated,
	}}}

	timeline, err := NewTimelineReader(posts).GetFeed(context.Background(), []int{1, 42}, cursor, 11)

	require.NoError(t, err)
	assert.Equal(t, []*domain.TimelinePost{{
		ID:        5,
		Author:    "alice",
		Title:     "Title",
		Content:   "Content",
		Tags:      []string{"go"},
		LikeCount: 3,
		CreatedAt: created,
		UpdatedAt: updated,
	}}, timeline)
	assert.Equal(t, []int{1, 42}, posts.gotUserIDs)
	assert.Same(t, cursor, posts.gotCursor)
	assert.Equal(t, 11, posts.gotLimit)
}

func TestNewTimelineReader_ReturnsEmptyTimeline(t *testing.T) {
	timeline, err := NewTimelineReader(&fakePublishingPosts{}).GetFeed(context.Background(), []int{1}, nil, 11)

	require.NoError(t, err)
	assert.NotNil(t, timeline)
	assert.Empty(t, timeline)
}

func TestNewTimelineReader_PassesThroughErrors(t *testing.T) {
	boom := errors.New("boom")

	_, err := NewTimelineReader(&fakePublishingPosts{err: boom}).GetFeed(context.Background(), []int{1}, nil, 11)

	assert.ErrorIs(t, err, boom)
}
