package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config holds all application configuration.
type Config struct {
	App      AppConfig      `mapstructure:"app"`
	Database DatabaseConfig `mapstructure:"database"`
	JWT      JWTConfig      `mapstructure:"jwt"`
}

// AppConfig holds application-level settings.
type AppConfig struct {
	Port int    `mapstructure:"port"`
	Env  string `mapstructure:"env"`
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
	for _, key := range []string{
		"app.port",
		"app.env",
		"database.host",
		"database.port",
		"database.user",
		"database.password",
		"database.dbname",
		"database.sslmode",
		"jwt.private_key_path",
		"jwt.public_key_path",
		"jwt.expiration_time",
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

	return &cfg, nil
}
