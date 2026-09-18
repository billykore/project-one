package usecase

import (
	"context"
	"testing"

	"github.com/billykore/project-one/internal/core/domain"
	"github.com/billykore/project-one/internal/core/ports/mocks"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestPostQueryUseCase_RepeatedReadsAreMutationFree(t *testing.T) {
	ctrl := gomock.NewController(t)
	posts := mocks.NewMockPostQueryRepository(ctrl)
	likes := mocks.NewMockLikeRepository(ctrl)
	log := mocks.NewMockLogger(ctrl)
	query := NewPostQueryUseCase(posts, likes, log)
	ctx := context.Background()

	posts.EXPECT().GetByID(ctx, 1).Return(&domain.Post{ID: 1, LikeCount: 2}, nil).Times(10)
	likes.EXPECT().Exists(ctx, 1, "reader").Return(true, nil).Times(10)
	for range 10 {
		liked, count, err := query.GetLikeStatus(ctx, 1, "reader")
		require.NoError(t, err)
		require.True(t, liked)
		require.Equal(t, 2, count)
	}
}

func TestPostQueryUseCase_PostAndPageReads(t *testing.T) {
	ctrl := gomock.NewController(t)
	posts := mocks.NewMockPostQueryRepository(ctrl)
	likes := mocks.NewMockLikeRepository(ctrl)
	log := mocks.NewMockLogger(ctrl)
	query := NewPostQueryUseCase(posts, likes, log)
	ctx := context.Background()

	posts.EXPECT().GetByID(ctx, 1).Return(&domain.Post{ID: 1}, nil).Times(10)
	for range 10 {
		post, err := query.GetPostByID(ctx, 1)
		require.NoError(t, err)
		require.Equal(t, 1, post.ID)
	}

	posts.EXPECT().GetUserPosts(ctx, "author", nil, 2).Return([]*domain.Post{{ID: 2}}, nil).Times(10)
	for range 10 {
		page, _, _, err := query.GetPosts(ctx, "author", nil, 1)
		require.NoError(t, err)
		require.Len(t, page, 1)
	}
}
