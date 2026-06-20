package main

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
)

func loadRSAKeyPair(privateKeyPath, publicKeyPath string) (*rsa.PrivateKey, *rsa.PublicKey, error) {
	privateKeyPEM, err := os.ReadFile(privateKeyPath)
	if err != nil {
		return nil, nil, fmt.Errorf("read private key: %w", err)
	}

	publicKeyPEM, err := os.ReadFile(publicKeyPath)
	if err != nil {
		return nil, nil, fmt.Errorf("read public key: %w", err)
	}

	privateKeyBlock, _ := pem.Decode(privateKeyPEM)
	if privateKeyBlock == nil {
		return nil, nil, fmt.Errorf("decode private key pem")
	}

	privateKey, err := x509.ParsePKCS1PrivateKey(privateKeyBlock.Bytes)
	if err != nil {
		parsedKey, parseErr := x509.ParsePKCS8PrivateKey(privateKeyBlock.Bytes)
		if parseErr != nil {
			return nil, nil, fmt.Errorf("parse private key: %w", err)
		}
		rsaPrivateKey, ok := parsedKey.(*rsa.PrivateKey)
		if !ok {
			return nil, nil, fmt.Errorf("parse private key: not rsa")
		}
		privateKey = rsaPrivateKey
	}

	publicKeyBlock, _ := pem.Decode(publicKeyPEM)
	if publicKeyBlock == nil {
		return nil, nil, fmt.Errorf("decode public key pem")
	}

	parsedPublicKey, err := x509.ParsePKIXPublicKey(publicKeyBlock.Bytes)
	if err != nil {
		return nil, nil, fmt.Errorf("parse public key: %w", err)
	}

	rsaPublicKey, ok := parsedPublicKey.(*rsa.PublicKey)
	if !ok {
		return nil, nil, fmt.Errorf("parse public key: not rsa")
	}

	return privateKey, rsaPublicKey, nil
}
