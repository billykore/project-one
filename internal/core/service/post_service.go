package service

import (
	"context"
	"errors"

	"github.com/billykore/project-one/internal/core/domain"
	"github.com/billykore/project-one/internal/core/ports"
)

type PostService struct {
	repo ports.PostRepository
	log  ports.Logger
}

func NewPostService(repo ports.PostRepository, log ports.Logger) *PostService {
	if repo == nil {
		panic("NewPostService: repo is required")
	}
	if log == nil {
		panic("NewPostService: log is required")
	}
	return &PostService{
		repo: repo,
		log:  log,
	}
}

func (s *PostService) CreatePost(ctx context.Context, userID int, title, content string, tags []string) (*domain.Post, error) {
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

func (s *PostService) GetPostByID(ctx context.Context, userID, id int) (*domain.Post, error) {
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
