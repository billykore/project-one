// Package config owns operations runtime settings.
package config

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

// MonitoringConfig holds the machine credential used to scrape operational
// metrics. It is unrelated to application-user authentication.
type MonitoringConfig struct {
	Username     string `mapstructure:"username"`
	PasswordFile string `mapstructure:"password_file"`
}

// Validate fails fast when the monitoring credential is half configured or
// points at a missing or empty secret file. Leaving both fields empty is valid
// and disables metrics scraping.
func (c MonitoringConfig) Validate() error {
	if (c.Username == "") != (c.PasswordFile == "") {
		return errors.New("monitoring username and password_file must be configured together")
	}
	if c.PasswordFile == "" {
		return nil
	}

	secret, err := os.ReadFile(c.PasswordFile)
	if err != nil {
		return fmt.Errorf("failed to read monitoring password file: %w", err)
	}
	if strings.TrimSpace(string(secret)) == "" {
		return fmt.Errorf("monitoring password file %s is empty", c.PasswordFile)
	}

	return nil
}
