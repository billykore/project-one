// Package config owns identity runtime settings.
package config

import "time"

// JWTConfig holds the key material locations and session lifetime used by
// identity's token adapter.
type JWTConfig struct {
	PrivateKeyPath string        `mapstructure:"private_key_path"`
	PublicKeyPath  string        `mapstructure:"public_key_path"`
	ExpirationTime time.Duration `mapstructure:"expiration_time"`
}
