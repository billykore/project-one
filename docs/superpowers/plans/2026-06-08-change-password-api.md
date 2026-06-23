# Change Password API Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement the backend API for password updates following Clean Architecture principles, allowing authenticated users to change their password by providing their old and new passwords.

**Architecture:** We will update ports, repository, usecases, DTOs, and handlers to support password updates. The request will be handled by a secure Echo handler, validated via the custom validator, and stored securely using GORM.

**Tech Stack:** Go 1.26+, Echo, GORM, Bcrypt, Validator v10, GoMock, Testify.

---

### Task 1: Update Ports Interface

**Files:**
- Modify: `internal/core/ports/user.go`

- [ ] **Step 1: Update ports.UserRepository interface**
  Add the `UpdateUser` signature to `UserRepository`:
  ```go
  // UpdateUser updates user details, including password if changed.
  UpdateUser(ctx context.Context, user *domain.User) error
  ```

- [ ] **Step 2: Update ports.UserUseCase interface**
  Add the `ChangePassword` signature to `UserUseCase`:
  ```go
  // ChangePassword verifies the old password and sets the new password.
  ChangePassword(ctx context.Context, username, oldPassword, newPassword string) error
  ```

- [ ] **Step 3: Run mock generation**
  Run: `make mock`
  Expected: Success, mocks in `internal/core/usecase/mocks/` are regenerated.

---

### Task 2: Implement UserRepository.UpdateUser

**Files:**
- Modify: `internal/adapters/repository/user_repository.go`

- [ ] **Step 1: Implement UpdateUser method**
  Add the `UpdateUser` method in `internal/adapters/repository/user_repository.go`:
  ```go
  func (r *userRepository) UpdateUser(ctx context.Context, user *domain.User) error {
  	m := userModel{
  		ID:        user.ID,
  		Email:     user.Email,
  		Username:  user.Username,
  		Password:  user.Password,
  		FirstName: user.FirstName,
  		LastName:  user.LastName,
  		CreatedAt: user.CreatedAt,
  		UpdatedAt: time.Now(),
  	}
  	// Save updates all fields by primary key
  	return r.db.WithContext(ctx).Save(&m).Error
  }
  ```

- [ ] **Step 2: Verify it compiles**
  Run: `go build ./internal/adapters/repository`
  Expected: No compilation errors.

---

### Task 3: Implement UserUseCase.ChangePassword & Tests

**Files:**
- Modify: `internal/core/usecase/user_usecase.go`
- Modify: `internal/core/usecase/user_usecase_test.go`

- [ ] **Step 1: Implement ChangePassword logic**
  Add `ChangePassword` method to `internal/core/usecase/user_usecase.go`:
  ```go
  func (u *userUseCase) ChangePassword(ctx context.Context, username, oldPassword, newPassword string) error {
  	// 1. Retrieve user
  	user, err := u.userRepo.GetUserByUsername(ctx, username)
  	if err != nil {
  		return fmt.Errorf("get user by username: %w", err)
  	}

  	// 2. Validate current password
  	if err := u.hasher.Compare(ctx, oldPassword, user.Password); err != nil {
  		return domain.ErrInvalidCredentials
  	}

  	// 3. Validate new password length (minimum 8 characters)
  	if len(newPassword) < 8 {
  		return fmt.Errorf("%w: password must be at least 8 characters", domain.ErrValidationFailed)
  	}

  	// 4. Hash new password
  	hashedPassword, err := u.hasher.Hash(ctx, newPassword)
  	if err != nil {
  		return fmt.Errorf("hash password: %w", err)
  	}

  	user.Password = hashedPassword

  	// 5. Save to repository
  	if err := u.userRepo.UpdateUser(ctx, user); err != nil {
  		return fmt.Errorf("update user: %w", err)
  	}

  	return nil
  }
  ```

