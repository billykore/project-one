package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/billykore/project-one/internal/core/domain"
	"github.com/billykore/project-one/internal/core/ports"
)

type commentUseCase struct {
	commentRepo ports.CommentRepository
	postRepo    ports.PostRepository
	userRepo    ports.UserRepository
	publisher   ports.Publisher
	log         ports.Logger
}

// NewCommentUseCase creates a new instance of ports.CommentUseCase.
func NewCommentUseCase(
	commentRepo ports.CommentRepository,
	postRepo ports.PostRepository,
	userRepo ports.UserRepository,
	publisher ports.Publisher,
	log ports.Logger,
) ports.CommentUseCase {
	// ponytail: simplified dependency validation to match NewPostUseCase
	if commentRepo == nil || postRepo == nil || userRepo == nil || publisher == nil || log == nil {
		panic("NewCommentUseCase: dependencies must not be nil")
	}
	return &commentUseCase{
		commentRepo: commentRepo,
		postRepo:    postRepo,
		userRepo:    userRepo,
		publisher:   publisher,
		log:         log,
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
		if errors.Is(err, domain.ErrPostNotFound) {
			return err
		}
		uc.log.Error(ctx, "failed to verify post existence for comment", "postID", postID, "error", err)
		return domain.ErrInternalServer
	}

	// 3. Create comment
	if err := uc.commentRepo.Create(ctx, comment); err != nil {
		uc.log.Error(ctx, "failed to create comment", "postID", postID, "username", username, "error", err)
		return domain.ErrInternalServer
	}

	uc.log.Info(ctx, "comment created successfully", "commentID", comment.ID, "postID", postID, "username", username)

	// ponytail: best-effort notification, flattened from nested if-else pyramid
	if post.Username != username {
		uc.publishCommentNotification(ctx, post, comment)
	}

	return nil
}

// ponytail: best-effort notification with early returns.
// Two user lookups needed for correct UserID/ActorID; errors logged, not returned.
func (uc *commentUseCase) publishCommentNotification(ctx context.Context, post *domain.Post, comment *domain.Comment) {
	postOwner, err := uc.userRepo.GetUserByUsername(ctx, post.Username)
	if err != nil {
		uc.log.Error(ctx, "failed to resolve post owner for comment notification", "username", post.Username, "error", err)
		return
	}
	if postOwner == nil {
		return
	}

	commenter, err := uc.userRepo.GetUserByUsername(ctx, comment.Username)
	if err != nil {
		uc.log.Error(ctx, "failed to resolve commenter for comment notification", "username", comment.Username, "error", err)
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
		ActorUsername: commenter.Username,
		CreatedAt:     comment.CreatedAt,
	}

	payload, err := json.Marshal(notification)
	if err != nil {
		uc.log.Error(ctx, "failed to marshal comment notification", "error", err)
		return
	}

	if err := uc.publisher.Publish(ctx, ports.Event{
		Topic:   postNotificationTopic,
		Key:     fmt.Sprintf("user:%d", commenter.ID),
		Payload: payload,
	}); err != nil {
		uc.log.Error(ctx, "failed to publish comment notification", "error", err)
	}
}

func (uc *commentUseCase) GetCommentsByPostID(ctx context.Context, postID int) ([]*domain.Comment, error) {
	comments, err := uc.commentRepo.GetByPostID(ctx, postID)
	if err != nil {
		uc.log.Error(ctx, "failed to get comments for post", "postID", postID, "error", err)
		return nil, domain.ErrInternalServer
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
		uc.log.Error(ctx, "failed to fetch comment for edit", "commentID", id, "error", err)
		return domain.ErrInternalServer
	}
	if comment == nil {
		return domain.ErrCommentNotFound
	}

	// 2. Authorize: only author can edit
	if comment.Username != username {
		uc.log.Warn(ctx, "unauthorized attempt to edit comment", "commentID", id, "attemptedBy", username, "actualAuthor", comment.Username)
		return domain.ErrUnauthorized
	}

	// 3. Update fields & Validate
	comment.Content = content
	if err := comment.Validate(); err != nil {
		return err
	}

	// 4. Persist changes
	if err := uc.commentRepo.Update(ctx, comment); err != nil {
		uc.log.Error(ctx, "failed to update comment in repository", "commentID", id, "error", err)
		return domain.ErrInternalServer
	}

	uc.log.Info(ctx, "comment updated successfully", "commentID", id, "username", username)
	return nil
}

func (uc *commentUseCase) DeleteComment(ctx context.Context, id int, username string) error {
	// 1. Fetch current comment
	comment, err := uc.commentRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, domain.ErrCommentNotFound) {
			return err
		}
		uc.log.Error(ctx, "failed to fetch comment for delete", "commentID", id, "error", err)
		return domain.ErrInternalServer
	}
	if comment == nil {
		return domain.ErrCommentNotFound
	}

	// 2. Authorize: only author can delete
	if comment.Username != username {
		uc.log.Warn(ctx, "unauthorized attempt to delete comment", "commentID", id, "attemptedBy", username, "actualAuthor", comment.Username)
		return domain.ErrUnauthorized
	}

	// 3. Persist deletion
	if err := uc.commentRepo.Delete(ctx, id); err != nil {
		uc.log.Error(ctx, "failed to delete comment in repository", "commentID", id, "error", err)
		return domain.ErrInternalServer
	}

	uc.log.Info(ctx, "comment deleted successfully", "commentID", id, "username", username)
	return nil
}
