package config

import (
	"fmt"

	"github.com/sirupsen/logrus"
)

// ConfigManager manages all application configurations
type ConfigManager struct {
	App         *AppConfig
	Integration IntegrationConfig
	logger      *logrus.Logger
}

// NewConfigManager creates a new configuration manager that loads all configs
func NewConfigManager() (*ConfigManager, error) {
	// Initialize logger first
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)

	// Load application config
	appConfig, err := LoadAppConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load app config: %w", err)
	}

	// Set logger level from config
	if level, err := logrus.ParseLevel(appConfig.Log.Level); err == nil {
		logger.SetLevel(level)
	}

	// Load integration config
	integrationConfig := LoadIntegrationConfig()

	// Validate all configurations
	if err := integrationConfig.Validate(); err != nil {
		return nil, fmt.Errorf("integration config validation failed: %w", err)
	}

	logger.Info("All configurations loaded and validated successfully")

	return &ConfigManager{
		App:         appConfig,
		Integration: integrationConfig,
		logger:      logger,
	}, nil
}

// GetLogger returns the configured logger
func (cm *ConfigManager) GetLogger() *logrus.Logger {
	return cm.logger
}

// ValidateAll validates all configurations
func (cm *ConfigManager) ValidateAll() error {
	if err := cm.Integration.Validate(); err != nil {
		return fmt.Errorf("integration config validation failed: %w", err)
	}

	cm.logger.Info("All configurations validated successfully")
	return nil
}

// GetDatabaseConnectionString returns the database connection string
func (cm *ConfigManager) GetDatabaseConnectionString() string {
	return cm.App.DB.DSN()
}

// GetIntegrationHealthStatus returns health status of all integrations
func (cm *ConfigManager) GetIntegrationHealthStatus() map[string]string {
	status := make(map[string]string)

	// Check Partner Service config
	if err := cm.Integration.Partner.Validate(); err != nil {
		status["partner"] = "unhealthy: " + err.Error()
	} else {
		status["partner"] = "healthy"
	}

	// Check Specification Service config
	if err := cm.Integration.Specification.Validate(); err != nil {
		status["specification"] = "unhealthy: " + err.Error()
	} else {
		status["specification"] = "healthy"
	}

	return status
}
