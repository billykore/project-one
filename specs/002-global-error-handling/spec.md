# Feature Specification: Global Error Handling Middleware

**Feature Branch**: `002-global-error-handling`

**Created**: 2026-07-13

**Status**: Draft

**Input**: User description: "Build a middleware for error management in this project. The middleware must implement the Global Error Handling Framework like what Google does"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Backend Developers Write Handlers Without Error-Mapping Boilerplate (Priority: P1)

A backend developer writes a new API handler. When the use case returns a domain error (e.g., `domain.ErrUserNotFound`, `domain.ErrValidationFailed`), the developer simply returns the error from the handler. The global error middleware automatically maps the domain error to the correct HTTP status code, formats a consistent error response, includes the request ID for traceability, and logs the error. The developer does not write any `c.JSON(statusCode, dto.ErrorResponse{...})` or manual `errors.Is()` checks in the handler.

**Why this priority**: This is the core value proposition — it eliminates ~150 lines of duplicated error-handling boilerplate across 6 handler files, enforces consistency for every API response, and makes handlers focus on the happy path. Every other story builds on this foundation.

**Independent Test**: Create a test handler that returns a known domain error (e.g., `domain.ErrUserNotFound`), send a request, and verify the response has HTTP 404, a structured error body with an error code, a human-readable message, and a request ID matching the `X-Request-Id` header.

**Acceptance Scenarios**:

1. **Given** a handler returns `domain.ErrUserNotFound`, **When** the middleware processes the error, **Then** the HTTP response has status 404 and a structured error body with code `USER_NOT_FOUND`.
2. **Given** a handler returns `domain.ErrValidationFailed`, **When** the middleware processes the error, **Then** the HTTP response has status 400 and a structured error body with code `VALIDATION_FAILED`.
3. **Given** a handler returns `domain.ErrInternalServer` (or any unexpected error), **When** the middleware processes the error, **Then** the HTTP response has status 500 and the error body contains a generic message (not leaking internal details) and the full error is logged server-side with the request ID.
4. **Given** a handler returns a wrapped error using `fmt.Errorf("...: %w", domain.ErrUserNotFound)`, **When** the middleware processes the error, **Then** it unwraps the chain and correctly identifies `ErrUserNotFound`, returning HTTP 404.
5. **Given** a handler returns `nil` (success), **When** the middleware processes it, **Then** the response passes through unchanged — the middleware does not interfere with successful responses.
6. **Given** any error response from the middleware, **When** the client inspects the response body, **Then** it includes a `request_id` field matching the `X-Request-Id` header set by Echo's RequestID middleware.

---

### User Story 2 - API Consumers Receive Structured, Machine-Readable Error Responses (Priority: P2)

An API consumer (frontend app, external service, or developer tool) receives an error response. Every error response follows a consistent JSON structure: an error code (machine-readable enum like `USER_NOT_FOUND`), a human-readable message, a request ID for support correlation, and optional field-level validation details. The frontend can write a single error-handling utility that works for all API calls without per-endpoint special cases.

**Why this priority**: Consistent error structure enables the frontend to build reusable error-handling logic (toast notifications, form field errors, retry logic). It is P2 because the middleware can technically work with a simpler response format (P1), but the structured format unlocks downstream productivity.

**Independent Test**: Call 3 different endpoints that produce different error types (validation, not-found, unauthorized) and verify all responses share the same JSON structure with the same top-level keys. Write a frontend utility that parses any error response and displays the message — verify it works for all 3 endpoints.

**Acceptance Scenarios**:

1. **Given** a validation error (e.g., username too short), **When** the client receives the error response, **Then** the response includes a `details` array with per-field error objects containing `field`, `reason`, and `message`.
2. **Given** an authentication error (expired/invalid token), **When** the client receives the error response, **Then** the response has code `UNAUTHORIZED` and HTTP 401, enabling the frontend to trigger a redirect to login.
3. **Given** a not-found error (user, post, or comment), **When** the client receives the error response, **Then** the response has code ending in `_NOT_FOUND` and HTTP 404, enabling the frontend to show a "not found" UI.
4. **Given** any error response, **When** the client parses the JSON, **Then** the structure is always `{"error": {"code": "...", "message": "...", "request_id": "...", "details": [...]}}`.

