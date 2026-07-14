# Feature Specification: RFC 9457 Problem Details Error Format

**Feature Branch**: `003-rfc9457-error-format`

**Created**: 2026-07-14

**Status**: Draft

**Input**: User description: "Update the global error handling middleware to be structured like the RFC 9457 (STD 97) Problem Details standard per https://datatracker.ietf.org/doc/html/rfc9457"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - API Consumers Receive RFC 9457-Compliant Error Responses (Priority: P1)

An API consumer (frontend app, external service, or developer tool) receives an error response from any endpoint. The response body follows the RFC 9457 Problem Details for HTTP APIs standard (STD 97): it contains the standard fields `type` (a URI identifying the problem type), `title` (a short human-readable summary), `status` (the HTTP status code), `detail` (a human-readable explanation specific to this occurrence), and `instance` (a URI path referencing the specific request). Extension fields (`code`, `request_id`, `errors`) carry project-specific diagnostic information. All existing error scenarios (not-found, validation, unauthorized, conflict, internal) produce this format.

**Why this priority**: This is the core deliverable — migrating the error response body from the current `{"error": {...}}` format to the RFC 9457 (STD 97) standard. Adopting the Internet Standard for problem details improves interoperability, enables off-the-shelf error parsers in client libraries, and aligns the API with industry conventions. Every other story depends on this structural change.

**Independent Test**: Call 3 different endpoints that produce different error types (validation, not-found, unauthorized) and verify ALL responses conform to RFC 9457: they contain `type`, `title`, `status`, `detail`, `instance` at the top level, plus the extension fields `code` and `request_id`. Verify the `Content-Type` header is `application/problem+json`.

**Acceptance Scenarios**:

1. **Given** a handler returns `domain.ErrUserNotFound`, **When** the middleware processes the error, **Then** the response body contains `"type": "https://api.project-one.dev/errors/not-found"`, `"title": "Not Found"`, `"status": 404`, `"detail": "User not found"`, `"instance"` matching the request path, and extension `"code": "NOT_FOUND"`.
2. **Given** a handler returns `domain.ErrValidationFailed` with field-level validator errors, **When** the middleware processes the error, **Then** the response body contains `"type": "https://api.project-one.dev/errors/invalid-argument"`, `"title": "Bad Request"`, `"status": 400`, and extension `"errors"` array with per-field detail objects.
3. **Given** a handler returns `domain.ErrInvalidCredentials`, **When** the middleware processes the error, **Then** the response body contains `"type": "https://api.project-one.dev/errors/unauthenticated"`, `"title": "Unauthorized"`, `"status": 401`.
4. **Given** any error response, **When** the client checks the `Content-Type` header, **Then** it is `application/problem+json` (not `application/json`).
5. **Given** a handler returns an unknown/unmapped error (not a domain sentinel), **When** the middleware processes the error, **Then** the response body contains `"type": "about:blank"`, `"title": "Internal Server Error"`, `"status": 500`, and extension `"code": "INTERNAL"`.
6. **Given** a handler returns `nil` (success), **When** the middleware runs, **Then** the response passes through unchanged — the middleware does not alter `Content-Type` or body for successful responses.

---

### User Story 2 - Frontend Error-Handling Utilities Adapt to RFC 9457 Format (Priority: P2)

The frontend's error-handling utilities are updated to parse RFC 9457 problem details responses instead of the previous `{"error": {...}}` structure. The frontend displays error messages from `detail`, uses `status` for HTTP-aware decisions (redirect to login on 401), uses `code` from extensions for programmatic dispatch, and displays per-field validation errors from the `errors` extension array on forms.

**Why this priority**: The backend format change is the primary deliverable (P1), but the frontend must be updated to consume it. This is P2 because a non-updated frontend would break on the new response format — the frontend update is a hard dependency for a working system.

**Independent Test**: Use the frontend's error parsing utility to handle a mocked RFC 9457 response and verify it extracts `detail` as the user-facing message, `status` as the HTTP code, and `errors` as field-level validation details. Verify the error modal and toast components display messages correctly.

**Acceptance Scenarios**:

