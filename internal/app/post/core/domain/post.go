package domain

import "time"

type Post struct {
	ID        int
	UserID    int
	Title     string
	Content   string
	Tags      []string
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt time.Time
}
