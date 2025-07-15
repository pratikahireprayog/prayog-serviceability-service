package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

// DBConfig contains database configuration
type DBConfig struct {
	Host            string        `mapstructure:"DB_HOST"`
	Port            int           `mapstructure:"DB_PORT"`
	User            string        `mapstructure:"DB_USER"`
	Password        string        `mapstructure:"DB_PASSWORD"`
	Name            string        `mapstructure:"DB_NAME"`
	SSLMode         string        `mapstructure:"DB_SSL_MODE"`
	MaxOpenConns    int           `mapstructure:"DB_MAX_OPEN_CONNS"`
	MaxIdleConns    int           `mapstructure:"DB_MAX_IDLE_CONNS"`
	ConnMaxLifetime time.Duration `mapstructure:"DB_CONN_MAX_LIFETIME"`
	QueryTimeout    time.Duration `mapstructure:"DB_QUERY_TIMEOUT"`
}

// LogConfig contains logging configuration
type LogConfig struct {
	Level      string `mapstructure:"LOG_LEVEL"`
	Format     string `mapstructure:"LOG_FORMAT"`
	OutputPath string `mapstructure:"LOG_OUTPUT_PATH"`
}

// ServerConfig contains HTTP server configuration
type ServerConfig struct {
	Port         int           `mapstructure:"SERVER_PORT"`
	ReadTimeout  time.Duration `mapstructure:"SERVER_READ_TIMEOUT"`
	WriteTimeout time.Duration `mapstructure:"SERVER_WRITE_TIMEOUT"`
	IdleTimeout  time.Duration `mapstructure:"SERVER_IDLE_TIMEOUT"`
}

// ServiceabilityConfig contains serviceability-specific configuration
type ServiceabilityConfig struct {
	ReturnOnlyServiceablePartners bool `mapstructure:"RETURN_ONLY_SERVICEABLE_PARTNERS"`
}

// AppConfig holds all application configuration
type AppConfig struct {
	DB             DBConfig
	Log            LogConfig
	Server         ServerConfig
	Serviceability ServiceabilityConfig
	Integration    IntegrationConfig
}

// DSN returns the PostgreSQL connection string
func (db *DBConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s dbname=%s password=%s sslmode=%s",
		db.Host, db.Port, db.User, db.Name, db.Password, db.SSLMode,
	)
}

// GetConnectionInfo returns a map of connection information for logging (without password)
func (db *DBConfig) GetConnectionInfo() map[string]interface{} {
	return map[string]interface{}{
		"host":              db.Host,
		"port":              db.Port,
		"database":          db.Name,
		"user":              db.User,
		"ssl_mode":          db.SSLMode,
		"max_open_conns":    db.MaxOpenConns,
		"max_idle_conns":    db.MaxIdleConns,
		"conn_max_lifetime": db.ConnMaxLifetime.String(),
		"query_timeout":     db.QueryTimeout.String(),
	}
}

// envVars holds the loaded environment variables from .env file and system
var envVars map[string]string

