// Package config holds the infrastructure settings shared by the composition
// root and technical adapters. Context-specific settings live in each bounded
// context's own config package.
package config

import (
	"errors"
	"fmt"
	"time"
)

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

// Validate reports missing connection settings.
func (c DatabaseConfig) Validate() error {
	if c.Host == "" || c.User == "" || c.DBName == "" {
		return errors.New("database host, user, and dbname cannot be empty")
	}
	return nil
}

// Validate reports an unsupported broker selection.
func (c MessageBrokerConfig) Validate() error {
	if c.Type != "kafka" && c.Type != "rabbitmq" {
		return fmt.Errorf("unsupported message broker type: %s", c.Type)
	}
	return nil
}
