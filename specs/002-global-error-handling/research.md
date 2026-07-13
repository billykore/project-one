# Research: Global Error Handling Middleware

**Feature**: 002-global-error-handling
**Date**: 2026-07-13

## Research Task 1: Echo Error Handling Mechanism

### Decision: Use `echo.HTTPErrorHandler` (custom error handler), NOT a traditional middleware

### Rationale

Echo provides two error-handling hooks:

1. **Middleware (`echo.MiddlewareFunc`)**: Runs in the middleware chain. Can intercept errors returned by `next(c)` **only if no response has been committed**. If a handler writes to the response and then returns an error, middleware cannot intercept. Also, Echo's default error handler still runs after all middleware if the error propagates.

2. **Custom `echo.HTTPErrorHandler`**: Echo calls this function with the error and context AFTER all middleware completes. This is the last stop before the response is sent. It can override any previously set response. Setting `e.HTTPErrorHandler = customHandler` replaces Echo's default entirely.

**Decision**: We use `echo.HTTPErrorHandler` for the core error transformation logic and a thin middleware wrapper for the "pass-through on success" behavior. Specifically:

- A middleware checks if `next(c)` returns an error. If so, it wraps it in an `echo.HTTPError` with the mapped status code and calls `c.Error(err)` to trigger the custom error handler.
- The custom `HTTPErrorHandler` does the actual JSON response formatting, logging, sanitization, and writing.

