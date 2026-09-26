package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/billykore/project-one/internal/core/domain"
	"github.com/billykore/project-one/internal/core/ports"
	vo "github.com/billykore/project-one/internal/core/valueobject"
)

type postQueryUseCase struct {
	postRepo ports.PostQueryRepository
	likeRepo ports.LikeRepository
	log      ports.Logger
}

func NewPostQueryUseCase(postRepo ports.PostQueryRepository, likeRepo ports.LikeRepository, log ports.Logger) ports.PostQueryUseCase {
	if postRepo == nil || likeRepo == nil || log == nil {
		panic("NewPostQueryUseCase: dependencies must not be nil")
	}
	return &postQueryUseCase{postRepo: postRepo, likeRepo: likeRepo, log: log}
}

func (uc *postQueryUseCase) GetPostByID(ctx context.Context, id int) (*domain.Post, error) {
	if id <= 0 {
		return nil, domain.ErrInvalidPost
	}
	post, err := uc.postRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, domain.ErrPostNotFound) {
			return nil, err
		}
		uc.log.Error(ctx, "failed to get post by id", "postID", id, "error", err)
		return nil, fmt.Errorf("get post by id: %w", domain.ErrRepositoryFailure)
	}
	return post, nil
}

func (uc *postQueryUseCase) GetPosts(ctx context.Context, userID int, cursor *vo.Cursor, limit int) ([]*domain.Post, *vo.Cursor, bool, error) {
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}
	posts, err := uc.postRepo.GetUserPosts(ctx, userID, cursor, limit+1)
	if err != nil {
		uc.log.Error(ctx, "failed to get posts for user", "userID", userID, "error", err)
		return nil, nil, false, fmt.Errorf("get posts for user: %w", domain.ErrRepositoryFailure)
	}
	var nextCursor *vo.Cursor
	hasMore := false
	if len(posts) > limit {
		hasMore = true
		posts = posts[:limit]
		last := posts[len(posts)-1]
		nextCursor = &vo.Cursor{CreatedAt: last.CreatedAt, ID: last.ID}
	}
	return posts, nextCursor, hasMore, nil
}

func (uc *postQueryUseCase) GetLikeStatus(ctx context.Context, postID int, userID int) (bool, int, error) {
	if postID <= 0 {
		return false, 0, domain.ErrInvalidPost
	}
	if userID <= 0 {
		return false, 0, domain.ErrInvalidUsername
	}
	post, err := uc.postRepo.GetByID(ctx, postID)
	if err != nil {
		if errors.Is(err, domain.ErrPostNotFound) {
			return false, 0, err
		}
		uc.log.Error(ctx, "failed to verify post existence for like status", "postID", postID, "error", err)
		return false, 0, fmt.Errorf("verify post existence: %w", err)
	}
	liked, err := uc.likeRepo.Exists(ctx, postID, userID)
	if err != nil {
		uc.log.Error(ctx, "failed to check like existence", "postID", postID, "userID", userID, "error", err)
		return false, 0, fmt.Errorf("check like existence: %w", err)
	}
	return liked, post.LikeCount, nil
}
