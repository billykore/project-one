package usecase

import (
	"context"
	"errors"
	"strings"

	"github.com/billykore/project-one/internal/core/domain"
	"github.com/billykore/project-one/internal/core/ports"
)

type postUseCase struct {
	postRepo  ports.PostRepository
	likeRepo  ports.LikeRepository
	userRepo  ports.UserRepository
	publisher ports.NotificationPublisher
	log       ports.Logger
}

// NewPostUseCase creates a new instance of ports.PostUseCase.
func NewPostUseCase(
	postRepo ports.PostRepository,
	likeRepo ports.LikeRepository,
	userRepo ports.UserRepository,
	publisher ports.NotificationPublisher,
	log ports.Logger,
) ports.PostUseCase {
	if postRepo == nil {
		panic("NewPostUseCase: postRepo is required")
	}
	if likeRepo == nil {
		panic("NewPostUseCase: likeRepo is required")
	}
	if userRepo == nil {
		panic("NewPostUseCase: userRepo is required")
	}
	if publisher == nil {
		panic("NewPostUseCase: publisher is required")
	}
	if log == nil {
		panic("NewPostUseCase: log is required")
	}
	return &postUseCase{
		postRepo:  postRepo,
		likeRepo:  likeRepo,
		userRepo:  userRepo,
		publisher: publisher,
		log:       log,
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

func (uc *postUseCase) GetPostByID(ctx context.Context, id int) (*domain.Post, error) {
	if id <= 0 {
		return nil, domain.ErrInvalidPost
	}

	post, err := uc.postRepo.GetByIDOnly(ctx, id)
	if err != nil {
		if errors.Is(err, domain.ErrPostNotFound) {
			return nil, err
		}
		uc.log.Error(ctx, "failed to get post by id", "postID", id, "error", err)
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

func (uc *postUseCase) LikePost(ctx context.Context, postID int, username string) (int, error) {
	if postID <= 0 {
		return 0, domain.ErrInvalidPost
	}
	if username == "" {
		return 0, domain.ErrValidationFailed
	}

	exists, err := uc.likeRepo.Exists(ctx, postID, username)
	if err != nil {
		uc.log.Error(ctx, "failed to check if like exists", "postID", postID, "username", username, "error", err)
		return 0, domain.ErrInternalServer
	}

	if exists {
		post, err := uc.postRepo.GetByIDOnly(ctx, postID)
		if err != nil {
			uc.log.Error(ctx, "failed to get post for like count", "postID", postID, "error", err)
			return 0, domain.ErrInternalServer
		}
		return post.LikeCount, nil
	}

	like := &domain.Like{
		PostID:   postID,
		Username: username,
	}
	if err := uc.likeRepo.Create(ctx, like); err != nil {
		if errors.Is(err, domain.ErrPostNotFound) {
			return 0, err
		}
		if errors.Is(err, domain.ErrAlreadyLiked) {
			post, err := uc.postRepo.GetByIDOnly(ctx, postID)
			if err != nil {
				return 0, domain.ErrInternalServer
			}
			return post.LikeCount, nil
		}
		uc.log.Error(ctx, "failed to create like", "postID", postID, "username", username, "error", err)
		return 0, domain.ErrInternalServer
	}

	if err := uc.postRepo.IncrementLikeCount(ctx, postID, 1); err != nil {
		uc.log.Error(ctx, "failed to increment like count", "postID", postID, "error", err)
		return 0, domain.ErrInternalServer
	}

	uc.log.Info(ctx, "post liked successfully", "postID", postID, "username", username)

	post, err := uc.postRepo.GetByIDOnly(ctx, postID)
	if err != nil {
		uc.log.Error(ctx, "failed to get post for like count", "postID", postID, "error", err)
		return 0, domain.ErrInternalServer
	}

	if post.Username != username {
		postOwner, err := uc.userRepo.GetUserByUsername(ctx, post.Username)
		if err == nil && postOwner != nil {
			liker, err := uc.userRepo.GetUserByUsername(ctx, username)
			if err == nil && liker != nil && postOwner.ID != liker.ID {
				notification := &domain.Notification{
					UserID:  postOwner.ID,
					ActorID: liker.ID,
					Type:    domain.NotificationTypeLike,
					PostID:  &post.ID,
				}
				if err := notification.Validate(); err != nil {
					uc.log.Error(ctx, "invalid like notification", "error", err)
				} else {
					if pErr := uc.publisher.Publish(ctx, notification); pErr != nil {
						uc.log.Error(ctx, "failed to publish like notification", "error", pErr)
					}
				}
			}
		}
	}

	return post.LikeCount, nil
}

func (uc *postUseCase) UnlikePost(ctx context.Context, postID int, username string) (int, error) {
	if postID <= 0 {
		return 0, domain.ErrInvalidPost
	}
	if username == "" {
		return 0, domain.ErrValidationFailed
	}

	exists, err := uc.likeRepo.Exists(ctx, postID, username)
	if err != nil {
		uc.log.Error(ctx, "failed to check if like exists", "postID", postID, "username", username, "error", err)
		return 0, domain.ErrInternalServer
	}

	if exists {
		if err := uc.likeRepo.Delete(ctx, postID, username); err != nil {
			uc.log.Error(ctx, "failed to delete like", "postID", postID, "username", username, "error", err)
			return 0, domain.ErrInternalServer
		}
		if err := uc.postRepo.IncrementLikeCount(ctx, postID, -1); err != nil {
			uc.log.Error(ctx, "failed to decrement like count", "postID", postID, "error", err)
			return 0, domain.ErrInternalServer
		}
		uc.log.Info(ctx, "post unliked successfully", "postID", postID, "username", username)
	}

	post, err := uc.postRepo.GetByIDOnly(ctx, postID)
	if err != nil {
		if errors.Is(err, domain.ErrPostNotFound) {
			return 0, err
		}
		uc.log.Error(ctx, "failed to get post for like count", "postID", postID, "error", err)
		return 0, domain.ErrInternalServer
	}

	return post.LikeCount, nil
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
