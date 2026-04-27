package repository

import (
	"context"
	"errors"
	"time"

	"github.com/billykore/project-one/internal/app/user/core/domain"
	"github.com/billykore/project-one/internal/app/user/core/ports"
	"gorm.io/gorm"
)

type userModel struct {
	ID        int    `gorm:"primaryKey;autoIncrement"`
	Email     string `gorm:"unique;notNull"`
	Password  string `gorm:"notNull"`
	FirstName string
	LastName  string
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
		FirstName: m.FirstName,
		LastName:  m.LastName,
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
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrUserNotFound
		}
		return nil, err
	}
	return m.toDomain(), nil
}

func (r *postgresUserRepository) GetUserByID(ctx context.Context, id int) (*domain.User, error) {
	var m userModel
	if err := r.db.WithContext(ctx).First(&m, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrUserNotFound
		}
		return nil, err
	}
	return m.toDomain(), nil
}
