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

// Config holds all configuration for the application
type Config struct {
	DB     DBConfig
	Log    LogConfig
	Server ServerConfig
}

// DSN returns the PostgreSQL connection string
func (db *DBConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		db.Host, db.Port, db.User, db.Password, db.Name, db.SSLMode,
	)
}

// LoadConfig loads configuration from system environment variables
func LoadConfig() (*Config, error) {
	// Load .env file - fail if it doesn't exist
	if err := godotenv.Load(); err != nil {
		return nil, fmt.Errorf("failed to load .env file: %w", err)
	}

	// Read configuration directly from system environment variables using os.Getenv()
	config := &Config{
		DB: DBConfig{
			Host:            getEnvOrError("DB_HOST"),
			Port:            getEnvAsIntOrDefault("DB_PORT", 5432),
			User:            getEnvOrError("DB_USER"),
			Password:        getEnvOrError("DB_PASSWORD"),
			Name:            getEnvOrError("DB_NAME"),
			SSLMode:         getEnvOrDefault("DB_SSL_MODE", "disable"),
			MaxOpenConns:    getEnvAsIntOrDefault("DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns:    getEnvAsIntOrDefault("DB_MAX_IDLE_CONNS", 25),
			ConnMaxLifetime: getEnvAsDurationOrDefault("DB_CONN_MAX_LIFETIME", 5*time.Minute),
		},
		Log: LogConfig{
			Level:      getEnvOrDefault("LOG_LEVEL", "info"),
			Format:     getEnvOrDefault("LOG_FORMAT", "json"),
			OutputPath: getEnvOrDefault("LOG_OUTPUT_PATH", "stdout"),
		},
		Server: ServerConfig{
			Port:         getEnvAsIntOrDefault("SERVER_PORT", 8080),
			ReadTimeout:  getEnvAsDurationOrDefault("SERVER_READ_TIMEOUT", 10*time.Second),
			WriteTimeout: getEnvAsDurationOrDefault("SERVER_WRITE_TIMEOUT", 10*time.Second),
			IdleTimeout:  getEnvAsDurationOrDefault("SERVER_IDLE_TIMEOUT", 120*time.Second),
		},
	}

	// Validate that all required environment variables are set
	var missingVars []string

	if config.DB.Host == "" {
		missingVars = append(missingVars, "DB_HOST")
	}
	if config.DB.User == "" {
		missingVars = append(missingVars, "DB_USER")
	}
	if config.DB.Password == "" {
		missingVars = append(missingVars, "DB_PASSWORD")
	}
	if config.DB.Name == "" {
		missingVars = append(missingVars, "DB_NAME")
	}

	// Return error if required environment variables are missing
	if len(missingVars) > 0 {
		return nil, fmt.Errorf("required environment variables not set: %v", missingVars)
	}

	return config, nil
}

// Helper functions using os.Getenv() directly
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsIntOrDefault(key string, defaultValue int) int {
	if valueStr := os.Getenv(key); valueStr != "" {
		if value, err := strconv.Atoi(valueStr); err == nil {
			return value
		}
	}
	return defaultValue
}

func getEnvAsDurationOrDefault(key string, defaultValue time.Duration) time.Duration {
	if valueStr := os.Getenv(key); valueStr != "" {
		if value, err := time.ParseDuration(valueStr); err == nil {
			return value
		}
	}
	return defaultValue
}

func getEnvOrError(key string) string {
	return os.Getenv(key)
}
