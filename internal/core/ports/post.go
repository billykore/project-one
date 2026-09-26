package ports

import (
	"context"

	"github.com/billykore/project-one/internal/core/domain"
	vo "github.com/billykore/project-one/internal/core/valueobject"
)

// PostCommandRepository is the driven port for state-changing post commands.
type PostCommandRepository interface {
	// Load returns an aggregate root for a command to act on.
	Load(ctx context.Context, id int) (*domain.Post, error)
	// Save persists the aggregate root, whether new or modified.
	Save(ctx context.Context, post *domain.Post) error
	// Delete removes an aggregate root selected by application/domain logic.
	Delete(ctx context.Context, post *domain.Post) error
}

// PostQueryRepository is the driven port for read-only post queries.
type PostQueryRepository interface {
	// GetByID retrieves a post by its identifier.
	GetByID(ctx context.Context, id int) (*domain.Post, error)
	// GetUserPosts retrieves a cursor-paginated page of posts for a user.
	GetUserPosts(ctx context.Context, userID int, cursor *vo.Cursor, limit int) ([]*domain.Post, error)
	// GetFeed retrieves a cursor-paginated page of posts authored by the given users.
	GetFeed(ctx context.Context, userIDs []int, cursor *vo.Cursor, limit int) ([]*domain.Post, error)
}

// PostCommandUseCase is the driving port for state-changing post actions.
type PostCommandUseCase interface {
	// CreatePost creates a new post with the given details.
	CreatePost(ctx context.Context, user *domain.User, title, content string, tags []string) (*domain.Post, error)
	// UpdatePost updates an existing post for a specific user.
	UpdatePost(ctx context.Context, userID int, postID int, title, content string) (*domain.Post, error)
	// DeletePost removes a post for a specific user.
	DeletePost(ctx context.Context, userID int, postID int) error
	// LikePost likes a post by the authenticated user. If already liked, it behaves idempotently.
	LikePost(ctx context.Context, postID int, actor *domain.User) (likeCount int, err error)
	// UnlikePost unlikes a post by the authenticated user. If not liked, it behaves idempotently.
	UnlikePost(ctx context.Context, postID int, actor *domain.User) (likeCount int, err error)
}

// PostQueryUseCase is the driving port for read-only post actions.
type PostQueryUseCase interface {
	// GetPostByID retrieves a post by its identifier.
	GetPostByID(ctx context.Context, postID int) (*domain.Post, error)
	// GetPosts retrieves posts and cursor-pagination metadata for a user.
	GetPosts(ctx context.Context, userID int, cursor *vo.Cursor, limit int) (posts []*domain.Post, nextCursor *vo.Cursor, hasMore bool, err error)
	// GetLikeStatus retrieves whether a user likes a post and its current like count.
	GetLikeStatus(ctx context.Context, postID int, userID int) (liked bool, likeCount int, err error)
}
