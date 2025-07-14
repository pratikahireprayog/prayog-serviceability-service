package services

import (
	"github.com/sirupsen/logrus"

	"prayog-serviceability-service/internal/shared/config"
	"prayog-serviceability-service/internal/shared/interfaces/v1"
)

// IntegrationFactory creates and manages service integration instances
type IntegrationFactory struct {
	config *config.AppConfig
	logger *logrus.Logger
}

// NewIntegrationFactory creates a new integration factory
func NewIntegrationFactory(appConfig *config.AppConfig, logger *logrus.Logger) (*IntegrationFactory, error) {
	return &IntegrationFactory{
		config: appConfig,
		logger: logger,
	}, nil
}

// CreatePartnerServiceClient creates a Partner Service integration client
func (f *IntegrationFactory) CreatePartnerServiceClient() (interfaces.PartnerServiceClient, error) {
	partnerConfig := f.config.Integration.Partner

	f.logger.WithField("base_url", partnerConfig.BaseURL).Info("🤝 Creating Partner Service integration client")

	return NewPartnerIntegrationService(partnerConfig), nil
}

// CreateSpecificationServiceClient creates a Specification Service integration client
func (f *IntegrationFactory) CreateSpecificationServiceClient() (interfaces.SpecificationServiceClient, error) {
	specConfig := f.config.Integration.Specification

	f.logger.WithField("base_url", specConfig.BaseURL).Info("📋 Creating Specification Service integration client")

	return NewSpecificationIntegrationService(specConfig), nil
}

// GetAllIntegrationClients returns all available integration clients
func (f *IntegrationFactory) GetAllIntegrationClients() (map[string]interface{}, error) {
	partnerClient, err := f.CreatePartnerServiceClient()
	if err != nil {
		return nil, err
	}

	specClient, err := f.CreateSpecificationServiceClient()
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"partner":       partnerClient,
		"specification": specClient,
	}, nil
}

// ValidateIntegrationConfig validates all integration configurations
func (f *IntegrationFactory) ValidateIntegrationConfig() error {
	if err := f.config.Integration.Validate(); err != nil {
		f.logger.WithError(err).Error("Integration configuration validation failed")
		return err
	}

	f.logger.Info("All integration configurations validated successfully")
	return nil
}

// GetIntegrationHealthStatus returns health status of all integrations
func (f *IntegrationFactory) GetIntegrationHealthStatus() map[string]string {
	status := make(map[string]string)

	// Check Partner Service integration
	partnerConfig := config.LoadIntegrationConfig().Partner
	if err := partnerConfig.Validate(); err != nil {
		status["partner"] = "unhealthy: " + err.Error()
	} else {
		status["partner"] = "healthy"
	}

	// Check Specification Service integration
	specConfig := config.LoadIntegrationConfig().Specification
	if err := specConfig.Validate(); err != nil {
		status["specification"] = "unhealthy: " + err.Error()
	} else {
		status["specification"] = "healthy"
	}

	return status
}
