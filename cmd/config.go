package main

import (
	"fmt"
	"strings"

	featureflagsconfig "github.com/billykore/project-one/internal/featureflags/config"
	identityconfig "github.com/billykore/project-one/internal/identity/config"
	operationsconfig "github.com/billykore/project-one/internal/operations/config"
	pkgconfig "github.com/billykore/project-one/internal/pkg/config"
	"github.com/spf13/viper"
)

// Config holds all application configuration.
type Config struct {
	App           pkgconfig.AppConfig                   `mapstructure:"app"`
	Database      pkgconfig.DatabaseConfig              `mapstructure:"database"`
	JWT           identityconfig.JWTConfig              `mapstructure:"jwt"`
	MessageBroker pkgconfig.MessageBrokerConfig         `mapstructure:"message_broker"`
	FeatureFlags  featureflagsconfig.FeatureFlagsConfig `mapstructure:"feature_flags"`
	Monitoring    operationsconfig.MonitoringConfig     `mapstructure:"monitoring"`
}

// ponytail: uses viper (already-installed dep). BindEnv needed so AutomaticEnv
// knows which keys to check (it only looks up env vars for registered keys).
func loadConfig(path string) (*Config, error) {
	v := viper.New()
	v.AddConfigPath(path)
	v.SetConfigName("config")
	v.SetConfigType("yaml")

	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	v.SetDefault("app.port", 8080)
	v.SetDefault("feature_flags.environment", "local")
	v.SetDefault("feature_flags.refresh_interval", "30s")
	v.SetDefault("feature_flags.operators", []string{})
	for _, key := range []string{
		"app.port",
		"app.env",
		"database.host",
		"database.port",
		"database.user",
		"database.password",
		"database.dbname",
		"database.sslmode",
		"app.error_type_base_url",
		"jwt.private_key_path",
		"jwt.public_key_path",
		"jwt.expiration_time",
		"message_broker.type",
		"message_broker.rabbitmq.url",
		"message_broker.rabbitmq.exchange",
		"message_broker.rabbitmq.queue",
		"feature_flags.environment",
		"feature_flags.refresh_interval",
		"feature_flags.operators",
		"monitoring.username",
		"monitoring.password_file",
	} {
		_ = v.BindEnv(key)
	}

	err := v.ReadInConfig()
	if err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	if err := cfg.JWT.Validate(); err != nil {
		return nil, err
	}
	if err := cfg.Database.Validate(); err != nil {
		return nil, err
	}
	if err := cfg.MessageBroker.Validate(); err != nil {
		return nil, err
	}
	if err := cfg.FeatureFlags.Normalize(); err != nil {
		return nil, err
	}
	if err := cfg.Monitoring.Validate(); err != nil {
		return nil, err
	}

	return &cfg, nil
}
