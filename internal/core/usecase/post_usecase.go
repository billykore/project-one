package usecase

import (
	"context"
	"errors"
	"strings"

	"github.com/billykore/project-one/internal/core/domain"
	"github.com/billykore/project-one/internal/core/ports"
)

type postUseCase struct {
	repo ports.PostRepository
	log  ports.Logger
}

// NewPostUseCase creates a new instance of ports.PostUseCase.
func NewPostUseCase(repo ports.PostRepository, log ports.Logger) ports.PostUseCase {
	if repo == nil {
		panic("NewPostUseCase: repo is required")
	}
	if log == nil {
		panic("NewPostUseCase: log is required")
	}
	return &postUseCase{
		repo: repo,
		log:  log,
	}
}

func (s *postUseCase) CreatePost(ctx context.Context, username string, title, content string, tags []string) (*domain.Post, error) {
	post := &domain.Post{
		Username: username,
		Title:    title,
		Content:  content,
		Tags:     tags,
	}

	if err := s.repo.Create(ctx, post); err != nil {
		s.log.Error(ctx, "failed to create post", "username", username, "error", err)
		return nil, domain.ErrInternalServer
	}

	s.log.Info(ctx, "post created successfully", "postID", post.ID, "username", username)
	return post, nil
}

func (s *postUseCase) GetPostByID(ctx context.Context, username string, id int) (*domain.Post, error) {
	if id <= 0 {
		return nil, domain.ErrInvalidPost
	}

	post, err := s.repo.GetByID(ctx, username, id)
	if err != nil {
		if errors.Is(err, domain.ErrPostNotFound) {
			return nil, err
		}
		s.log.Error(ctx, "failed to get post by id", "postID", id, "username", username, "error", err)
		return nil, domain.ErrInternalServer
	}

	return post, nil
}

func (s *postUseCase) GetPosts(ctx context.Context, username string, limit, offset int) ([]*domain.Post, error) {
	posts, err := s.repo.GetUserPosts(ctx, username, limit, offset)
	if err != nil {
		s.log.Error(ctx, "failed to get posts for user", "username", username, "error", err)
		return nil, domain.ErrInternalServer
	}
	return posts, nil
}

func (s *postUseCase) UpdatePost(ctx context.Context, username string, postID int, title, content string) (*domain.Post, error) {
	if postID <= 0 {
		return nil, domain.ErrInvalidPost
	}

	post, err := s.repo.GetByID(ctx, username, postID)
	if err != nil {
		if errors.Is(err, domain.ErrPostNotFound) {
			return nil, err
		}
		s.log.Error(ctx, "failed to get post for update", "postID", postID, "username", username, "error", err)
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

	if err := s.repo.Update(ctx, username, post); err != nil {
		s.log.Error(ctx, "failed to update post", "postID", postID, "error", err)
		return nil, domain.ErrInternalServer
	}

	s.log.Info(ctx, "post updated successfully", "postID", post.ID, "username", username)
	return post, nil
}

func (s *postUseCase) DeletePost(ctx context.Context, username string, postID int) error {
	if postID <= 0 {
		return domain.ErrInvalidPost
	}

	if err := s.repo.Delete(ctx, username, postID); err != nil {
		s.log.Error(ctx, "failed to delete post", "postID", postID, "error", err)
		return domain.ErrInternalServer
	}

	s.log.Info(ctx, "post deleted successfully", "postID", postID, "username", username)
	return nil
}
