package handlers

import (
	"context"
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"

	services "prayog-serviceability-service/internal/services/v1/data"
	"prayog-serviceability-service/internal/services/v2/orchestrators"
	"prayog-serviceability-service/internal/shared/errors"
	modelsv1 "prayog-serviceability-service/internal/shared/models/v1"
)

// ServiceabilityV2Handler handles V2 serviceability check requests
type ServiceabilityV2Handler struct {
	v2Orchestrator     orchestrators.ServiceabilityOrchestrator
	geolocationService services.GeolocationService
	validator          *validator.Validate
	logger             *logrus.Logger
	errorHandler       *ErrorHandler
}

// NewServiceabilityV2Handler creates a new V2 serviceability handler
func NewServiceabilityV2Handler(
	v2Orchestrator orchestrators.ServiceabilityOrchestrator,
	geolocationService services.GeolocationService,
	validator *validator.Validate,
	logger *logrus.Logger,
) *ServiceabilityV2Handler {
	return &ServiceabilityV2Handler{
		v2Orchestrator:     v2Orchestrator,
		geolocationService: geolocationService,
		validator:          validator,
		logger:             logger,
		errorHandler:       NewErrorHandler(logger),
	}
}

// CheckServiceability handles POST /check (V2 single serviceability check)
func (h *ServiceabilityV2Handler) CheckServiceability(c *fiber.Ctx) error {
	h.logger.Debug("V2 single serviceability check requested")

	// Parse request body
	var request modelsv1.ServiceabilityV2Request
	if err := c.BodyParser(&request); err != nil {
		h.logger.WithError(err).Error("Failed to parse request body")
		return h.errorHandler.HandleParsingError(c, err)
	}

	h.logger.Debug("Request parsed successfully")

	// Validate request
	if err := h.validator.Struct(&request); err != nil {
		h.logger.WithError(err).Error("Request validation failed")
		return h.errorHandler.HandleValidationError(c, err)
	}

	h.logger.Debug("Request struct validation passed")

	// Skip enhanced validation - let orchestrator and partners handle postal code validation
	h.logger.Debug("Skipping enhanced validation - letting partners handle postal code validation")

	// Call V2 orchestrator
	h.logger.Debug("Calling V2 orchestrator")
	response, err := h.v2Orchestrator.CheckServiceability(c.Context(), &request)
	if err != nil {
		h.logger.WithError(err).Error("V2 orchestrator failed")

		// Handle structured errors with appropriate HTTP status codes
		if serviceErr, ok := err.(*errors.ServiceError); ok {
			return c.Status(serviceErr.HTTPStatus).JSON(modelsv1.ServiceabilityV2Response{
				Success:  false,
				Partners: []modelsv1.PartnerV2Response{},
				Error: &modelsv1.ErrorResponse{
					Code:    serviceErr.Code,
					Message: serviceErr.Message,
				},
			})
		}

		// Fallback to generic error handler for non-structured errors
		return h.errorHandler.HandleServiceError(c, ErrorCodeServiceabilityCheckFailed, "Failed to check V2 serviceability", err)
	}

	h.logger.Debug("V2 orchestrator completed successfully")

	// Return response with 200 OK status
	// Note: Success=false in response doesn't mean HTTP error - it means no serviceable partners found
	// This is a valid business response that should return 200 OK
	return c.Status(fiber.StatusOK).JSON(response)
}

