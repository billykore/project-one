package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	identitydomain "github.com/billykore/project-one/internal/identity/domain"
	notificationdomain "github.com/billykore/project-one/internal/notifications/domain"
	platformports "github.com/billykore/project-one/internal/platform/ports"
	"github.com/billykore/project-one/internal/platform/problem"
	"github.com/billykore/project-one/internal/publishing/domain"
	"github.com/billykore/project-one/internal/publishing/ports"
)

type commentUseCase struct {
	commentRepo ports.CommentRepository
	postRepo    ports.PostCommandRepository
	userRepo    ports.UserLookup
	publisher   platformports.Publisher
}

// NewCommentUseCase creates a new instance of ports.CommentUseCase.
func NewCommentUseCase(
	commentRepo ports.CommentRepository,
	postRepo ports.PostCommandRepository,
	userRepo ports.UserLookup,
	publisher platformports.Publisher,
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

func (uc *commentUseCase) AddComment(ctx context.Context, postID int, author *identitydomain.User, content string) error {
	if author == nil || author.ID <= 0 {
		return problem.ErrInvalidUsername
	}
	comment := &domain.Comment{
		PostID:   postID,
		UserID:   author.ID,
		Username: author.Username,
		Content:  content,
	}

	// 1. Validate domain entity
	if err := comment.Validate(); err != nil {
		return err
	}

	// 2. Verify post exists
	post, err := uc.postRepo.Load(ctx, postID)
	if err != nil {
		return fmt.Errorf("failed to fetch post for comment: %w", err)
	}

	// 3. Create comment
	if err := uc.commentRepo.Create(ctx, comment); err != nil {
		return fmt.Errorf("failed to create comment: %w", err)
	}

	if post.UserID != author.ID {
		uc.publishCommentNotification(ctx, post, comment)
	}

	return nil
}

func (uc *commentUseCase) publishCommentNotification(ctx context.Context, post *domain.Post, comment *domain.Comment) {
	postOwner, err := uc.userRepo.GetUserByID(ctx, post.UserID)
	if err != nil {
		return
	}
	if postOwner == nil {
		return
	}

	notification := &notificationdomain.Notification{
		UserID:        postOwner.ID,
		ActorID:       comment.UserID,
		Type:          notificationdomain.NotificationTypeComment,
		PostID:        post.ID,
		CommentID:     comment.ID,
		ActorUsername: comment.Username,
		CreatedAt:     comment.CreatedAt,
	}

	notificationEvent := notificationdomain.NotificationEvent{
		EventID:       fmt.Sprintf("backend-%d", time.Now().UnixNano()),
		SchemaVersion: "1.0",
		Timestamp:     time.Now().UTC().Format(time.RFC3339Nano),
		Notification:  *notification,
	}

	payload, err := json.Marshal(notificationEvent)
	if err != nil {
		return
	}

	if err := uc.publisher.Publish(ctx, platformports.Event{
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
		return nil, fmt.Errorf("failed to get comments for post: %w", problem.ErrRepositoryFailure)
	}
	return comments, nil
}

func (uc *commentUseCase) EditComment(ctx context.Context, id int, userID int, content string) error {
	// 1. Fetch current comment
	comment, err := uc.commentRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, problem.ErrCommentNotFound) {
			return err
		}
		return fmt.Errorf("failed to fetch comment for edit: %w", err)
	}
	if comment == nil {
		return problem.ErrCommentNotFound
	}

	// 2. Authorize: only author can edit
	if comment.UserID != userID {
		return problem.ErrCommentNotOwned
	}

	// 3. Update fields & Validate
	comment.Content = content
	if err := comment.Validate(); err != nil {
		return fmt.Errorf("%w: %v", problem.ErrInvalidComment, err)
	}

	// 4. Persist changes
	if err := uc.commentRepo.Update(ctx, comment); err != nil {
		return fmt.Errorf("failed to update comment: %w", problem.ErrRepositoryFailure)
	}

	return nil
}

func (uc *commentUseCase) DeleteComment(ctx context.Context, id int, userID int) error {
	// 1. Fetch current comment
	comment, err := uc.commentRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, problem.ErrCommentNotFound) {
			return err
		}
		return fmt.Errorf("failed to fetch comment for delete: %w", problem.ErrRepositoryFailure)
	}
	if comment == nil {
		return problem.ErrCommentNotFound
	}

	// 2. Authorize: only author can delete
	if comment.UserID != userID {
		return problem.ErrCommentNotOwned
	}

	// 3. Persist deletion
	if err := uc.commentRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete comment: %w", problem.ErrRepositoryFailure)
	}

	return nil
}
