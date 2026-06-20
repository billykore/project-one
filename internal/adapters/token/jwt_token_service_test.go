package token

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"testing"
	"time"

	"github.com/billykore/project-one/internal/core/domain"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newRSAKeyPair(t *testing.T) (*rsa.PrivateKey, *rsa.PublicKey) {
	t.Helper()

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	publicKeyDER, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	require.NoError(t, err)
	publicBlock, _ := pem.Decode(pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: publicKeyDER}))
	require.NotNil(t, publicBlock)

	parsedPublicKey, err := x509.ParsePKIXPublicKey(publicBlock.Bytes)
	require.NoError(t, err)

	rsaPublicKey, ok := parsedPublicKey.(*rsa.PublicKey)
	require.True(t, ok)

	return privateKey, rsaPublicKey
}

func TestJWTTokenService_GenerateAndValidateWithRSA(t *testing.T) {
	privateKey, publicKey := newRSAKeyPair(t)
	svc := NewJWTTokenService(privateKey, publicKey, time.Hour)

	token, err := svc.GenerateTokens(context.Background(), &domain.User{Username: "alice"})
	require.NoError(t, err)
	require.NotNil(t, token)

	username, err := svc.ValidateToken(context.Background(), token.Token)
	require.NoError(t, err)
	assert.Equal(t, "alice", username)
	assert.WithinDuration(t, time.Now().Add(time.Hour), token.ExpiresAt, time.Minute)
}

func TestJWTTokenService_RejectsWrongPublicKey(t *testing.T) {
	privateKey, _ := newRSAKeyPair(t)
	_, wrongPublicKey := newRSAKeyPair(t)
	svc := NewJWTTokenService(privateKey, wrongPublicKey, time.Hour)

	token, err := svc.GenerateTokens(context.Background(), &domain.User{Username: "alice"})
	require.NoError(t, err)

	username, err := svc.ValidateToken(context.Background(), token.Token)
	assert.Error(t, err)
	assert.Empty(t, username)
}

func TestJWTTokenService_RejectsHS256Token(t *testing.T) {
	privateKey, publicKey := newRSAKeyPair(t)
	svc := NewJWTTokenService(privateKey, publicKey, time.Hour)

	tokenString, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": "alice",
		"exp":      time.Now().Add(time.Hour).Unix(),
	}).SignedString([]byte("secret"))
	require.NoError(t, err)

	username, err := svc.ValidateToken(context.Background(), tokenString)
	assert.Error(t, err)
	assert.Empty(t, username)
}
