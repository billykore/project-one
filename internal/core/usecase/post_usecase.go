package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/billykore/project-one/internal/core/domain"
	"github.com/billykore/project-one/internal/core/ports"
)

const postNotificationTopic = "notifications"

type postUseCase struct {
	log       ports.Logger
	postRepo  ports.PostRepository
	likeRepo  ports.LikeRepository
	userRepo  ports.UserRepository
	publisher ports.Publisher
}

// NewPostUseCase creates a new instance of ports.PostUseCase.
func NewPostUseCase(
	postRepo ports.PostRepository,
	likeRepo ports.LikeRepository,
	userRepo ports.UserRepository,
	publisher ports.Publisher,
	log ports.Logger,
) ports.PostUseCase {
	// ponytail: simplified dependency validation to match NewFollowUseCase
	if postRepo == nil || likeRepo == nil || userRepo == nil || publisher == nil || log == nil {
		panic("NewPostUseCase: dependencies must not be nil")
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
		return nil, fmt.Errorf("create post: %w", domain.ErrRepositoryFailure)
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
		return nil, fmt.Errorf("get post by id: %w", domain.ErrRepositoryFailure)
	}

	return post, nil
}

func (uc *postUseCase) GetPosts(ctx context.Context, username string, limit, offset int) ([]*domain.Post, error) {
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	posts, err := uc.postRepo.GetUserPosts(ctx, username, limit, offset)
	if err != nil {
		uc.log.Error(ctx, "failed to get posts for user", "username", username, "error", err)
		return nil, fmt.Errorf("get posts for user: %w", domain.ErrRepositoryFailure)
	}
	return posts, nil
}

func (uc *postUseCase) UpdatePost(ctx context.Context, username string, postID int, title, content string) (*domain.Post, error) {
	if postID <= 0 {
		return nil, domain.ErrInvalidPost
	}

	post, err := uc.postRepo.GetByID(ctx, username, postID)
	if err != nil {
		return nil, fmt.Errorf("get post for update: %w", err)
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
		return nil, fmt.Errorf("update post: %w", err)
	}

	uc.log.Info(ctx, "post updated successfully", "postID", postID, "username", username)
	return post, nil
}

func (uc *postUseCase) DeletePost(ctx context.Context, username string, postID int) error {
	if postID <= 0 {
		return domain.ErrInvalidPost
	}

	if err := uc.postRepo.Delete(ctx, username, postID); err != nil {
		uc.log.Error(ctx, "failed to delete post", "postID", postID, "error", err)
		return fmt.Errorf("delete post: %w", domain.ErrRepositoryFailure)
	}

	uc.log.Info(ctx, "post deleted successfully", "postID", postID, "username", username)
	return nil
}

// ponytail: best-effort notification, flattened from nested if-else pyramid.
// Two user lookups needed for correct UserID/ActorID; errors logged, not returned.
func (uc *postUseCase) publishLikeNotification(ctx context.Context, post *domain.Post, like *domain.Like) {
	postOwner, err := uc.userRepo.GetUserByUsername(ctx, post.Username)
	if err != nil {
		uc.log.Error(ctx, "failed to resolve post owner for like notification", "username", post.Username, "error", err)
		return
	}
	if postOwner == nil {
		return
	}

	liker, err := uc.userRepo.GetUserByUsername(ctx, like.Username)
	if err != nil {
		uc.log.Error(ctx, "failed to resolve liker for like notification", "username", like.Username, "error", err)
		return
	}
	if liker == nil {
		return
	}

	notification := &domain.Notification{
		UserID:        postOwner.ID,
		ActorID:       liker.ID,
		Type:          domain.NotificationTypeLike,
		PostID:        post.ID,
		ActorUsername: liker.Username,
		CreatedAt:     like.CreatedAt,
	}

	payload, err := json.Marshal(notification)
	if err != nil {
		uc.log.Error(ctx, "failed to marshal like notification", "error", err)
		return
	}

	if err := uc.publisher.Publish(ctx, ports.Event{
		Topic:   postNotificationTopic,
		Key:     fmt.Sprintf("user:%d", liker.ID),
		Payload: payload,
	}); err != nil {
		uc.log.Error(ctx, "failed to publish like notification", "error", err)
	}
}

