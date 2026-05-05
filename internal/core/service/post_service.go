package service

import (
	"context"

	"github.com/billykore/project-one/internal/core/domain"
	"github.com/billykore/project-one/internal/core/ports"
)

type postService struct {
	repo ports.PostRepository
	log  ports.Logger
}

func NewPostService(repo ports.PostRepository, log ports.Logger) ports.PostService {
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
