package config_test

import (
	"fmt" // Added import
	"os"
	"path/filepath"
	"strings" // Added for string assertion in last test
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

	// Create a mock user_config.yaml file
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
app:
  port: 8080
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
	assert.Equal(t, 8080, cfg.App.Port)
}

func TestLoadConfig_SuccessFromEnv(t *testing.T) {
	tempDir, cleanup := setupTestEnvironment(t)
	defer cleanup()

	// Set environment variables
	t.Setenv("DATABASE_HOST", "env_db_host")
	t.Setenv("DATABASE_PORT", "1234")
	t.Setenv("DATABASE_USER", "envuser")
	t.Setenv("DATABASE_PASSWORD", "envpassword")
	t.Setenv("DATABASE_DBNAME", "env_dbname")
	t.Setenv("DATABASE_SSLMODE", "require")
	t.Setenv("JWT_SECRET_KEY", "env_secret_key")
	t.Setenv("JWT_EXPIRATION_TIME", "2h") // Viper can parse duration strings
	t.Setenv("APP_PORT", "9000")

	// Debugging: Print environment variable to verify t.Setenv
	fmt.Printf("DEBUG: JWT_SECRET_KEY from env: %s\n", os.Getenv("JWT_SECRET_KEY"))

	cfg, err := config.LoadConfig(tempDir) // No user_config.yaml file, should rely on env
	assert.NoError(t, err)
	assert.NotNil(t, cfg)

	assert.Equal(t, "env_db_host", cfg.Database.Host)
	assert.Equal(t, 1234, cfg.Database.Port)
	assert.Equal(t, "envuser", cfg.Database.User)
	assert.Equal(t, "envpassword", cfg.Database.Password)
	assert.Equal(t, "env_dbname", cfg.Database.DBName)
	assert.Equal(t, "require", cfg.Database.SSLMode)
	assert.Equal(t, "env_secret_key", cfg.JWT.SecretKey)
	assert.Equal(t, time.Hour*2, cfg.JWT.ExpirationTime)
	assert.Equal(t, 9000, cfg.App.Port)
}

func TestLoadConfig_EnvOverridesFile(t *testing.T) {
	tempDir, cleanup := setupTestEnvironment(t)
	defer cleanup()

	// Create a mock user_config.yaml file
	yamlContent := []byte(`
database:
  host: "file_db_host"
  port: 5432
  user: "fileuser"
  password: "filepassword"
  dbname: "file_dbname"
  sslmode: "disable"
jwt:
  secret_key: "file_secret_key"
  expiration_time: 1h
app:
  port: 8080
`)
	err := os.WriteFile(filepath.Join(tempDir, "user_config.yaml"), yamlContent, 0644)
	assert.NoError(t, err)

	// Set environment variables that should override file values
	t.Setenv("DATABASE_HOST", "env_override_host")
	t.Setenv("JWT_SECRET_KEY", "env_override_secret")
	t.Setenv("APP_PORT", "9001")

	cfg, err := config.LoadConfig(tempDir)
	assert.NoError(t, err)
	assert.NotNil(t, cfg)

	assert.Equal(t, "env_override_host", cfg.Database.Host)
	assert.Equal(t, "env_override_secret", cfg.JWT.SecretKey)
	assert.Equal(t, 5432, cfg.Database.Port)
	assert.Equal(t, time.Hour, cfg.JWT.ExpirationTime)
	assert.Equal(t, 9001, cfg.App.Port)
}

func TestLoadConfig_MissingJWTSecretKey(t *testing.T) {
	tempDir, cleanup := setupTestEnvironment(t)
	defer cleanup()

	// Create a mock user_config.yaml file with missing secret_key
	yamlContent := []byte(`
database:
  host: "test_db_host"
  user: "testuser"
  dbname: "test_dbname"
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

func TestLoadConfig_MissingDatabaseConfig(t *testing.T) {
	tempDir, cleanup := setupTestEnvironment(t)
	defer cleanup()

	// Create a mock user_config.yaml file with missing database host
	yamlContent := []byte(`
database:
  # host is missing
  user: "testuser"
  dbname: "test_dbname"
jwt:
  secret_key: "test_secret_key"
  expiration_time: 1h
`)
	err := os.WriteFile(filepath.Join(tempDir, "user_config.yaml"), yamlContent, 0644)
	assert.NoError(t, err)

	cfg, err := config.LoadConfig(tempDir)
	assert.Error(t, err)
	assert.Nil(t, cfg)
	assert.Contains(t, err.Error(), "database host, user, and dbname cannot be empty")
}

func TestLoadConfig_InvalidYaml(t *testing.T) {
	tempDir, cleanup := setupTestEnvironment(t)
	defer cleanup()

	// Create an invalid user_config.yaml file
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

	// No user_config.yaml file and no relevant env vars
	cfg, err := config.LoadConfig(tempDir)
	assert.Error(t, err)
	assert.Nil(t, cfg)
	// Now it can fail for JWT secret or database config
	assert.True(t, strings.Contains(err.Error(), "JWT secret key cannot be empty") || strings.Contains(err.Error(), "database host, user, and dbname cannot be empty"))
}
