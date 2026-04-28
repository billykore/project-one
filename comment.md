### Summary of Changes
This pull request implements the user registration endpoint (`POST /user/register`) following Domain-Driven Design (DDD) and Clean Architecture principles. It includes the necessary domain validation, service logic for duplicate email checking and password hashing, the GORM repository method for user creation, and the Echo HTTP handler with DTOs.

### Critical Issues

**1. Brittle Error Handling (String Matching)**
In `user_handler.go`, domain validation errors are identified using `strings.Contains(errMsg, "required")`, etc. This is fragile and could cause unexpected internal server errors (500s) to be incorrectly reported as client errors (400s) if the error message happens to contain one of these keywords.

**2. Check-Then-Act Race Condition**
In `user_service.go`, the code checks if an email exists before creating the user. In a highly concurrent environment, two requests for the same email can pass the check simultaneously. The database's `UNIQUE` constraint will catch this, but `CreateUser` in the repository currently returns the raw DB error. This leads to a `500 Internal Server Error` instead of mapping it to `domain.ErrEmailAlreadyRegistered`.

### Suggestions for Improvement

**1. Robust Validation Error Handling**
Avoid string matching by wrapping validation errors with a sentinel error in `domain/errors.go` (e.g., `ErrValidationFailed`) or creating a custom error type. 

**2. Handle Database Unique Constraint Violations**
Map the GORM/Postgres unique constraint error in the repository layer directly to `domain.ErrEmailAlreadyRegistered`.

**3. Input Normalization**
Trim whitespace from names and emails, and normalize the email to lowercase before processing it to prevent duplicate accounts due to casing differences.

### Step-by-step Plan for Improvement

**Step 1: Normalize Input**
In `internal/app/user/adapters/handler/user_handler.go`, normalize the request data before mapping it to the domain entity:
```go
user := &domain.User{
	FirstName: strings.TrimSpace(req.FirstName),
	LastName:  strings.TrimSpace(req.LastName),
	Email:     strings.ToLower(strings.TrimSpace(req.Email)),
	Password:  req.Password,
}
```

**Step 2: Define a Sentinel Error for Domain Validation**
In `internal/app/user/core/domain/errors.go`, add a new error:
```go
var ErrValidationFailed = errors.New("validation failed")
```

**Step 3: Wrap Domain Validation Errors**
In `internal/app/user/core/domain/user.go`, wrap the validation errors:
```go
import "fmt"
// ...
if u.FirstName == "" {
	return fmt.Errorf("%w: first name is required", ErrValidationFailed)
}
// Apply this to all validation checks...
```

**Step 4: Improve Handler Error Logic**
In `internal/app/user/adapters/handler/user_handler.go`, replace the `strings.Contains` checks with `errors.Is`:
```go
if err := h.userSvc.Register(c.Request().Context(), user); err != nil {
	if errors.Is(err, domain.ErrEmailAlreadyRegistered) {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "email is already registered"})
	}
	if errors.Is(err, domain.ErrValidationFailed) {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
	}
	return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "something went wrong"})
}
```

**Step 5: Catch Database Unique Constraint Errors**
In `internal/app/user/adapters/repository/postgres_user_repository.go`, map the unique constraint violation:
```go
import "errors"
// ...
if err := r.db.WithContext(ctx).Create(&m).Error; err != nil {
	if errors.Is(err, gorm.ErrDuplicatedKey) { // Or check pgconn.PgError code "23505"
		return domain.ErrEmailAlreadyRegistered
	}
	return err
}
```