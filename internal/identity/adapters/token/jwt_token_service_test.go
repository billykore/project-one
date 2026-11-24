package token

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"testing"
	"time"

	"github.com/billykore/project-one/internal/identity/domain"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
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

	token, err := svc.GenerateTokens(context.Background(), &domain.User{ID: 1, Username: "alice"})
	require.NoError(t, err)
	require.NotNil(t, token)

	user, err := svc.ValidateToken(context.Background(), token.Token)
	require.NoError(t, err)
	require.NotNil(t, user)
	assert.Empty(t, user.Username)
	assert.Equal(t, 1, user.ID)
	assert.WithinDuration(t, time.Now().Add(time.Hour), token.ExpiresAt, time.Minute)
}

func TestJWTTokenService_RejectsWrongPublicKey(t *testing.T) {
	privateKey, _ := newRSAKeyPair(t)
	_, wrongPublicKey := newRSAKeyPair(t)
	svc := NewJWTTokenService(privateKey, wrongPublicKey, time.Hour)

	token, err := svc.GenerateTokens(context.Background(), &domain.User{ID: 1, Username: "alice"})
	require.NoError(t, err)

	user, err := svc.ValidateToken(context.Background(), token.Token)
	assert.Error(t, err)
	assert.Nil(t, user)
}

func TestJWTTokenService_RejectsHS256Token(t *testing.T) {
	privateKey, publicKey := newRSAKeyPair(t)
	svc := NewJWTTokenService(privateKey, publicKey, time.Hour)

	tokenString, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": "alice",
		"exp":      time.Now().Add(time.Hour).Unix(),
	}).SignedString([]byte("secret"))
	require.NoError(t, err)

	user, err := svc.ValidateToken(context.Background(), tokenString)
	assert.Error(t, err)
	assert.Nil(t, user)
}

func TestJWTTokenService_IssuesUniqueTokensWithJTIAndIAT(t *testing.T) {
	privateKey, publicKey := newRSAKeyPair(t)
	svc := NewJWTTokenService(privateKey, publicKey, time.Hour)
	user := &domain.User{ID: 1, Username: "alice"}

	// Same user, same second: before jti these signatures were identical and
	// collided on the user_tokens.token uniqueness constraint.
	first, err := svc.GenerateTokens(context.Background(), user)
	require.NoError(t, err)
	second, err := svc.GenerateTokens(context.Background(), user)
	require.NoError(t, err)
	assert.NotEqual(t, first.Token, second.Token)

	claims := jwt.MapClaims{}
	_, err = jwt.ParseWithClaims(first.Token, claims, func(*jwt.Token) (any, error) { return publicKey, nil })
	require.NoError(t, err)

	jti, ok := claims["jti"].(string)
	require.True(t, ok, "jti claim must be a string")
	_, err = uuid.Parse(jti)
	assert.NoError(t, err, "jti must be a UUID")

	iat, err := claims.GetIssuedAt()
	require.NoError(t, err)
	require.NotNil(t, iat)
	assert.WithinDuration(t, time.Now(), iat.Time, time.Minute)
}
