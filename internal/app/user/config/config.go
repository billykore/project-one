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
func LoadConfig(path string) (cfg *Config, err error) { // Changed named return to cfg
	v := viper.New() // Use a new Viper instance for each call
	v.AddConfigPath(path)
	v.SetConfigName("user_config")
	v.SetConfigType("yaml")

	// Set EnvKeyReplacer to automatically map env vars like DATABASE_HOST to database.host
	replacer := strings.NewReplacer(".", "_")
	v.SetEnvKeyReplacer(replacer)
	v.AutomaticEnv() // read in environment variables that match

	// Explicitly bind environment variables for all fields to ensure they are picked up reliably
	_ = v.BindEnv("app.port", "APP_PORT")
	_ = v.BindEnv("database.host", "DATABASE_HOST")
	_ = v.BindEnv("database.port", "DATABASE_PORT")
	_ = v.BindEnv("database.user", "DATABASE_USER")
	_ = v.BindEnv("database.password", "DATABASE_PASSWORD")
	_ = v.BindEnv("database.dbname", "DATABASE_DBNAME")
	_ = v.BindEnv("database.sslmode", "DATABASE_SSLMODE")
	_ = v.BindEnv("jwt.secret_key", "JWT_SECRET_KEY")
	_ = v.BindEnv("jwt.expiration_time", "JWT_EXPIRATION_TIME")

	// Set default values
	v.SetDefault("app.port", 8080) // Added default for app.port

	err = v.ReadInConfig()
	if err != nil {
		// If config file not found, but env vars are present, it's okay.
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}

	var loadedConfig Config // Changed variable name
	err = v.Unmarshal(&loadedConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Basic validation for critical fields
	if loadedConfig.JWT.SecretKey == "" {
		return nil, fmt.Errorf("JWT secret key cannot be empty")
	}
	if loadedConfig.Database.Host == "" || loadedConfig.Database.User == "" || loadedConfig.Database.DBName == "" { // Added database validation
		return nil, fmt.Errorf("database host, user, and dbname cannot be empty")
	}

	return &loadedConfig, nil // Return the renamed variable
}
