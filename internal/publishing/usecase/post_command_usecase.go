package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	featureflagdomain "github.com/billykore/project-one/internal/featureflags/domain"
	identitydomain "github.com/billykore/project-one/internal/identity/domain"
	notificationdomain "github.com/billykore/project-one/internal/notifications/domain"
	platformports "github.com/billykore/project-one/internal/platform/ports"
	"github.com/billykore/project-one/internal/platform/problem"
	"github.com/billykore/project-one/internal/publishing/domain"
	"github.com/billykore/project-one/internal/publishing/ports"
)

const postNotificationTopic = "notifications"

type postCommandUseCase struct {
	postRepo  ports.PostCommandRepository
	likeRepo  ports.LikeRepository
	userRepo  ports.UserLookup
	publisher platformports.Publisher
	log       platformports.Logger
	evaluator ports.FeatureEvaluator
}

func NewPostCommandUseCase(postRepo ports.PostCommandRepository, likeRepo ports.LikeRepository, userRepo ports.UserLookup, publisher platformports.Publisher, log platformports.Logger, evaluator ports.FeatureEvaluator) ports.PostCommandUseCase {
	if postRepo == nil || likeRepo == nil || userRepo == nil || publisher == nil || log == nil || evaluator == nil {
		panic("NewPostCommandUseCase: dependencies must not be nil")
	}
	return &postCommandUseCase{postRepo: postRepo, likeRepo: likeRepo, userRepo: userRepo, publisher: publisher, log: log, evaluator: evaluator}
}

func (uc *postCommandUseCase) CreatePost(ctx context.Context, user *identitydomain.User, title, content string, tags []string) (*domain.Post, error) {
	decision := uc.evaluator.Evaluate(ctx, "post-creation", user.Username)
	if decision.Source == featureflagdomain.SourceUnknown || !decision.Enabled {
		return nil, problem.ErrFeatureDisabled
	}
	post := &domain.Post{UserID: user.ID, Username: user.Username, Title: title, Content: content, Tags: tags}
	if err := uc.postRepo.Save(ctx, post); err != nil {
		uc.log.Error(ctx, "failed to create post", "username", user.Username, "error", err)
		return nil, fmt.Errorf("create post: %w", problem.ErrRepositoryFailure)
	}
	uc.log.Info(ctx, "post created successfully", "postID", post.ID, "username", user.Username)
	return post, nil
}

func (uc *postCommandUseCase) UpdatePost(ctx context.Context, userID int, postID int, title, content string) (*domain.Post, error) {
	if postID <= 0 {
		return nil, problem.ErrInvalidPost
	}
	post, err := uc.postRepo.Load(ctx, postID)
	if err != nil {
		return nil, fmt.Errorf("get post for update: %w", err)
	}
	if post.UserID != userID {
		return nil, problem.ErrPostNotOwned
	}
	post.Update(title, content)
	if err := uc.postRepo.Save(ctx, post); err != nil {
		return nil, fmt.Errorf("update post: %w", err)
	}
	uc.log.Info(ctx, "post updated successfully", "postID", postID, "userID", userID)
	return post, nil
}

func (uc *postCommandUseCase) DeletePost(ctx context.Context, userID int, postID int) error {
	if postID <= 0 {
		return problem.ErrInvalidPost
	}
	post, err := uc.postRepo.Load(ctx, postID)
	if err != nil {
		return fmt.Errorf("get post for delete: %w", err)
	}
	if post.UserID != userID {
		return problem.ErrPostNotOwned
	}
	if err := uc.postRepo.Delete(ctx, post); err != nil {
		uc.log.Error(ctx, "failed to delete post", "postID", postID, "error", err)
		return fmt.Errorf("delete post: %w", problem.ErrRepositoryFailure)
	}
	uc.log.Info(ctx, "post deleted successfully", "postID", postID, "userID", userID)
	return nil
}

