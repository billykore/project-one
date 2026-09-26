// Package config owns feature-flags runtime settings.
package config

import "time"

// FeatureFlagsConfig controls flag evaluation and administration for this
// deployment environment.
type FeatureFlagsConfig struct {
	Environment     string        `mapstructure:"environment"`
	RefreshInterval time.Duration `mapstructure:"refresh_interval"`
	Operators       []string      `mapstructure:"operators"`
}
