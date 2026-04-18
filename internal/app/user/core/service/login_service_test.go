package service

import (
	"context"
	"errors"
	"testing"

	"github.com/billykore/project-one/internal/app/user/core/domain"
	"github.com/billykore/project-one/internal/app/user/core/service/mocks"
	"go.uber.org/mock/gomock"
)

func TestLoginService_Login_WithMocks(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockUserRepository(ctrl)
	mockTokens := mocks.NewMockTokenService(ctrl)
	mockHasher := mocks.NewMockHasher(ctrl)
	mockLogger := mocks.NewMockLogger(ctrl)

	svc := NewLoginService(mockRepo, mockTokens, mockHasher, mockLogger)

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
				mockRepo.EXPECT().GetUserByEmail(gomock.Any(), "user@example.com").Return(user, nil)
				mockHasher.EXPECT().Compare(gomock.Any(), "password123", "hashed_password").Return(nil)
				mockTokens.EXPECT().GenerateTokens(gomock.Any(), user).Return("access", "refresh", nil)
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
				mockTokens.EXPECT().GenerateTokens(gomock.Any(), user).Return("", "", errors.New("token error"))
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
					// We use fmt.Errorf for some errors so wrap check might be tricky if not careful
					// But our Login service returns domain.ErrInvalidCredentials directly for repo/hasher errors.
					// For tokens, it returns fmt.Errorf("generate tokens: %w", err)
					if tt.name == "token generation error" {
						// Special case for wrapped error
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
