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

// LoginRateLimitConfig bounds login attempts per account email. Setting
// MaxAttempts or Window to zero disables the limit.
type LoginRateLimitConfig struct {
	MaxAttempts int           `mapstructure:"max_attempts"`
	Window      time.Duration `mapstructure:"window"`
}

// Validate rejects negative limits, which are most likely configuration typos.
func (c LoginRateLimitConfig) Validate() error {
	if c.MaxAttempts < 0 {
		return errors.New("login rate limit max attempts cannot be negative")
	}
	if c.Window < 0 {
		return errors.New("login rate limit window cannot be negative")
	}
	return nil
}
