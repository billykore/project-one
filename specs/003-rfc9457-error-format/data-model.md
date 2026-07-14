# Data Model: RFC 9457 Problem Details

**Feature**: 003-rfc9457-error-format
**Date**: 2026-07-14

## Overview

This document describes the data structures that change as part of migrating error responses to RFC 9457. No database schema changes — all changes are Go DTO types and TypeScript interfaces.

## 1. Go DTO: ProblemDetail (replaces StructuredError)

### Current (pre-RFC 9457)

```go
// dto/error_dto.go
type StructuredError struct {
    Code      string        `json:"code"`
    Message   string        `json:"message"`
    RequestID string        `json:"request_id"`
    Details   []ErrorDetail `json:"details,omitempty"`
}

type APIErrorResponse struct {
    Error StructuredError `json:"error"`
}

type ErrorDetail struct {
    Field   string `json:"field"`
    Reason  string `json:"reason"`
    Message string `json:"message"`
}
```

### New (RFC 9457)

```go
// dto/error_dto.go
type ProblemDetail struct {
    Type      string             `json:"type"`
    Title     string             `json:"title"`
    Status    int                `json:"status"`
    Detail    string             `json:"detail"`
    Instance  string             `json:"instance"`
    // Extension members
    Code      string             `json:"code,omitempty"`
    RequestID string             `json:"request_id,omitempty"`
    Errors    []ValidationError  `json:"errors,omitempty"`
}

type ValidationError struct {
    Field   string `json:"field"`
    Reason  string `json:"reason"`
    Message string `json:"message"`
}
```

### Field Mapping

| Old Field | New Field | Notes |
|-----------|-----------|-------|
| `error.code` | `code` (extension) | Moved from nested `error` to top-level extension |
| `error.message` | `detail` | Standard RFC 9457 field for occurrence-specific message |
| `error.request_id` | `request_id` (extension) | Moved from nested `error` to top-level extension |
| `error.details` | `errors` (extension) | Renamed to `errors` to match RFC 9457 convention |
| *(new)* | `type` | Standard RFC 9457 field: problem type URI |
| *(new)* | `title` | Standard RFC 9457 field: short summary of problem category |
| *(new)* | `status` | Standard RFC 9457 field: HTTP status code |
| *(new)* | `instance` | Standard RFC 9457 field: request path URI |

### Example Response: Not Found

```json
{
  "type": "https://api.project-one.dev/errors/not-found",
  "title": "Not Found",
  "status": 404,
  "detail": "User not found",
  "instance": "/users/nonexistent",
  "code": "NOT_FOUND",
  "request_id": "req_abc123def456"
}
```

### Example Response: Validation Error

```json
{
  "type": "https://api.project-one.dev/errors/invalid-argument",
  "title": "Bad Request",
  "status": 400,
  "detail": "Validation failed",
  "instance": "/auth/register",
  "code": "INVALID_ARGUMENT",
  "request_id": "req_xyz789",
  "errors": [
    {
      "field": "username",
      "reason": "min",
      "message": "username must be at least 3 characters"
    },
    {
      "field": "email",
      "reason": "required",
      "message": "email is required"
    }
  ]
}
```

### Example Response: Unknown/Unmapped Error

```json
{
  "type": "about:blank",
  "title": "Internal Server Error",
  "status": 500,
  "detail": "Internal server error",
  "instance": "/some/path",
  "code": "INTERNAL",
  "request_id": "req_panic_001"
}
```

## 2. ErrorMapping Changes (middleware/error_registry.go)

### Current

```go
type ErrorMapping struct {
    Status  int
    Code    string
    Message string
}
```

### New

```go
type ErrorMapping struct {
    Status    int
    Code      string
    TypeSlug  string  // URI path segment after base URL, e.g., "not-found"
    Title     string  // Short human-readable summary, e.g., "Not Found"
    Detail    string  // Human-readable detail message, e.g., "User not found"
}
```

### Migration

The existing `Message` field is renamed to `Detail` to match RFC 9457 terminology. `TypeSlug` and `Title` are new fields. Status and Code remain unchanged.

### Registry Updates

```go
var errorMappings = map[error]ErrorMapping{
    domain.ErrUserNotFound:           {http.StatusNotFound, domain.CodeNotFound, "not-found", "Not Found", "User not found"},
    domain.ErrInvalidCredentials:     {http.StatusUnauthorized, domain.CodeUnauthenticated, "unauthenticated", "Unauthorized", "Invalid email or password"},
    domain.ErrUnauthorized:           {http.StatusUnauthorized, domain.CodeUnauthenticated, "unauthenticated", "Unauthorized", "Unauthorized"},
    domain.ErrInternalServer:         {http.StatusInternalServerError, domain.CodeInternal, "", "Internal Server Error", "Internal server error"},
    domain.ErrEmailAlreadyRegistered: {http.StatusConflict, domain.CodeAlreadyExists, "already-exists", "Conflict", "Email is already registered"},
    // ... etc for all 18 sentinels
}
```

**Note on `ErrInternalServer`**: The `TypeSlug` is empty — when the `type` URI is constructed and the slug is empty, the handler uses `about:blank` per RFC 9457 §3.1.

## 3. TypeScript Error Interface (web/lib/errors.ts)

### Current

```typescript
export class ApiError extends Error {
  status: number;
  constructor(message: string, status: number) {
    super(message);
    this.status = status;
    this.name = "ApiError";
  }
}
```

### New

```typescript
export interface ProblemDetail {
  type: string;
  title: string;
  status: number;
  detail: string;
  instance: string;
  code?: string;
  request_id?: string;
  errors?: ValidationError[];
}

export interface ValidationError {
  field: string;
  reason: string;
  message: string;
}

export class ApiError extends Error {
  status: number;
  code?: string;
  type?: string;
  instance?: string;
  validationErrors?: ValidationError[];

  constructor(problem: ProblemDetail) {
    super(problem.detail || problem.title);
    this.status = problem.status;
    this.code = problem.code;
    this.type = problem.type;
    this.instance = problem.instance;
    this.validationErrors = problem.errors;
    this.name = "ApiError";
  }
}
```

### handleApiResponse Update

The `handleApiResponse` function in `web/lib/errors.ts` is updated to:

1. Check for `content-type: application/problem+json` (with fallback to `application/json` for transition safety)
2. Parse the response body as `ProblemDetail`
3. Construct `ApiError` from the parsed problem detail

## 4. Config Changes

### config.go

```go
type AppConfig struct {
    Port            int    `mapstructure:"port"`
    Env             string `mapstructure:"env"`
    ErrorTypeBaseURL string `mapstructure:"error_type_base_url"`
}
```

### config.yaml

```yaml
app:
  port: 8080
  env: "development"
  error_type_base_url: ""  # empty defaults to http://localhost:{port}/errors/
```

## 5. Validation Rules

- `type` MUST be a valid URI (or `about:blank`)
- `title` MUST be non-empty
- `status` MUST be a valid HTTP status code (100-599)
- `detail` MUST be non-empty in production responses
- `instance` MUST be a valid URI reference (relative path is acceptable)
- `code` extension MUST match the existing code constants pattern (`^[A-Z][A-Z0-9]*(_[A-Z][A-Z0-9]*)*$`)
- `errors` array items MUST have `field`, `reason`, and `message` when present