- [ ] **Step 2: Add unit tests for ChangePassword**
  Add `TestUserUseCase_ChangePassword` in `internal/core/usecase/user_usecase_test.go`:
  ```go
  func TestUserUseCase_ChangePassword(t *testing.T) {
  	ctrl := gomock.NewController(t)
  	defer ctrl.Finish()

  	mockRepo := mocks.NewMockUserRepository(ctrl)
  	mockTokenRepo := mocks.NewMockTokenRepository(ctrl)
  	mockHasher := mocks.NewMockHasher(ctrl)
  	svc := NewUserUseCase(mockRepo, mockTokenRepo, mockHasher)

  	ctx := context.Background()
  	username := "testuser"

  	t.Run("success", func(t *testing.T) {
  		existingUser := &domain.User{
  			ID:       1,
  			Username: username,
  			Password: "hashed_old_password",
  		}
  		mockRepo.EXPECT().GetUserByUsername(ctx, username).Return(existingUser, nil)
  		mockHasher.EXPECT().Compare(ctx, "old_password", "hashed_old_password").Return(nil)
  		mockHasher.EXPECT().Hash(ctx, "new_password_123").Return("hashed_new_password", nil)
  		mockRepo.EXPECT().UpdateUser(ctx, gomock.Any()).DoAndReturn(func(ctx context.Context, u *domain.User) error {
  			assert.Equal(t, "hashed_new_password", u.Password)
  			return nil
  		})

  		err := svc.ChangePassword(ctx, username, "old_password", "new_password_123")
  		assert.NoError(t, err)
  	})

  	t.Run("user not found", func(t *testing.T) {
  		mockRepo.EXPECT().GetUserByUsername(ctx, username).Return(nil, domain.ErrUserNotFound)

  		err := svc.ChangePassword(ctx, username, "old_password", "new_password_123")
  		assert.ErrorIs(t, err, domain.ErrUserNotFound)
  	})

  	t.Run("invalid credentials", func(t *testing.T) {
  		existingUser := &domain.User{
  			ID:       1,
  			Username: username,
  			Password: "hashed_old_password",
  		}
  		mockRepo.EXPECT().GetUserByUsername(ctx, username).Return(existingUser, nil)
  		mockHasher.EXPECT().Compare(ctx, "wrong_old_password", "hashed_old_password").Return(domain.ErrInvalidCredentials)

  		err := svc.ChangePassword(ctx, username, "wrong_old_password", "new_password_123")
  		assert.ErrorIs(t, err, domain.ErrInvalidCredentials)
  	})

  	t.Run("validation failed - too short", func(t *testing.T) {
  		existingUser := &domain.User{
  			ID:       1,
  			Username: username,
  			Password: "hashed_old_password",
  		}
  		mockRepo.EXPECT().GetUserByUsername(ctx, username).Return(existingUser, nil)
  		mockHasher.EXPECT().Compare(ctx, "old_password", "hashed_old_password").Return(nil)

  		err := svc.ChangePassword(ctx, username, "old_password", "short")
  		assert.ErrorIs(t, err, domain.ErrValidationFailed)
  	})
  }
  ```

- [ ] **Step 3: Run usecase tests**
  Run: `go test -v ./internal/core/usecase -run TestUserUseCase_ChangePassword`
  Expected: PASS

---

### Task 4: Add DTO for ChangePassword

**Files:**
- Modify: `internal/api/dto/user_dto.go`

- [ ] **Step 1: Add ChangePasswordRequest struct**
  Append the following to `internal/api/dto/user_dto.go`:
  ```go
  // ChangePasswordRequest is the request body for updating the password.
  type ChangePasswordRequest struct {
  	OldPassword string `json:"old_password" validate:"required,min=8"`
  	NewPassword string `json:"new_password" validate:"required,min=8"`
  }
  ```

---

### Task 5: Implement UserHandler.HandleChangePassword & Handler Tests

**Files:**
- Modify: `internal/api/handler/user_handler.go`
- Create: `internal/api/handler/user_handler_test.go`

