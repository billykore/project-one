# Quickstart: RFC 9457 Error Format Validation

**Feature**: 003-rfc9457-error-format
**Date**: 2026-07-14

## Prerequisites

- Go 1.26+ and project dependencies (`go mod download`)
- Node.js and frontend dependencies (`cd web && npm install`)
- PostgreSQL running (or test DB)
- `make` tooling (swag CLI, golang-migrate, golangci-lint)

## Step 1: Verify Backend Test Suite

Run the existing backend tests to establish a baseline before making changes:

```bash
# Run all backend tests
make test

# Or run error-specific tests
go test ./internal/api/middleware/... -v
go test ./internal/api/dto/... -v
```

**Expected**: All tests pass. Note any test output that references the current `{"error": {...}}` format — these will need updating.

## Step 2: Validate New JSON Schema

After implementing the new RFC 9457 DTOs, validate the response contract:

```bash
# Install a JSON Schema validator (one-time)
go install github.com/santhosh-tekuri/jsonschema/cmd/jv@latest

# Start the server
go run ./cmd/main.go

# In another terminal, trigger various errors and capture responses:

# Not Found (404)
curl -s -D - http://localhost:8080/users/nonexistent | jq .

# Validation Error (400) — send invalid registration
curl -s -D - -X POST http://localhost:8080/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"bad","password":"short"}' | jq .

# Unauthorized (401) — hit an auth-required endpoint without token
curl -s -D - http://localhost:8080/posts | jq .

# Internal Error (500) — trigger via unknown route or malformed request
curl -s -D - -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d 'not-json' | jq .
```

**Validate each response**:

1. Content-Type header is `application/problem+json`
2. Body contains `type`, `title`, `status`, `detail`, `instance` at top level
3. Extension fields `code` and `request_id` are present
4. Validation error responses include `errors` array with `field`, `reason`, `message`

**Schema validation** (after implementing):

```bash
# Save a response to file and validate against schema
curl -s http://localhost:8080/users/nonexistent > /tmp/error.json
jv specs/003-rfc9457-error-format/contracts/problem-detail.schema.json /tmp/error.json
```

## Step 3: Run Updated Backend Tests

After implementing changes, verify all backend tests pass with the new format:

```bash
# Run error-specific tests
go test ./internal/api/middleware/... -v -run TestErrorHandler
go test ./internal/api/dto/... -v

# Full test suite
make test
```

**Expected assertions in tests**:

- `Content-Type: application/problem+json` (not `application/json`)
- Response body has `type`, `title`, `status`, `detail`, `instance` fields
- No `{"error": {...}}` wrapper
- Status codes remain unchanged (404, 400, 401, 409, 500)

## Step 4: Verify Frontend Error Parsing

After updating `web/lib/errors.ts`, run frontend tests:

```bash
cd web
npm test
# Or run error-specific tests
npx vitest run tests/lib/errors.test.ts
npx vitest run tests/hooks/use-error-modal.test.tsx
```

**Manual frontend verification**:

1. Start both backend and frontend
2. Navigate to a non-existent user profile (e.g., `/users/doesnotexist`)
3. Verify the error message is displayed correctly (uses `detail` field)
4. Attempt login with invalid credentials — verify error toast shows `detail`
5. Submit a registration form with invalid data — verify field-level errors appear next to form fields

## Step 5: Verify Swagger Documentation

Regenerate API docs and verify RFC 9457 schema:

```bash
# Regenerate Swagger docs
make docs

# Start server and open Swagger UI
open http://localhost:8080/swagger/index.html
```

**Verify in Swagger UI**:

1. Any endpoint's error responses show `application/problem+json` content type
2. The response schema shows ProblemDetail structure with `type`, `title`, `status`, `detail`, `instance`
3. Extension fields (`code`, `request_id`, `errors`) are documented

## Step 6: Full Integration Check

Run the complete CI pipeline locally:

```bash
# Backend
make check    # Runs docs + vet + lint + test

# Frontend
cd web && npm test && npm run lint
```

**Expected**: All checks pass with no regressions.

## Troubleshooting

| Symptom | Likely Cause | Fix |
|---------|-------------|-----|
| `Content-Type` still `application/json` | Error handler not setting new Content-Type | Check `c.JSON()` call in `error_handler.go` — Echo's JSON method sets `application/json` by default; may need explicit header set |
| Frontend shows raw JSON error | `handleApiResponse` not parsing `application/problem+json` | Update content-type check in `web/lib/errors.ts` to include `application/problem+json` |
| `type` URI uses wrong base | Config not loaded or default not applied | Verify `error_type_base_url` in config.yaml and that handler reads from config |
| Test assertions fail on removed `error` wrapper | Tests still checking `body["error"]` | Update test assertions to read top-level fields directly |
| Swagger docs show old format | `make docs` not run after handler annotation updates | Run `make docs` and check that ProblemDetail struct is referenced in `@Failure` annotations |
