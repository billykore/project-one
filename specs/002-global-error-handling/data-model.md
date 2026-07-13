# Data Model: Global Error Handling Middleware

**Feature**: 002-global-error-handling
**Date**: 2026-07-13

## Entities

### 1. ErrorCode

Machine-readable string constant identifying an error type. Defined in `internal/core/domain/errors.go`.

**Fields**:
| Field | Type | Description |
|-------|------|-------------|
| (value) | `string` | UPPER_SNAKE_CASE identifier, e.g., `"USER_NOT_FOUND"` |

**Validation**: Must be unique across all error codes. Must follow pattern `^[A-Z][A-Z0-9]*(_[A-Z][A-Z0-9]*)*$`.

**Relationships**: Each `ErrorCode` maps to exactly one domain sentinel `error` via the `ErrorMapping` registry.

**State**: Immutable constant. All 18 codes (17 domain + 1 default) defined at compile time.

---

### 2. ErrorMapping

A registry entry associating a domain sentinel error with its HTTP representation.

**Fields**:
| Field | Type | Description |
|-------|------|-------------|
| Status | `int` | HTTP status code (e.g., 400, 404, 500) |
| Code | `ErrorCode` | Machine-readable error code string |
| Message | `string` | Default human-readable message for API consumers |

**Validation**:
- `Status` must be a valid HTTP status code (100–599)
- `Code` must not be empty
- `Message` must not be empty and must not contain internal implementation details

**Relationships**: Stored in `ErrorRegistry.mappings` map keyed by sentinel `error` value. Lookup walks error chain with `errors.Is()`.

**State transitions**: None. Mappings are registered once at initialization and never modified at runtime.

---

### 3. ErrorResponse

The JSON structure returned to API clients on error. Replaces the current flat `dto.ErrorResponse`.

**Fields**:
| Field | Type | JSON Key | Description |
|-------|------|----------|-------------|
| Code | `ErrorCode` | `code` | Machine-readable error identifier |
| Message | `string` | `message` | Human-readable description (sanitized in production) |
| RequestID | `string` | `request_id` | Correlation ID from `X-Request-Id` header |
| Details | `[]ErrorDetail` | `details` | Optional per-field validation details (omitempty) |

**JSON Schema**: See `contracts/error-response.schema.json`.

**Validation**:
- All fields except `Details` are required
- `RequestID` must not be empty (Echo's RequestID middleware guarantees this)
- In production, `Message` must be the registry-default message (not the raw `err.Error()`)

**State transitions**: Created per-request in the error handler. Not persisted.

---

### 4. ErrorDetail

Structured information about a single field-level validation failure.

**Fields**:
| Field | Type | JSON Key | Description |
|-------|------|----------|-------------|
| Field | `string` | `field` | The struct field or JSON field name that failed validation |
| Reason | `string` | `reason` | The validation tag that failed (e.g., "required", "min", "max") |
| Message | `string` | `message` | Human-readable description of the failure |

**Validation**:
- All fields required when present
- `Field` should use the JSON tag name (snake_case) when available, not the Go struct field name

**State transitions**: Derived from `validator.ValidationErrors` in the error handler. Not persisted.

---

### 5. ErrorRegistry

The in-memory mapping of domain sentinel errors to their HTTP representations.

**Fields**:
| Field | Type | Description |
|-------|------|-------------|
| mappings | `map[error]ErrorMapping` | Sentinel → HTTP mapping |
| defaultMapping | `ErrorMapping` | Fallback for unrecognized errors (Status: 500, Code: "INTERNAL_SERVER", Message: "Internal server error") |

**Relationships**: Contains 17 domain-specific mappings + 1 default.

**State transitions**: Mappings are registered via `Register(err error, status int, code ErrorCode, message string)` during initialization. Thread-safe (read-only after init).

---

## Error-to-HTTP Mapping Table

| Domain Sentinel Error | HTTP Status | Error Code |
|----------------------|-------------|------------|
| `ErrUserNotFound` | 404 | `NOT_FOUND` |
| `ErrInvalidCredentials` | 401 | `UNAUTHENTICATED` |
| `ErrUnauthorized` | 401 | `UNAUTHENTICATED` |
| `ErrInternalServer` | 500 | `INTERNAL` |
| `ErrEmailAlreadyRegistered` | 409 | `ALREADY_EXISTS` |
| `ErrValidationFailed` | 400 | `INVALID_ARGUMENT` |
| `ErrAlreadyFollowing` | 409 | `ALREADY_EXISTS` |
| `ErrCannotFollowSelf` | 422 | `INVALID_ARGUMENT` |
| `ErrNotFollowing` | 404 | `NOT_FOUND` |
| `ErrCannotUnfollowSelf` | 422 | `INVALID_ARGUMENT` |
| `ErrUsernameAlreadyTaken` | 409 | `ALREADY_EXISTS` |
| `ErrPostNotFound` | 404 | `NOT_FOUND` |
| `ErrInvalidPost` | 400 | `INVALID_ARGUMENT` |
| `ErrAlreadyLiked` | 409 | `ALREADY_EXISTS` |
| `ErrNotLiked` | 404 | `NOT_FOUND` |
| `ErrCommentNotFound` | 404 | `NOT_FOUND` |
| `ErrNotificationNotFound` | 404 | `NOT_FOUND` |
| `ErrInvalidNotification` | 400 | `INVALID_ARGUMENT` |
| *(any unrecognized error)* | 500 | `INTERNAL` |

## Response Examples

### Validation Error (400)
```json
{
  "error": {
    "code": "VALIDATION_FAILED",
    "message": "Validation failed",
    "request_id": "req_abc123def456",
    "details": [
      {"field": "username", "reason": "min", "message": "username must be at least 3 characters"},
      {"field": "first_name", "reason": "required", "message": "first name is required"}
    ]
  }
}
```

### Not Found Error (404)
```json
{
  "error": {
    "code": "USER_NOT_FOUND",
    "message": "User not found",
    "request_id": "req_abc123def456"
  }
}
```

### Internal Server Error (500, production mode)
```json
{
  "error": {
    "code": "INTERNAL_SERVER",
    "message": "Internal server error",
    "request_id": "req_abc123def456"
  }
}
```
*(Raw Go error message is logged server-side but not exposed to client.)*