- [ ] **Step 1: Implement HandleChangePassword**
  Add the method to `internal/api/handler/user_handler.go`:
  ```go
  // HandleChangePassword handles the PUT /users/password endpoint.
  //
  //	@Summary		Change password
  //	@Description	Verify old password and update to new password.
  //	@Tags			users
  //	@Accept			json
  //	@Produce		json
  //	@Param			request	body		dto.ChangePasswordRequest	true	"Change password request details"
  //	@Success		200		{object}	dto.MessageResponse
  //	@Failure		400		{object}	dto.ErrorResponse
  //	@Failure		401		{object}	dto.ErrorResponse
  //	@Failure		500		{object}	dto.ErrorResponse
  //	@Security		BearerAuth
  //	@Router			/users/password [put]
  func (h *UserHandler) HandleChangePassword(c echo.Context) error {
  	username, ok := c.Get("username").(string)
  	if !ok {
  		return c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "Unauthorized"})
  	}

  	var req dto.ChangePasswordRequest
  	if err := c.Bind(&req); err != nil {
  		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "Invalid request body"})
  	}

  	if err := h.validator.Validate(req); err != nil {
  		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
  	}

  	err := h.userUseCase.ChangePassword(c.Request().Context(), username, req.OldPassword, req.NewPassword)
  	if err != nil {
  		if errors.Is(err, domain.ErrInvalidCredentials) {
  			return c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "Invalid current password"})
  		}
  		if errors.Is(err, domain.ErrValidationFailed) {
  			return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
  		}
  		h.log.Error(c.Request().Context(), "failed to change password", "username", username, "error", err)
  		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "Internal server error"})
  	}

  	return c.JSON(http.StatusOK, dto.MessageResponse{Message: "Password updated successfully"})
  }
  ```

