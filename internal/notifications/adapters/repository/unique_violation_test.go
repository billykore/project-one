package repository

import (
	"errors"
	"fmt"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestIsUniqueViolation(t *testing.T) {
	assert.True(t, isUniqueViolation(gorm.ErrDuplicatedKey), "translated by GORM")
	assert.True(t, isUniqueViolation(fmt.Errorf("insert: %w", &pgconn.PgError{Code: "23505"})), "raw PostgreSQL error")
	assert.False(t, isUniqueViolation(&pgconn.PgError{Code: "23503"}), "foreign key violation")
	assert.False(t, isUniqueViolation(errors.New("duplicate key value violates unique constraint")), "message text alone")
}
