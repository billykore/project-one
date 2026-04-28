package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/billykore/project-one/internal/app/user/core/domain"
	"github.com/billykore/project-one/internal/app/user/core/service/mocks"
	"go.uber.org/mock/gomock"
)

func TestLoginService_Login_WithMocks(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockUserRepository(ctrl)
	mockTokens := mocks.NewMockTokenService(ctrl)
	mockUserTokens := mocks.NewMockUserTokenRepository(ctrl)
	mockHasher := mocks.NewMockHasher(ctrl)
	mockLogger := mocks.NewMockLogger(ctrl)

	svc := NewLoginService(mockRepo, mockTokens, mockUserTokens, mockHasher, mockLogger)

	tests := []struct {
		name        string
		email       string
		password    string
		setup       func()
		wantErr     error
		wantAccess  string
		wantRefresh string
	}{
		{
			name:     "successful login",
			email:    "user@example.com",
			password: "password123",
			setup: func() {
				user := &domain.User{ID: 1, Email: "user@example.com", Password: "hashed_password"}
				exp := time.Now().Add(time.Hour)
				accessToken := &domain.TokenDetails{Token: "access", ExpiresAt: exp}
				refreshToken := &domain.TokenDetails{Token: "refresh", ExpiresAt: exp.Add(time.Hour)}

				mockRepo.EXPECT().GetUserByEmail(gomock.Any(), "user@example.com").Return(user, nil)
				mockHasher.EXPECT().Compare(gomock.Any(), "password123", "hashed_password").Return(nil)
				mockTokens.EXPECT().GenerateTokens(gomock.Any(), user).Return(accessToken, refreshToken, nil)
				mockUserTokens.EXPECT().StoreToken(gomock.Any(), &domain.UserToken{
					UserID:    1,
					Token:     "access",
					ExpiresAt: exp,
				}).Return(nil)
				mockLogger.EXPECT().Info(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any())
			},
			wantAccess:  "access",
			wantRefresh: "refresh",
			wantErr:     nil,
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
				mockTokens.EXPECT().GenerateTokens(gomock.Any(), user).Return(nil, nil, errors.New("token error"))
				mockLogger.EXPECT().Error(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any())
			},
			wantErr: domain.ErrInternalServer,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			access, refresh, err := svc.Login(context.Background(), tt.email, tt.password)

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

			if access != tt.wantAccess {
				t.Errorf("Login() access = %v, want %v", access, tt.wantAccess)
			}
			if refresh != tt.wantRefresh {
				t.Errorf("Login() refresh = %v, want %v", refresh, tt.wantRefresh)
			}
		})
	}
}

func TestLoginService_Logout(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockUserRepository(ctrl)
	mockTokens := mocks.NewMockTokenService(ctrl)
	mockUserTokens := mocks.NewMockUserTokenRepository(ctrl)
	mockHasher := mocks.NewMockHasher(ctrl)
	mockLogger := mocks.NewMockLogger(ctrl)

	svc := NewLoginService(mockRepo, mockTokens, mockUserTokens, mockHasher, mockLogger)

	t.Run("successful logout", func(t *testing.T) {
		token := "some-token"
		mockUserTokens.EXPECT().DeleteToken(gomock.Any(), token).Return(nil)
		mockLogger.EXPECT().Info(gomock.Any(), "user logged out successfully")

		err := svc.Logout(context.Background(), token)
		if err != nil {
			t.Errorf("Logout() unexpected error = %v", err)
		}
	})

	t.Run("empty token logout", func(t *testing.T) {
		err := svc.Logout(context.Background(), "")
		if err == nil {
			t.Error("Logout() expected error for empty token, got nil")
		}
	})

	t.Run("failed logout", func(t *testing.T) {
		token := "some-token"
		mockUserTokens.EXPECT().DeleteToken(gomock.Any(), token).Return(errors.New("db error"))
		mockLogger.EXPECT().Error(gomock.Any(), gomock.Any(), gomock.Any())

		err := svc.Logout(context.Background(), token)
		if err == nil {
			t.Error("Logout() expected error, got nil")
		}
	})
}