// LoadAppConfig loads all application configuration from environment variables
func LoadAppConfig() (*AppConfig, error) {
	// Load .env file first and prioritize its values
	envVars = make(map[string]string)

	// Load .env file values
	if envMap, err := godotenv.Read(); err == nil {
		for key, value := range envMap {
			envVars[key] = value
		}
	}

	// Load system environment variables only if not present in .env
	for _, env := range os.Environ() {
		if len(env) > 0 {
			if idx := findIndex(env, '='); idx > 0 {
				key := env[:idx]
				value := env[idx+1:]
				// Only use system env var if not already set from .env file
				if _, exists := envVars[key]; !exists {
					envVars[key] = value
				}
			}
		}
	}

	config := &AppConfig{
		DB: DBConfig{
			Host:            getEnvOrDefault("DB_HOST", "localhost"),
			Port:            getEnvAsIntOrDefault("DB_PORT", 5432),
			User:            getEnvOrDefault("DB_USER", "postgres"),
			Password:        getEnvOrDefault("DB_PASSWORD", "postgres"),
			Name:            getEnvOrDefault("DB_NAME", "serviceability_db"),
			SSLMode:         enforceSSLMode(),
			MaxOpenConns:    getEnvAsIntOrDefault("DB_MAX_OPEN_CONNS", 0),    // 0 = unlimited
			MaxIdleConns:    getEnvAsIntOrDefault("DB_MAX_IDLE_CONNS", 1000), // High idle pool
			ConnMaxLifetime: getEnvAsDurationOrDefault("DB_CONN_MAX_LIFETIME", 5*time.Minute),
			QueryTimeout:    getEnvAsDurationOrDefault("DB_QUERY_TIMEOUT", 30*time.Second),
		},
		Log: LogConfig{
			Level:      getEnvOrDefault("LOG_LEVEL", "info"),
			Format:     getEnvOrDefault("LOG_FORMAT", "json"),
			OutputPath: getEnvOrDefault("LOG_OUTPUT_PATH", "stdout"),
		},
		Server: ServerConfig{
			Port:         getEnvAsIntOrDefault("SERVER_PORT", 9022),
			ReadTimeout:  getEnvAsDurationOrDefault("SERVER_READ_TIMEOUT", 10*time.Second),
			WriteTimeout: getEnvAsDurationOrDefault("SERVER_WRITE_TIMEOUT", 10*time.Second),
			IdleTimeout:  getEnvAsDurationOrDefault("SERVER_IDLE_TIMEOUT", 120*time.Second),
		},
		Serviceability: ServiceabilityConfig{
			ReturnOnlyServiceablePartners: getEnvAsBoolOrDefault("RETURN_ONLY_SERVICEABLE_PARTNERS", true),
		},
		Integration: LoadIntegrationConfig(),
	}

	return config, nil
}

// findIndex finds the first occurrence of a character in a string
func findIndex(s string, char rune) int {
	for i, c := range s {
		if c == char {
			return i
		}
	}
	return -1
}

// Helper functions for environment variables using the prioritized envVars map
func getEnvOrDefault(key, defaultValue string) string {
	if value, exists := envVars[key]; exists && value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsIntOrDefault(key string, defaultValue int) int {
	if valueStr := getEnvOrDefault(key, ""); valueStr != "" {
		if value, err := strconv.Atoi(valueStr); err == nil {
			return value
		}
	}
	return defaultValue
}

func getEnvAsDurationOrDefault(key string, defaultValue time.Duration) time.Duration {
	if valueStr := getEnvOrDefault(key, ""); valueStr != "" {
		if value, err := time.ParseDuration(valueStr); err == nil {
			return value
		}
	}
	return defaultValue
}

func getEnvAsBoolOrDefault(key string, defaultValue bool) bool {
	if valueStr := getEnvOrDefault(key, ""); valueStr != "" {
		if value, err := strconv.ParseBool(valueStr); err == nil {
			return value
		}
	}
	return defaultValue
}

func getEnvOrError(key string) string {
	return getEnvOrDefault(key, "")
}

// enforceSSLMode ensures SSL is always required for database connections
// This is enforced at the code level for security reasons, especially for cloud databases like AWS RDS
func enforceSSLMode() string {
	// Get the environment variable but validate it
	envSSLMode := getEnvOrDefault("DB_SSL_MODE", "require")

	// For production and cloud databases, SSL should always be required
	// Only allow "disable" for local development with explicit override
	switch envSSLMode {
	case "disable":
		// Only allow SSL disable for localhost development
		host := getEnvOrDefault("DB_HOST", "localhost")
		if host == "localhost" || host == "127.0.0.1" {
			return "disable"
		}
		// For non-localhost connections, force SSL
		return "require"
	case "allow", "prefer", "require", "verify-ca", "verify-full":
		// These are all valid SSL modes, prefer "require" as minimum
		if envSSLMode == "allow" {
			return "require" // Upgrade "allow" to "require" for security
		}
		return envSSLMode
	default:
		// Unknown/invalid SSL mode, default to require
		return "require"
	}
}
