package token

import (
	"context"
	"time"

	"github.com/billykore/project-one/internal/core/domain"
	"github.com/billykore/project-one/internal/core/ports"
	"github.com/golang-jwt/jwt/v5"
)

type jwtTokenService struct {
	secret           []byte
	accessExpiration time.Duration
}

// NewJWTTokenService creates a new instance of TokenService.
func NewJWTTokenService(secret string, accessExpiration time.Duration) ports.TokenService {
	return &jwtTokenService{
		secret:           []byte(secret),
		accessExpiration: accessExpiration,
	}
}

func (s *jwtTokenService) GenerateTokens(_ context.Context, user *domain.User) (*domain.UserToken, error) {
	// Access token
	accessExp := time.Now().Add(s.accessExpiration)
	accessClaims := jwt.MapClaims{
		"username": user.Username,
		"exp":      accessExp.Unix(),
	}
	accessToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims).SignedString(s.secret)
	if err != nil {
		return nil, err
	}

	return &domain.UserToken{Token: accessToken, ExpiresAt: accessExp}, nil
}

func (s *jwtTokenService) ValidateToken(_ context.Context, tokenString string) (string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return "", domain.ErrUnauthorized
		}
		return s.secret, nil
	})

	if err != nil || !token.Valid {
		return "", domain.ErrUnauthorized
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", domain.ErrUnauthorized
	}

	username, ok := claims["username"].(string)
	if !ok {
		return "", domain.ErrUnauthorized
	}

	return username, nil
}
