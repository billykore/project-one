# GET /users/:username Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Create a public API endpoint `GET /users/:username` to retrieve user profile information.

**Architecture:** Following Clean Architecture, we'll extend the `UserUseCase` port, implement the logic in `userUseCase`, add a handler method in `UserHandler`, and register the route in `cmd/main.go`.

**Tech Stack:** Go, Echo, GORM, Testify, GoMock.

---

### Task 1: Update UserUseCase Port

**Files:**
- Modify: `internal/core/ports/user.go`

- [ ] **Step 1: Add GetUserProfile to UserUseCase interface**

```go
type UserUseCase interface {
	GetCurrentUser(ctx context.Context, username string) (*domain.User, error)
	Register(ctx context.Context, user *domain.User) error
	// Add this:
	GetUserProfile(ctx context.Context, username string) (*domain.User, error)
}
```

- [ ] **Step 2: Commit**

```bash
git add internal/core/ports/user.go
git commit -m "port: add GetUserProfile to UserUseCase"
```

---

### Task 2: Implement UserUseCase.GetUserProfile

**Files:**
- Modify: `internal/core/usecase/user_usecase.go`
- Test: `internal/core/usecase/user_usecase_test.go`

- [ ] **Step 1: Write failing test in user_usecase_test.go**

```go
func TestUserUseCase_GetUserProfile(t *testing.T) {
    ctrl := gomock.NewController(t)
    defer ctrl.Finish()

    mockUserRepo := mocks.NewMockUserRepository(ctrl)
    mockTokenRepo := mocks.NewMockTokenRepository(ctrl)
    mockHasher := mocks.NewMockHasher(ctrl)
    
    uc := NewUserUseCase(mockUserRepo, mockTokenRepo, mockHasher)
    ctx := context.Background()
    username := "testuser"

    t.Run("success", func(t *testing.T) {
        expectedUser := &domain.User{Username: username}
        mockUserRepo.EXPECT().GetUserByUsername(ctx, username).Return(expectedUser, nil)

        user, err := uc.GetUserProfile(ctx, username)
        assert.NoError(t, err)
        assert.Equal(t, expectedUser, user)
    })

    t.Run("not found", func(t *testing.T) {
        mockUserRepo.EXPECT().GetUserByUsername(ctx, username).Return(nil, domain.ErrUserNotFound)

        user, err := uc.GetUserProfile(ctx, username)
        assert.ErrorIs(t, err, domain.ErrUserNotFound)
        assert.Nil(t, user)
    })
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test -v internal/core/usecase/user_usecase_test.go`
Expected: Compilation error (GetUserProfile not implemented)

- [ ] **Step 3: Implement GetUserProfile in user_usecase.go**

```go
func (s *userUseCase) GetUserProfile(ctx context.Context, username string) (*domain.User, error) {
	user, err := s.userRepo.GetUserByUsername(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("get user by username: %w", err)
	}
	return user, nil
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test -v ./internal/core/usecase/...`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/core/usecase/user_usecase.go internal/core/usecase/user_usecase_test.go
git commit -m "usecase: implement GetUserProfile"
```

---

### Task 3: Implement UserHandler.GetProfile

**Files:**
- Modify: `internal/api/handler/user_handler.go`
- Test: `internal/api/handler/user_handler_test.go` (Create if not exists)

- [ ] **Step 1: Write failing test for GetProfile**

```go
func TestUserHandler_GetProfile(t *testing.T) {
    // Setup Echo and Mocks
    e := echo.New()
    mockUC := mocks.NewMockUserUseCase(ctrl)
    h := NewUserHandler(mockUC, ...)

    t.Run("success", func(t *testing.T) {
        username := "johndoe"
        req := httptest.NewRequest(http.MethodGet, "/users/"+username, nil)
        rec := httptest.NewRecorder()
        c := e.NewContext(req, rec)
        c.SetParamNames("username")
        c.SetParamValues(username)

        mockUC.EXPECT().GetUserProfile(gomock.Any(), username).Return(&domain.User{
            Username: username,
            Email: "johndoe@gmail.com",
            FirstName: "John",
            LastName: "Doe",
        }, nil)

        if assert.NoError(t, h.GetProfile(c)) {
            assert.Equal(t, http.StatusOK, rec.Code)
            // Verify JSON body
        }
    })
}
```

- [ ] **Step 2: Run test to verify it fails**

Expected: Compilation error (GetProfile not defined)

- [ ] **Step 3: Implement GetProfile in user_handler.go**

```go
func (h *UserHandler) GetProfile(c echo.Context) error {
	username := c.Param("username")
	if username == "" {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "Invalid username"})
	}

	user, err := h.userUseCase.GetUserProfile(c.Request().Context(), username)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return c.JSON(http.StatusNotFound, dto.ErrorResponse{Error: fmt.Sprintf("User %s not found", username)})
		}
		h.log.Error(c.Request().Context(), "failed to get user profile", "username", username, "error", err)
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "Something went wrong"})
	}

	res := dto.UserResponse{
		Username: user.Username,
		Email:    user.Email,
		Name:     user.FirstName + " " + user.LastName,
	}

	return c.JSON(http.StatusOK, res)
}
```

- [ ] **Step 4: Run test to verify it passes**

- [ ] **Step 5: Commit**

```bash
git add internal/api/handler/user_handler.go
git commit -m "api: implement GetProfile handler"
```

---

### Task 4: Register Route

**Files:**
- Modify: `cmd/main.go` (or wherever routes are registered)

- [ ] **Step 1: Find route registration and add the new endpoint**

Search for `v1 := e.Group("/api/v1")` or similar.
Add: `e.GET("/users/:username", userHandler.GetProfile)`

- [ ] **Step 2: Verify with manual curl or integration test**

- [ ] **Step 3: Commit**

```bash
git add cmd/main.go
git commit -m "api: register GET /users/:username route"
```

---

### Task 5: Update Swagger Docs

- [ ] **Step 1: Add Swagger annotations to GetProfile**
- [ ] **Step 2: Run `make docs`**
- [ ] **Step 3: Commit**

```bash
make docs
git add internal/api/handler/user_handler.go api/swagger/
git commit -m "docs: update swagger for GetProfile"
```
