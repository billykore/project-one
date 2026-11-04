package valueobject

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCursorEncodeDecodeRoundTrip(t *testing.T) {
	original := Cursor{
		CreatedAt: time.Date(2026, time.September, 13, 14, 30, 0, 0, time.UTC),
		ID:        42,
		Key:       "alice",
		Rank:      2,
		Score:     0.875,
	}

	decoded, err := DecodeCursor(original.Encode())
	require.NoError(t, err)
	assert.Equal(t, original, decoded)
}

func TestDecodeCursorRejectsInvalidValue(t *testing.T) {
	_, err := DecodeCursor("not-a-cursor")
	assert.Error(t, err)
}
