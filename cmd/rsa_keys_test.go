package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func writeRSAKeyPair(t *testing.T, dir string) (string, string) {
	t.Helper()

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	privateDER := x509.MarshalPKCS1PrivateKey(privateKey)
	privateFile := filepath.Join(dir, "private.pem")
	require.NoError(t, os.WriteFile(privateFile, pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: privateDER}), 0600))

	publicDER, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	require.NoError(t, err)
	publicFile := filepath.Join(dir, "public.pem")
	require.NoError(t, os.WriteFile(publicFile, pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: publicDER}), 0644))

	return privateFile, publicFile
}

func TestLoadRSAKeyPair_Success(t *testing.T) {
	dir := t.TempDir()
	privateFile, publicFile := writeRSAKeyPair(t, dir)

	privateKey, publicKey, err := loadRSAKeyPair(privateFile, publicFile)
	require.NoError(t, err)
	require.NotNil(t, privateKey)
	require.NotNil(t, publicKey)
	assert.Equal(t, privateKey.N.String(), publicKey.N.String())
}

func TestLoadRSAKeyPair_FailsOnInvalidPEM(t *testing.T) {
	dir := t.TempDir()
	privateFile := filepath.Join(dir, "private.pem")
	publicFile := filepath.Join(dir, "public.pem")
	require.NoError(t, os.WriteFile(privateFile, []byte("not pem"), 0600))
	require.NoError(t, os.WriteFile(publicFile, []byte("not pem"), 0644))

	privateKey, publicKey, err := loadRSAKeyPair(privateFile, publicFile)
	assert.Error(t, err)
	assert.Nil(t, privateKey)
	assert.Nil(t, publicKey)
}