func (uc *postCommandUseCase) LikePost(ctx context.Context, postID int, actor *identitydomain.User) (int, error) {
	if postID <= 0 {
		return 0, problem.ErrInvalidPostID
	}
	if actor == nil || actor.ID <= 0 {
		return 0, problem.ErrInvalidUsername
	}
	post, err := uc.postRepo.Load(ctx, postID)
	if err != nil {
		if errors.Is(err, problem.ErrPostNotFound) {
			return 0, err
		}
		uc.log.Error(ctx, "failed to verify post existence for like", "postID", postID, "error", err)
		return 0, fmt.Errorf("verify post existence: %w", err)
	}
	likeCount, changed, err := uc.likeRepo.SetLiked(ctx, postID, actor.ID, true)
	if err != nil {
		uc.log.Error(ctx, "failed to set like state", "postID", postID, "userID", actor.ID, "error", err)
		return 0, fmt.Errorf("set like state: %w", err)
	}
	post.LikeCount = likeCount
	if !changed {
		return likeCount, nil
	}
	uc.log.Info(ctx, "post liked successfully", "postID", postID, "userID", actor.ID)
	if post.UserID != actor.ID {
		uc.publishLikeNotification(ctx, post, actor)
	}
	return likeCount, nil
}

func (uc *postCommandUseCase) UnlikePost(ctx context.Context, postID int, actor *identitydomain.User) (int, error) {
	if postID <= 0 {
		return 0, problem.ErrInvalidPost
	}
	if actor == nil || actor.ID <= 0 {
		return 0, problem.ErrInvalidUsername
	}
	post, err := uc.postRepo.Load(ctx, postID)
	if err != nil {
		if errors.Is(err, problem.ErrPostNotFound) {
			return 0, err
		}
		uc.log.Error(ctx, "failed to get post for unlike", "postID", postID, "error", err)
		return 0, fmt.Errorf("get post for unlike: %w", err)
	}
	likeCount, changed, err := uc.likeRepo.SetLiked(ctx, postID, actor.ID, false)
	if err != nil {
		uc.log.Error(ctx, "failed to set like state", "postID", postID, "userID", actor.ID, "error", err)
		return 0, fmt.Errorf("set like state: %w", err)
	}
	post.LikeCount = likeCount
	if !changed {
		return likeCount, nil
	}
	uc.log.Info(ctx, "post unliked successfully", "postID", postID, "userID", actor.ID)
	return likeCount, nil
}

func (uc *postCommandUseCase) publishLikeNotification(ctx context.Context, post *domain.Post, actor *identitydomain.User) {
	owner, err := uc.userRepo.GetUserByID(ctx, post.UserID)
	if err != nil {
		uc.log.Error(ctx, "failed to resolve post owner for like notification", "userID", post.UserID, "error", err)
		return
	}
	if owner == nil {
		return
	}
	notification := notificationdomain.Notification{UserID: owner.ID, ActorID: actor.ID, Type: notificationdomain.NotificationTypeLike, PostID: post.ID, ActorUsername: actor.Username, CreatedAt: time.Now().UTC()}
	event := notificationdomain.NotificationEvent{EventID: fmt.Sprintf("backend-%d", time.Now().UnixNano()), SchemaVersion: "1.0", Timestamp: time.Now().UTC().Format(time.RFC3339Nano), Notification: notification}
	payload, err := json.Marshal(event)
	if err != nil {
		uc.log.Error(ctx, "failed to marshal like notification", "error", err)
		return
	}
	if err := uc.publisher.Publish(ctx, platformports.Event{Topic: postNotificationTopic, Key: fmt.Sprintf("user:%d", owner.ID), Payload: payload, Metadata: map[string]string{"event_id": event.EventID, "schema_version": event.SchemaVersion, "timestamp": event.Timestamp}}); err != nil {
		uc.log.Error(ctx, "failed to publish like notification", "error", err)
	}
}
