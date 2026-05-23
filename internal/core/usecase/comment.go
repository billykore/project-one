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
	log         ports.Logger
}

// NewCommentUseCase creates a new instance of ports.CommentUseCase.
func NewCommentUseCase(
	commentRepo ports.CommentRepository,
	postRepo ports.PostRepository,
	userRepo ports.UserRepository,
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
	if log == nil {
		panic("NewCommentUseCase: log is required")
	}
	return &commentUseCase{
		commentRepo: commentRepo,
		postRepo:    postRepo,
		userRepo:    userRepo,
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
	_, err := u.postRepo.GetByIDOnly(ctx, int(postID))
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
	return nil
}