- [ ] **Step 2: Create unit tests for HandleChangePassword**
  Create `internal/api/handler/user_handler_test.go`:
  ```go
  package handler

  import (
  	"bytes"
  	"encoding/json"
  	"errors"
  	"net/http"
  	"net/http/httptest"
  	"strings"
  	"testing"

  	"github.com/billykore/project-one/internal/api/dto"
  	"github.com/billykore/project-one/internal/core/domain"
  	"github.com/billykore/project-one/internal/core/ports/mocks"
  	"github.com/labstack/echo/v4"
  	"github.com/stretchr/testify/assert"
  	"go.uber.org/mock/gomock"
  )

  func TestUserHandler_HandleChangePassword(t *testing.T) {
  	ctrl := gomock.NewController(t)
  	defer ctrl.Finish()

  	mockUserUC := mocks.NewMockUserUseCase(ctrl)
  	mockLoginUC := mocks.NewMockLoginUseCase(ctrl)
  	mockFollowUC := mocks.NewMockFollowUseCase(ctrl)
  	mockPostUC := mocks.NewMockPostUseCase(ctrl)
  	mockValidator := mocks.NewMockValidator(ctrl)
  	mockLogger := mocks.NewMockLogger(ctrl)

  	h := NewUserHandler(mockUserUC, mockLoginUC, mockFollowUC, mockPostUC, mockValidator, mockLogger)

  	t.Run("success", func(t *testing.T) {
  		e := echo.New()
  		reqBody := dto.ChangePasswordRequest{
  			OldPassword: "oldpassword123",
  			NewPassword: "newpassword123",
  		}
  		body, _ := json.Marshal(reqBody)
  		req := httptest.NewRequest(http.MethodPut, "/users/password", bytes.NewReader(body))
  		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
  		rec := httptest.NewRecorder()
  		c := e.NewContext(req, rec)
  		c.Set("username", "testuser")

  		mockValidator.EXPECT().Validate(reqBody).Return(nil)
  		mockUserUC.EXPECT().
  			ChangePassword(gomock.Any(), "testuser", "oldpassword123", "newpassword123").
  			Return(nil)

  		err := h.HandleChangePassword(c)
  		assert.NoError(t, err)
  		assert.Equal(t, http.StatusOK, rec.Code)

  		var resp dto.MessageResponse
  		err = json.Unmarshal(rec.Body.Bytes(), &resp)
  		assert.NoError(t, err)
  		assert.Equal(t, "Password updated successfully", resp.Message)
  	})

  	t.Run("unauthorized - no user context", func(t *testing.T) {
  		e := echo.New()
  		req := httptest.NewRequest(http.MethodPut, "/users/password", nil)
  		rec := httptest.NewRecorder()
  		c := e.NewContext(req, rec)

  		err := h.HandleChangePassword(c)
  		assert.NoError(t, err)
  		assert.Equal(t, http.StatusUnauthorized, rec.Code)
  	})

  	t.Run("bad request - invalid json", func(t *testing.T) {
  		e := echo.New()
  		req := httptest.NewRequest(http.MethodPut, "/users/password", strings.NewReader("invalid json"))
  		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
  		rec := httptest.NewRecorder()
  		c := e.NewContext(req, rec)
  		c.Set("username", "testuser")

  		err := h.HandleChangePassword(c)
  		assert.NoError(t, err)
  		assert.Equal(t, http.StatusBadRequest, rec.Code)
  	})

  	t.Run("bad request - validation error", func(t *testing.T) {
  		e := echo.New()
  		reqBody := dto.ChangePasswordRequest{
  			OldPassword: "short",
  			NewPassword: "newpassword123",
  		}
  		body, _ := json.Marshal(reqBody)
  		req := httptest.NewRequest(http.MethodPut, "/users/password", bytes.NewReader(body))
  		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
  		rec := httptest.NewRecorder()
  		c := e.NewContext(req, rec)
  		c.Set("username", "testuser")

  		validationErr := errors.New("validation failed")
  		mockValidator.EXPECT().Validate(reqBody).Return(validationErr)

  		err := h.HandleChangePassword(c)
  		assert.NoError(t, err)
  		assert.Equal(t, http.StatusBadRequest, rec.Code)
  	})

  	t.Run("unauthorized - invalid current password", func(t *testing.T) {
  		e := echo.New()
  		reqBody := dto.ChangePasswordRequest{
  			OldPassword: "wrongpassword",
  			NewPassword: "newpassword123",
  		}
  		body, _ := json.Marshal(reqBody)
  		req := httptest.NewRequest(http.MethodPut, "/users/password", bytes.NewReader(body))
  		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
  		rec := httptest.NewRecorder()
  		c := e.NewContext(req, rec)
  		c.Set("username", "testuser")

  		mockValidator.EXPECT().Validate(reqBody).Return(nil)
  		mockUserUC.EXPECT().
  			ChangePassword(gomock.Any(), "testuser", "wrongpassword", "newpassword123").
  			Return(domain.ErrInvalidCredentials)

  		err := h.HandleChangePassword(c)
  		assert.NoError(t, err)
  		assert.Equal(t, http.StatusUnauthorized, rec.Code)
  	})

  	t.Run("internal server error", func(t *testing.T) {
  		e := echo.New()
  		reqBody := dto.ChangePasswordRequest{
  			OldPassword: "oldpassword123",
  			NewPassword: "newpassword123",
  		}
  		body, _ := json.Marshal(reqBody)
  		req := httptest.NewRequest(http.MethodPut, "/users/password", bytes.NewReader(body))
  		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
  		rec := httptest.NewRecorder()
  		c := e.NewContext(req, rec)
  		c.Set("username", "testuser")

  		mockValidator.EXPECT().Validate(reqBody).Return(nil)
  		mockUserUC.EXPECT().
  			ChangePassword(gomock.Any(), "testuser", "oldpassword123", "newpassword123").
  			Return(errors.New("db error"))
  		mockLogger.EXPECT().Error(gomock.Any(), "failed to change password", "username", "testuser", "error", gomock.Any())

  		err := h.HandleChangePassword(c)
  		assert.NoError(t, err)
  		assert.Equal(t, http.StatusInternalServerError, rec.Code)
  	})
  }
  ```

- [ ] **Step 3: Run handler tests**
  Run: `go test -v ./internal/api/handler`
  Expected: PASS

---

### Task 6: Register Route

**Files:**
- Modify: `cmd/main.go`

- [ ] **Step 1: Register PUT /password**
  Add the route in `cmd/main.go` inside the `usersAuth` group:
  ```go
  usersAuth.PUT("/password", userHdl.HandleChangePassword)
  ```

---

### Task 7: Verification

- [ ] **Step 1: Regenerate Swagger documentation**
  Run: `make docs`
  Expected: Success

- [ ] **Step 2: Run all checks**
  Run: `make check`
  Expected: SUCCESS (all tests pass, lint/vet pass, compilation succeeds)
