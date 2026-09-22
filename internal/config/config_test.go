package config_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/billykore/project-one/internal/config"
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
		_ = os.RemoveAll(tempDir)
		viper.Reset()
	}
	return tempDir, cleanup
}

func TestLoad_SuccessFromFile(t *testing.T) {
	tempDir, cleanup := setupTestEnvironment(t)
	defer cleanup()

	yamlContent := []byte(`
database:
  host: "test_db_host"
  port: 5432
  user: "testuser"
  password: "testpassword"
  dbname: "test_dbname"
  sslmode: "disable"
jwt:
	private_key_path: "/tmp/test-private.pem"
	public_key_path: "/tmp/test-public.pem"
  expiration_time: 1h
app:
  port: 8080
message_broker:
  type: "rabbitmq"
  kafka:
    brokers:
      - "localhost:9092"
    topic_prefix: "project1"
    consumer_group: "notification-consumer"
    tls_enabled: false
  rabbitmq:
    url: "amqp://guest:guest@localhost:5672/"
    exchange: "project1.notifications"
    queue: "notifications"
feature_flags:
  environment: "staging"
  refresh_interval: "15s"
  operators:
    - "operator1"
    - "operator2"

`)
	yamlContent = []byte(strings.ReplaceAll(string(yamlContent), "\t", "  "))
	err := os.WriteFile(filepath.Join(tempDir, "config.yaml"), yamlContent, 0644)
	assert.NoError(t, err)

	cfg, err := config.Load(tempDir)
	assert.NoError(t, err)
	assert.NotNil(t, cfg)

	assert.Equal(t, "test_db_host", cfg.Database.Host)
	assert.Equal(t, 5432, cfg.Database.Port)
	assert.Equal(t, "testuser", cfg.Database.User)
	assert.Equal(t, "testpassword", cfg.Database.Password)
	assert.Equal(t, "test_dbname", cfg.Database.DBName)
	assert.Equal(t, "disable", cfg.Database.SSLMode)
	assert.Equal(t, "/tmp/test-private.pem", cfg.JWT.PrivateKeyPath)
	assert.Equal(t, "/tmp/test-public.pem", cfg.JWT.PublicKeyPath)
	assert.Equal(t, time.Hour, cfg.JWT.ExpirationTime)
	assert.Equal(t, 8080, cfg.App.Port)
	assert.Equal(t, "amqp://guest:guest@localhost:5672/", cfg.MessageBroker.RabbitMQ.URL)
	assert.Equal(t, "project1.notifications", cfg.MessageBroker.RabbitMQ.Exchange)
	assert.Equal(t, "notifications", cfg.MessageBroker.RabbitMQ.Queue)
	assert.Equal(t, "staging", cfg.FeatureFlags.Environment)
	assert.Equal(t, 15*time.Second, cfg.FeatureFlags.RefreshInterval)
	assert.Equal(t, []string{"operator1", "operator2"}, cfg.FeatureFlags.Operators)
}

func TestLoad_SuccessFromEnv(t *testing.T) {
	tempDir, cleanup := setupTestEnvironment(t)
	defer cleanup()

	t.Setenv("DATABASE_HOST", "env_db_host")
	t.Setenv("DATABASE_PORT", "1234")
	t.Setenv("DATABASE_USER", "envuser")
	t.Setenv("DATABASE_PASSWORD", "envpassword")
	t.Setenv("DATABASE_DBNAME", "env_dbname")
	t.Setenv("DATABASE_SSLMODE", "require")
	t.Setenv("JWT_PRIVATE_KEY_PATH", "/env/private.pem")
	t.Setenv("JWT_PUBLIC_KEY_PATH", "/env/public.pem")
	t.Setenv("JWT_EXPIRATION_TIME", "2h") // Viper can parse duration strings
	t.Setenv("APP_PORT", "9000")
	t.Setenv("MESSAGE_BROKER_TYPE", "rabbitmq")
	t.Setenv("MESSAGE_BROKER_RABBITMQ_URL", "amqp://guest:guest@localhost:5672/")
	t.Setenv("MESSAGE_BROKER_RABBITMQ_EXCHANGE", "project1.notifications")
	t.Setenv("MESSAGE_BROKER_RABBITMQ_QUEUE", "notifications")

	cfg, err := config.Load(tempDir)
	assert.NoError(t, err)
	assert.NotNil(t, cfg)

	assert.Equal(t, "env_db_host", cfg.Database.Host)
	assert.Equal(t, 1234, cfg.Database.Port)
	assert.Equal(t, "envuser", cfg.Database.User)
	assert.Equal(t, "envpassword", cfg.Database.Password)
	assert.Equal(t, "env_dbname", cfg.Database.DBName)
	assert.Equal(t, "require", cfg.Database.SSLMode)
	assert.Equal(t, "/env/private.pem", cfg.JWT.PrivateKeyPath)
	assert.Equal(t, "/env/public.pem", cfg.JWT.PublicKeyPath)
	assert.Equal(t, time.Hour*2, cfg.JWT.ExpirationTime)
	assert.Equal(t, 9000, cfg.App.Port)
}

