package usecase

import (
	"context"
	"errors"

	"github.com/billykore/project-one/internal/core/domain"
	"github.com/billykore/project-one/internal/core/ports"
)

type commentUseCase struct {
	commentRepo ports.CommentRepository
	postRepo    ports.PostRepository
	userRepo    ports.UserRepository
	publisher   ports.NotificationPublisher
	log         ports.Logger
}

// NewCommentUseCase creates a new instance of ports.CommentUseCase.
func NewCommentUseCase(
	commentRepo ports.CommentRepository,
	postRepo ports.PostRepository,
	userRepo ports.UserRepository,
	publisher ports.NotificationPublisher,
	log ports.Logger,
) ports.CommentUseCase {
	if commentRepo == nil {
		panic("NewCommentUseCase: commentRepo is required")
	}
	if postRepo == nil {
		panic("NewCommentUseCase: postRepo is required")
	}
	if userRepo == nil {
		panic("NewCommentUseCase: userRepo is required")
	}
	if publisher == nil {
		panic("NewCommentUseCase: publisher is required")
	}
	if log == nil {
		panic("NewCommentUseCase: log is required")
	}
	return &commentUseCase{
		commentRepo: commentRepo,
		postRepo:    postRepo,
		userRepo:    userRepo,
		publisher:   publisher,
		log:         log,
	}
}

func (u *commentUseCase) AddComment(ctx context.Context, postID int, username string, content string) error {
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
	post, err := u.postRepo.GetByIDOnly(ctx, int(postID))
	if err != nil {
		if errors.Is(err, domain.ErrPostNotFound) {
			return err
		}
		u.log.Error(ctx, "failed to verify post existence for comment", "postID", postID, "error", err)
		return domain.ErrInternalServer
	}

	// 3. Create comment
	if err := u.commentRepo.Create(ctx, comment); err != nil {
		u.log.Error(ctx, "failed to create comment", "postID", postID, "username", username, "error", err)
		return domain.ErrInternalServer
	}

	u.log.Info(ctx, "comment created successfully", "commentID", comment.ID, "postID", postID, "username", username)

	postOwner, err := u.userRepo.GetUserByUsername(ctx, post.Username)
	if err == nil {
		commenter, err := u.userRepo.GetUserByUsername(ctx, username)
		if err == nil && postOwner.ID != commenter.ID {
			notification := &domain.Notification{
				UserID:    postOwner.ID,
				ActorID:   commenter.ID,
				Type:      domain.NotificationTypeComment,
				PostID:    &post.ID,
				CommentID: &comment.ID,
			}
			if pErr := u.publisher.Publish(ctx, notification); pErr != nil {
				u.log.Error(ctx, "failed to publish comment notification", "error", pErr)
			}
		}
	}

	return nil
}

func (u *commentUseCase) GetCommentsByPostID(ctx context.Context, postID int) ([]*domain.Comment, error) {
	comments, err := u.commentRepo.GetByPostID(ctx, postID)
	if err != nil {
		u.log.Error(ctx, "failed to get comments for post", "postID", postID, "error", err)
		return nil, domain.ErrInternalServer
	}
	return comments, nil
}

func (u *commentUseCase) EditComment(ctx context.Context, id int, username string, content string) error {
	// 1. Fetch current comment
	comment, err := u.commentRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, domain.ErrCommentNotFound) {
			return err
		}
		u.log.Error(ctx, "failed to fetch comment for edit", "commentID", id, "error", err)
		return domain.ErrInternalServer
	}
	if comment == nil {
		return domain.ErrCommentNotFound
	}

	// 2. Authorize: only author can edit
	if comment.Username != username {
		u.log.Warn(ctx, "unauthorized attempt to edit comment", "commentID", id, "attemptedBy", username, "actualAuthor", comment.Username)
		return domain.ErrUnauthorized
	}

	// 3. Update fields & Validate
	comment.Content = content
	if err := comment.Validate(); err != nil {
		return err
	}

	// 4. Persist changes
	if err := u.commentRepo.Update(ctx, comment); err != nil {
		u.log.Error(ctx, "failed to update comment in repository", "commentID", id, "error", err)
		return domain.ErrInternalServer
	}

	u.log.Info(ctx, "comment updated successfully", "commentID", id, "username", username)
	return nil
}

func (u *commentUseCase) DeleteComment(ctx context.Context, id int, username string) error {
	// 1. Fetch current comment
	comment, err := u.commentRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, domain.ErrCommentNotFound) {
			return err
		}
		u.log.Error(ctx, "failed to fetch comment for delete", "commentID", id, "error", err)
		return domain.ErrInternalServer
	}
	if comment == nil {
		return domain.ErrCommentNotFound
	}

	// 2. Authorize: only author can delete
	if comment.Username != username {
		u.log.Warn(ctx, "unauthorized attempt to delete comment", "commentID", id, "attemptedBy", username, "actualAuthor", comment.Username)
		return domain.ErrUnauthorized
	}

	// 3. Persist deletion
	if err := u.commentRepo.Delete(ctx, id); err != nil {
		u.log.Error(ctx, "failed to delete comment in repository", "commentID", id, "error", err)
		return domain.ErrInternalServer
	}

	u.log.Info(ctx, "comment deleted successfully", "commentID", id, "username", username)
	return nil
}
