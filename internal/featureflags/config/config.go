// Package config owns feature-flags runtime settings.
package config

import (
	"fmt"
	"time"
)

// FeatureFlagsConfig controls flag evaluation and administration for this
// deployment environment.
type FeatureFlagsConfig struct {
	Environment     string        `mapstructure:"environment"`
	RefreshInterval time.Duration `mapstructure:"refresh_interval"`
	Operators       []string      `mapstructure:"operators"`
}

// DefaultRefreshInterval is used when no positive refresh interval is configured.
const DefaultRefreshInterval = 30 * time.Second

// Normalize validates the evaluation environment and fills in a missing
// refresh interval.
func (c *FeatureFlagsConfig) Normalize() error {
	switch c.Environment {
	case "local", "test", "staging", "production":
	default:
		return fmt.Errorf("unsupported feature flags environment: %s", c.Environment)
	}
	if c.RefreshInterval <= 0 {
		c.RefreshInterval = DefaultRefreshInterval
	}
	return nil
}