---

### User Story 3 - Operations Engineers Debug Issues with Request-Correlated Error Logs (Priority: P3)

An operations engineer investigates a production incident. A user reports error ID `req_abc123`. The engineer searches the logs for `req_abc123` and finds a single structured log entry containing the full error chain, the HTTP method and path, the authenticated user (if any), the error code that was returned to the client, and — in non-production environments — a stack trace. The engineer can trace the request end-to-end without grep'ing through unstructured log lines.

**Why this priority**: Debuggability is essential for production systems but is a supporting concern — the middleware already works without it (P1). Structured logging with correlation IDs is the difference between minutes and hours when investigating incidents.

**Independent Test**: Send a request that triggers a 500 error, capture the `request_id` from the response, then verify the application log contains exactly one log entry for that request ID with the error message, HTTP path, and status code.

**Acceptance Scenarios**:

1. **Given** an error occurs during request processing, **When** the middleware logs the error, **Then** the log entry includes `request_id`, `method`, `path`, `status` (HTTP status code), `error_code`, `error` (message), and `user` (authenticated username, or "anonymous").
2. **Given** the application is running in a non-production environment, **When** an error occurs, **Then** the log entry includes a `stack_trace` field showing the goroutine's call stack at the point the error was returned.
3. **Given** the application is running in production, **When** an error occurs, **Then** the log entry does NOT include a stack trace (to avoid log volume and information leakage), but all other fields are present.
4. **Given** a handler panics (unrecovered panic), **When** the middleware (or Echo's Recover) catches it, **Then** it responds with HTTP 500, logs the panic with the request ID, and does not crash the server.

---

### Edge Cases

- What happens when a handler returns an error that is NOT a known domain sentinel error (e.g., a bare `fmt.Errorf("something broke")`)? → The middleware treats unknown errors as HTTP 500 Internal Server Error with a generic public message, but logs the full error details server-side.
- What happens when an error wraps multiple layers (e.g., `fmt.Errorf("create post: %w", fmt.Errorf("db insert: %w", domain.ErrInternalServer))`)? → The middleware walks the error chain using `errors.Is()` and maps to the first recognized domain sentinel, falling back to 500 if none match.
- What happens when a handler writes a response body AND returns an error? → Echo's built-in behavior applies: if a response was already committed, the error is logged but the client sees whatever was written first. The middleware documents this as a handler bug pattern to avoid.
- What happens when the client sets `Accept: text/html` instead of `Accept: application/json`? → The middleware respects content negotiation: returns JSON for API clients, but could return a plain-text or HTML error for browser requests (configurable behavior, JSON by default for this project).
- What happens when the error occurs in a WebSocket upgrade request? → WebSocket upgrades happen before the error middleware runs; upgrade errors are handled by the existing WebSocket handler. The middleware does not apply to established WebSocket connections.
- What happens when an error contains sensitive data (passwords, tokens, PII) in its message? → The middleware never exposes the raw `err.Error()` string to clients. It always uses the mapped human-readable message. Production logs still contain the full error — developers are responsible for not putting secrets in error messages.
- What happens when multiple middleware layers all return errors for the same request? → The outermost error middleware catches the first error returned and processes it; subsequent errors in the chain are not reached. This is standard Echo middleware behavior.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The system MUST provide an Echo middleware that intercepts errors returned by downstream handlers and transforms them into consistent HTTP JSON responses.
- **FR-002**: The middleware MUST maintain a mapping between domain sentinel errors (defined in `internal/core/domain/`) and HTTP status codes. Unknown errors MUST default to HTTP 500.
- **FR-003**: The middleware MUST support error unwrapping — if an error wraps a domain sentinel via `fmt.Errorf("...: %w", sentinelErr)`, the middleware MUST identify the sentinel using `errors.Is()`.
- **FR-004**: The error response body MUST follow a consistent JSON structure with at minimum: `code` (machine-readable error code string), `message` (human-readable description), and `request_id` (correlation ID from Echo's RequestID middleware).
- **FR-005**: The middleware MUST log every error it processes using structured logging (`log/slog`) at the appropriate level (client errors at Warn level, server errors at Error level), including `request_id`, `method`, `path`, `status`, `error_code`, `error`, and `user` fields.
- **FR-006**: The middleware MUST respond with the HTTP `Content-Type: application/json` header for all error responses.
- **FR-007**: The middleware MUST NOT interfere with successful (nil error) responses — it passes them through unchanged.
- **FR-008**: The middleware MUST expose a way for handlers to return custom HTTP status codes with domain errors when the default mapping is insufficient (e.g., a handler can wrap an error with a specific HTTP status override).
- **FR-009**: The middleware MUST include per-field validation details in the error response when the error is a validation error with structured field information.
- **FR-010**: The middleware MUST sanitize error messages in production — internal error details (stack traces, raw Go error messages, SQL errors) MUST NOT be exposed to API clients.
- **FR-011**: The middleware MUST be registered in the Echo middleware chain before route registration, after Echo's built-in `RequestID` and `Recover` middleware.
- **FR-012**: Existing handlers MUST be refactored to remove manual `c.JSON(statusCode, dto.ErrorResponse{...})` error handling and instead return errors directly to the middleware.

### Key Entities

- **Error Code**: A machine-readable string identifier for an error type (e.g., `USER_NOT_FOUND`, `VALIDATION_FAILED`, `UNAUTHORIZED`). Derived from domain sentinel errors. Used by API consumers for programmatic error handling.
- **Error Mapping**: A registry associating each domain sentinel error with an HTTP status code and a human-readable default message. The middleware consults this mapping to transform errors into responses.
- **Error Detail**: Optional structured information about specific validation failures, keyed by field name with a reason code and human-readable message. Allows clients to highlight individual form fields with errors.
- **Error Response**: The JSON body returned to the client, containing `code`, `message`, `request_id`, and optional `details`. This replaces the current flat `dto.ErrorResponse`.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: All error responses across the API share the same JSON structure — verified by automated tests that call every endpoint with invalid input and validate the response schema.
- **SC-002**: Backend handler code eliminates at least 80% of manual `c.JSON(statusCode, dto.ErrorResponse{...})` calls currently present across handler files (from approximately 50+ instances to fewer than 10).
- **SC-003**: A developer can add a new domain sentinel error, register it in the error mapping, and have all handlers that return it produce correct HTTP responses without modifying any handler code — verified by a documented 3-step process.
- **SC-004**: Every error response includes a `request_id` that matches the `X-Request-Id` response header — verified by an integration test.
- **SC-005**: No raw Go error messages, stack traces, or internal implementation details appear in error responses when the server is running in production mode — verified by sending malformed requests and inspecting response bodies.
- **SC-006**: All existing API endpoints continue to return semantically equivalent HTTP status codes after the refactoring (e.g., 404 for not-found, 400 for validation, 401 for unauthorized) — verified by running the existing test suite.

## Assumptions

- The middleware integrates with Echo's built-in error-handling mechanism (`echo.HTTPError` and the `echo.HTTPErrorHandler` hook), extending rather than replacing it.
- The existing sentinel errors in `internal/core/domain/` are the authoritative set of domain errors. New errors will follow the same pattern.
- The frontend (`web/`) will be updated in a separate feature to consume the new structured error format. This spec covers only the backend middleware.
- The project's structured logger (`log/slog`) remains the logging backend. The middleware uses the existing `ports.Logger` interface for testability.
- The Echo `RequestID` middleware is already registered and generates unique request IDs — the error middleware reads from it, it does not generate its own IDs.
- Validation errors from `go-playground/validator/v10` need to be parsed into structured field-level details. The middleware will include a converter for this specific validator library.
- The refactoring of existing handlers (FR-012) is scoped to this feature — it is not a separate follow-up.
- "Production mode" is determined by the existing `config.App.Env` setting (string `"production"`).