// BulkCheckServiceability handles POST /bulk-check (V2 bulk serviceability check)
func (h *ServiceabilityV2Handler) BulkCheckServiceability(c *fiber.Ctx) error {
	h.logger.Debug("V2 bulk serviceability check requested")

	// Parse request body
	var request modelsv1.BulkServiceabilityV2Request
	if err := c.BodyParser(&request); err != nil {
		return h.errorHandler.HandleParsingError(c, err)
	}

	// Validate request
	if err := h.validator.Struct(&request); err != nil {
		return h.errorHandler.HandleValidationError(c, err)
	}

	// Validate each individual request in the bulk
	for i, req := range request.Requests {
		if err := h.validateV2Request(c.Context(), &req); err != nil {
			h.logger.WithError(err).Errorf("V2 bulk request item %d validation failed", i)

			// Handle structured errors with appropriate HTTP status codes
			if serviceErr, ok := err.(*errors.ServiceError); ok {
				return c.Status(serviceErr.HTTPStatus).JSON(modelsv1.BulkServiceabilityV2Response{
					Success: false,
					Data:    []modelsv1.ServiceabilityV2Response{},
					Error: &modelsv1.ErrorResponse{
						Code:    serviceErr.Code,
						Message: fmt.Sprintf("Validation failed for request item %d: %s", i+1, serviceErr.Message),
					},
				})
			}

			// Fallback to generic error handler for non-structured errors
			return h.errorHandler.HandleBusinessLogicError(c, ErrorCodeInvalidRequestType, "Invalid request data in bulk request", err)
		}
	}

	// Call V2 orchestrator
	response, err := h.v2Orchestrator.BulkCheckServiceability(c.Context(), &request)
	if err != nil {
		// Handle structured errors with appropriate HTTP status codes
		if serviceErr, ok := err.(*errors.ServiceError); ok {
			return c.Status(serviceErr.HTTPStatus).JSON(modelsv1.BulkServiceabilityV2Response{
				Success: false,
				Data:    []modelsv1.ServiceabilityV2Response{},
				Error: &modelsv1.ErrorResponse{
					Code:    serviceErr.Code,
					Message: serviceErr.Message,
				},
			})
		}

		// Fallback to generic error handler for non-structured errors
		return h.errorHandler.HandleBulkServiceError(c, ErrorCodeBulkServiceabilityCheckFailed, "Failed to process V2 bulk serviceability check", err)
	}

	// Return response with appropriate status code
	// Multi-status (207) is appropriate for bulk operations where some may succeed and others fail
	statusCode := fiber.StatusOK
	if !response.Success {
		statusCode = fiber.StatusMultiStatus
	}

	return c.Status(statusCode).JSON(response)
}

// GetHealthStatus handles GET /health (V2 health check)
func (h *ServiceabilityV2Handler) GetHealthStatus(c *fiber.Ctx) error {
	h.logger.Debug("V2 health check requested")

	// Create health response
	healthResponse := &modelsv1.ServiceabilityV2Response{
		Success: true,
		Partners: []modelsv1.PartnerV2Response{
			{
				PartnerID:     "system",
				PartnerName:   "V2 System Health",
				Rating:        5.0,
				IsServiceable: true,
				Services: []modelsv1.ServiceV2{
					{
						ServiceName: "health-check",
						TATDays:     1,
						IsCOD:       false,
						Pickup:      true,
						Delivery:    true,
					},
				},
			},
		},
		Metadata: &modelsv1.V2ResponseMetadata{
			TotalPartners:    1,
			ServiceableCount: 1,
			Filters: modelsv1.V2Filters{
				CountryCode: func() *string { s := "IN"; return &s }(),
			},
		},
	}

	return c.Status(fiber.StatusOK).JSON(healthResponse)
}

// validateV2Request validates V2 serviceability request data with enhanced logic
func (h *ServiceabilityV2Handler) validateV2Request(ctx context.Context, request *modelsv1.ServiceabilityV2Request) error {
	// Basic validation first
	if err := h.validateBasicRequest(request); err != nil {
		return err
	}

	// Enhanced validation for international parcels
	if err := h.validateInternationalRequest(ctx, request); err != nil {
		return err
	}

	return nil
}

// validateBasicRequest performs basic validation on the request
func (h *ServiceabilityV2Handler) validateBasicRequest(request *modelsv1.ServiceabilityV2Request) error {
	// Remove postal code existence validation - let partners handle invalid postal codes
	// This allows the system to continue processing even with invalid postal codes

	// Validate country code if provided
	if request.CountryCode != nil && len(*request.CountryCode) != 2 {
		return errors.ErrInvalidCountryCode(*request.CountryCode)
	}

	// Validate parcel category if provided
	if request.ParcelCategory != nil && *request.ParcelCategory != "" {
		validCategories := []string{"ecomm", "courier", "cargo", "international", "hyperlocal"}
		isValidCategory := false
		for _, category := range validCategories {
			if *request.ParcelCategory == category {
				isValidCategory = true
				break
			}
		}
		if !isValidCategory {
			return errors.ErrInvalidParcelCategory(*request.ParcelCategory)
		}
	}

	return nil
}

