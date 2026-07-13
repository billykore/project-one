# Quickstart: Global Error Handling Middleware

**Feature**: 002-global-error-handling
**Date**: 2026-07-13

## Prerequisites

- Go 1.26+ installed
- Project dependencies: `go mod download`
- Database running (PostgreSQL via `docker-compose up -d` from `deployments/`)
- Migrations applied: `make migrate-up dsn="postgres://user:pass@localhost:5432/project1?sslmode=disable"`

## Setup

No additional setup beyond the existing project. The middleware uses existing dependencies (Echo, log/slog, validator).

## Validation Scenarios

### Scenario 1: Verify Middleware Is Registered and Active

**Command:**
```bash
make run
```

**Then:**
```bash
# Trigger a known error — request a non-existent user
curl -s -w "\nHTTP %{http_code}\n" http://localhost:8080/users/nonexistent_user | jq .
```

**Expected outcome:**
```json
{
  "error": {
    "code": "USER_NOT_FOUND",
    "message": "User not found",
    "request_id": "req_..."
  }
}
```
HTTP Status: **404**

---

### Scenario 2: Verify Validation Error with Field Details

**Command:**
```bash
# Register with invalid data (username too short)
curl -s -w "\nHTTP %{http_code}\n" -X POST http://localhost:8080/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","username":"ab","password":"password123","first_name":"","last_name":""}' | jq .
```

**Expected outcome:**
```json
{
  "error": {
    "code": "VALIDATION_FAILED",
    "message": "Validation failed",
    "request_id": "req_...",
    "details": [
      {"field": "first_name", "reason": "required", "message": "first name is required"},
      {"field": "last_name", "reason": "required", "message": "last name is required"},
      {"field": "username", "reason": "min", "message": "username must be at least 3 characters"}
    ]
  }
}
```
HTTP Status: **400**

---

### Scenario 3: Verify Unauthorized Access

**Command:**
```bash
# Access protected endpoint without auth token
curl -s -w "\nHTTP %{http_code}\n" http://localhost:8080/feeds | jq .
```

**Expected outcome:**
```json
{
  "error": {
    "code": "UNAUTHORIZED",
    "message": "Unauthorized",
    "request_id": "req_..."
  }
}
```
HTTP Status: **401**

---

### Scenario 4: Verify 500 Error with Generic Message in Production-Like Mode

**Command:**
```bash
# Set production mode in config.yaml (app.env: "production"), restart
# Then trigger an internal error — e.g., malformed input that bypasses validation
curl -s -w "\nHTTP %{http_code}\n" -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"", "password":""}' | jq .
```

**Expected outcome:**
- Response contains `"code": "INTERNAL_SERVER"` or `"code": "VALIDATION_FAILED"` (generic)
- Response `message` is the registered default, NOT a raw Go error or stack trace
- Server logs (stderr) contain the full error details with `request_id`

---

### Scenario 5: Verify All Endpoints Still Return Correct Status Codes

**Command:**
```bash
# Run existing test suite — no status code regressions
make test
```

**Expected outcome:** All existing tests pass. No test assertions about error response body format need updating (tests that check error bodies will be updated as part of handler refactoring).

---

### Scenario 6: Verify Request ID Correlation

**Command:**
```bash
# Capture both response body and headers
curl -s -i http://localhost:8080/users/nonexistent_user
```

**Expected outcome:**
- `X-Request-Id` header is present in the response
- `error.request_id` in the JSON body matches `X-Request-Id` header value
- Server log line for this error includes the same request ID

---

### Scenario 7: Verify WebSocket Errors Are Unaffected

**Command:**
```bash
# Attempt WebSocket upgrade without token
curl -s -w "\nHTTP %{http_code}\n" -H "Upgrade: websocket" -H "Connection: Upgrade" \
  http://localhost:8080/websocket | jq .
```

**Expected outcome:** HTTP **401** (handled by existing `Authorize` middleware, not the error middleware). WebSocket endpoints are not affected by the new middleware.

---

### Scenario 8: Verify Custom HTTP Status Override (FR-008)

**Command:**
```bash
# This tests a new mechanism: handler returns echo.NewHTTPError(422, "custom message")
# Triggered by attempting to follow yourself
TOKEN=$(curl -s -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"alice@example.com","password":"password123"}' | jq -r '.access_token // empty')
# If token obtained, try self-follow
curl -s -w "\nHTTP %{http_code}\n" -X POST http://localhost:8080/users/alice/followers \
  -H "Authorization: Bearer $TOKEN" | jq .
```

**Expected outcome:**
```json
{
  "error": {
    "code": "CANNOT_FOLLOW_SELF",
    "message": "Cannot follow yourself",
    "request_id": "req_..."
  }
}
```
HTTP Status: **422**

---

## Key Assertions

| # | Assertion | Verification Method |
|---|-----------|-------------------|
| 1 | All error responses follow `{"error": {"code", "message", "request_id", "details"}}` schema | JSON Schema validation against `contracts/error-response.schema.json` |
| 2 | Every error response includes `request_id` matching `X-Request-Id` header | Manual curl + `-i` flag |
| 3 | Production mode never exposes raw Go errors in response body | Inspect response body in production config |
| 4 | All existing tests pass | `make test` |
| 5 | Success responses (2xx) are unaffected | `curl -s -w "%{http_code}" http://localhost:8080/users/alice` returns 200 with normal user JSON |
| 6 | Unknown errors default to 500 with generic message | Trigger a database error (stop PostgreSQL) and call any endpoint |