This two-layer approach gives us:
- Full control over the error response (no race with Echo's default handler)
- Ability to handle errors from anywhere in the middleware chain (not just handlers)
- Compatibility with `echo.HTTPError` propagation for middleware that want to set specific status codes (FR-008)

### Alternatives Considered

- **Pure middleware only**: Would miss errors after response commit and conflict with Echo's default handler. Rejected.
- **Replace `Recover()` with custom panic handler**: Echo's `Recover()` already works correctly. We layer on top of it rather than replacing it. The custom `HTTPErrorHandler` receives panics as `echo.HTTPError{Code: 500}` from `Recover()`.

---

## Research Task 2: Error-to-HTTP Mapping Strategy

### Decision: In-memory registry map using `errors.Is` for chain walking

### Rationale

Google's error model (used in gRPC, Google Cloud APIs, and internal frameworks) uses:

1. **Canonical error codes**: A fixed set of codes (like `NotFound`, `InvalidArgument`, `Unauthenticated`)
2. **Error details**: Rich structured metadata attached to errors
3. **Centralized mapping**: Error codes map to protocol-specific status codes (HTTP, gRPC)

Our adaptation for this Go/Echo project:

```go
// Registry: maps sentinel errors to HTTP config
type ErrorMapping struct {
    Status  int    // HTTP status code
    Code    string // Machine-readable error code (e.g., "USER_NOT_FOUND")
    Message string // Default human-readable message
}
```

The registry is a `map[error]ErrorMapping` where keys are the sentinel error values (e.g., `domain.ErrUserNotFound`). Lookup uses `errors.Is()` to walk the error chain:

```go
func (r *ErrorRegistry) Lookup(err error) (ErrorMapping, bool) {
    for sentinel, mapping := range r.mappings {
        if errors.Is(err, sentinel) {
            return mapping, true
        }
    }
    return r.defaultMapping, false // 500 fallback
}
```

Each domain sentinel error gets a unique machine-readable code derived from the Go variable name (e.g., `ErrUserNotFound` → `USER_NOT_FOUND`). These codes are defined as constants in `internal/core/domain/errors.go` to keep them close to the sentinel definitions.

### Alternatives Considered

- **Custom error type with embedded HTTP status**: Would require every use case to know about HTTP. Violates Clean Architecture by leaking transport concerns into the domain. Rejected.
- **Interface-based mapping**: A `HTTPStatus() int` interface on errors. Same domain pollution problem. Rejected.
- **Reflection-based code generation from error variable names**: Fragile, hard to debug, non-obvious. Rejected in favor of explicit constants.

---

## Research Task 3: Validator Error Parsing

### Decision: Convert `validator.ValidationErrors` to structured `ErrorDetail` objects in the middleware

### Rationale

`go-playground/validator/v10` returns `validator.ValidationErrors` which is a slice of `FieldError`. Each `FieldError` has:
- `Field()` — the struct field name
- `Tag()` — the validation tag that failed (e.g., "required", "min", "max")
- `Param()` — the tag parameter (e.g., "3" for `min=3`)
- `Translate()` — human-readable error message (requires a translator)

The middleware detects validation errors via `errors.As(err, &validator.ValidationErrors{})` and converts them to a `[]ErrorDetail`:

```json
{
  "error": {
    "code": "VALIDATION_FAILED",
    "message": "Validation failed",
    "request_id": "req_abc123",
    "details": [
      {"field": "first_name", "reason": "min", "message": "first name must be at least 3 characters"}
    ]
  }
}
```

We do NOT use `Translate()` (which requires setting up a translator with the `en` package). Instead, we generate messages from `field + tag + param` using a simple lookup table for common tags. This avoids pulling in the `go-playground/validator/translations/en` dependency.

### Alternatives Considered

- **Return raw `err.Error()` string**: The current pattern in most handlers. No structured field info, poor frontend UX. Rejected.
- **Custom validation library**: Overkill. The existing validator is wired in. Rejected.
- **Use `Translate()` with English translator**: Adds dependency, more complex setup. Rejected for ponytail simplicity.

---

## Research Task 4: Error Code Naming Convention

### Decision: UPPER_SNAKE_CASE derived from Go sentinel error variable names

### Rationale

All 17 domain sentinel errors follow the pattern `Err<Description>`:

| Sentinel | Error Code |
|----------|-----------|
| `ErrUserNotFound` | `USER_NOT_FOUND` |
| `ErrInvalidCredentials` | `INVALID_CREDENTIALS` |
| `ErrUnauthorized` | `UNAUTHORIZED` |
| `ErrInternalServer` | `INTERNAL_SERVER` |
| `ErrEmailAlreadyRegistered` | `EMAIL_ALREADY_REGISTERED` |
| `ErrValidationFailed` | `VALIDATION_FAILED` |
| `ErrAlreadyFollowing` | `ALREADY_FOLLOWING` |
| `ErrCannotFollowSelf` | `CANNOT_FOLLOW_SELF` |
| `ErrNotFollowing` | `NOT_FOLLOWING` |
| `ErrCannotUnfollowSelf` | `CANNOT_UNFOLLOW_SELF` |
| `ErrUsernameAlreadyTaken` | `USERNAME_ALREADY_TAKEN` |
| `ErrPostNotFound` | `POST_NOT_FOUND` |
| `ErrInvalidPost` | `INVALID_POST` |
| `ErrAlreadyLiked` | `ALREADY_LIKED` |
| `ErrNotLiked` | `NOT_LIKED` |
| `ErrCommentNotFound` | `COMMENT_NOT_FOUND` |
| `ErrNotificationNotFound` | `NOTIFICATION_NOT_FOUND` |
| `ErrInvalidNotification` | `INVALID_NOTIFICATION` |

Code constants are defined in `internal/core/domain/errors.go` alongside the sentinels:
```go
const (
    CodeUserNotFound          = "USER_NOT_FOUND"
    CodeInvalidCredentials    = "INVALID_CREDENTIALS"
    // ...
)
```

### Alternatives Considered

- **Numeric error codes**: Harder to debug (need lookup table), less self-documenting in logs. Rejected.
- **Dot-notation codes** (`user.not_found`): Google uses this in some APIs. Adds hierarchy complexity that isn't needed with 17 errors. Rejected for simplicity.
- **Auto-generate from sentinel string**: `strings.TrimPrefix(err.Error(), "err")` — fragile, error messages can drift. Rejected.

---

## Research Task 5: Stack Trace Capture Strategy

### Decision: Use `runtime/debug.Stack()` in non-production only

### Rationale

Go does not natively capture stack traces at error creation time (unlike Java or .NET). Options:

1. **Capture stack at error return point (in middleware)**: The goroutine stack at the point the middleware receives the error may be deep in the Echo/router internals, not at the actual error site. Not useful for debugging the root cause.

2. **Capture stack at sentinel error creation**: Would require modifying all `errors.New()` calls or wrapping them in a custom error type. Too invasive.

3. **Capture stack in the middleware on the error path (when env != production)**: Provides the call stack from the handler through use case to adapter. While not pinpointing the exact error creation line, it shows the full request-processing call chain, which is sufficient for debugging most issues.

**Decision**: Option 3. A `runtime/debug.Stack()` capture in the middleware's error path, included in log output when `config.App.Env != "production"`. This is a "good enough" compromise that doesn't require invasive changes to the existing error creation pattern.

### Alternatives Considered

- **`pkg/errors` or custom error types with stack capture**: Would require rewriting all sentinel error definitions and error returns. Massive scope creep. Rejected.
- **No stack traces at all**: Makes debugging hard in dev/staging without adding observability. Rejected.

---

## Research Task 6: Existing Handler Refactoring Strategy

### Decision: Phased refactoring — middleware first, then handlers one file at a time

### Rationale

The middleware can coexist with existing manual error handling because:
- Handlers that still call `c.JSON(...)` and return `nil` will not trigger the error middleware (it's a no-op on nil).
- Handlers that return errors will have them intercepted by the middleware.

This allows incremental migration:
1. Deploy middleware (existing handlers unchanged — they still call `c.JSON` and return `nil`, so middleware is dormant for them).
2. Refactor handlers one at a time (replace `c.JSON` → `return err`).
3. Each refactored handler immediately benefits from centralized logging and consistent formatting.

### Handler Refactoring Pattern

Before:
```go
func (h *UserHandler) GetUser(c echo.Context) error {
    username := c.Param("username")
    if username == "" {
        return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "Invalid username"})
    }
    user, err := h.userUseCase.GetUser(c.Request().Context(), username)
    if err != nil {
        if errors.Is(err, domain.ErrUserNotFound) {
            return c.JSON(http.StatusNotFound, dto.ErrorResponse{Error: fmt.Sprintf("User %s not found", username)})
        }
        h.log.Error(c.Request().Context(), "failed to get user profile", "username", username, "error", err)
        return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "Internal server error"})
    }
    return c.JSON(http.StatusOK, toUserResponse(user))
}
```

After:
```go
func (h *UserHandler) GetUser(c echo.Context) error {
    username := c.Param("username")
    if username == "" {
        return domain.ErrValidationFailed // middleware maps to 400
    }
    user, err := h.userUseCase.GetUser(c.Request().Context(), username)
    if err != nil {
        return err // middleware handles mapping, logging, and response
    }
    return c.JSON(http.StatusOK, toUserResponse(user))
}
```

Note: Success responses (`c.JSON(200, ...)`) are NOT changed — handlers still write their own success responses. Only error paths switch from `c.JSON(status, ErrorResponse{...})` to `return err`.

The `username == ""` edge case maps to `domain.ErrValidationFailed` which the middleware returns as 400. For cases where a handler needs a specific status code not in the default mapping (FR-008), a thin wrapper type `echo.HTTPError` can be used: `return echo.NewHTTPError(http.StatusBadRequest, "Invalid username")`. The middleware detects `*echo.HTTPError` and uses its status code directly.

### Alternatives Considered

- **Big-bang refactoring all handlers at once**: High risk of regressions, hard to test, violates incremental delivery principle. Rejected.
- **Only middleware, no handler refactoring**: The middleware would never be triggered because handlers return nil after writing their own error responses. Defeats the purpose. Rejected.

---

## Summary of All Decisions

| # | Decision | Key Rationale |
|---|----------|---------------|
| 1 | Use `echo.HTTPErrorHandler` + thin middleware wrapper | Full control, no race with default handler |
| 2 | In-memory registry map with `errors.Is` chain walking | Clean, O(1) lookup, no domain pollution |
| 3 | Parse `validator.ValidationErrors` to `[]ErrorDetail` | Structured field errors for frontend |
| 4 | UPPER_SNAKE_CASE error codes as domain constants | Self-documenting, close to sentinel definitions |
| 5 | `runtime/debug.Stack()` in non-production only | Good-enough debugging without invasive changes |
| 6 | Phased handler refactoring (middleware first, then per-file) | Incremental, low-risk, revertible |
