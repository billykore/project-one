package repository

import (
	"github.com/billykore/project-one/internal/core/domain"
	"github.com/lib/pq"
	"gorm.io/gorm"
)

type postModel struct {
	gorm.Model
	UserID    int            `gorm:"default:0"`
	Username  string         `gorm:"size:255;notNull"`
	Title     string         `gorm:"size:255;notNull"`
	Content   string         `gorm:"type:text;notNull"`
	Tags      pq.StringArray `gorm:"type:text[]"`
	LikeCount int            `gorm:"default:0"`
}

func (m *postModel) TableName() string { return "posts" }

func (m *postModel) fromDomain(post *domain.Post) {
	m.UserID = post.UserID
	m.Username = post.Username
	m.Title = post.Title
	m.Content = post.Content
	m.Tags = pq.StringArray(post.Tags)
	m.LikeCount = post.LikeCount
}

func (m *postModel) toDomain() *domain.Post {
	return &domain.Post{ID: int(m.ID), UserID: m.UserID, Username: m.Username, Title: m.Title, Content: m.Content, Tags: []string(m.Tags), LikeCount: m.LikeCount, CreatedAt: m.CreatedAt, UpdatedAt: m.UpdatedAt}
}
