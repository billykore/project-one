package token

import (
	"context"
	"crypto/rsa"
	"time"

	"github.com/billykore/project-one/internal/core/domain"
	"github.com/billykore/project-one/internal/core/ports"
	"github.com/golang-jwt/jwt/v5"
)

type jwtTokenService struct {
	privateKey       *rsa.PrivateKey
	publicKey        *rsa.PublicKey
	accessExpiration time.Duration
}

// NewJWTTokenService creates a new instance of TokenService.
func NewJWTTokenService(privateKey *rsa.PrivateKey, publicKey *rsa.PublicKey, accessExpiration time.Duration) ports.TokenService {
	return &jwtTokenService{
		privateKey:       privateKey,
		publicKey:        publicKey,
		accessExpiration: accessExpiration,
	}
}

func (s *jwtTokenService) GenerateTokens(_ context.Context, user *domain.User) (*domain.UserToken, error) {
	// Access token
	accessExp := time.Now().Add(s.accessExpiration)
	accessClaims := jwt.MapClaims{
		"userID": user.ID,
		"exp":    accessExp.Unix(),
	}
	accessToken, err := jwt.NewWithClaims(jwt.SigningMethodRS256, accessClaims).SignedString(s.privateKey)
	if err != nil {
		return nil, err
	}

	return &domain.UserToken{
		UserID:    user.ID,
		Username:  user.Username,
		Token:     accessToken,
		ExpiresAt: accessExp,
	}, nil
}

func (s *jwtTokenService) ValidateToken(_ context.Context, tokenString string) (*domain.User, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, domain.ErrUntrustedToken
		}
		return s.publicKey, nil
	})

	if err != nil || !token.Valid {
		return nil, domain.ErrUntrustedToken
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, domain.ErrUntrustedToken
	}

	userID, ok := claimInt(claims["userID"])
	if !ok {
		return nil, domain.ErrUntrustedToken
	}

	return &domain.User{
		ID: userID,
	}, nil
}

func claimInt(value any) (int, bool) {
	switch v := value.(type) {
	case int:
		return v, true
	case int64:
		return int(v), true
	case float64:
		return int(v), true
	default:
		return 0, false
	}
}
