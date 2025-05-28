package config

import (
	"os"
	"testing"
	"time"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestDBConfigDSN(t *testing.T) {
	// Test the DSN method of DBConfig
	config := &DBConfig{
		Host:     "testhost",
		Port:     5555,
		User:     "testuser",
		Password: "testpass",
		Name:     "testdb",
		SSLMode:  "disable",
	}

	expected := "host=testhost port=5555 user=testuser password=testpass dbname=testdb sslmode=disable"
	assert.Equal(t, expected, config.DSN(), "DSN should be formatted correctly")
}

func TestLoadConfigDefaults(t *testing.T) {
	// Save original viper instance and restore after test
	oldViper := viper.GetViper()
	defer func() {
		// Reset viper to original state
		viper.Reset()
		*viper.GetViper() = *oldViper
	}()

	// Create a fresh viper instance for testing
	viper.Reset()

	// Create temporary .env file for testing
	tempEnvFile := ".env.test"
	defer os.Remove(tempEnvFile)

	// Set config file to non-existent file to test defaults
	viper.SetConfigFile(tempEnvFile + ".nonexistent")

	// Load config
	cfg, err := LoadConfig()

	// Check if config is correctly loaded
	assert.NoError(t, err, "LoadConfig should not return an error")
	assert.NotNil(t, cfg, "Config should not be nil")

	// Check default DB values
	assert.Equal(t, "localhost", cfg.DB.Host, "Default DB host should be localhost")
	assert.Equal(t, 5432, cfg.DB.Port, "Default DB port should be 5432")
	assert.Equal(t, "postgres", cfg.DB.User, "Default DB user should be postgres")
	assert.Equal(t, "postgres", cfg.DB.Password, "Default DB password should be postgres")
	assert.Equal(t, "serviceability", cfg.DB.Name, "Default DB name should be serviceability")
	assert.Equal(t, "disable", cfg.DB.SSLMode, "Default DB SSL mode should be disable")
	assert.Equal(t, 25, cfg.DB.MaxOpenConns, "Default max open connections should be 25")
	assert.Equal(t, 25, cfg.DB.MaxIdleConns, "Default max idle connections should be 25")
	assert.Equal(t, 5*time.Minute, cfg.DB.ConnMaxLifetime, "Default connection max lifetime should be 5m")

	// Check default Log values
	assert.Equal(t, "info", cfg.Log.Level, "Default log level should be info")
	assert.Equal(t, "json", cfg.Log.Format, "Default log format should be json")
	assert.Equal(t, "stdout", cfg.Log.OutputPath, "Default log output path should be stdout")

	// Check default Server values
	assert.Equal(t, 8080, cfg.Server.Port, "Default server port should be 8080")
	assert.Equal(t, 10*time.Second, cfg.Server.ReadTimeout, "Default read timeout should be 10s")
	assert.Equal(t, 10*time.Second, cfg.Server.WriteTimeout, "Default write timeout should be 10s")
	assert.Equal(t, 120*time.Second, cfg.Server.IdleTimeout, "Default idle timeout should be 120s")
}

func TestLoadConfigFromEnv(t *testing.T) {
	// Save original viper instance and restore after test
	oldViper := viper.GetViper()
	defer func() {
		// Reset viper to original state
		viper.Reset()
		*viper.GetViper() = *oldViper
	}()

	// Create a fresh viper instance for testing
	viper.Reset()

	// Create temporary .env file for testing
	tempEnvFile := ".env.test"
	defer os.Remove(tempEnvFile)

	// Create temporary .env file with test values
	envContent := `
DB_HOST=envhost
DB_PORT=6543
DB_USER=envuser
DB_PASSWORD=envpass
DB_NAME=envdb
DB_SSL_MODE=require
DB_MAX_OPEN_CONNS=50
DB_MAX_IDLE_CONNS=10
DB_CONN_MAX_LIFETIME=10m
LOG_LEVEL=debug
LOG_FORMAT=console
LOG_OUTPUT_PATH=stderr
SERVER_PORT=9090
SERVER_READ_TIMEOUT=15s
SERVER_WRITE_TIMEOUT=20s
SERVER_IDLE_TIMEOUT=60s
`
	err := os.WriteFile(tempEnvFile, []byte(envContent), 0644)
	assert.NoError(t, err, "Failed to create test .env file")

	// Set config file
	viper.SetConfigFile(tempEnvFile)

	// Load config
	cfg, err := LoadConfig()

	// Check if config is correctly loaded
	assert.NoError(t, err, "LoadConfig should not return an error")
	assert.NotNil(t, cfg, "Config should not be nil")

	// Check environment values for DB config
	assert.Equal(t, "envhost", cfg.DB.Host, "DB host should match env value")
	assert.Equal(t, 6543, cfg.DB.Port, "DB port should match env value")
	assert.Equal(t, "envuser", cfg.DB.User, "DB user should match env value")
	assert.Equal(t, "envpass", cfg.DB.Password, "DB password should match env value")
	assert.Equal(t, "envdb", cfg.DB.Name, "DB name should match env value")
	assert.Equal(t, "require", cfg.DB.SSLMode, "DB SSL mode should match env value")
	assert.Equal(t, 50, cfg.DB.MaxOpenConns, "Max open connections should match env value")
	assert.Equal(t, 10, cfg.DB.MaxIdleConns, "Max idle connections should match env value")
	assert.Equal(t, 10*time.Minute, cfg.DB.ConnMaxLifetime, "Connection max lifetime should match env value")

	// Check environment values for Log config
	assert.Equal(t, "debug", cfg.Log.Level, "Log level should match env value")
	assert.Equal(t, "console", cfg.Log.Format, "Log format should match env value")
	assert.Equal(t, "stderr", cfg.Log.OutputPath, "Log output path should match env value")

	// Check environment values for Server config
	assert.Equal(t, 9090, cfg.Server.Port, "Server port should match env value")
	assert.Equal(t, 15*time.Second, cfg.Server.ReadTimeout, "Read timeout should match env value")
	assert.Equal(t, 20*time.Second, cfg.Server.WriteTimeout, "Write timeout should match env value")
	assert.Equal(t, 60*time.Second, cfg.Server.IdleTimeout, "Idle timeout should match env value")
}

func TestLoadConfigError(t *testing.T) {
	// Save original viper instance and restore after test
	oldViper := viper.GetViper()
	defer func() {
		// Reset viper to original state
		viper.Reset()
		*viper.GetViper() = *oldViper
	}()

	// Create a fresh viper instance for testing
	viper.Reset()

	// Create temporary .env file for testing
	tempEnvFile := ".env.test"
	defer os.Remove(tempEnvFile)

	// Create malformed .env file to test error handling
	envContent := `
DB_HOST=envhost
DB_PORT=invalid-port
`
	err := os.WriteFile(tempEnvFile, []byte(envContent), 0644)
	assert.NoError(t, err, "Failed to create test .env file")

	// Set config file
	viper.SetConfigFile(tempEnvFile)

	// Load config should fail due to invalid port
	_, err = LoadConfig()

	// Check if error is returned
	assert.Error(t, err, "LoadConfig should return an error for invalid values")
	assert.Contains(t, err.Error(), "unable to decode config", "Error message should indicate decoding issue")
}
