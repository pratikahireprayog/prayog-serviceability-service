package services

import (
	"context"
	"log"
	"time"

	"prayog-serviceability-service/internal/shared/constants/v1"
)

// PartnerValidationService defines the interface for partner validation operations
type PartnerValidationService interface {
	// ValidatePartner validates if a partner exists and is active
	// Returns INVALID_PARTNER error if validation fails
	ValidatePartner(ctx context.Context, partnerID string) error
}

// PartnerValidationServiceImpl implements the PartnerValidationService interface
type PartnerValidationServiceImpl struct {
	baseURL string
	client  HTTPClient
	logger  *log.Logger
}

// HTTPClient defines the interface for HTTP operations
type HTTPClient interface {
	Get(ctx context.Context, url string) (*HTTPResponse, error)
}

// HTTPResponse represents an HTTP response
type HTTPResponse struct {
	StatusCode int
	Body       []byte
	Headers    map[string]string
}

// PartnerValidationConfig holds configuration for the partner validation service
type PartnerValidationConfig struct {
	BaseURL        string
	RequestTimeout time.Duration
	RetryAttempts  int
	RetryDelay     time.Duration
}

// NewPartnerValidationService creates a new instance of PartnerValidationService
func NewPartnerValidationService(config PartnerValidationConfig, client HTTPClient, logger *log.Logger) PartnerValidationService {
	if config.RequestTimeout == 0 {
		config.RequestTimeout = 30 * time.Second
	}
	if config.RetryAttempts == 0 {
		config.RetryAttempts = 3
	}
	if config.RetryDelay == 0 {
		config.RetryDelay = 1 * time.Second
	}

	return &PartnerValidationServiceImpl{
		baseURL: config.BaseURL,
		client:  client,
		logger:  logger,
	}
}

// ValidatePartner validates if a partner exists and is active by calling the Partner API
func (s *PartnerValidationServiceImpl) ValidatePartner(ctx context.Context, partnerID string) error {
	if partnerID == "" {
		if s.logger != nil {
			s.logger.Printf("Partner validation failed: empty partner ID")
		}
		return NewPartnerValidationError(constants.ErrorCodeInvalidPartner, "Partner ID cannot be empty", nil)
	}

	url := s.baseURL + "/partner/v1/partners/" + partnerID
	if s.logger != nil {
		s.logger.Printf("Validating partner %s at URL: %s", partnerID, url)
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	response, err := s.client.Get(ctx, url)
	if err != nil {
		if s.logger != nil {
			s.logger.Printf("Partner validation failed for %s: %v", partnerID, err)
		}
		return NewPartnerValidationError(constants.ErrorCodePartnerServiceError, "Failed to validate partner", err)
	}

	// Handle response based on status code
	switch response.StatusCode {
	case constants.StatusOK:
		if s.logger != nil {
			s.logger.Printf("Partner %s validation successful", partnerID)
		}
		return nil
	case constants.StatusNotFound:
		if s.logger != nil {
			s.logger.Printf("Partner %s not found", partnerID)
		}
		return NewPartnerValidationError(constants.ErrorCodeInvalidPartner, "Partner not found", nil)
	case constants.StatusUnauthorized:
		if s.logger != nil {
			s.logger.Printf("Unauthorized access when validating partner %s", partnerID)
		}
		return NewPartnerValidationError(constants.ErrorCodeAuthFailed, "Unauthorized access to partner service", nil)
	default:
		if s.logger != nil {
			s.logger.Printf("Partner validation failed for %s with status %d", partnerID, response.StatusCode)
		}
		return NewPartnerValidationError(constants.ErrorCodePartnerServiceError, "Partner service returned error", nil)
	}
}

// PartnerValidationError represents an error that occurred during partner validation
type PartnerValidationError struct {
	Code    string
	Message string
	Cause   error
}

// Error implements the error interface
func (e *PartnerValidationError) Error() string {
	if e.Cause != nil {
		return e.Message + ": " + e.Cause.Error()
	}
	return e.Message
}

// NewPartnerValidationError creates a new PartnerValidationError
func NewPartnerValidationError(code, message string, cause error) *PartnerValidationError {
	return &PartnerValidationError{
		Code:    code,
		Message: message,
		Cause:   cause,
	}
}

// GetErrorCode returns the error code for the validation error
func (e *PartnerValidationError) GetErrorCode() string {
	return e.Code
}

// Unwrap returns the underlying cause error
func (e *PartnerValidationError) Unwrap() error {
	return e.Cause
}