func TestLoadConfig_EnvOverridesFile(t *testing.T) {
	tempDir, cleanup := setupTestEnvironment(t)
	defer cleanup()

	yamlContent := []byte(`
database:
  host: "file_db_host"
  port: 5432
  user: "fileuser"
  password: "filepassword"
  dbname: "file_dbname"
  sslmode: "disable"
jwt:
	private_key_path: "/file/private.pem"
	public_key_path: "/file/public.pem"
  expiration_time: 1h
app:
  port: 8080
`)
	yamlContent = []byte(strings.ReplaceAll(string(yamlContent), "\t", "  "))
	err := os.WriteFile(filepath.Join(tempDir, "config.yaml"), yamlContent, 0644)
	assert.NoError(t, err)

	// Set environment variables that should override file values
	t.Setenv("DATABASE_HOST", "env_override_host")
	t.Setenv("JWT_PRIVATE_KEY_PATH", "/env/override-private.pem")
	t.Setenv("JWT_PUBLIC_KEY_PATH", "/env/override-public.pem")
	t.Setenv("APP_PORT", "9001")
	t.Setenv("MESSAGE_BROKER_TYPE", "rabbitmq")
	t.Setenv("MESSAGE_BROKER_RABBITMQ_URL", "amqp://guest:guest@localhost:5672/")
	t.Setenv("MESSAGE_BROKER_RABBITMQ_EXCHANGE", "project1.notifications")
	t.Setenv("MESSAGE_BROKER_RABBITMQ_QUEUE", "notifications")

	cfg, err := config.Load(tempDir)
	assert.NoError(t, err)
	assert.NotNil(t, cfg)

	assert.Equal(t, "env_override_host", cfg.Database.Host)
	assert.Equal(t, "/env/override-private.pem", cfg.JWT.PrivateKeyPath)
	assert.Equal(t, "/env/override-public.pem", cfg.JWT.PublicKeyPath)
	assert.Equal(t, 5432, cfg.Database.Port)
	assert.Equal(t, time.Hour, cfg.JWT.ExpirationTime)
	assert.Equal(t, 9001, cfg.App.Port)
}

func TestLoadConfig_MissingJWTPrivateKeyPath(t *testing.T) {
	tempDir, cleanup := setupTestEnvironment(t)
	defer cleanup()

	yamlContent := []byte(`
database:
  host: "test_db_host"
  user: "testuser"
  dbname: "test_dbname"
jwt:
  public_key_path: "/tmp/test-public.pem"
  expiration_time: 1h
`)
	yamlContent = []byte(strings.ReplaceAll(string(yamlContent), "\t", "  "))
	err := os.WriteFile(filepath.Join(tempDir, "config.yaml"), yamlContent, 0644)
	assert.NoError(t, err)

	cfg, err := config.Load(tempDir)
	assert.Error(t, err)
	assert.Nil(t, cfg)
	assert.Contains(t, err.Error(), "JWT private key path cannot be empty")
}

func TestLoadConfig_MissingJWTPublicKeyPath(t *testing.T) {
	tempDir, cleanup := setupTestEnvironment(t)
	defer cleanup()

	yamlContent := []byte(`
database:
  host: "test_db_host"
  user: "testuser"
  dbname: "test_dbname"
jwt:
  private_key_path: "/tmp/test-private.pem"
  expiration_time: 1h
`)
	yamlContent = []byte(strings.ReplaceAll(string(yamlContent), "\t", "  "))
	err := os.WriteFile(filepath.Join(tempDir, "config.yaml"), yamlContent, 0644)
	assert.NoError(t, err)

	cfg, err := config.Load(tempDir)
	assert.Error(t, err)
	assert.Nil(t, cfg)
	assert.Contains(t, err.Error(), "JWT public key path cannot be empty")
}

