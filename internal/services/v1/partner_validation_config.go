package services

import (
	"os"
	"strconv"
	"time"

	"prayog-serviceability-service/internal/shared/constants/v1"
)

// LoadPartnerValidationConfig loads configuration from environment variables
func LoadPartnerValidationConfig() PartnerValidationConfig {
	config := PartnerValidationConfig{
		BaseURL:        getEnvString(constants.EnvPartnerServiceBaseURL, ""),
		RequestTimeout: getEnvDuration("PARTNER_VALIDATION_TIMEOUT", 30*time.Second),
		RetryAttempts:  getEnvInt("PARTNER_VALIDATION_RETRY_ATTEMPTS", 3),
		RetryDelay:     getEnvDuration("PARTNER_VALIDATION_RETRY_DELAY", 1*time.Second),
	}

	// Validate required configuration
	if config.BaseURL == "" {
		config.BaseURL = "http://localhost:9024" // Default for development
	}

	return config
}

// getEnvString returns environment variable value as string with fallback
func getEnvString(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

// getEnvInt returns environment variable value as int with fallback
func getEnvInt(key string, fallback int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return fallback
}

// getEnvDuration returns environment variable value as time.Duration with fallback
func getEnvDuration(key string, fallback time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return fallback
}

// GetDefaultConfig returns a default configuration for testing and development
func GetDefaultConfig() PartnerValidationConfig {
	return PartnerValidationConfig{
		BaseURL:        "http://localhost:9024",
		RequestTimeout: 30 * time.Second,
		RetryAttempts:  3,
		RetryDelay:     1 * time.Second,
	}
}
