package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/billykore/project-one/internal/core/domain"
	"github.com/billykore/project-one/internal/core/ports"
	"gorm.io/gorm"
)

type userTokenModel struct {
	ID        int       `gorm:"primaryKey;autoIncrement"`
	UserID    int       `gorm:"notNull"`
	Username  string    `gorm:"notNull"`
	Token     string    `gorm:"notNull"`
	ExpiresAt time.Time `gorm:"notNull"`
	CreatedAt time.Time
}

func (m *userTokenModel) TableName() string {
	return "user_tokens"
}

func fromDomainUserToken(token *domain.UserToken) *userTokenModel {
	return &userTokenModel{
		ID:        token.ID,
		UserID:    token.UserID,
		Username:  token.Username,
		Token:     token.Token,
		ExpiresAt: token.ExpiresAt,
	}
}

type userTokenRepository struct {
	db *gorm.DB
}

// NewUserTokenRepository creates a new instance of TokenRepository.
func NewUserTokenRepository(db *gorm.DB) ports.TokenRepository {
	return &userTokenRepository{db: db}
}

func (r *userTokenRepository) StoreToken(ctx context.Context, token *domain.UserToken) error {
	m := fromDomainUserToken(token)
	err := r.db.WithContext(ctx).Create(m).Error
	if err != nil {
		return fmt.Errorf("%w: %v", domain.ErrRepositoryFailure, err)
	}
	return nil
}

// IsActive verifies that the presented JWT belongs to an active, unexpired
// persisted session for the same user. JWT verification remains the
// responsibility of the token service; this repository only supplies the
// server-side revocation check.
func (r *userTokenRepository) IsActive(ctx context.Context, token string, userID int) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&userTokenModel{}).
		Where("token = ? AND user_id = ? AND expires_at > ?", token, userID, time.Now()).
		Count(&count).Error
	if err != nil {
		return false, fmt.Errorf("%w: %v", domain.ErrRepositoryFailure, err)
	}
	return count == 1, nil
}

func (r *userTokenRepository) GetTokenByUsername(ctx context.Context, username string) (*domain.UserToken, error) {
	var m userTokenModel
	err := r.db.WithContext(ctx).Where("username = ?", username).First(&m).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil // No token found for this user
		}
		return nil, fmt.Errorf("%w: %v", domain.ErrRepositoryFailure, err)
	}
	return &domain.UserToken{
		ID:        m.ID,
		UserID:    m.UserID,
		Username:  m.Username,
		Token:     m.Token,
		ExpiresAt: m.ExpiresAt,
	}, nil
}

func (r *userTokenRepository) DeleteTokenByUsername(ctx context.Context, username string) error {
	err := r.db.WithContext(ctx).Where("username = ?", username).Delete(&userTokenModel{}).Error
	if err != nil {
		return fmt.Errorf("%w: %v", domain.ErrRepositoryFailure, err)
	}
	return nil
}
