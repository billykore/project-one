package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/billykore/project-one/internal/identity/domain"
	"github.com/billykore/project-one/internal/pkg/problem"
	"github.com/billykore/project-one/internal/testkit/mocks"
	"go.uber.org/mock/gomock"
)

func TestLoginUseCase_Login_WithMocks(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockUserRepository(ctrl)
	mockTokens := mocks.NewMockTokenService(ctrl)
	mockUserTokens := mocks.NewMockTokenRepository(ctrl)
	mockHasher := mocks.NewMockHasher(ctrl)
	mockLimiter := mocks.NewMockLoginAttemptLimiter(ctrl)
	mockLogger := mocks.NewMockLogger(ctrl)

	svc := NewLoginUseCase(mockRepo, mockTokens, mockUserTokens, mockHasher, mockLimiter, mockLogger)

	tests := []struct {
		name         string
		email        string
		password     string
		setup        func()
		wantErr      error
		wantAccess   *domain.UserToken
		wantUsername string
	}{
		{
			name:     "successful login",
			email:    "user@example.com",
			password: "password123",
			setup: func() {
				user := &domain.User{ID: 1, Username: "user1", Email: "user@example.com", Password: "hashed_password"}
				exp := time.Now().Add(time.Hour)
				accessToken := &domain.UserToken{Token: "access", ExpiresAt: exp}

				mockRepo.EXPECT().GetUserByEmail(gomock.Any(), "user@example.com").Return(user, nil)
				mockHasher.EXPECT().Compare(gomock.Any(), "password123", "hashed_password").Return(nil)
				mockTokens.EXPECT().GenerateTokens(gomock.Any(), user).Return(accessToken, nil)
				mockUserTokens.EXPECT().StoreToken(gomock.Any(), &domain.UserToken{
					UserID:    user.ID,
					Username:  user.Username,
					Token:     "access",
					ExpiresAt: exp,
				}).Return(nil)
				mockLogger.EXPECT().Info(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any())
			},
			wantAccess: &domain.UserToken{Token: "access"},
			wantErr:    nil,
		},
		{
			name:     "user not found",
			email:    "notfound@example.com",
			password: "password123",
			setup: func() {
				mockRepo.EXPECT().GetUserByEmail(gomock.Any(), "notfound@example.com").Return(nil, problem.ErrUserNotFound)
				mockLogger.EXPECT().Error(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any())
			},
			wantErr: domain.ErrInvalidCredentials,
		},
		{
			name:     "password mismatch",
			email:    "user@example.com",
			password: "wrongpassword",
			setup: func() {
				user := &domain.User{ID: 1, Email: "user@example.com", Password: "hashed_password"}
				mockRepo.EXPECT().GetUserByEmail(gomock.Any(), "user@example.com").Return(user, nil)
				mockHasher.EXPECT().Compare(gomock.Any(), "wrongpassword", "hashed_password").Return(errors.New("mismatch"))
				mockLogger.EXPECT().Error(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any())
			},
			wantErr: domain.ErrInvalidCredentials,
		},
		{
			name:     "token generation error",
			email:    "user@example.com",
			password: "password123",
			setup: func() {
				user := &domain.User{ID: 1, Email: "user@example.com", Password: "hashed_password"}
				mockRepo.EXPECT().GetUserByEmail(gomock.Any(), "user@example.com").Return(user, nil)
				mockHasher.EXPECT().Compare(gomock.Any(), "password123", "hashed_password").Return(nil)
				mockTokens.EXPECT().GenerateTokens(gomock.Any(), user).Return(nil, errors.New("token error"))
				mockLogger.EXPECT().Error(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any())
			},
			wantErr: problem.ErrRepositoryFailure,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockLimiter.EXPECT().Allow(gomock.Any(), tt.email).Return(true, time.Duration(0))
			tt.setup()
			access, err := svc.Login(context.Background(), tt.email, tt.password)

			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) && err.Error() != tt.wantErr.Error() {
					if tt.name == "token generation error" {
						return
					}
					t.Errorf("Login() error = %v, wantErr %v", err, tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Errorf("Login() unexpected error = %v", err)
				return
			}

			if access == nil || access.Token != tt.wantAccess.Token {
				t.Errorf("Login() access = %v, want %v", access, tt.wantAccess)
			}
		})
	}
}