1. **Given** the API returns an RFC 9457 error response, **When** the frontend's error parser processes it, **Then** the user sees the `detail` field as the error message.
2. **Given** the API returns a 401 RFC 9457 response, **When** the frontend's error handler processes it, **Then** the user is redirected to the login page (existing behavior preserved).
3. **Given** the API returns a validation error with `errors` extension array, **When** the frontend renders a form, **Then** individual field-level errors are displayed next to the corresponding form fields.
4. **Given** the API returns an RFC 9457 error response, **When** the error toast/notification displays it, **Then** the `title` field is shown as the heading and `detail` as the body (or just `detail` if compact).

---

### User Story 3 - RFC 9457 Error Schema Is Documented in API Documentation (Priority: P3)

The API documentation is updated to reflect the new RFC 9457 error response schema. Every endpoint's documented error responses reference the `application/problem+json` content type and the RFC 9457 problem details schema, so API consumers can discover the format from the docs without reading source code.

**Why this priority**: Documentation is important for API consumers but is a follow-up concern — the format is discoverable from actual responses. The API works without updated docs (P1), but good documentation improves developer experience.

**Independent Test**: Open the API documentation, navigate to any endpoint, and verify the "Responses" section shows error responses with content type `application/problem+json` and the RFC 9457 schema structure.

**Acceptance Scenarios**:

1. **Given** the API docs are generated, **When** viewing any endpoint's 400/401/404/500 responses, **Then** the response schema shows the RFC 9457 problem details object with `type`, `title`, `status`, `detail`, `instance`, and extension fields.
2. **Given** a developer reads the API docs, **When** they look at the error response format, **Then** they see documented extension fields (`code`, `request_id`, `errors`) so they know how to programmatically dispatch on errors.

---

### Edge Cases

