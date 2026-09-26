package repository

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/billykore/project-one/internal/platform/problem"
	"github.com/billykore/project-one/internal/publishing/domain"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestPostCommandAndQueryRepositoryNotFoundAndSoftDelete(t *testing.T) {
	dsn := os.Getenv("POSTS_TEST_DSN")
	if dsn == "" {
		dsn = os.Getenv("FEATURE_FLAGS_TEST_DSN")
	}
	if dsn == "" {
		t.Skip("POSTS_TEST_DSN is not configured")
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{TranslateError: true})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	defer func() { require.NoError(t, sqlDB.Close()) }()

	command := NewPostCommandRepository(db)
	query := NewPostQueryRepository(db)
	ctx := context.Background()
	username := "cqrs-repository-test-" + time.Now().UTC().Format("20060102150405.000000000")
	post := &domain.Post{Username: username, Title: "CQRS", Content: "test"}
	t.Cleanup(func() { _ = db.Unscoped().Where("username = ?", username).Delete(&postModel{}).Error })

	require.NoError(t, command.Save(ctx, post))
	loaded, err := query.GetByID(ctx, post.ID)
	require.NoError(t, err)
	require.Equal(t, post.ID, loaded.ID)

	require.NoError(t, command.Delete(ctx, post))
	_, err = query.GetByID(ctx, post.ID)
	require.ErrorIs(t, err, problem.ErrPostNotFound)
	posts, err := query.GetUserPosts(ctx, post.UserID, nil, 10)
	require.NoError(t, err)
	require.Empty(t, posts)

	_, err = command.Load(ctx, post.ID)
	require.True(t, errors.Is(err, problem.ErrPostNotFound))
}
