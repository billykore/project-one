package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/billykore/project-one/internal/core/domain"
	"github.com/billykore/project-one/internal/core/ports"
	"gorm.io/gorm"
)

type notificationModel struct {
	ID        uint      `gorm:"primaryKey;autoIncrement"`
	UserID    uint      `gorm:"column:user_id;notNull"`
	ActorID   uint      `gorm:"column:actor_id;notNull"`
	Type      string    `gorm:"column:type;size:50;notNull"`
	PostID    uint      `gorm:"column:post_id;default:null"`
	CommentID uint      `gorm:"column:comment_id;default:null"`
	IsRead    bool      `gorm:"column:is_read;default:false"`
	CreatedAt time.Time `gorm:"column:created_at;notNull"`
}

func (notificationModel) TableName() string {
	return "notifications"
}

func (m *notificationModel) toDomain() *domain.Notification {
	return &domain.Notification{
		ID:        int(m.ID),
		UserID:    int(m.UserID),
		ActorID:   int(m.ActorID),
		Type:      domain.NotificationType(m.Type),
		PostID:    int(m.PostID),
		CommentID: int(m.CommentID),
		IsRead:    m.IsRead,
		CreatedAt: m.CreatedAt,
	}
}

func (m *notificationModel) fromDomain(n *domain.Notification) {
	m.ID = uint(n.ID)
	m.UserID = uint(n.UserID)
	m.ActorID = uint(n.ActorID)
	m.Type = string(n.Type)
	m.PostID = uint(n.PostID)
	m.CommentID = uint(n.CommentID)
	m.IsRead = n.IsRead
	m.CreatedAt = n.CreatedAt
}

type notificationRepository struct {
	db *gorm.DB
}

func NewNotificationRepository(db *gorm.DB) ports.NotificationRepository {
	return &notificationRepository{db: db}
}

func (r *notificationRepository) Create(ctx context.Context, n *domain.Notification) error {
	var m notificationModel
	m.fromDomain(n)
	if m.CreatedAt.IsZero() {
		m.CreatedAt = time.Now()
	}
	if err := r.db.WithContext(ctx).Create(&m).Error; err != nil {
		return fmt.Errorf("%w: %v", domain.ErrRepositoryFailure, err)
	}
	n.ID = int(m.ID)
	n.CreatedAt = m.CreatedAt
	return nil
}

func (r *notificationRepository) GetByID(ctx context.Context, id int) (*domain.Notification, error) {
	var m notificationModel
	if err := r.db.WithContext(ctx).First(&m, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotificationNotFound
		}
		return nil, fmt.Errorf("%w: %v", domain.ErrRepositoryFailure, err)
	}
	return m.toDomain(), nil
}

func (r *notificationRepository) GetByUserID(ctx context.Context, userID int, limit, offset int) ([]*domain.Notification, error) {
	var models []notificationModel
	query := r.db.WithContext(ctx).Where("user_id = ?", userID).Order("created_at DESC")
	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}
	if err := query.Find(&models).Error; err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrRepositoryFailure, err)
	}

	notifications := make([]*domain.Notification, len(models))
	for i, m := range models {
		notifications[i] = m.toDomain()
	}
	return notifications, nil
}

func (r *notificationRepository) MarkAsRead(ctx context.Context, id int) error {
	err := r.db.WithContext(ctx).Model(&notificationModel{}).Where("id = ?", id).Update("is_read", true).Error
	if err != nil {
		return fmt.Errorf("%w: %v", domain.ErrRepositoryFailure, err)
	}
	return nil
}

func (r *notificationRepository) MarkAllAsRead(ctx context.Context, userID int) error {
	err := r.db.WithContext(ctx).Model(&notificationModel{}).Where("user_id = ?", userID).Update("is_read", true).Error
	if err != nil {
		return fmt.Errorf("%w: %v", domain.ErrRepositoryFailure, err)
	}
	return nil
}
