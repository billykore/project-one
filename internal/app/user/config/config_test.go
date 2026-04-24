package config_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/billykore/project-one/internal/app/user/config" // Alias the package to avoid name collision with the test package
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

// setupTestEnvironment creates a temporary directory and sets up Viper for testing.
func setupTestEnvironment(t *testing.T) (string, func()) {
	// Create a temporary directory for config files
	tempDir, err := os.MkdirTemp("", "config_test")
	assert.NoError(t, err)

	// Reset Viper before each test to ensure a clean state
	viper.Reset()

	// Cleanup function
	cleanup := func() {
		_ = os.RemoveAll(tempDir) // Clean up the temporary directory
		viper.Reset()             // Reset Viper again after tests
	}
	return tempDir, cleanup
}

func TestLoadConfig_SuccessFromFile(t *testing.T) {
	tempDir, cleanup := setupTestEnvironment(t)
	defer cleanup()

	// Create a mock config.yaml file
	yamlContent := []byte(`
database:
  host: "test_db_host"
  port: 5432
  user: "testuser"
  password: "testpassword"
  dbname: "test_dbname"
  sslmode: "disable"
jwt:
  secret_key: "test_secret_key"
  expiration_time: 1h
`)
	err := os.WriteFile(filepath.Join(tempDir, "user_config.yaml"), yamlContent, 0644)
	assert.NoError(t, err)

	cfg, err := config.LoadConfig(tempDir)
	assert.NoError(t, err)
	assert.NotNil(t, cfg)

	assert.Equal(t, "test_db_host", cfg.Database.Host)
	assert.Equal(t, 5432, cfg.Database.Port)
	assert.Equal(t, "testuser", cfg.Database.User)
	assert.Equal(t, "testpassword", cfg.Database.Password)
	assert.Equal(t, "test_dbname", cfg.Database.DBName)
	assert.Equal(t, "disable", cfg.Database.SSLMode)
	assert.Equal(t, "test_secret_key", cfg.JWT.SecretKey)
	assert.Equal(t, time.Hour, cfg.JWT.ExpirationTime)
}

func TestLoadConfig_MissingJWTSecretKey(t *testing.T) {
	tempDir, cleanup := setupTestEnvironment(t)
	defer cleanup()

	// Create a mock config.yaml file with missing secret_key
	yamlContent := []byte(`
database:
  host: "test_db_host"
jwt:
  # secret_key is missing
  expiration_time: 1h
`)
	err := os.WriteFile(filepath.Join(tempDir, "user_config.yaml"), yamlContent, 0644)
	assert.NoError(t, err)

	cfg, err := config.LoadConfig(tempDir)
	assert.Error(t, err)
	assert.Nil(t, cfg)
	assert.Contains(t, err.Error(), "JWT secret key cannot be empty")
}

func TestLoadConfig_InvalidYaml(t *testing.T) {
	tempDir, cleanup := setupTestEnvironment(t)
	defer cleanup()

	// Create an invalid config.yaml file
	yamlContent := []byte(`
database:
  host: "test_db_host"
  port: "invalid_port" # This should cause an error
jwt:
  secret_key: "test_secret_key"
`)
	err := os.WriteFile(filepath.Join(tempDir, "user_config.yaml"), yamlContent, 0644)
	assert.NoError(t, err)

	cfg, err := config.LoadConfig(tempDir)
	assert.Error(t, err)
	assert.Nil(t, cfg)
	assert.Contains(t, err.Error(), "failed to unmarshal config")
}

func TestLoadConfig_ConfigFileNotFoundAndNoEnvFallback(t *testing.T) {
	tempDir, cleanup := setupTestEnvironment(t)
	defer cleanup()

	// No config.yaml file and no relevant env vars
	cfg, err := config.LoadConfig(tempDir)
	assert.Error(t, err) // Should error because JWT_SECRET_KEY is empty and no file provides it
	assert.Nil(t, cfg)
	assert.Contains(t, err.Error(), "JWT secret key cannot be empty")
}
