package usecase

import (
	"context"
	"errors"
	"strings"

	"github.com/billykore/project-one/internal/core/domain"
	"github.com/billykore/project-one/internal/core/ports"
)

type postUseCase struct {
	postRepo ports.PostRepository
	likeRepo ports.LikeRepository
	log      ports.Logger
}

// NewPostUseCase creates a new instance of ports.PostUseCase.
func NewPostUseCase(postRepo ports.PostRepository, likeRepo ports.LikeRepository, log ports.Logger) ports.PostUseCase {
	if postRepo == nil {
		panic("NewPostUseCase: postRepo is required")
	}
	if likeRepo == nil {
		panic("NewPostUseCase: likeRepo is required")
	}
	if log == nil {
		panic("NewPostUseCase: log is required")
	}
	return &postUseCase{
		postRepo: postRepo,
		likeRepo: likeRepo,
		log:      log,
	}
}

func (uc *postUseCase) CreatePost(ctx context.Context, username string, title, content string, tags []string) (*domain.Post, error) {
	post := &domain.Post{
		Username: username,
		Title:    title,
		Content:  content,
		Tags:     tags,
	}

	if err := uc.postRepo.Create(ctx, post); err != nil {
		uc.log.Error(ctx, "failed to create post", "username", username, "error", err)
		return nil, domain.ErrInternalServer
	}

	uc.log.Info(ctx, "post created successfully", "postID", post.ID, "username", username)
	return post, nil
}

func (uc *postUseCase) GetPostByID(ctx context.Context, username string, id int) (*domain.Post, error) {
	if id <= 0 {
		return nil, domain.ErrInvalidPost
	}

	post, err := uc.postRepo.GetByID(ctx, username, id)
	if err != nil {
		if errors.Is(err, domain.ErrPostNotFound) {
			return nil, err
		}
		uc.log.Error(ctx, "failed to get post by id", "postID", id, "username", username, "error", err)
		return nil, domain.ErrInternalServer
	}

	return post, nil
}

func (uc *postUseCase) GetPosts(ctx context.Context, username string, limit, offset int) ([]*domain.Post, error) {
	posts, err := uc.postRepo.GetUserPosts(ctx, username, limit, offset)
	if err != nil {
		uc.log.Error(ctx, "failed to get posts for user", "username", username, "error", err)
		return nil, domain.ErrInternalServer
	}
	return posts, nil
}

func (uc *postUseCase) UpdatePost(ctx context.Context, username string, postID int, title, content string) (*domain.Post, error) {
	if postID <= 0 {
		return nil, domain.ErrInvalidPost
	}

	post, err := uc.postRepo.GetByID(ctx, username, postID)
	if err != nil {
		if errors.Is(err, domain.ErrPostNotFound) {
			return nil, err
		}
		uc.log.Error(ctx, "failed to get post for update", "postID", postID, "username", username, "error", err)
		return nil, domain.ErrInternalServer
	}

	title = strings.TrimSpace(title)
	content = strings.TrimSpace(content)

	if title != "" {
		post.Title = title
	}
	if content != "" {
		post.Content = content
	}

	if err := uc.postRepo.Update(ctx, username, post); err != nil {
		uc.log.Error(ctx, "failed to update post", "postID", postID, "error", err)
		return nil, domain.ErrInternalServer
	}

	uc.log.Info(ctx, "post updated successfully", "postID", post.ID, "username", username)
	return post, nil
}

func (uc *postUseCase) DeletePost(ctx context.Context, username string, postID int) error {
	if postID <= 0 {
		return domain.ErrInvalidPost
	}

	if err := uc.postRepo.Delete(ctx, username, postID); err != nil {
		uc.log.Error(ctx, "failed to delete post", "postID", postID, "error", err)
		return domain.ErrInternalServer
	}

	uc.log.Info(ctx, "post deleted successfully", "postID", postID, "username", username)
	return nil
}

func (uc *postUseCase) ToggleLike(ctx context.Context, postID int, username string) (bool, int, error) {
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
	if err := uc.likeRepo.Create(ctx, like); err != nil {
		if errors.Is(err, domain.ErrAlreadyLiked) {
			// Already liked, so unlike
			if err := uc.likeRepo.Delete(ctx, postID, username); err != nil {
				if errors.Is(err, domain.ErrNotLiked) {
					return false, 0, err
				}
				uc.log.Error(ctx, "failed to delete like", "postID", postID, "username", username, "error", err)
				return false, 0, domain.ErrInternalServer
			}
			liked = false
			// Decrement count
			if err := uc.postRepo.IncrementLikeCount(ctx, postID, -1); err != nil {
				uc.log.Error(ctx, "failed to decrement like count", "postID", postID, "error", err)
			}
			uc.log.Info(ctx, "post unliked successfully", "postID", postID, "username", username)
		} else if errors.Is(err, domain.ErrPostNotFound) {
			return false, 0, err
		} else {
			uc.log.Error(ctx, "failed to create like", "postID", postID, "username", username, "error", err)
			return false, 0, domain.ErrInternalServer
		}
	} else {
		// Like created successfully
		liked = true
		if err := uc.postRepo.IncrementLikeCount(ctx, postID, 1); err != nil {
			uc.log.Error(ctx, "failed to increment like count", "postID", postID, "error", err)
		}
		uc.log.Info(ctx, "post liked successfully", "postID", postID, "username", username)
	}

	// 3. Get updated count from post.
	post, err := uc.postRepo.GetByIDOnly(ctx, postID)
	if err != nil {
		uc.log.Error(ctx, "failed to get post for like count", "postID", postID, "error", err)
		return false, 0, domain.ErrInternalServer
	}

	return liked, post.LikeCount, nil
}

func (uc *postUseCase) GetLikeStatus(ctx context.Context, postID int, username string) (bool, int, error) {
	// 1. Validate input.
	if postID <= 0 {
		return false, 0, domain.ErrInvalidPost
	}
	if username == "" {
		return false, 0, domain.ErrValidationFailed
	}

	// 2. Verify post exists.
	post, err := uc.postRepo.GetByIDOnly(ctx, postID)
	if err != nil {
		if errors.Is(err, domain.ErrPostNotFound) {
			return false, 0, err
		}
		uc.log.Error(ctx, "failed to verify post existence for like status", "postID", postID, "error", err)
		return false, 0, domain.ErrInternalServer
	}

	// 3. Check if user liked the post.
	liked, err := uc.likeRepo.Exists(ctx, postID, username)
	if err != nil {
		uc.log.Error(ctx, "failed to check like existence", "postID", postID, "username", username, "error", err)
		return false, 0, domain.ErrInternalServer
	}

	return liked, post.LikeCount, nil
}
