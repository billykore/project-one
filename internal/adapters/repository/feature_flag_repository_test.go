package repository

import (
	"context"
	"os"
	"testing"

	"github.com/billykore/project-one/internal/core/domain"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestFeatureFlagRepositoryPersistenceAndRevisionConflict(t *testing.T) {
	dsn := os.Getenv("FEATURE_FLAGS_TEST_DSN")
	if dsn == "" {
		t.Skip("FEATURE_FLAGS_TEST_DSN is not configured")
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{TranslateError: true})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	defer func() { require.NoError(t, sqlDB.Close()) }()
	require.NoError(t, sqlDB.Ping())

	repository := NewFeatureFlagRepository(db)
	ctx := context.Background()
	flag := &domain.FeatureFlag{
		Key:         "integration-feature-flag",
		Name:        "Integration Flag",
		Purpose:     "Repository integration test",
		Owner:       "test",
		Lifecycle:   domain.LifecycleActive,
		SafeDefault: false,
	}

	_ = db.Where("lower(trim(key)) = ?", flag.Key).Delete(&featureFlagModel{}).Error
	t.Cleanup(func() { _ = db.Where("lower(trim(key)) = ?", flag.Key).Delete(&featureFlagModel{}).Error })

	require.NoError(t, repository.Create(ctx, flag))
	loaded, err := repository.GetByKey(ctx, " INTEGRATION-FEATURE-FLAG ")
	require.NoError(t, err)
	require.Equal(t, flag.ID, loaded.ID)

	setting := &domain.EnvironmentSetting{
		FlagID:            flag.ID,
		Environment:       domain.EnvironmentTest,
		Mode:              domain.ModeDisabledAll,
		RolloutPercentage: 0,
	}
	require.NoError(t, repository.UpsertSetting(ctx, setting))
	stale := *setting
	stale.Mode = domain.ModeEnabledAll
	require.NoError(t, repository.UpsertSetting(ctx, setting))
	require.ErrorIs(t, repository.UpsertSetting(ctx, &stale), domain.ErrRevisionConflict)
}
