package token

import (
	"context"
	"time"

	"github.com/billykore/project-one/internal/app/login/core/domain"
	"github.com/billykore/project-one/internal/app/login/core/ports"
	"github.com/golang-jwt/jwt/v5"
)

type jwtTokenService struct {
	secret            []byte
	accessExpiration  time.Duration
	refreshExpiration time.Duration
}

// NewJWTTokenService creates a new instance of TokenService.
func NewJWTTokenService(secret string) ports.TokenService {
	return &jwtTokenService{
		secret:            []byte(secret),
		accessExpiration:  time.Hour * 1,
		refreshExpiration: time.Hour * 24 * 7,
	}
}

func (s *jwtTokenService) GenerateTokens(_ context.Context, user *domain.User) (string, string, error) {
	// Access token
	accessClaims := jwt.MapClaims{
		"userID": user.ID,
		"exp":    time.Now().Add(s.accessExpiration).Unix(),
	}
	accessToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims).SignedString(s.secret)
	if err != nil {
		return "", "", err
	}

	// Refresh token
	refreshClaims := jwt.MapClaims{
		"userID": user.ID,
		"exp":    time.Now().Add(s.refreshExpiration).Unix(),
	}
	refreshToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims).SignedString(s.secret)
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}
