# Design Document: GET /users/:username API

**Date:** 2026-05-15
**Status:** Approved
**Topic:** Create an API for getting user by username.

## 1. Overview
The goal is to provide a public API endpoint that allows clients to retrieve basic user profile information (username, email, and full name) using a username.

## 2. Architecture
The implementation follows Clean Architecture principles:
- **API Layer**: Echo handler to manage HTTP request/response.
- **UseCase Layer**: Business logic to orchestrate the retrieval.
- **Ports Layer**: Interface definition for the UseCase.
- **Domain Layer**: User entity.
- **Repository Layer**: Data access (already implemented).

## 3. Detailed Components

### 3.1 Ports (`internal/core/ports/user.go`)
Extend the `UserUseCase` interface:
```go
type UserUseCase interface {
    // ... existing methods
    GetUserProfile(ctx context.Context, username string) (*domain.User, error)
}
```

### 3.2 UseCase (`internal/core/usecase/user_usecase.go`)
Implement `GetUserProfile`:
```go
func (s *userUseCase) GetUserProfile(ctx context.Context, username string) (*domain.User, error) {
    user, err := s.userRepo.GetUserByUsername(ctx, username)
    if err != nil {
        return nil, fmt.Errorf("get user by username: %w", err)
    }
    return user, nil
}
```

### 3.3 Handler (`internal/api/handler/user_handler.go`)
Add `GetProfile` method:
- **Endpoint**: `GET /users/:username`
- **Logic**:
  1. Extract `username` from path param.
  2. Call `h.userUseCase.GetUserProfile`.
  3. Map `domain.User` to `dto.UserResponse`.
- **Response Mapping**:
  - `Username` -> `user.Username`
  - `Email` -> `user.Email`
  - `Name` -> `user.FirstName + " " + user.LastName`

### 3.4 Routing
The route will be registered in the public group (no JWT middleware required):
```go
e.GET("/users/:username", userHandler.GetProfile)
```

## 4. Error Handling
- **404 Not Found**: Returned when `userRepo.GetUserByUsername` returns `domain.ErrUserNotFound`. Message: `"User {username} not found"`.
- **500 Internal Server Error**: Returned for any other database or unexpected error. Message: `"Something went wrong"`.

## 5. Testing Strategy
- **Unit Test (UseCase)**: Mock `UserRepository` to test successful retrieval and "not found" scenario.
- **Unit Test (Handler)**: Mock `UserUseCase` to verify status codes (200, 404, 500) and correct DTO mapping.
