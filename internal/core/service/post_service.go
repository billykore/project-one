package service

import (
	"context"
	"errors"
	"strings"

	"github.com/billykore/project-one/internal/core/domain"
	"github.com/billykore/project-one/internal/core/ports"
)

type postService struct {
	repo ports.PostRepository
	log  ports.Logger
}

// NewPostService creates a new instance of ports.PostService.
func NewPostService(repo ports.PostRepository, log ports.Logger) ports.PostService {
	if repo == nil {
		panic("NewPostService: repo is required")
	}
	if log == nil {
		panic("NewPostService: log is required")
	}
	return &postService{
		repo: repo,
		log:  log,
	}
}

func (s *postService) CreatePost(ctx context.Context, userID int, title, content string, tags []string) (*domain.Post, error) {
	post := &domain.Post{
		UserID:  userID,
		Title:   title,
		Content: content,
		Tags:    tags,
	}

	if err := s.repo.Create(ctx, post); err != nil {
		s.log.Error(ctx, "failed to create post", "userID", userID, "error", err)
		return nil, domain.ErrInternalServer
	}

	s.log.Info(ctx, "post created successfully", "postID", post.ID, "userID", userID)
	return post, nil
}

func (s *postService) GetPostByID(ctx context.Context, userID, id int) (*domain.Post, error) {
	if id <= 0 {
		return nil, domain.ErrInvalidPost
	}

	post, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, domain.ErrPostNotFound) {
			return nil, err
		}
		s.log.Error(ctx, "failed to get post by id", "postID", id, "error", err)
		return nil, domain.ErrInternalServer
	}

	if post.UserID != userID {
		s.log.Error(ctx, "unauthorized access to post", "postID", id, "userID", userID)
		return nil, domain.ErrUnauthorized
	}

	return post, nil
}

func (s *postService) UpdatePost(ctx context.Context, userID, postID int, title, content string) (*domain.Post, error) {
	if postID <= 0 {
		return nil, domain.ErrInvalidPost
	}

	post, err := s.repo.GetByID(ctx, postID)
	if err != nil {
		if errors.Is(err, domain.ErrPostNotFound) {
			return nil, err
		}
		s.log.Error(ctx, "failed to get post for update", "postID", postID, "error", err)
		return nil, domain.ErrInternalServer
	}

	if post.UserID != userID {
		s.log.Error(ctx, "unauthorized update attempt", "postID", postID, "userID", userID)
		return nil, domain.ErrUnauthorized
	}

	title = strings.TrimSpace(title)
	content = strings.TrimSpace(content)

	if title != "" {
		post.Title = title
	}
	if content != "" {
		post.Content = content
	}

	if err := s.repo.Update(ctx, post); err != nil {
		s.log.Error(ctx, "failed to update post", "postID", postID, "error", err)
		return nil, domain.ErrInternalServer
	}

	s.log.Info(ctx, "post updated successfully", "postID", post.ID, "userID", userID)
	return post, nil
}