func TestLoad_Success_MissingDatabaseConfig(t *testing.T) {
	tempDir, cleanup := setupTestEnvironment(t)
	defer cleanup()

	yamlContent := []byte(`
database:
  # host is missing
  user: "testuser"
  dbname: "test_dbname"
jwt:
	private_key_path: "/tmp/test-private.pem"
	public_key_path: "/tmp/test-public.pem"
  expiration_time: 1h
`)
	yamlContent = []byte(strings.ReplaceAll(string(yamlContent), "\t", "  "))
	err := os.WriteFile(filepath.Join(tempDir, "config.yaml"), yamlContent, 0644)
	assert.NoError(t, err)

	cfg, err := config.Load(tempDir)
	assert.Error(t, err)
	assert.Nil(t, cfg)
	assert.Contains(t, err.Error(), "database host, user, and dbname cannot be empty")
}

func TestLoad_Success_InvalidYaml(t *testing.T) {
	tempDir, cleanup := setupTestEnvironment(t)
	defer cleanup()

	yamlContent := []byte(`
database:
  host: "test_db_host"
  port: "invalid_port" # This should cause an error
jwt:
  private_key_path: "/tmp/test-private.pem"
  public_key_path: "/tmp/test-public.pem"
`)
	yamlContent = []byte(strings.ReplaceAll(string(yamlContent), "\t", "  "))
	err := os.WriteFile(filepath.Join(tempDir, "config.yaml"), yamlContent, 0644)
	assert.NoError(t, err)

	cfg, err := config.Load(tempDir)
	assert.Error(t, err)
	assert.Nil(t, cfg)
	assert.Contains(t, err.Error(), "failed to unmarshal config")
}

func TestLoad_Success_ConfigFileNotFoundAndNoEnvFallback(t *testing.T) {
	tempDir, cleanup := setupTestEnvironment(t)
	defer cleanup()

	// No user_config.yaml file and no relevant env vars
	cfg, err := config.Load(tempDir)
	assert.Error(t, err)
	assert.Nil(t, cfg)
	assert.Contains(t, err.Error(), "JWT private key path cannot be empty")
}

func TestLoad_FeatureFlagsFromEnv(t *testing.T) {
	tempDir, cleanup := setupTestEnvironment(t)
	defer cleanup()

	yamlContent := []byte(`
database:
  host: "test_db_host"
  user: "testuser"
  dbname: "test_dbname"
jwt:
  private_key_path: "/tmp/test-private.pem"
  public_key_path: "/tmp/test-public.pem"
  expiration_time: 1h
message_broker:
  type: "rabbitmq"
  rabbitmq:
    url: "amqp://guest:guest@localhost:5672/"
    exchange: "project1.notifications"
    queue: "notifications"
`)
	yamlContent = []byte(strings.ReplaceAll(string(yamlContent), "\t", "  "))
	err := os.WriteFile(filepath.Join(tempDir, "config.yaml"), yamlContent, 0644)
	assert.NoError(t, err)

	t.Setenv("FEATURE_FLAGS_ENVIRONMENT", "production")
	t.Setenv("FEATURE_FLAGS_REFRESH_INTERVAL", "45s")
	t.Setenv("FEATURE_FLAGS_OPERATORS", "op1,op2")

	cfg, err := config.Load(tempDir)
	assert.NoError(t, err)
	assert.Equal(t, "production", cfg.FeatureFlags.Environment)
	assert.Equal(t, 45*time.Second, cfg.FeatureFlags.RefreshInterval)
	assert.Equal(t, []string{"op1", "op2"}, cfg.FeatureFlags.Operators)
}

func TestLoad_InvalidFeatureFlagsEnvironment(t *testing.T) {
	tempDir, cleanup := setupTestEnvironment(t)
	defer cleanup()

	yamlContent := []byte(`
database:
  host: "test_db_host"
  user: "testuser"
  dbname: "test_dbname"
jwt:
  private_key_path: "/tmp/test-private.pem"
  public_key_path: "/tmp/test-public.pem"
  expiration_time: 1h
message_broker:
  type: "rabbitmq"
  rabbitmq:
    url: "amqp://guest:guest@localhost:5672/"
    exchange: "project1.notifications"
    queue: "notifications"
feature_flags:
  environment: "outer-space"
`)
	yamlContent = []byte(strings.ReplaceAll(string(yamlContent), "\t", "  "))
	err := os.WriteFile(filepath.Join(tempDir, "config.yaml"), yamlContent, 0644)
	assert.NoError(t, err)

	cfg, err := config.Load(tempDir)
	assert.Error(t, err)
	assert.Nil(t, cfg)
	assert.Contains(t, err.Error(), "unsupported feature flags environment")
}

// baseConfigWithoutMonitoring is the smallest valid configuration; monitoring
// credentials are optional so every deployment that does not scrape metrics
// still starts.
const baseConfigWithoutMonitoring = `
database:
  host: "test_db_host"
  user: "testuser"
  dbname: "test_dbname"
jwt:
  private_key_path: "/tmp/test-private.pem"
  public_key_path: "/tmp/test-public.pem"
  expiration_time: 1h
message_broker:
  type: "rabbitmq"
  rabbitmq:
    url: "amqp://guest:guest@localhost:5672/"
    exchange: "project1.notifications"
    queue: "notifications"
`

