package repository

import (
	"context"
	"fmt"

	vo "github.com/billykore/project-one/internal/platform/pagination"
	"github.com/billykore/project-one/internal/platform/problem"
	"github.com/billykore/project-one/internal/publishing/domain"
	"github.com/billykore/project-one/internal/publishing/ports"
	"gorm.io/gorm"
)

type postQueryRepository struct{ db *gorm.DB }

func NewPostQueryRepository(db *gorm.DB) ports.PostQueryRepository {
	return &postQueryRepository{db: db}
}

func (r *postQueryRepository) GetByID(ctx context.Context, id int) (*domain.Post, error) {
	model, err := findPost(ctx, r.db.Where("id = ?", id))
	if err != nil {
		return nil, err
	}
	return model.toDomain(), nil
}

func (r *postQueryRepository) GetUserPosts(ctx context.Context, userID int, cursor *vo.Cursor, limit int) ([]*domain.Post, error) {
	return findPosts(r.db.WithContext(ctx).Where("user_id = ?", userID).Where("deleted_at IS NULL"), cursor, limit)
}

func (r *postQueryRepository) GetFeed(ctx context.Context, userIDs []int, cursor *vo.Cursor, limit int) ([]*domain.Post, error) {
	if len(userIDs) == 0 {
		return []*domain.Post{}, nil
	}
	return findPosts(r.db.WithContext(ctx).Model(&postModel{}).Where("user_id IN ?", userIDs).Where("deleted_at IS NULL"), cursor, limit)
}

func findPost(ctx context.Context, query *gorm.DB) (*postModel, error) {
	var model postModel
	if err := query.WithContext(ctx).First(&model).Error; err != nil {
		return nil, mapPostReadError(err)
	}
	return &model, nil
}

func findPosts(query *gorm.DB, cursor *vo.Cursor, limit int) ([]*domain.Post, error) {
	if cursor != nil && !cursor.CreatedAt.IsZero() && cursor.ID > 0 {
		query = query.Where("(created_at, id) < (?, ?)", cursor.CreatedAt, cursor.ID)
	}
	var models []postModel
	if err := query.Order("created_at DESC, id DESC").Limit(limit).Find(&models).Error; err != nil {
		return nil, fmt.Errorf("%w: %v", problem.ErrRepositoryFailure, err)
	}
	posts := make([]*domain.Post, 0, len(models))
	for _, model := range models {
		posts = append(posts, model.toDomain())
	}
	return posts, nil
}
