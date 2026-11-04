// Package config owns identity runtime settings.
package config

import (
	"errors"
	"time"
)

// JWTConfig holds the key material locations and session lifetime used by
// identity's token adapter.
type JWTConfig struct {
	PrivateKeyPath string        `mapstructure:"private_key_path"`
	PublicKeyPath  string        `mapstructure:"public_key_path"`
	ExpirationTime time.Duration `mapstructure:"expiration_time"`
}

// Validate reports a missing key path before identity tries to load the keys.
func (c JWTConfig) Validate() error {
	if c.PrivateKeyPath == "" {
		return errors.New("JWT private key path cannot be empty")
	}
	if c.PublicKeyPath == "" {
		return errors.New("JWT public key path cannot be empty")
	}
	return nil
}
