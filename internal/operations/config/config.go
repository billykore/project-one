// Package config owns operations runtime settings.
package config

// MonitoringConfig holds the machine credential used to scrape operational
// metrics. It is unrelated to application-user authentication.
type MonitoringConfig struct {
	Username     string `mapstructure:"username"`
	PasswordFile string `mapstructure:"password_file"`
}