// validateInternationalRequest validates international parcel requirements
func (h *ServiceabilityV2Handler) validateInternationalRequest(ctx context.Context, request *modelsv1.ServiceabilityV2Request) error {
	// Check if context is nil
	if ctx == nil {
		return fmt.Errorf("context cannot be nil")
	}

	// Check if geolocation service is nil
	if h.geolocationService == nil {
		h.logger.Warn("Geolocation service is nil, skipping international validation")
		return nil
	}

	// Check if this is an international request
	isInternational := false

	// Method 1: Check if parcel_category is explicitly set to "international"
	if request.ParcelCategory != nil && *request.ParcelCategory == "international" {
		isInternational = true
		h.logger.Debugf("Request marked as international via parcel_category")
	}

	// Method 2: If no explicit category, try to check destination postal code for international shipping
	// But make it non-blocking - if postal code validation fails, just skip the geo check
	if !isInternational && request.DestinationPostalCode != nil && *request.DestinationPostalCode != "" {
		// Use geolocation service to determine if destination is international
		isDestinationInternational, countryCode, err := h.geolocationService.ValidateInternationalRequest(ctx, *request.DestinationPostalCode)
		if err != nil {
			// Don't fail early for postal code validation errors - just log and continue
			h.logger.WithError(err).Warnf("Could not validate destination postal code %s for international check, continuing with processing", *request.DestinationPostalCode)
		} else {
			isInternational = isDestinationInternational
			// Set country code in request if not already provided
			if request.CountryCode == nil && countryCode != nil {
				request.CountryCode = countryCode
				h.logger.Debugf("Set country code from geolocation: %s", *countryCode)
			}
		}
	}

	// Method 3: Check if postal_code (generic) is international
	// But make it non-blocking - if postal code validation fails, just skip the geo check
	if !isInternational && request.PostalCode != nil && *request.PostalCode != "" {
		isDestinationInternational, countryCode, err := h.geolocationService.ValidateInternationalRequest(ctx, *request.PostalCode)
		if err != nil {
			// Don't fail early for postal code validation errors - just log and continue
			h.logger.WithError(err).Warnf("Could not validate postal code %s for international check, continuing with processing", *request.PostalCode)
		} else {
			isInternational = isDestinationInternational
			// Set country code in request if not already provided
			if request.CountryCode == nil && countryCode != nil {
				request.CountryCode = countryCode
				h.logger.Debugf("Set country code from geolocation: %s", *countryCode)
			}
		}
	}

	// If this is an international request, validate package requirements
	if isInternational {
		h.logger.Debugf("Validating international request requirements")

		// Package information is mandatory for international shipments
		if request.Package == nil {
			h.logger.Error("Package information is missing for international request")
			return errors.ErrInternationalPackageRequired()
		}

		h.logger.Debugf("Package is not nil, checking weight")

		// Weight is mandatory for international shipments
		if request.Package.Weight == nil {
			h.logger.Error("Package weight is missing for international request")
			return errors.ErrInternationalPackageRequired()
		}

		h.logger.Debugf("Weight is not nil, checking dimensions")

		// Dimensions are mandatory for international shipments
		if request.Package.Dimensions == nil {
			h.logger.Error("Package dimensions are missing for international request")
			return errors.ErrInternationalPackageRequired()
		}

		h.logger.Debugf("Dimensions is not nil, validating weight value")

		// Validate weight values
		if request.Package.Weight.Value <= 0 {
			h.logger.Error("Package weight value is invalid for international request")
			return errors.ErrInvalidPackageWeight(request.Package.Weight.Value)
		}

		h.logger.Debugf("Weight value is valid, validating dimension values")

		// Validate dimension values
		if request.Package.Dimensions.Length <= 0 || request.Package.Dimensions.Width <= 0 || request.Package.Dimensions.Height <= 0 {
			h.logger.Error("Package dimensions are invalid for international request")
			return errors.ErrInvalidPackageDimensions()
		}

		h.logger.Debugf("All international request validations passed")
	}

	return nil
}

// Legacy method names for backward compatibility
func (h *ServiceabilityV2Handler) CheckServiceabilityV2(c *fiber.Ctx) error {
	return h.CheckServiceability(c)
}

func (h *ServiceabilityV2Handler) BulkCheckServiceabilityV2(c *fiber.Ctx) error {
	return h.BulkCheckServiceability(c)
}

func (h *ServiceabilityV2Handler) GetHealthV2(c *fiber.Ctx) error {
	return h.GetHealthStatus(c)
}
