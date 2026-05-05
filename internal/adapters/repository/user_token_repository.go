package repository

import (
	"context"
	"time"

	"github.com/billykore/project-one/internal/core/domain"
	"github.com/billykore/project-one/internal/core/ports"
	"gorm.io/gorm"
)

type userTokenModel struct {
	ID        int       `gorm:"primaryKey;autoIncrement"`
	UserID    int       `gorm:"notNull"`
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
	return r.db.WithContext(ctx).Create(m).Error
}

func (r *userTokenRepository) GetTokenByUserID(ctx context.Context, userID int) (*domain.UserToken, error) {
	var m userTokenModel
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&m).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil // No token found for this user
		}
		return nil, err
	}
	return &domain.UserToken{
		ID:        m.ID,
		UserID:    m.UserID,
		Token:     m.Token,
		ExpiresAt: m.ExpiresAt,
	}, nil
}

func (r *userTokenRepository) DeleteTokenByUserID(ctx context.Context, userID int) error {
	return r.db.WithContext(ctx).Where("user_id = ?", userID).Delete(&userTokenModel{}).Error
}
