package usecase

import (
	"context"
	"errors"

	"github.com/billykore/project-one/internal/core/domain"
	"github.com/billykore/project-one/internal/core/ports"
)

type likeUseCase struct {
	likeRepo ports.LikeRepository
	postRepo ports.PostRepository
	log      ports.Logger
}

// NewLikeUseCase creates a new instance of ports.LikeUseCase.
func NewLikeUseCase(likeRepo ports.LikeRepository, postRepo ports.PostRepository, log ports.Logger) ports.LikeUseCase {
	if likeRepo == nil {
		panic("NewLikeUseCase: likeRepo is required")
	}
	if postRepo == nil {
		panic("NewLikeUseCase: postRepo is required")
	}
	if log == nil {
		panic("NewLikeUseCase: log is required")
	}
	return &likeUseCase{
		likeRepo: likeRepo,
		postRepo: postRepo,
		log:      log,
	}
}

func (u *likeUseCase) ToggleLike(ctx context.Context, postID int, username string) (bool, int, error) {
	// 1. Validate input.
	if postID <= 0 {
		return false, 0, domain.ErrInvalidPost
	}
	if username == "" {
		return false, 0, domain.ErrValidationFailed
	}

	// 2. Try to create the like.
	var liked bool
	like := &domain.Like{
		PostID:   postID,
		Username: username,
	}
	if err := u.likeRepo.Create(ctx, like); err != nil {
		if errors.Is(err, domain.ErrAlreadyLiked) {
			// Already liked, so unlike
			if err := u.likeRepo.Delete(ctx, postID, username); err != nil {
				if errors.Is(err, domain.ErrNotLiked) {
					return false, 0, err
				}
				u.log.Error(ctx, "failed to delete like", "postID", postID, "username", username, "error", err)
				return false, 0, domain.ErrInternalServer
			}
			liked = false
			// Decrement count
			if err := u.postRepo.IncrementLikeCount(ctx, postID, -1); err != nil {
				u.log.Error(ctx, "failed to decrement like count", "postID", postID, "error", err)
			}
			u.log.Info(ctx, "post unliked successfully", "postID", postID, "username", username)
		} else if errors.Is(err, domain.ErrPostNotFound) {
			return false, 0, err
		} else {
			u.log.Error(ctx, "failed to create like", "postID", postID, "username", username, "error", err)
			return false, 0, domain.ErrInternalServer
		}
	} else {
		// Like created successfully
		liked = true
		if err := u.postRepo.IncrementLikeCount(ctx, postID, 1); err != nil {
			u.log.Error(ctx, "failed to increment like count", "postID", postID, "error", err)
		}
		u.log.Info(ctx, "post liked successfully", "postID", postID, "username", username)
	}

	// 3. Get updated count from post.
	post, err := u.postRepo.GetByIDOnly(ctx, postID)
	if err != nil {
		u.log.Error(ctx, "failed to get post for like count", "postID", postID, "error", err)
		return false, 0, domain.ErrInternalServer
	}

	return liked, post.LikeCount, nil
}

func (u *likeUseCase) GetLikeStatus(ctx context.Context, postID int, username string) (bool, int, error) {
	// 1. Validate input.
	if postID <= 0 {
		return false, 0, domain.ErrInvalidPost
	}
	if username == "" {
		return false, 0, domain.ErrValidationFailed
	}

	// 2. Verify post exists.
	post, err := u.postRepo.GetByIDOnly(ctx, postID)
	if err != nil {
		if errors.Is(err, domain.ErrPostNotFound) {
			return false, 0, err
		}
		u.log.Error(ctx, "failed to verify post existence for like status", "postID", postID, "error", err)
		return false, 0, domain.ErrInternalServer
	}

	// 3. Check if user liked the post.
	liked, err := u.likeRepo.Exists(ctx, postID, username)
	if err != nil {
		u.log.Error(ctx, "failed to check like existence", "postID", postID, "username", username, "error", err)
		return false, 0, domain.ErrInternalServer
	}

	return liked, post.LikeCount, nil
}
