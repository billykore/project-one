package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/billykore/project-one/internal/core/domain"
	"github.com/billykore/project-one/internal/core/ports"
)

const postNotificationTopic = "notifications"

type postCommandUseCase struct {
	postRepo  ports.PostCommandRepository
	likeRepo  ports.LikeRepository
	userRepo  ports.UserRepository
	publisher ports.Publisher
	log       ports.Logger
	evaluator ports.FeatureFlagEvaluator
}

func NewPostCommandUseCase(postRepo ports.PostCommandRepository, likeRepo ports.LikeRepository, userRepo ports.UserRepository, publisher ports.Publisher, log ports.Logger, evaluator ports.FeatureFlagEvaluator) ports.PostCommandUseCase {
	if postRepo == nil || likeRepo == nil || userRepo == nil || publisher == nil || log == nil || evaluator == nil {
		panic("NewPostCommandUseCase: dependencies must not be nil")
	}
	return &postCommandUseCase{postRepo: postRepo, likeRepo: likeRepo, userRepo: userRepo, publisher: publisher, log: log, evaluator: evaluator}
}

func (uc *postCommandUseCase) CreatePost(ctx context.Context, user *domain.User, title, content string, tags []string) (*domain.Post, error) {
	decision := uc.evaluator.Evaluate(ctx, "post-creation", user.Username)
	if decision.Source == domain.SourceUnknown || !decision.Enabled {
		return nil, domain.ErrFeatureDisabled
	}
	post := &domain.Post{UserID: user.ID, Username: user.Username, Title: title, Content: content, Tags: tags}
	if err := uc.postRepo.Save(ctx, post); err != nil {
		uc.log.Error(ctx, "failed to create post", "username", user.Username, "error", err)
		return nil, fmt.Errorf("create post: %w", domain.ErrRepositoryFailure)
	}
	uc.log.Info(ctx, "post created successfully", "postID", post.ID, "username", user.Username)
	return post, nil
}

func (uc *postCommandUseCase) UpdatePost(ctx context.Context, username string, postID int, title, content string) (*domain.Post, error) {
	if postID <= 0 {
		return nil, domain.ErrInvalidPost
	}
	post, err := uc.postRepo.Load(ctx, postID)
	if err != nil {
		return nil, fmt.Errorf("get post for update: %w", err)
	}
	if post.Username != username {
		return nil, domain.ErrPostNotOwned
	}
	post.Update(title, content)
	if err := uc.postRepo.Save(ctx, post); err != nil {
		return nil, fmt.Errorf("update post: %w", err)
	}
	uc.log.Info(ctx, "post updated successfully", "postID", postID, "username", username)
	return post, nil
}

func (uc *postCommandUseCase) DeletePost(ctx context.Context, username string, postID int) error {
	if postID <= 0 {
		return domain.ErrInvalidPost
	}
	post, err := uc.postRepo.Load(ctx, postID)
	if err != nil {
		return fmt.Errorf("get post for delete: %w", err)
	}
	if post.Username != username {
		return domain.ErrPostNotOwned
	}
	if err := uc.postRepo.Delete(ctx, post); err != nil {
		uc.log.Error(ctx, "failed to delete post", "postID", postID, "error", err)
		return fmt.Errorf("delete post: %w", domain.ErrRepositoryFailure)
	}
	uc.log.Info(ctx, "post deleted successfully", "postID", postID, "username", username)
	return nil
}

func (uc *postCommandUseCase) LikePost(ctx context.Context, postID int, username string) (int, error) {
	if postID <= 0 {
		return 0, domain.ErrInvalidPostID
	}
	if username == "" {
		return 0, domain.ErrInvalidUsername
	}
	post, err := uc.postRepo.Load(ctx, postID)
	if err != nil {
		if errors.Is(err, domain.ErrPostNotFound) {
			return 0, err
		}
		uc.log.Error(ctx, "failed to verify post existence for like", "postID", postID, "error", err)
		return 0, fmt.Errorf("verify post existence: %w", err)
	}
	likeCount, changed, err := uc.likeRepo.SetLiked(ctx, postID, username, true)
	if err != nil {
		uc.log.Error(ctx, "failed to set like state", "postID", postID, "username", username, "error", err)
		return 0, fmt.Errorf("set like state: %w", err)
	}
	post.LikeCount = likeCount
	if !changed {
		return likeCount, nil
	}
	uc.log.Info(ctx, "post liked successfully", "postID", postID, "username", username)
	if post.Username != username {
		uc.publishLikeNotification(ctx, post, &domain.Like{PostID: postID, Username: username})
	}
	return likeCount, nil
}

func (uc *postCommandUseCase) UnlikePost(ctx context.Context, postID int, username string) (int, error) {
	if postID <= 0 {
		return 0, domain.ErrInvalidPost
	}
	if username == "" {
		return 0, domain.ErrInvalidUsername
	}
	post, err := uc.postRepo.Load(ctx, postID)
	if err != nil {
		if errors.Is(err, domain.ErrPostNotFound) {
			return 0, err
		}
		uc.log.Error(ctx, "failed to get post for unlike", "postID", postID, "error", err)
		return 0, fmt.Errorf("get post for unlike: %w", err)
	}
	likeCount, changed, err := uc.likeRepo.SetLiked(ctx, postID, username, false)
	if err != nil {
		uc.log.Error(ctx, "failed to set like state", "postID", postID, "username", username, "error", err)
		return 0, fmt.Errorf("set like state: %w", err)
	}
	post.LikeCount = likeCount
	if !changed {
		return likeCount, nil
	}
	uc.log.Info(ctx, "post unliked successfully", "postID", postID, "username", username)
	return likeCount, nil
}

func (uc *postCommandUseCase) publishLikeNotification(ctx context.Context, post *domain.Post, like *domain.Like) {
	owner, err := uc.userRepo.GetUserByUsername(ctx, post.Username)
	if err != nil {
		uc.log.Error(ctx, "failed to resolve post owner for like notification", "username", post.Username, "error", err)
		return
	}
	if owner == nil {
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
	notification := domain.Notification{UserID: owner.ID, ActorID: liker.ID, Type: domain.NotificationTypeLike, PostID: post.ID, ActorUsername: liker.Username, CreatedAt: like.CreatedAt}
	event := domain.NotificationEvent{EventID: fmt.Sprintf("backend-%d", time.Now().UnixNano()), SchemaVersion: "1.0", Timestamp: time.Now().UTC().Format(time.RFC3339Nano), Notification: notification}
	payload, err := json.Marshal(event)
	if err != nil {
		uc.log.Error(ctx, "failed to marshal like notification", "error", err)
		return
	}
	if err := uc.publisher.Publish(ctx, ports.Event{Topic: postNotificationTopic, Key: fmt.Sprintf("user:%d", owner.ID), Payload: payload, Metadata: map[string]string{"event_id": event.EventID, "schema_version": event.SchemaVersion, "timestamp": event.Timestamp}}); err != nil {
		uc.log.Error(ctx, "failed to publish like notification", "error", err)
	}
}
