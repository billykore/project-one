package config

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config holds all application configuration.
type Config struct {
	App           AppConfig           `mapstructure:"app"`
	Database      DatabaseConfig      `mapstructure:"database"`
	JWT           JWTConfig           `mapstructure:"jwt"`
	MessageBroker MessageBrokerConfig `mapstructure:"message_broker"`
	FeatureFlags  FeatureFlagsConfig  `mapstructure:"feature_flags"`
	Monitoring    MonitoringConfig    `mapstructure:"monitoring"`
}

// AppConfig holds application-level settings.
type AppConfig struct {
	Port             int    `mapstructure:"port"`
	Env              string `mapstructure:"env"`
	ErrorTypeBaseURL string `mapstructure:"error_type_base_url"`
}

// DatabaseConfig holds database connection settings.
type DatabaseConfig struct {
	Host            string        `mapstructure:"host"`
	Port            int           `mapstructure:"port"`
	User            string        `mapstructure:"user"`
	Password        string        `mapstructure:"password"`
	DBName          string        `mapstructure:"dbname"`
	SSLMode         string        `mapstructure:"sslmode"`
	MaxIdleConns    int           `mapstructure:"max_idle_conns"`
	MaxOpenConns    int           `mapstructure:"max_open_conns"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"`
}

// JWTConfig holds JWT-related settings.
type JWTConfig struct {
	PrivateKeyPath string        `mapstructure:"private_key_path"`
	PublicKeyPath  string        `mapstructure:"public_key_path"`
	ExpirationTime time.Duration `mapstructure:"expiration_time"`
}

// MessageBrokerConfig holds the message broker selection and configuration.
type MessageBrokerConfig struct {
	Type     string               `mapstructure:"type"`
	Kafka    KafkaBrokerConfig    `mapstructure:"kafka"`
	RabbitMQ RabbitMQBrokerConfig `mapstructure:"rabbitmq"`
}

// KafkaBrokerConfig holds Kafka-specific connection settings.
type KafkaBrokerConfig struct {
	Brokers       []string `mapstructure:"brokers"`
	TopicPrefix   string   `mapstructure:"topic_prefix"`
	ConsumerGroup string   `mapstructure:"consumer_group"`
	TLSEnabled    bool     `mapstructure:"tls_enabled"`
}

// RabbitMQBrokerConfig holds RabbitMQ-specific connection settings.
type RabbitMQBrokerConfig struct {
	URL      string `mapstructure:"url"`
	Exchange string `mapstructure:"exchange"`
	Queue    string `mapstructure:"queue"`
}

// FeatureFlagsConfig holds runtime feature-flag settings.
type FeatureFlagsConfig struct {
	// Environment is the deployment stage the server is running in:
	// local, test, staging, or production.
	Environment string `mapstructure:"environment"`
	// RefreshInterval is how often the evaluator reloads flag state.
	RefreshInterval time.Duration `mapstructure:"refresh_interval"`
	// Operators is the allowlist of usernames permitted to administer flags.
	Operators []string `mapstructure:"operators"`
}

// MonitoringConfig holds the dedicated machine credentials that Prometheus uses
// to scrape /metrics. It is deliberately separate from user authentication.
// Both fields are empty when scraping is not configured, and /metrics then
// rejects every request rather than exposing measurements.
type MonitoringConfig struct {
	// Username is the fixed machine identity; it is not an application account.
	Username string `mapstructure:"username"`
	// PasswordFile is the path of the mounted secret holding the matching
	// password. The secret content is never logged, returned, or committed.
	PasswordFile string `mapstructure:"password_file"`
}

// ponytail: uses viper (already-installed dep). BindEnv needed so AutomaticEnv
// knows which keys to check (it only looks up env vars for registered keys).
func Load(path string) (*Config, error) {
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
		"message_broker.kafka.brokers",
		"message_broker.kafka.topic_prefix",
		"message_broker.kafka.consumer_group",
		"message_broker.kafka.tls_enabled",
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

	if cfg.JWT.PrivateKeyPath == "" {
		return nil, fmt.Errorf("JWT private key path cannot be empty")
	}
	if cfg.JWT.PublicKeyPath == "" {
		return nil, fmt.Errorf("JWT public key path cannot be empty")
	}
	if cfg.Database.Host == "" || cfg.Database.User == "" || cfg.Database.DBName == "" {
		return nil, fmt.Errorf("database host, user, and dbname cannot be empty")
	}
	if cfg.MessageBroker.Type != "kafka" && cfg.MessageBroker.Type != "rabbitmq" {
		return nil, fmt.Errorf("unsupported message broker type: %s", cfg.MessageBroker.Type)
	}
	switch cfg.FeatureFlags.Environment {
	case "local", "test", "staging", "production":
	default:
		return nil, fmt.Errorf("unsupported feature flags environment: %s", cfg.FeatureFlags.Environment)
	}
	if cfg.FeatureFlags.RefreshInterval <= 0 {
		cfg.FeatureFlags.RefreshInterval = 30 * time.Second
	}
	if err := validateMonitoring(cfg.Monitoring); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// validateMonitoring fails fast when the monitoring credential is half
// configured or points at a missing or empty secret file. Leaving both fields
// empty is valid and disables metrics scraping.
func validateMonitoring(monitoring MonitoringConfig) error {
	if (monitoring.Username == "") != (monitoring.PasswordFile == "") {
		return fmt.Errorf("monitoring username and password_file must be configured together")
	}
	if monitoring.PasswordFile == "" {
		return nil
	}

	secret, err := os.ReadFile(monitoring.PasswordFile)
	if err != nil {
		return fmt.Errorf("failed to read monitoring password file: %w", err)
	}
	if strings.TrimSpace(string(secret)) == "" {
		return fmt.Errorf("monitoring password file %s is empty", monitoring.PasswordFile)
	}

	return nil
}
