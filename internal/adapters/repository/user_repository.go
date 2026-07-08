package repository

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/billykore/project-one/internal/core/domain"
	"github.com/billykore/project-one/internal/core/ports"
	"gorm.io/gorm"
)

type userModel struct {
	ID        int    `gorm:"primaryKey;autoIncrement"`
	Email     string `gorm:"unique;notNull"`
	Username  string `gorm:"unique;notNull"`
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
		Username:  m.Username,
		Password:  m.Password,
		FirstName: m.FirstName,
		LastName:  m.LastName,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}
}

type userRepository struct {
	db *gorm.DB
}

// NewUserRepository creates a new instance of UserRepository.
func NewUserRepository(db *gorm.DB) ports.UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	var m userModel
	if err := r.db.WithContext(ctx).Where("email = ?", email).First(&m).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrUserNotFound
		}
		return nil, err
	}
	return m.toDomain(), nil
}

func (r *userRepository) GetUserByUsername(ctx context.Context, username string) (*domain.User, error) {
	var m userModel
	if err := r.db.WithContext(ctx).Where("username = ?", username).First(&m).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrUserNotFound
		}
		return nil, err
	}
	return m.toDomain(), nil
}

func (r *userRepository) GetUserByID(ctx context.Context, id int) (*domain.User, error) {
	var m userModel
	if err := r.db.WithContext(ctx).First(&m, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrUserNotFound
		}
		return nil, err
	}
	return m.toDomain(), nil
}

func (r *userRepository) CreateUser(ctx context.Context, user *domain.User) error {
	m := userModel{
		Email:     user.Email,
		Username:  user.Username,
		Password:  user.Password,
		FirstName: user.FirstName,
		LastName:  user.LastName,
	}
	if err := r.db.WithContext(ctx).Create(&m).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			if strings.Contains(err.Error(), "username") {
				return domain.ErrUsernameAlreadyTaken
			}
			return domain.ErrEmailAlreadyRegistered
		}
		return err
	}
	user.ID = m.ID
	return nil
}

func (r *userRepository) UpdateUser(ctx context.Context, user *domain.User) error {
	m := userModel{
		ID:        user.ID,
		Email:     user.Email,
		Username:  user.Username,
		Password:  user.Password,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		CreatedAt: user.CreatedAt,
		UpdatedAt: time.Now(),
	}
	// Save updates all fields by primary key
	return r.db.WithContext(ctx).Save(&m).Error
}

// UpdateProfile updates the user's profile fields and cascades the username
// change to all denormalized columns within a single database transaction.
// oldUsername is the user's current username before the update; it is used
// in WHERE clauses to find rows that need the cascade.
func (r *userRepository) UpdateProfile(ctx context.Context, oldUsername string, user *domain.User) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Update the users row.
		result := tx.Model(&userModel{}).Where("id = ?", user.ID).Updates(map[string]interface{}{
			"first_name": user.FirstName,
			"last_name":  user.LastName,
			"username":   user.Username,
			"updated_at": time.Now(),
		})
		if result.Error != nil {
			if errors.Is(result.Error, gorm.ErrDuplicatedKey) {
				return domain.ErrUsernameAlreadyTaken
			}
			return result.Error
		}

		// Only cascade if the username actually changed.
		if oldUsername == user.Username {
			return nil
		}

		// Cascade username to follows table (both follower and followed columns).
		if err := tx.Model(&followModel{}).Where("follower_username = ?", oldUsername).
			Update("follower_username", user.Username).Error; err != nil {
			return err
		}
		if err := tx.Model(&followModel{}).Where("followed_username = ?", oldUsername).
			Update("followed_username", user.Username).Error; err != nil {
			return err
		}

		// Cascade username to posts table.
		if err := tx.Model(&postModel{}).Where("username = ?", oldUsername).
			Update("username", user.Username).Error; err != nil {
			return err
		}

		// Cascade username to user_tokens table.
		if err := tx.Model(&userTokenModel{}).Where("username = ?", oldUsername).
			Update("username", user.Username).Error; err != nil {
			return err
		}

		// Cascade username to comments table.
		if err := tx.Model(&commentModel{}).Where("username = ?", oldUsername).
			Update("username", user.Username).Error; err != nil {
			return err
		}

		// Cascade username to post_likes table.
		if err := tx.Model(&likeModel{}).Where("username = ?", oldUsername).
			Update("username", user.Username).Error; err != nil {
			return err
		}

		return nil
	})
}
