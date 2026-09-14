package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/billykore/project-one/internal/core/domain"
	"github.com/billykore/project-one/internal/core/ports"
	vo "github.com/billykore/project-one/internal/core/valueobject"
	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

type featureFlagModel struct {
	ID          uint   `gorm:"primaryKey"`
	Key         string `gorm:"size:128;notNull"`
	Name        string `gorm:"size:255;notNull"`
	Purpose     string `gorm:"type:text;notNull"`
	Owner       string `gorm:"size:255;notNull"`
	Lifecycle   string `gorm:"size:16;notNull;default:active"`
	SafeDefault bool   `gorm:"notNull;default:false"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (m *featureFlagModel) TableName() string { return "feature_flags" }

type featureFlagSettingModel struct {
	ID                uint   `gorm:"primaryKey"`
	FlagID            uint   `gorm:"notNull"`
	Environment       string `gorm:"size:16;notNull"`
	Mode              string `gorm:"size:16;notNull;default:disabled_all"`
	RolloutPercentage int    `gorm:"notNull;default:0"`
	Revision          int    `gorm:"notNull;default:1"`
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

func (m *featureFlagSettingModel) TableName() string { return "feature_flag_settings" }

type featureFlagOverrideModel struct {
	ID          uint   `gorm:"primaryKey"`
	FlagID      uint   `gorm:"notNull"`
	Environment string `gorm:"size:16;notNull"`
	Username    string `gorm:"size:255;notNull"`
	Type        string `gorm:"size:8;notNull"`
	CreatedAt   time.Time
}

func (m *featureFlagOverrideModel) TableName() string { return "feature_flag_overrides" }

type featureFlagAuditModel struct {
	ID            uint    `gorm:"primaryKey"`
	FlagID        uint    `gorm:"notNull"`
	Environment   *string `gorm:"size:16"`
	Field         string  `gorm:"size:64;notNull"`
	PreviousValue string  `gorm:"type:text"`
	NewValue      string  `gorm:"type:text"`
	Actor         string  `gorm:"size:255;notNull"`
	Reason        string  `gorm:"type:text;notNull"`
	CreatedAt     time.Time
}

func (m *featureFlagAuditModel) TableName() string { return "feature_flag_audit_logs" }

type featureFlagRepository struct {
	db *gorm.DB
}

const maxFeatureFlagSnapshotRows = 10000

// NewFeatureFlagRepository creates a new instance of FeatureFlagRepository.
func NewFeatureFlagRepository(db *gorm.DB) ports.FeatureFlagRepository {
	return &featureFlagRepository{db: db}
}

func normalizeKey(key string) string {
	return strings.ToLower(strings.TrimSpace(key))
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}

func (r *featureFlagRepository) Create(ctx context.Context, flag *domain.FeatureFlag) error {
	m := featureFlagModel{
		Key:         normalizeKey(flag.Key),
		Name:        flag.Name,
		Purpose:     flag.Purpose,
		Owner:       flag.Owner,
		Lifecycle:   string(flag.Lifecycle),
		SafeDefault: flag.SafeDefault,
	}
	if err := r.db.WithContext(ctx).Create(&m).Error; err != nil {
		if isUniqueViolation(err) {
			return domain.ErrFlagKeyExists
		}
		return fmt.Errorf("%w: %v", domain.ErrRepositoryFailure, err)
	}
	flag.ID = int(m.ID)
	flag.Key = m.Key
	flag.CreatedAt = m.CreatedAt
	flag.UpdatedAt = m.UpdatedAt
	return nil
}

func (r *featureFlagRepository) GetByKey(ctx context.Context, key string) (*domain.FeatureFlag, error) {
	var m featureFlagModel
	err := r.db.WithContext(ctx).
		Where("lower(trim(key)) = ?", normalizeKey(key)).
		First(&m).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrFlagNotFound
		}
		return nil, fmt.Errorf("%w: %v", domain.ErrRepositoryFailure, err)
	}
	return m.toDomain(), nil
}

func (r *featureFlagRepository) List(ctx context.Context) ([]domain.FeatureFlag, error) {
	var models []featureFlagModel
	if err := r.db.WithContext(ctx).Order("key asc").Find(&models).Error; err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrRepositoryFailure, err)
	}
	flags := make([]domain.FeatureFlag, 0, len(models))
	for i := range models {
		flags = append(flags, *models[i].toDomain())
	}
	return flags, nil
}

func (r *featureFlagRepository) UpdateMetadata(ctx context.Context, flag *domain.FeatureFlag) error {
	m := featureFlagModel{
		ID:          uint(flag.ID),
		Name:        flag.Name,
		Purpose:     flag.Purpose,
		Owner:       flag.Owner,
		SafeDefault: flag.SafeDefault,
	}
	res := r.db.WithContext(ctx).Model(&m).
		Select("Name", "Purpose", "Owner", "SafeDefault", "UpdatedAt").
		Where("id = ?", flag.ID).
		Updates(&m)
	if res.Error != nil {
		return fmt.Errorf("%w: %v", domain.ErrRepositoryFailure, res.Error)
	}
	flag.UpdatedAt = m.UpdatedAt
	return nil
}

func (r *featureFlagRepository) Archive(ctx context.Context, key string) error {
	res := r.db.WithContext(ctx).Model(&featureFlagModel{}).
		Where("lower(trim(key)) = ?", normalizeKey(key)).
		Update("lifecycle", string(domain.LifecycleArchived))
	if res.Error != nil {
		return fmt.Errorf("%w: %v", domain.ErrRepositoryFailure, res.Error)
	}
	return nil
}

func (r *featureFlagRepository) GetSetting(ctx context.Context, flagID int, env domain.Environment) (*domain.EnvironmentSetting, error) {
	var m featureFlagSettingModel
	err := r.db.WithContext(ctx).
		Where("flag_id = ? AND environment = ?", flagID, string(env)).
		First(&m).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrFlagSettingNotFound
		}
		return nil, fmt.Errorf("%w: %v", domain.ErrRepositoryFailure, err)
	}
	return m.toDomain(), nil
}

func (r *featureFlagRepository) ListSettings(ctx context.Context, flagID int) ([]domain.EnvironmentSetting, error) {
	var models []featureFlagSettingModel
	if err := r.db.WithContext(ctx).
		Where("flag_id = ?", flagID).
		Order("environment asc").
		Find(&models).Error; err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrRepositoryFailure, err)
	}
	settings := make([]domain.EnvironmentSetting, 0, len(models))
	for i := range models {
		settings = append(settings, *models[i].toDomain())
	}
	return settings, nil
}

func (r *featureFlagRepository) UpsertSetting(ctx context.Context, setting *domain.EnvironmentSetting) error {
	if setting.ID == 0 {
		m := featureFlagSettingModel{
			FlagID:            uint(setting.FlagID),
			Environment:       string(setting.Environment),
			Mode:              string(setting.Mode),
			RolloutPercentage: setting.RolloutPercentage,
			Revision:          1,
		}
		if err := r.db.WithContext(ctx).Create(&m).Error; err != nil {
			if isUniqueViolation(err) {
				return domain.ErrRevisionConflict
			}
			return fmt.Errorf("%w: %v", domain.ErrRepositoryFailure, err)
		}
		setting.ID = int(m.ID)
		setting.Revision = m.Revision
		setting.CreatedAt = m.CreatedAt
		setting.UpdatedAt = m.UpdatedAt
		return nil
	}

	res := r.db.WithContext(ctx).Model(&featureFlagSettingModel{}).
		Where("id = ? AND revision = ?", setting.ID, setting.Revision).
		Updates(map[string]any{
			"mode":               string(setting.Mode),
			"rollout_percentage": setting.RolloutPercentage,
			"revision":           gorm.Expr("revision + 1"),
			"updated_at":         gorm.Expr("now()"),
		})
	if res.Error != nil {
		return fmt.Errorf("%w: %v", domain.ErrRepositoryFailure, res.Error)
	}
	if res.RowsAffected == 0 {
		return domain.ErrRevisionConflict
	}
	setting.Revision = setting.Revision + 1
	return nil
}

func (r *featureFlagRepository) ListOverrides(ctx context.Context, flagID int, env domain.Environment) ([]domain.UserOverride, error) {
	var models []featureFlagOverrideModel
	if err := r.db.WithContext(ctx).
		Where("flag_id = ? AND environment = ?", flagID, string(env)).
		Order("username asc").
		Find(&models).Error; err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrRepositoryFailure, err)
	}
	overrides := make([]domain.UserOverride, 0, len(models))
	for i := range models {
		overrides = append(overrides, *models[i].toDomain())
	}
	return overrides, nil
}

func (r *featureFlagRepository) SetOverrides(ctx context.Context, flagID int, env domain.Environment, includes, excludes []string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("flag_id = ? AND environment = ?", flagID, string(env)).
			Delete(&featureFlagOverrideModel{}).Error; err != nil {
			return fmt.Errorf("%w: %v", domain.ErrRepositoryFailure, err)
		}
		for _, username := range includes {
			if err := tx.Create(&featureFlagOverrideModel{
				FlagID:      uint(flagID),
				Environment: string(env),
				Username:    username,
				Type:        string(domain.OverrideInclude),
			}).Error; err != nil {
				return fmt.Errorf("%w: %v", domain.ErrRepositoryFailure, err)
			}
		}
		for _, username := range excludes {
			if err := tx.Create(&featureFlagOverrideModel{
				FlagID:      uint(flagID),
				Environment: string(env),
				Username:    username,
				Type:        string(domain.OverrideExclude),
			}).Error; err != nil {
				return fmt.Errorf("%w: %v", domain.ErrRepositoryFailure, err)
			}
		}
		return nil
	})
}

func (r *featureFlagRepository) AppendAudit(ctx context.Context, record *domain.AuditRecord) error {
	var env *string
	if record.Environment != nil {
		s := string(*record.Environment)
		env = &s
	}
	m := featureFlagAuditModel{
		FlagID:        uint(record.FlagID),
		Environment:   env,
		Field:         record.Field,
		PreviousValue: record.PreviousValue,
		NewValue:      record.NewValue,
		Actor:         record.Actor,
		Reason:        record.Reason,
	}
	if err := r.db.WithContext(ctx).Create(&m).Error; err != nil {
		return fmt.Errorf("%w: %v", domain.ErrRepositoryFailure, err)
	}
	record.ID = int(m.ID)
	record.CreatedAt = m.CreatedAt
	return nil
}

func (r *featureFlagRepository) ListAudit(ctx context.Context, flagID int, cursor *vo.Cursor, limit int) ([]domain.AuditRecord, bool, error) {
	if limit <= 0 {
		limit = 20
	}
	query := r.db.WithContext(ctx).Where("flag_id = ?", flagID)
	if cursor != nil && cursor.ID > 0 {
		query = query.Where("id < ?", cursor.ID)
	}
	var models []featureFlagAuditModel
	if err := query.Order("id desc").Limit(limit + 1).Find(&models).Error; err != nil {
		return nil, false, fmt.Errorf("%w: %v", domain.ErrRepositoryFailure, err)
	}
	hasMore := len(models) > limit
	if hasMore {
		models = models[:limit]
	}
	records := make([]domain.AuditRecord, 0, len(models))
	for i := range models {
		records = append(records, *models[i].toDomain())
	}
	return records, hasMore, nil
}

func (r *featureFlagRepository) LoadSnapshot(ctx context.Context, env domain.Environment) ([]domain.FlagSnapshot, error) {
	var flags []featureFlagModel
	if err := r.db.WithContext(ctx).Order("key asc").Limit(maxFeatureFlagSnapshotRows + 1).Find(&flags).Error; err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrRepositoryFailure, err)
	}
	if len(flags) > maxFeatureFlagSnapshotRows {
		return nil, fmt.Errorf("%w: feature flag snapshot exceeds %d rows", domain.ErrRepositoryFailure, maxFeatureFlagSnapshotRows)
	}
	if len(flags) == 0 {
		return []domain.FlagSnapshot{}, nil
	}

	var settings []featureFlagSettingModel
	if err := r.db.WithContext(ctx).Where("environment = ?", string(env)).Limit(maxFeatureFlagSnapshotRows + 1).Find(&settings).Error; err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrRepositoryFailure, err)
	}
	if len(settings) > maxFeatureFlagSnapshotRows {
		return nil, fmt.Errorf("%w: feature flag settings snapshot exceeds %d rows", domain.ErrRepositoryFailure, maxFeatureFlagSnapshotRows)
	}
	settingsByFlag := map[uint]*domain.EnvironmentSetting{}
	for i := range settings {
		s := settings[i].toDomain()
		settingsByFlag[settings[i].FlagID] = s
	}

	var overrides []featureFlagOverrideModel
	if err := r.db.WithContext(ctx).Where("environment = ?", string(env)).Limit(maxFeatureFlagSnapshotRows + 1).Find(&overrides).Error; err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrRepositoryFailure, err)
	}
	if len(overrides) > maxFeatureFlagSnapshotRows {
		return nil, fmt.Errorf("%w: feature flag overrides snapshot exceeds %d rows", domain.ErrRepositoryFailure, maxFeatureFlagSnapshotRows)
	}
	overridesByFlag := map[uint][]domain.UserOverride{}
	for i := range overrides {
		o := overrides[i].toDomain()
		overridesByFlag[overrides[i].FlagID] = append(overridesByFlag[overrides[i].FlagID], *o)
	}

	snapshots := make([]domain.FlagSnapshot, 0, len(flags))
	for i := range flags {
		flag := flags[i].toDomain()
		snapshots = append(snapshots, domain.FlagSnapshot{
			Key:         flag.Key,
			Lifecycle:   flag.Lifecycle,
			SafeDefault: flag.SafeDefault,
			Setting:     settingsByFlag[flags[i].ID],
			Overrides:   overridesByFlag[flags[i].ID],
		})
	}
	return snapshots, nil
}

func (m *featureFlagModel) toDomain() *domain.FeatureFlag {
	return &domain.FeatureFlag{
		ID:          int(m.ID),
		Key:         m.Key,
		Name:        m.Name,
		Purpose:     m.Purpose,
		Owner:       m.Owner,
		Lifecycle:   domain.LifecycleState(m.Lifecycle),
		SafeDefault: m.SafeDefault,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}
}

func (m *featureFlagSettingModel) toDomain() *domain.EnvironmentSetting {
	return &domain.EnvironmentSetting{
		ID:                int(m.ID),
		FlagID:            int(m.FlagID),
		Environment:       domain.Environment(m.Environment),
		Mode:              domain.AvailabilityMode(m.Mode),
		RolloutPercentage: m.RolloutPercentage,
		Revision:          m.Revision,
		CreatedAt:         m.CreatedAt,
		UpdatedAt:         m.UpdatedAt,
	}
}

func (m *featureFlagOverrideModel) toDomain() *domain.UserOverride {
	return &domain.UserOverride{
		ID:          int(m.ID),
		FlagID:      int(m.FlagID),
		Environment: domain.Environment(m.Environment),
		Username:    m.Username,
		Type:        domain.OverrideType(m.Type),
		CreatedAt:   m.CreatedAt,
	}
}

func (m *featureFlagAuditModel) toDomain() *domain.AuditRecord {
	var env *domain.Environment
	if m.Environment != nil {
		e := domain.Environment(*m.Environment)
		env = &e
	}
	return &domain.AuditRecord{
		ID:            int(m.ID),
		FlagID:        int(m.FlagID),
		Environment:   env,
		Field:         m.Field,
		PreviousValue: m.PreviousValue,
		NewValue:      m.NewValue,
		Actor:         m.Actor,
		Reason:        m.Reason,
		CreatedAt:     m.CreatedAt,
	}
}
