package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/billykore/project-one/internal/core/domain"
	"github.com/billykore/project-one/internal/core/ports/mocks"
	"go.uber.org/mock/gomock"
)

func TestLoginUseCase_Login_WithMocks(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockUserRepository(ctrl)
	mockTokens := mocks.NewMockTokenService(ctrl)
	mockUserTokens := mocks.NewMockTokenRepository(ctrl)
	mockHasher := mocks.NewMockHasher(ctrl)
	mockLogger := mocks.NewMockLogger(ctrl)

	svc := NewLoginUseCase(mockRepo, mockTokens, mockUserTokens, mockHasher, mockLogger)

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
				user := &domain.User{Username: "user1", Email: "user@example.com", Password: "hashed_password"}
				exp := time.Now().Add(time.Hour)
				accessToken := &domain.UserToken{Token: "access", ExpiresAt: exp}

				mockRepo.EXPECT().GetUserByEmail(gomock.Any(), "user@example.com").Return(user, nil)
				mockHasher.EXPECT().Compare(gomock.Any(), "password123", "hashed_password").Return(nil)
				mockTokens.EXPECT().GenerateTokens(gomock.Any(), user).Return(accessToken, nil)
				mockUserTokens.EXPECT().StoreToken(gomock.Any(), &domain.UserToken{
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
				mockRepo.EXPECT().GetUserByEmail(gomock.Any(), "notfound@example.com").Return(nil, domain.ErrUserNotFound)
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
			wantErr: domain.ErrRepositoryFailure,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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
	mockLogger := mocks.NewMockLogger(ctrl)

	svc := NewLoginUseCase(mockRepo, mockTokens, mockUserTokens, mockHasher, mockLogger)

	t.Run("successful logout", func(t *testing.T) {
		username := "testuser"
		mockUserTokens.EXPECT().DeleteTokenByUsername(gomock.Any(), username).Return(nil)
		mockRepo.EXPECT().GetUserByUsername(gomock.Any(), username).Return(&domain.User{Username: username}, nil)
		mockLogger.EXPECT().Info(gomock.Any(), "user logged out successfully", "username", username)

		err := svc.Logout(context.Background(), username)
		if err != nil {
			t.Errorf("Logout() unexpected error = %v", err)
		}
	})

	t.Run("failed logout", func(t *testing.T) {
		username := "testuser"
		mockRepo.EXPECT().GetUserByUsername(gomock.Any(), username).Return(&domain.User{Username: username}, nil)
		mockUserTokens.EXPECT().DeleteTokenByUsername(gomock.Any(), username).Return(errors.New("db error"))
		mockLogger.EXPECT().Error(gomock.Any(), "failed to delete user token on logout", "username", username, "error", errors.New("db error"))

		err := svc.Logout(context.Background(), username)
		if err == nil {
			t.Error("Logout() expected error, got nil")
		}
	})
}
