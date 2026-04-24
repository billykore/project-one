package config

import (
	"fmt"
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
	Port int `mapstructure:"port"`
}

// DatabaseConfig holds database connection settings.
type DatabaseConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	DBName   string `mapstructure:"dbname"`
	SSLMode  string `mapstructure:"sslmode"`
}

// JWTConfig holds JWT-related settings.
type JWTConfig struct {
	SecretKey      string        `mapstructure:"secret_key"`
	ExpirationTime time.Duration `mapstructure:"expiration_time"`
}

// LoadConfig reads configuration from file or environment variables.
func LoadConfig(path string) (config *Config, err error) {
	v := viper.New() // Use a new Viper instance for each call
	v.AddConfigPath(path)
	v.SetConfigName("user_config")
	v.SetConfigType("yaml")
	v.AutomaticEnv() // read in environment variables that match

	err = v.ReadInConfig()
	if err != nil {
		// If config file not found, but env vars are present, it's okay.
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}

	err = v.Unmarshal(&config)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Basic validation for critical fields
	if config.JWT.SecretKey == "" {
		return nil, fmt.Errorf("JWT secret key cannot be empty")
	}

	return config, nil
}