func TestLoginUseCase_Logout(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockUserRepository(ctrl)
	mockTokens := mocks.NewMockTokenService(ctrl)
	mockUserTokens := mocks.NewMockTokenRepository(ctrl)
	mockHasher := mocks.NewMockHasher(ctrl)
	mockLimiter := mocks.NewMockLoginAttemptLimiter(ctrl)
	mockLogger := mocks.NewMockLogger(ctrl)

	svc := NewLoginUseCase(mockRepo, mockTokens, mockUserTokens, mockHasher, mockLimiter, mockLogger)

	t.Run("successful logout", func(t *testing.T) {
		userID := 1
		mockUserTokens.EXPECT().DeleteTokensByUserID(gomock.Any(), userID).Return(nil)
		mockLogger.EXPECT().Info(gomock.Any(), "user logged out successfully", "userID", userID)

		err := svc.Logout(context.Background(), userID)
		if err != nil {
			t.Errorf("Logout() unexpected error = %v", err)
		}
	})

	t.Run("failed logout", func(t *testing.T) {
		userID := 1
		mockUserTokens.EXPECT().DeleteTokensByUserID(gomock.Any(), userID).Return(errors.New("db error"))
		mockLogger.EXPECT().Error(gomock.Any(), "failed to revoke user sessions on logout", "userID", userID, "error", errors.New("db error"))

		err := svc.Logout(context.Background(), userID)
		if err == nil {
			t.Error("Logout() expected error, got nil")
		}
	})

	t.Run("invalid user id", func(t *testing.T) {
		err := svc.Logout(context.Background(), 0)
		if !errors.Is(err, problem.ErrInvalidUser) {
			t.Errorf("Logout() error = %v, want %v", err, problem.ErrInvalidUser)
		}
	})
}

func TestLoginUseCase_Login_RateLimited(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockRepo := mocks.NewMockUserRepository(ctrl)
	mockTokens := mocks.NewMockTokenService(ctrl)
	mockUserTokens := mocks.NewMockTokenRepository(ctrl)
	mockHasher := mocks.NewMockHasher(ctrl)
	mockLimiter := mocks.NewMockLoginAttemptLimiter(ctrl)
	mockLogger := mocks.NewMockLogger(ctrl)

	svc := NewLoginUseCase(mockRepo, mockTokens, mockUserTokens, mockHasher, mockLimiter, mockLogger)

	// The key is normalized, and a throttled attempt never reaches the
	// repository, hasher, or token service (gomock fails on unexpected calls).
	mockLimiter.EXPECT().Allow(gomock.Any(), "user@example.com").Return(false, 42*time.Second)
	mockLogger.EXPECT().Warn(gomock.Any(), "login attempt rejected by rate limit", "retry_after_seconds", 42)

	access, err := svc.Login(context.Background(), "  User@Example.COM ", "password123")

	if access != nil {
		t.Errorf("Login() access = %v, want nil", access)
	}
	if !errors.Is(err, domain.ErrTooManyLoginAttempts) {
		t.Fatalf("Login() error = %v, want %v", err, domain.ErrTooManyLoginAttempts)
	}
	throttled, ok := errors.AsType[*domain.LoginThrottledError](err)
	if !ok {
		t.Fatalf("Login() error = %T, want *domain.LoginThrottledError", err)
	}
	if throttled.RetryAfter != 42*time.Second {
		t.Errorf("RetryAfter = %v, want %v", throttled.RetryAfter, 42*time.Second)
	}
}
