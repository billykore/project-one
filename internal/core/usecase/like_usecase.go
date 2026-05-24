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

	// 2. Verify post exists (uses GetByIDOnly — no owner check needed).
	_, err := u.postRepo.GetByIDOnly(ctx, postID)
	if err != nil {
		if errors.Is(err, domain.ErrPostNotFound) {
			return false, 0, err
		}
		u.log.Error(ctx, "failed to verify post existence for like", "postID", postID, "error", err)
		return false, 0, domain.ErrInternalServer
	}

	// 3. Check if user already liked the post.
	alreadyLiked, err := u.likeRepo.Exists(ctx, postID, username)
	if err != nil {
		u.log.Error(ctx, "failed to check like existence", "postID", postID, "username", username, "error", err)
		return false, 0, domain.ErrInternalServer
	}

	// 4. Toggle: unlike if already liked, like if not.
	if alreadyLiked {
		// Unlike
		if err := u.likeRepo.Delete(ctx, postID, username); err != nil {
			u.log.Error(ctx, "failed to delete like", "postID", postID, "username", username, "error", err)
			return false, 0, domain.ErrInternalServer
		}
		u.log.Info(ctx, "post unliked successfully", "postID", postID, "username", username)
	} else {
		// Like
		like := &domain.Like{
			PostID:   postID,
			Username: username,
		}
		if err := u.likeRepo.Create(ctx, like); err != nil {
			u.log.Error(ctx, "failed to create like", "postID", postID, "username", username, "error", err)
			return false, 0, domain.ErrInternalServer
		}
		u.log.Info(ctx, "post liked successfully", "postID", postID, "username", username)
	}

	// 5. Get updated count.
	count, err := u.likeRepo.CountByPostID(ctx, postID)
	if err != nil {
		u.log.Error(ctx, "failed to count likes", "postID", postID, "error", err)
		return false, 0, domain.ErrInternalServer
	}

	return !alreadyLiked, count, nil
}

func (u *likeUseCase) GetLikeStatus(ctx context.Context, postID int, username string) (bool, int, error) {
	// 1. Validate input.
	if postID <= 0 {
		return false, 0, domain.ErrInvalidPost
	}

	// 2. Verify post exists.
	_, err := u.postRepo.GetByIDOnly(ctx, postID)
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

	// 4. Get count.
	count, err := u.likeRepo.CountByPostID(ctx, postID)
	if err != nil {
		u.log.Error(ctx, "failed to count likes", "postID", postID, "error", err)
		return false, 0, domain.ErrInternalServer
	}

	return liked, count, nil
}
