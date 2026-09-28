// Package database carries an active database transaction through a context.
package database

import (
	"context"

	"gorm.io/gorm"
)

type transactionKey struct{}

// WithTransaction returns a child context that makes Repository return tx.
func WithTransaction(ctx context.Context, tx *gorm.DB) context.Context {
	return context.WithValue(ctx, transactionKey{}, tx)
}

// FromContext returns the active transaction when present, otherwise fallback.
func FromContext(ctx context.Context, fallback *gorm.DB) *gorm.DB {
	if tx, ok := ctx.Value(transactionKey{}).(*gorm.DB); ok && tx != nil {
		return tx
	}
	return fallback
}