func (uc *postUseCase) LikePost(ctx context.Context, postID int, username string) (int, error) {
	if postID <= 0 {
		return 0, domain.ErrInvalidPost
	}
	if username == "" {
		return 0, domain.ErrInvalidUsername
	}

	// Check if post exists
	post, err := uc.postRepo.GetByIDOnly(ctx, postID)
	if err != nil {
		if errors.Is(err, domain.ErrPostNotFound) {
			return 0, err
		}
		uc.log.Error(ctx, "failed to verify post existence for like", "postID", postID, "error", err)
		return 0, fmt.Errorf("verify post existence: %w", err)
	}

	like := &domain.Like{
		PostID:   postID,
		Username: username,
	}
	// ponytail: calling Create directly instead of checking Exists first saves a DB roundtrip
	if err := uc.likeRepo.Create(ctx, like); err != nil {
		if errors.Is(err, domain.ErrPostNotFound) {
			return 0, err
		}
		if errors.Is(err, domain.ErrAlreadyLiked) {
			return post.LikeCount, nil
		}
		uc.log.Error(ctx, "failed to create like", "postID", postID, "username", username, "error", err)
		return 0, fmt.Errorf("create like: %w", err)
	}

	if err := uc.postRepo.IncrementLikeCount(ctx, postID, 1); err != nil {
		uc.log.Error(ctx, "failed to increment like count", "postID", postID, "error", err)
		return 0, fmt.Errorf("increment like count: %w", err)
	}

	uc.log.Info(ctx, "post liked successfully", "postID", postID, "username", username)

	// ponytail: best-effort notification, flattened from nested if-else pyramid
	if post.Username != username {
		uc.publishLikeNotification(ctx, post, like)
	}

	return post.LikeCount + 1, nil
}

func (uc *postUseCase) UnlikePost(ctx context.Context, postID int, username string) (int, error) {
	if postID <= 0 {
		return 0, domain.ErrInvalidPost
	}
	if username == "" {
		return 0, domain.ErrInvalidUsername
	}

	// ponytail: one fetch for existence + count, then delete — same DB calls as before but no re-fetch
	post, err := uc.postRepo.GetByIDOnly(ctx, postID)
	if err != nil {
		if errors.Is(err, domain.ErrPostNotFound) {
			return 0, err
		}
		uc.log.Error(ctx, "failed to get post for unlike", "postID", postID, "error", err)
		return 0, fmt.Errorf("get post for unlike: %w", err)
	}

	// ponytail: calling Delete directly instead of checking Exists first saves a DB roundtrip
	if err := uc.likeRepo.Delete(ctx, postID, username); err != nil {
		if errors.Is(err, domain.ErrNotLiked) {
			return post.LikeCount, nil
		}
		uc.log.Error(ctx, "failed to delete like", "postID", postID, "username", username, "error", err)
		return 0, fmt.Errorf("delete like: %w", err)
	}

	if err := uc.postRepo.IncrementLikeCount(ctx, postID, -1); err != nil {
		uc.log.Error(ctx, "failed to decrement like count", "postID", postID, "error", err)
		return 0, fmt.Errorf("decrement like count: %w", err)
	}
	uc.log.Info(ctx, "post unliked successfully", "postID", postID, "username", username)

	return post.LikeCount - 1, nil
}

func (uc *postUseCase) GetLikeStatus(ctx context.Context, postID int, username string) (bool, int, error) {
	// 1. Validate input.
	if postID <= 0 {
		return false, 0, domain.ErrInvalidPost
	}
	if username == "" {
		return false, 0, domain.ErrInvalidUsername
	}

	// 2. Verify post exists.
	post, err := uc.postRepo.GetByIDOnly(ctx, postID)
	if err != nil {
		if errors.Is(err, domain.ErrPostNotFound) {
			return false, 0, err
		}
		uc.log.Error(ctx, "failed to verify post existence for like status", "postID", postID, "error", err)
		return false, 0, fmt.Errorf("verify post existence: %w", err)
	}

	// 3. Check if user liked the post.
	liked, err := uc.likeRepo.Exists(ctx, postID, username)
	if err != nil {
		uc.log.Error(ctx, "failed to check like existence", "postID", postID, "username", username, "error", err)
		return false, 0, fmt.Errorf("check like existence: %w", err)
	}

	return liked, post.LikeCount, nil
}