- What happens when `debug.Stack()` info is added to the `detail` field? → It MUST NOT be added. The `detail` field in RFC 9457 is human-readable and MUST be sanitized in production. Stack traces remain log-only (as in the current implementation).
- What happens with the `type` URI when the API is deployed to different environments? → The `type` URI base (`https://api.project-one.dev/errors/`) is configurable. In development, it may point to a local documentation page; in production, to the public API docs.
- What happens when `instance` is constructed for requests without a well-known path? → `instance` is set to the request URI path (e.g., `/api/v1/users/123`). For errors that occur before routing (e.g., middleware-level auth failures), `instance` reflects the original request path.
- What happens when an `*echo.HTTPError` is thrown with a custom message? → The custom message is used as the `detail` field, and the `type` URI is derived from the status code if no domain sentinel is found in the chain.
- What about the existing JSON Schema contract at `specs/002-global-error-handling/contracts/error-response.schema.json`? → It will be superseded by a new schema in this feature's `contracts/` directory reflecting RFC 9457 structure.
- What happens when a response body has already been committed before the error handler runs? → Same as current behavior: the error is logged but no body is written. The middleware checks `c.Response().Committed` before writing (existing FR).

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The error response body MUST conform to RFC 9457 (Problem Details for HTTP APIs, STD 97), containing the standard fields: `type` (URI string), `title` (string), `status` (number), `detail` (string), and `instance` (string).
- **FR-002**: The `Content-Type` header for all error responses MUST be `application/problem+json` instead of the current `application/json`.
- **FR-003**: The `type` field MUST be a URI reference identifying the problem type. For known domain errors, it MUST follow the pattern `{base_url}/errors/{error-type-slug}` (e.g., `https://api.project-one.dev/errors/not-found`). For unknown errors, it MUST be `about:blank` (per RFC 9457 §3.1).
- **FR-004**: The `title` field MUST be a short, human-readable summary of the problem type, derived from the HTTP status text (e.g., "Not Found" for 404, "Bad Request" for 400, "Internal Server Error" for 500).
- **FR-005**: The `status` field MUST match the HTTP response status code.
- **FR-006**: The `detail` field MUST contain the human-readable error message specific to this occurrence (e.g., "User not found", "Invalid email or password"). In production, it MUST NOT contain raw Go error strings, stack traces, or internal details.
- **FR-007**: The `instance` field MUST be a URI reference identifying the specific occurrence, set to the request path (e.g., `/api/v1/users/123`).
- **FR-008**: The response body MUST include project-specific extension members: `code` (machine-readable error code, e.g., `NOT_FOUND`) and `request_id` (correlation ID from Echo's RequestID middleware).
- **FR-009**: Validation errors (from `go-playground/validator/v10`) MUST include an `errors` extension array containing per-field detail objects with `field`, `reason`, and `message` keys. This replaces the current `details` field inside `error`.
- **FR-010**: The `type` URI base URL MUST be configurable (via application config), defaulting to `https://api.project-one.dev/errors/` for production and `http://localhost:8080/errors/` for development.
- **FR-011**: All existing error-to-status mappings in the error registry MUST be updated to include the RFC 9457 `type` URI slug and `title` string alongside the existing status and code.
- **FR-012**: The existing error response DTOs MUST be replaced with an RFC 9457-compliant data structure containing the standard problem detail fields plus extension members.
- **FR-013**: The error logging behavior MUST remain unchanged — structured logs continue to include `request_id`, `method`, `path`, `status`, `error_code`, `error`, and `user` fields, with stack traces in non-production environments.
- **FR-014**: The frontend's error-handling utilities, error modal, and toast components MUST be updated to parse RFC 9457 problem details responses instead of the previous custom error format.
- **FR-015**: All existing error handler and error DTO tests MUST be updated to validate RFC 9457 response structure.
- **FR-016**: The API documentation MUST reflect the `application/problem+json` content type and RFC 9457 response schema for all error responses.

### Key Entities

- **Problem Detail (RFC 9457 Object)**: The top-level JSON response body with standard fields (`type`, `title`, `status`, `detail`, `instance`) and extension fields (`code`, `request_id`, `errors`). This replaces the current `{"error": {...}}` wrapper.
- **Problem Type URI**: A URI that identifies the category of problem (e.g., `https://api.project-one.dev/errors/not-found`). Each domain error type maps to a unique URI. Consumers can dereference the URI for human-readable documentation.
- **Error Extension Fields**: Project-specific fields added to the RFC 9457 object: `code` (machine-readable, stable across releases), `request_id` (correlation identifier), and `errors` (validation details array). These are extensions per RFC 9457 §3.2.
- **Validation Error Item**: An object in the `errors` array describing a single field-level validation failure, containing `field` (JSON field name), `reason` (validator tag), and `message` (human-readable description). Equivalent to the current `ErrorDetail` type.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Every error response from any API endpoint validates against the RFC 9457 JSON schema — verified by an automated contract test that calls all endpoint error paths and validates with a JSON Schema validator.
- **SC-002**: The `Content-Type` header is `application/problem+json` for 100% of error responses — verified by an integration test.
- **SC-003**: All existing API endpoints continue to return semantically equivalent HTTP status codes after the format change (404 for not-found, 400 for validation, 401 for unauthorized, 409 for conflict, 500 for internal) — verified by the existing test suite passing without modification to status code assertions.
- **SC-004**: The frontend's error-handling utilities parse the new format without errors — verified by frontend unit tests passing for error parsing, error modal, and toast components.
- **SC-005**: The API documentation shows `application/problem+json` as the error response content type for all documented endpoints — verified by inspecting the generated API spec.
- **SC-006**: The `type` URI base is configurable and changes between development and production builds — verified by checking the error response `type` field in both environments.

## Assumptions

- The RFC 9457 (STD 97) standard is the target format — this is a structural migration of the existing error handling, not a greenfield implementation. The middleware architecture (Echo HTTPErrorHandler, sentinel error mapping, structured logging) remains intact.
- The `type` URI base URL will be sourced from application configuration. If the config value is empty, it defaults to the development URL.
- The existing domain sentinel errors and their error code constants (`CodeNotFound`, `CodeInvalidArgument`, etc.) remain unchanged — only the HTTP response format changes.
- The frontend update is scoped to error-parsing utilities; UI components (modal, toast) only need wiring changes, not visual redesign.
- The API documentation update uses the project's existing documentation generation pattern. No new documentation infrastructure is introduced.
- The `instance` field uses the request URL path (no query string) as the default value. This is a pragmatic choice — full URI construction with scheme and host would require Echo server configuration access.
- RFC 9457 allows `type` to be `about:blank` for generic errors where no more specific type URI is known (per §3.1). We use this for unmapped/unexpected errors.
