package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/spf13/viper"
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

// LoadConfig loads configuration from .env file and environment variables
func LoadConfig() (*Config, error) {
	// Initialize new viper instance
	v := viper.New()

	// Enable viper to read environment variables
	v.AutomaticEnv()

	// Try to load from .env file
	v.SetConfigFile(".env")
	v.AddConfigPath(".")

	// It's okay if we can't find the .env file - just ignore any errors
	_ = v.ReadInConfig()

	// Set default values
	// setDefaultsForViper(v)

	// Create config instance
	config := &Config{
		DB: DBConfig{
			Host:            v.GetString("DB_HOST"),
			Port:            v.GetInt("DB_PORT"),
			User:            v.GetString("DB_USER"),
			Password:        v.GetString("DB_PASSWORD"),
			Name:            v.GetString("DB_NAME"),
			SSLMode:         v.GetString("DB_SSL_MODE"),
			MaxOpenConns:    v.GetInt("DB_MAX_OPEN_CONNS"),
			MaxIdleConns:    v.GetInt("DB_MAX_IDLE_CONNS"),
			ConnMaxLifetime: v.GetDuration("DB_CONN_MAX_LIFETIME"),
		},
		Log: LogConfig{
			Level:      v.GetString("LOG_LEVEL"),
			Format:     v.GetString("LOG_FORMAT"),
			OutputPath: v.GetString("LOG_OUTPUT_PATH"),
		},
		Server: ServerConfig{
			Port:         v.GetInt("SERVER_PORT"),
			ReadTimeout:  v.GetDuration("SERVER_READ_TIMEOUT"),
			WriteTimeout: v.GetDuration("SERVER_WRITE_TIMEOUT"),
			IdleTimeout:  v.GetDuration("SERVER_IDLE_TIMEOUT"),
		},
	}

	return config, nil
}

// setDefaultsForViper sets default values for configuration
func setDefaultsForViper(v *viper.Viper) {
	// Database defaults
	v.SetDefault("DB_HOST", "localhost")
	v.SetDefault("DB_PORT", 5432)
	v.SetDefault("DB_USER", "postgres")
	v.SetDefault("DB_PASSWORD", "postgres")
	v.SetDefault("DB_NAME", "serviceability")
	v.SetDefault("DB_SSL_MODE", "disable")
	v.SetDefault("DB_MAX_OPEN_CONNS", 25)
	v.SetDefault("DB_MAX_IDLE_CONNS", 25)
	v.SetDefault("DB_CONN_MAX_LIFETIME", 5*time.Minute)

	// Logging defaults
	v.SetDefault("LOG_LEVEL", "info")
	v.SetDefault("LOG_FORMAT", "json")
	v.SetDefault("LOG_OUTPUT_PATH", "stdout")

	// Server defaults
	v.SetDefault("SERVER_PORT", 9022)
	v.SetDefault("SERVER_READ_TIMEOUT", 10*time.Second)
	v.SetDefault("SERVER_WRITE_TIMEOUT", 10*time.Second)
	v.SetDefault("SERVER_IDLE_TIMEOUT", 120*time.Second)
}

// Helper functions for reading environment variables
func getEnv(key, defaultValue string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		return defaultValue
	}
	return value
}

func getEnvAsInt(key string, defaultValue int) int {
	valueStr := getEnv(key, strconv.Itoa(defaultValue))
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return defaultValue
	}
	return value
}

func getEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
	valueStr := getEnv(key, defaultValue.String())
	value, err := time.ParseDuration(valueStr)
	if err != nil {
		return defaultValue
	}
	return value
}
