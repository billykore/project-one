package dto

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProblemDetail_Marshal_Full(t *testing.T) {
	resp := ProblemDetail{
		Type:      "https://api.project-one.dev/errors/invalid-argument",
		Title:     "Bad Request",
		Status:    400,
		Detail:    "Validation failed",
		Instance:  "/auth/register",
		Code:      "INVALID_ARGUMENT",
		RequestID: "req_abc123",
		Errors: []ValidationError{
			{Field: "username", Reason: "min", Message: "username must be at least 3 characters"},
		},
	}

	b, err := json.Marshal(resp)
	assert.NoError(t, err)

	var m map[string]interface{}
	err = json.Unmarshal(b, &m)
	assert.NoError(t, err)

	assert.Equal(t, "https://api.project-one.dev/errors/invalid-argument", m["type"])
	assert.Equal(t, "Bad Request", m["title"])
	assert.Equal(t, float64(400), m["status"])
	assert.Equal(t, "Validation failed", m["detail"])
	assert.Equal(t, "INVALID_ARGUMENT", m["code"])
	assert.Equal(t, "req_abc123", m["request_id"])
	assert.NotNil(t, m["errors"])
}

func TestProblemDetail_Marshal_NoExtensions(t *testing.T) {
	resp := ProblemDetail{
		Type:      "https://api.project-one.dev/errors/not-found",
		Title:     "Not Found",
		Status:    404,
		Detail:    "User not found",
		Instance:  "/users/nonexistent",
		Code:      "NOT_FOUND",
		RequestID: "req_xyz",
	}

	b, err := json.Marshal(resp)
	assert.NoError(t, err)

	var m map[string]interface{}
	err = json.Unmarshal(b, &m)
	assert.NoError(t, err)

	assert.Equal(t, "Not Found", m["title"])
	assert.Equal(t, float64(404), m["status"])
	// errors must be absent, not empty array
	_, hasErrors := m["errors"]
	assert.False(t, hasErrors)
}

func TestProblemDetail_UnknownError(t *testing.T) {
	resp := ProblemDetail{
		Type:      "about:blank",
		Title:     "Internal Server Error",
		Status:    500,
		Detail:    "Internal server error",
		Instance:  "/some/path",
		Code:      "INTERNAL",
		RequestID: "req_panic_001",
	}

	b, err := json.Marshal(resp)
	assert.NoError(t, err)

	var m map[string]interface{}
	err = json.Unmarshal(b, &m)
	assert.NoError(t, err)

	assert.Equal(t, "about:blank", m["type"])
	assert.Equal(t, "Internal Server Error", m["title"])
	assert.Equal(t, float64(500), m["status"])
}
