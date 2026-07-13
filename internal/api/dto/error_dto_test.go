package dto

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAPIErrorResponse_Marshal_Full(t *testing.T) {
	resp := APIErrorResponse{
		Error: StructuredError{
			Code:      "INVALID_ARGUMENT",
			Message:   "Validation failed",
			RequestID: "req_abc123",
			Details: []ErrorDetail{
				{Field: "username", Reason: "min", Message: "username must be at least 3 characters"},
			},
		},
	}

	b, err := json.Marshal(resp)
	assert.NoError(t, err)

	var m map[string]interface{}
	err = json.Unmarshal(b, &m)
	assert.NoError(t, err)

	errObj := m["error"].(map[string]interface{})
	assert.Equal(t, "INVALID_ARGUMENT", errObj["code"])
	assert.Equal(t, "req_abc123", errObj["request_id"])
	assert.NotNil(t, errObj["details"])
}

func TestAPIErrorResponse_Marshal_NoDetails(t *testing.T) {
	resp := APIErrorResponse{
		Error: StructuredError{
			Code:      "NOT_FOUND",
			Message:   "User not found",
			RequestID: "req_xyz",
		},
	}

	b, err := json.Marshal(resp)
	assert.NoError(t, err)

	var m map[string]interface{}
	err = json.Unmarshal(b, &m)
	assert.NoError(t, err)

	errObj := m["error"].(map[string]interface{})
	// details must be absent, not empty array
	_, hasDetails := errObj["details"]
	assert.False(t, hasDetails)
}