// writeConfig writes a config file into a fresh temp directory and returns the directory.
func writeConfig(t *testing.T, content string) (string, func()) {
	t.Helper()

	tempDir, cleanup := setupTestEnvironment(t)
	err := os.WriteFile(filepath.Join(tempDir, "config.yaml"), []byte(content), 0o600)
	assert.NoError(t, err)
	return tempDir, cleanup
}

func TestLoad_MonitoringCredentialsFromFile(t *testing.T) {
	secretPath := filepath.Join(t.TempDir(), "metrics-password")
	assert.NoError(t, os.WriteFile(secretPath, []byte("local-monitoring-secret\n"), 0o600))

	tempDir, cleanup := writeConfig(t, baseConfigWithoutMonitoring+`
monitoring:
  username: "projectone-metrics"
  password_file: "`+secretPath+`"
`)
	defer cleanup()

	cfg, err := config.Load(tempDir)
	assert.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.Equal(t, "projectone-metrics", cfg.Monitoring.Username)
	assert.Equal(t, secretPath, cfg.Monitoring.PasswordFile)
}

func TestLoad_MonitoringCredentialsFromEnv(t *testing.T) {
	secretPath := filepath.Join(t.TempDir(), "metrics-password")
	assert.NoError(t, os.WriteFile(secretPath, []byte("env-monitoring-secret"), 0o600))

	tempDir, cleanup := writeConfig(t, baseConfigWithoutMonitoring)
	defer cleanup()

	t.Setenv("MONITORING_USERNAME", "env-metrics-user")
	t.Setenv("MONITORING_PASSWORD_FILE", secretPath)

	cfg, err := config.Load(tempDir)
	assert.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.Equal(t, "env-metrics-user", cfg.Monitoring.Username)
	assert.Equal(t, secretPath, cfg.Monitoring.PasswordFile)
}

func TestLoad_MonitoringDisabledWhenUnconfigured(t *testing.T) {
	tempDir, cleanup := writeConfig(t, baseConfigWithoutMonitoring)
	defer cleanup()

	cfg, err := config.Load(tempDir)
	assert.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.Empty(t, cfg.Monitoring.Username)
	assert.Empty(t, cfg.Monitoring.PasswordFile)
}

func TestLoad_MonitoringRejectsMissingPasswordFile(t *testing.T) {
	tempDir, cleanup := writeConfig(t, baseConfigWithoutMonitoring+`
monitoring:
  username: "projectone-metrics"
  password_file: "/nonexistent/metrics-password"
`)
	defer cleanup()

	cfg, err := config.Load(tempDir)
	assert.Error(t, err)
	assert.Nil(t, cfg)
	assert.Contains(t, err.Error(), "monitoring password file")
}

func TestLoad_MonitoringRejectsEmptyPasswordFile(t *testing.T) {
	secretPath := filepath.Join(t.TempDir(), "metrics-password")
	assert.NoError(t, os.WriteFile(secretPath, []byte("\n"), 0o600))

	tempDir, cleanup := writeConfig(t, baseConfigWithoutMonitoring+`
monitoring:
  username: "projectone-metrics"
  password_file: "`+secretPath+`"
`)
	defer cleanup()

	cfg, err := config.Load(tempDir)
	assert.Error(t, err)
	assert.Nil(t, cfg)
	assert.Contains(t, err.Error(), "monitoring password file")
}

func TestLoad_MonitoringRequiresUsernameAndPasswordFileTogether(t *testing.T) {
	tempDir, cleanup := writeConfig(t, baseConfigWithoutMonitoring+`
monitoring:
  username: "projectone-metrics"
`)
	defer cleanup()

	cfg, err := config.Load(tempDir)
	assert.Error(t, err)
	assert.Nil(t, cfg)
	assert.Contains(t, err.Error(), "monitoring username and password_file")

	secretPath := filepath.Join(t.TempDir(), "metrics-password")
	assert.NoError(t, os.WriteFile(secretPath, []byte("secret"), 0o600))

	tempDirPeer, cleanupPeer := writeConfig(t, baseConfigWithoutMonitoring+`
monitoring:
  password_file: "`+secretPath+`"
`)
	defer cleanupPeer()

	peerCfg, peerErr := config.Load(tempDirPeer)
	assert.Error(t, peerErr)
	assert.Nil(t, peerCfg)
	assert.Contains(t, peerErr.Error(), "monitoring username and password_file")
}
