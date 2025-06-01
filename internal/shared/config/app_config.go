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

// AppConfig holds all application configuration
type AppConfig struct {
	DB          DBConfig
	Log         LogConfig
	Server      ServerConfig
	Integration IntegrationConfig
}

// DSN returns the PostgreSQL connection string
func (db *DBConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s dbname=%s password=%s sslmode=require",
		db.Host, fmt.Sprintf("%d", db.Port), db.User, db.Name, db.Password,
	)
}

// LoadAppConfig loads all application configuration from environment variables
func LoadAppConfig() (*AppConfig, error) {
	// Load .env file
	godotenv.Load()

	config := &AppConfig{
		DB: DBConfig{
			Host:            getAppEnvOrDefault("DB_HOST", "localhost"),
			Port:            getAppEnvAsIntOrDefault("DB_PORT", 5432),
			User:            getAppEnvOrDefault("DB_USER", "postgres"),
			Password:        getAppEnvOrDefault("DB_PASSWORD", "postgres"),
			Name:            getAppEnvOrDefault("DB_NAME", "serviceability_db"),
			SSLMode:         getAppEnvOrDefault("DB_SSL_MODE", "disable"),
			MaxOpenConns:    getAppEnvAsIntOrDefault("DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns:    getAppEnvAsIntOrDefault("DB_MAX_IDLE_CONNS", 25),
			ConnMaxLifetime: getAppEnvAsDurationOrDefault("DB_CONN_MAX_LIFETIME", 5*time.Minute),
		},
		Log: LogConfig{
			Level:      getAppEnvOrDefault("LOG_LEVEL", "info"),
			Format:     getAppEnvOrDefault("LOG_FORMAT", "json"),
			OutputPath: getAppEnvOrDefault("LOG_OUTPUT_PATH", "stdout"),
		},
		Server: ServerConfig{
			Port:         getAppEnvAsIntOrDefault("SERVER_PORT", 9022),
			ReadTimeout:  getAppEnvAsDurationOrDefault("SERVER_READ_TIMEOUT", 10*time.Second),
			WriteTimeout: getAppEnvAsDurationOrDefault("SERVER_WRITE_TIMEOUT", 10*time.Second),
			IdleTimeout:  getAppEnvAsDurationOrDefault("SERVER_IDLE_TIMEOUT", 120*time.Second),
		},
		Integration: LoadIntegrationConfig(),
	}

	return config, nil
}

// Helper functions for environment variables
func getAppEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getAppEnvAsIntOrDefault(key string, defaultValue int) int {
	if valueStr := os.Getenv(key); valueStr != "" {
		if value, err := strconv.Atoi(valueStr); err == nil {
			return value
		}
	}
	return defaultValue
}

func getAppEnvAsDurationOrDefault(key string, defaultValue time.Duration) time.Duration {
	if valueStr := os.Getenv(key); valueStr != "" {
		if value, err := time.ParseDuration(valueStr); err == nil {
			return value
		}
	}
	return defaultValue
}

func getAppEnvAsBoolOrDefault(key string, defaultValue bool) bool {
	if valueStr := os.Getenv(key); valueStr != "" {
		if value, err := strconv.ParseBool(valueStr); err == nil {
			return value
		}
	}
	return defaultValue
}

func getAppEnvOrError(key string) string {
	return os.Getenv(key)
}
