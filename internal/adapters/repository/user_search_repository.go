package repository

import (
	"context"
	"fmt"

	"github.com/billykore/project-one/internal/core/domain"
	"github.com/billykore/project-one/internal/core/ports"
	vo "github.com/billykore/project-one/internal/core/valueobject"
	"gorm.io/gorm"
)

type userSearchRepository struct {
	db *gorm.DB
}

// NewUserSearchRepository creates a new instance of UserSearchRepository.
func NewUserSearchRepository(db *gorm.DB) ports.UserSearchRepository {
	return &userSearchRepository{db: db}
}

// Search executes a hybrid query that ranks results:
// 1. exact match (username = query)
// 2. prefix match (username LIKE 'query%')
// 3. trigram fuzzy match (pg_trgm similarity)
// Within each tier, results are ordered by trigram similarity descending.
// cursor.ID is used as the offset (nil cursor = offset 0).
func (r *userSearchRepository) Search(ctx context.Context, query string, cursor *vo.Cursor, limit int) ([]domain.SearchResult, *vo.Cursor, bool, error) {
	offset := 0
	if cursor != nil {
		offset = cursor.ID
	}

	type row struct {
		Username  string
		FirstName string
		LastName  string
	}

	// Fetch limit+1 to detect has_more.
	var rows []row
	err := r.db.WithContext(ctx).Raw(`
SELECT username, first_name, last_name
FROM users
WHERE username % ?
   OR username LIKE ? || '%'
ORDER BY
	CASE WHEN username = ? THEN 0
	     WHEN username LIKE ? || '%' THEN 1
	     ELSE 2
	END,
	similarity(username, ?) DESC
LIMIT ? OFFSET ?
	`, query, query, query, query, query, limit+1, offset).Scan(&rows).Error
	if err != nil {
		return nil, nil, false, fmt.Errorf("%w: %v", domain.ErrRepositoryFailure, err)
	}

	hasMore := len(rows) > limit
	if hasMore {
		rows = rows[:limit]
	}

	results := make([]domain.SearchResult, 0, len(rows))
	for _, r := range rows {
		results = append(results, domain.SearchResult{
			Username:  r.Username,
			FirstName: r.FirstName,
			LastName:  r.LastName,
		})
	}

	var nextCursor *vo.Cursor
	if hasMore {
		nextCursor = &vo.Cursor{ID: offset + limit}
	}

	return results, nextCursor, hasMore, nil
}
