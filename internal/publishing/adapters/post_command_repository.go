package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/billykore/project-one/internal/platform/problem"
	"github.com/billykore/project-one/internal/publishing/domain"
	"github.com/billykore/project-one/internal/publishing/ports"
	"gorm.io/gorm"
)

type postCommandRepository struct{ db *gorm.DB }

func NewPostCommandRepository(db *gorm.DB) ports.PostCommandRepository {
	return &postCommandRepository{db: db}
}

func (r *postCommandRepository) Load(ctx context.Context, id int) (*domain.Post, error) {
	model, err := findPost(ctx, r.db.Where("id = ?", id))
	if err != nil {
		return nil, err
	}
	return model.toDomain(), nil
}

func (r *postCommandRepository) Save(ctx context.Context, post *domain.Post) error {
	if post.ID == 0 {
		var model postModel
		model.fromDomain(post)
		if err := r.db.WithContext(ctx).Create(&model).Error; err != nil {
			return fmt.Errorf("%w: %v", problem.ErrRepositoryFailure, err)
		}
		post.ID = int(model.ID)
		post.CreatedAt = model.CreatedAt
		post.UpdatedAt = model.UpdatedAt
		return nil
	}

	model := postModel{Model: gorm.Model{ID: uint(post.ID)}}
	model.fromDomain(post)
	result := r.db.WithContext(ctx).Model(&model).Select("Title", "Content", "Tags", "LikeCount").Updates(&model)
	if result.Error != nil {
		return fmt.Errorf("%w: %v", problem.ErrRepositoryFailure, result.Error)
	}
	if result.RowsAffected == 0 {
		return problem.ErrPostNotFound
	}
	post.UpdatedAt = model.UpdatedAt
	return nil
}

func (r *postCommandRepository) Delete(ctx context.Context, post *domain.Post) error {
	result := r.db.WithContext(ctx).Delete(&postModel{}, post.ID)
	if result.Error != nil {
		return fmt.Errorf("%w: %v", problem.ErrRepositoryFailure, result.Error)
	}
	if result.RowsAffected == 0 {
		return problem.ErrPostNotFound
	}
	return nil
}

func mapPostReadError(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return problem.ErrPostNotFound
	}
	return fmt.Errorf("%w: %v", problem.ErrRepositoryFailure, err)
}
