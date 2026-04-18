package repository

import (
	"context"
	"time"

	"github.com/billykore/project-one/internal/app/user/core/domain"
	"github.com/billykore/project-one/internal/app/user/core/ports"
	"gorm.io/gorm"
)

type userModel struct {
	ID        int    `gorm:"primaryKey;autoIncrement"`
	Email     string `gorm:"unique;notNull"`
	Password  string `gorm:"notNull"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (m *userModel) TableName() string {
	return "users"
}

func (m *userModel) toDomain() *domain.User {
	return &domain.User{
		ID:        m.ID,
		Email:     m.Email,
		Password:  m.Password,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}
}

type postgresUserRepository struct {
	db *gorm.DB
}

// NewPostgresUserRepository creates a new instance of UserRepository.
func NewPostgresUserRepository(db *gorm.DB) ports.UserRepository {
	return &postgresUserRepository{db: db}
}

func (r *postgresUserRepository) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	var m userModel
	if err := r.db.WithContext(ctx).Where("email = ?", email).First(&m).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, domain.ErrUserNotFound
		}
		return nil, err
	}
	return m.toDomain(), nil
}
