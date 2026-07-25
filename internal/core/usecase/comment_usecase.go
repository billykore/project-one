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

type commentUseCase struct {
	commentRepo ports.CommentRepository
	postRepo    ports.PostRepository
	userRepo    ports.UserRepository
	publisher   ports.Publisher
}

// NewCommentUseCase creates a new instance of ports.CommentUseCase.
func NewCommentUseCase(
	commentRepo ports.CommentRepository,
	postRepo ports.PostRepository,
	userRepo ports.UserRepository,
	publisher ports.Publisher,
) ports.CommentUseCase {
	if commentRepo == nil || postRepo == nil || userRepo == nil || publisher == nil {
		panic("NewCommentUseCase: dependencies must not be nil")
	}
	return &commentUseCase{
		commentRepo: commentRepo,
		postRepo:    postRepo,
		userRepo:    userRepo,
		publisher:   publisher,
	}
}

func (uc *commentUseCase) AddComment(ctx context.Context, postID int, username string, content string) error {
	comment := &domain.Comment{
		PostID:   postID,
		Username: username,
		Content:  content,
	}

	// 1. Validate domain entity
	if err := comment.Validate(); err != nil {
		return err
	}

	// 2. Verify post exists
	post, err := uc.postRepo.GetByIDOnly(ctx, postID)
	if err != nil {
		return fmt.Errorf("failed to fetch post for comment: %w", err)
	}

	// 3. Create comment
	if err := uc.commentRepo.Create(ctx, comment); err != nil {
		return fmt.Errorf("failed to create comment: %w", err)
	}

	if post.Username != username {
		uc.publishCommentNotification(ctx, post, comment)
	}

	return nil
}

func (uc *commentUseCase) publishCommentNotification(ctx context.Context, post *domain.Post, comment *domain.Comment) {
	postOwner, err := uc.userRepo.GetUserByUsername(ctx, post.Username)
	if err != nil {
		return
	}
	if postOwner == nil {
		return
	}

	commenter, err := uc.userRepo.GetUserByUsername(ctx, comment.Username)
	if err != nil {
		return
	}
	if commenter == nil {
		return
	}

	notification := &domain.Notification{
		UserID:        postOwner.ID,
		ActorID:       commenter.ID,
		Type:          domain.NotificationTypeComment,
		PostID:        post.ID,
		CommentID:     comment.ID,
		ActorUsername: commenter.Username,
		CreatedAt:     comment.CreatedAt,
	}

	notificationEvent := domain.NotificationEvent{
		EventID:       fmt.Sprintf("backend-%d", time.Now().UnixNano()),
		SchemaVersion: "1.0",
		Timestamp:     time.Now().UTC().Format(time.RFC3339Nano),
		Notification:  *notification,
	}

	payload, err := json.Marshal(notificationEvent)
	if err != nil {
		return
	}

	if err := uc.publisher.Publish(ctx, ports.Event{
		Topic:   postNotificationTopic,
		Key:     fmt.Sprintf("user:%d", postOwner.ID),
		Payload: payload,
		Metadata: map[string]string{
			"event_id":       notificationEvent.EventID,
			"schema_version": notificationEvent.SchemaVersion,
			"timestamp":      notificationEvent.Timestamp,
		},
	}); err != nil {
		return
	}
}

func (uc *commentUseCase) GetCommentsByPostID(ctx context.Context, postID int) ([]*domain.Comment, error) {
	comments, err := uc.commentRepo.GetByPostID(ctx, postID)
	if err != nil {
		return nil, fmt.Errorf("failed to get comments for post: %w", domain.ErrRepositoryFailure)
	}
	return comments, nil
}

func (uc *commentUseCase) EditComment(ctx context.Context, id int, username string, content string) error {
	// 1. Fetch current comment
	comment, err := uc.commentRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, domain.ErrCommentNotFound) {
			return err
		}
		return fmt.Errorf("failed to fetch comment for edit: %w", err)
	}
	if comment == nil {
		return domain.ErrCommentNotFound
	}

	// 2. Authorize: only author can edit
	if comment.Username != username {
		return domain.ErrCommentNotOwned
	}

	// 3. Update fields & Validate
	comment.Content = content
	if err := comment.Validate(); err != nil {
		return fmt.Errorf("%w: %v", domain.ErrInvalidComment, err)
	}

	// 4. Persist changes
	if err := uc.commentRepo.Update(ctx, comment); err != nil {
		return fmt.Errorf("failed to update comment: %w", domain.ErrRepositoryFailure)
	}

	return nil
}

func (uc *commentUseCase) DeleteComment(ctx context.Context, id int, username string) error {
	// 1. Fetch current comment
	comment, err := uc.commentRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, domain.ErrCommentNotFound) {
			return err
		}
		return fmt.Errorf("failed to fetch comment for delete: %w", domain.ErrRepositoryFailure)
	}
	if comment == nil {
		return domain.ErrCommentNotFound
	}

	// 2. Authorize: only author can delete
	if comment.Username != username {
		return domain.ErrCommentNotOwned
	}

	// 3. Persist deletion
	if err := uc.commentRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete comment: %w", domain.ErrRepositoryFailure)
	}

	return nil
}
