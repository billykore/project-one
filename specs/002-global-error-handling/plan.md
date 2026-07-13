# Implementation Plan: Global Error Handling Middleware

**Branch**: `002-global-error-handling` | **Date**: 2026-07-13 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `/specs/002-global-error-handling/spec.md`

**Note**: This template is filled in by the `/speckit.plan` command. See `.specify/templates/plan-template.md` for the execution workflow.

## Summary

Build an Echo middleware that implements a Google-style Global Error Handling Framework. The middleware intercepts errors returned by handlers, maps domain sentinel errors to HTTP status codes via a centralized registry, produces consistent structured JSON error responses (`{"error": {"code", "message", "request_id", "details"}}`), logs every error with request-correlated structured fields, and sanitizes internal details in production. Existing handlers are refactored to return errors directly instead of calling `c.JSON(statusCode, dto.ErrorResponse{...})`, eliminating ~126 repetitive error-mapping blocks across 6 handler files while preserving semantically equivalent HTTP status codes.

## Technical Context

**Language/Version**: Go 1.26+

**Primary Dependencies**: Echo v4.15.1 (HTTP framework), log/slog (structured logging, standard library), go-playground/validator v10 (input validation), GORM (ORM — not directly touched by this feature)

**Storage**: N/A (middleware has no persistent state; error mapping registry is an in-memory Go map)

**Testing**: GoMock (use case mocks), Testify (assertions), Go standard `testing` package; `go test ./...` with `-race` flag

**Target Platform**: Linux server (Docker deployment via `deployments/docker-compose.yml`)

**Project Type**: web-service backend (Clean Architecture: domain → ports → usecase → adapters → api)

**Performance Goals**: Middleware overhead <1ms per request (error path only); no measurable impact on success-path latency (the middleware is a no-op for nil errors)

**Constraints**: Must preserve all existing HTTP status code semantics (no status code regressions); Must integrate with existing Echo middleware chain (after `RequestID`, before routes); Must use existing `ports.Logger` interface; Must not introduce new external dependencies

**Scale/Scope**: 17 sentinel errors across 5 domain files; 6 handler files to refactor (~126 error-mapping blocks); 1 new middleware file + error registry + DTO restructure; ~15 new unit tests; existing ~200 tests must continue passing

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

### I. Code Quality & Clean Architecture ✅

- **Domain layer**: No changes. Existing sentinel errors in `internal/core/domain/` remain pure — zero new imports. The middleware does not introduce framework dependencies into the domain.
- **Ports layer**: No new interfaces needed. The middleware uses existing `ports.Logger`. If an `ErrorMapper` port is warranted for testability, it will be defined in `internal/core/ports/`.
- **UseCase layer**: No changes. Error wrapping (`fmt.Errorf("...: %w", sentinel)`) is already the established pattern.
- **Adapters layer**: The error mapping registry is an adapter concern — it translates domain errors to protocol-level (HTTP) representations. Located in `internal/api/middleware/` alongside the middleware itself, keeping it close to the HTTP boundary.
- **API layer**: New middleware in `internal/api/middleware/`. DTO restructure in `internal/api/dto/`. Handler refactoring stays within `internal/api/handler/`.
- **Linting**: `make check` will pass before merge. No commented-out code or dead imports.

### II. Testing Standards ✅

- Unit tests for the error middleware (mocking `ports.Logger`) using GoMock
- Unit tests for the error registry (mapping correctness, fallback behavior)
- Integration tests that exercise the Echo middleware chain with real HTTP requests
- Existing handler tests (e.g., `comment_usecase_test.go`, `feed_usecase_test.go`) continue to pass — they test at the use case level, below the middleware
- Coverage must not decrease

### III. User Experience Consistency ✅

- Frontend is out of scope for this feature, but the structured error response format is designed for frontend consumption — the `details` array enables field-level error highlighting, and the `code` field enables programmatic routing (e.g., `UNAUTHORIZED` → redirect to login)

### IV. Performance Requirements ✅

- Middleware is a no-op on the success path (nil error returned: `return next(c)` unchanged)
- Error path: single map lookup (O(1)), single `errors.Is` chain walk (bounded by wrapping depth, typically 1–3), one structured log call, one JSON marshal. Total overhead <1ms.
- No database queries, no allocations beyond the error response DTO

### Security & Authentication ✅

- Error responses in production sanitize internal details (no stack traces, no raw Go error messages, no SQL errors) — enforced by middleware behavior, not handler discipline
- The middleware reads the authenticated username from `c.Get("username")` for logging but does not expose it in error responses (unnecessary information disclosure)
- No new authentication or authorization logic

### Development Workflow & Quality Gates ✅

- Commit convention: `feat(middleware): add global error handling middleware`
- `make check` and `go test ./...` must pass before merge
- Branch: `002-global-error-handling`

### Gate Result: PASS — No violations. All principles satisfied.

## Project Structure

### Documentation (this feature)

```text
specs/002-global-error-handling/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
├── contracts/           # Phase 1 output
│   └── error-response.schema.json
└── tasks.md             # Phase 2 output (NOT created by /speckit.plan)
```

### Source Code (repository root)

```text
internal/
├── api/
│   ├── middleware/
│   │   ├── authorization.go          # [EXISTING] JWT auth middleware
│   │   ├── error_registry.go         # [NEW] Error-to-HTTP mapping registry
│   │   ├── error_registry_test.go    # [NEW] Registry unit tests
│   │   ├── error_handler.go          # [NEW] Global error handling middleware
│   │   └── error_handler_test.go     # [NEW] Middleware integration tests
│   ├── dto/
│   │   ├── error_dto.go           # [MODIFY] Restructure ErrorResponse → structured format
│   │   └── error_dto_test.go      # [NEW] DTO serialization tests
│   └── handler/
│       ├── user_handler.go        # [REFACTOR] Remove c.JSON error boilerplate
│       ├── post_handler.go        # [REFACTOR] Remove c.JSON error boilerplate; add logger
│       ├── comment_handler.go     # [REFACTOR] Remove c.JSON error boilerplate
│       ├── feed_handler.go        # [REFACTOR] Remove c.JSON error boilerplate
│       ├── notification_handler.go # [REFACTOR] Remove c.JSON error boilerplate
│       └── websocket_handler.go   # [REFACTOR] Remove c.JSON error boilerplate
├── core/
│   └── domain/
│       ├── errors.go              # [NEW] Centralized error code constants (machine-readable codes)
│       └── errors_test.go         # [NEW] Error code stability tests
└── adapters/
    └── logger/
        └── slog_logger.go         # [EXISTING] No changes needed

cmd/
└── main.go                        # [MODIFY] Register error middleware in Echo chain
```

**Structure Decision**: Single backend project (Option 1). The feature is entirely scoped to the `internal/api/` layer with a minor touchpoint in `cmd/main.go` for middleware registration and `internal/core/domain/` for error code constants. No new top-level directories. Frontend is out of scope.

## Complexity Tracking

> No violations to justify — all constitution principles are satisfied.

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| N/A | N/A | N/A |
