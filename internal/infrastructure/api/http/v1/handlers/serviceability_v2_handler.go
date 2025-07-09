package handlers

import (
	"context"
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"

	"prayog-serviceability-service/internal/services/v2/orchestrators"
	modelsv1 "prayog-serviceability-service/internal/shared/models/v1"
	"prayog-serviceability-service/internal/shared/services/v1"
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

	// Enhanced validation with geolocation support
	h.logger.Debug("Starting enhanced validation")
	if err := h.validateV2Request(c.Context(), &request); err != nil {
		h.logger.WithError(err).Error("Enhanced validation failed")
		return h.errorHandler.HandleBusinessLogicError(c, ErrorCodeInvalidRequestType, "Invalid request data", err)
	}

	h.logger.Debug("Enhanced validation passed")

	// Call V2 orchestrator
	h.logger.Debug("Calling V2 orchestrator")
	response, err := h.v2Orchestrator.CheckServiceability(c.Context(), &request)
	if err != nil {
		h.logger.WithError(err).Error("V2 orchestrator failed")
		return h.errorHandler.HandleServiceError(c, ErrorCodeServiceabilityCheckFailed, "Failed to check V2 serviceability", err)
	}

	h.logger.Debug("V2 orchestrator completed successfully")

	// Return response with appropriate status code
	statusCode := fiber.StatusOK
	if !response.Success {
		statusCode = fiber.StatusInternalServerError
	}

	return c.Status(statusCode).JSON(response)
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
			return h.errorHandler.HandleBusinessLogicError(c, ErrorCodeInvalidRequestType, "Invalid request data in bulk request", err)
		}
	}

	// Call V2 orchestrator
	response, err := h.v2Orchestrator.BulkCheckServiceability(c.Context(), &request)
	if err != nil {
		return h.errorHandler.HandleBulkServiceError(c, ErrorCodeBulkServiceabilityCheckFailed, "Failed to process V2 bulk serviceability check", err)
	}

	// Return response with appropriate status code
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
	// Validate source postal code format if provided
	if request.SourcePostalCode != nil && *request.SourcePostalCode != "" {
		// Basic validation - just check if it's not empty
		// TODO: Add proper postal code format validation
	}

	// Validate destination postal code format if provided
	if request.DestinationPostalCode != nil && *request.DestinationPostalCode != "" {
		// Basic validation - just check if it's not empty
		// TODO: Add proper postal code format validation
	}

	// Validate country code if provided
	if request.CountryCode != nil && len(*request.CountryCode) != 2 {
		return fiber.NewError(fiber.StatusBadRequest, "Country code must be 2 characters")
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
			return fiber.NewError(fiber.StatusBadRequest, "Invalid parcel category")
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

	// Method 2: If no explicit category, check destination postal code for international shipping
	if !isInternational && request.DestinationPostalCode != nil && *request.DestinationPostalCode != "" {
		// Use geolocation service to determine if destination is international
		isDestinationInternational, countryCode, err := h.geolocationService.ValidateInternationalRequest(ctx, *request.DestinationPostalCode)
		if err != nil {
			// If we can't determine country code, log warning but don't fail the request
			h.logger.WithError(err).Warnf("Failed to get country code for postal code %s", *request.DestinationPostalCode)
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
	if !isInternational && request.PostalCode != nil && *request.PostalCode != "" {
		isDestinationInternational, countryCode, err := h.geolocationService.ValidateInternationalRequest(ctx, *request.PostalCode)
		if err != nil {
			h.logger.WithError(err).Warnf("Failed to get country code for postal code %s", *request.PostalCode)
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
			return fiber.NewError(fiber.StatusBadRequest, "Package information (weight and dimensions) is required for international shipments")
		}

		h.logger.Debugf("Package is not nil, checking weight")

		// Weight is mandatory for international shipments
		if request.Package.Weight == nil {
			h.logger.Error("Package weight is missing for international request")
			return fiber.NewError(fiber.StatusBadRequest, "Package weight is required for international shipments")
		}

		h.logger.Debugf("Weight is not nil, checking dimensions")

		// Dimensions are mandatory for international shipments
		if request.Package.Dimensions == nil {
			h.logger.Error("Package dimensions are missing for international request")
			return fiber.NewError(fiber.StatusBadRequest, "Package dimensions are required for international shipments")
		}

		h.logger.Debugf("Dimensions is not nil, validating weight value")

		// Validate weight values
		if request.Package.Weight.Value <= 0 {
			h.logger.Error("Package weight value is invalid for international request")
			return fiber.NewError(fiber.StatusBadRequest, "Package weight must be greater than 0")
		}

		h.logger.Debugf("Weight value is valid, validating dimension values")

		// Validate dimension values
		if request.Package.Dimensions.Length <= 0 || request.Package.Dimensions.Width <= 0 || request.Package.Dimensions.Height <= 0 {
			h.logger.Error("Package dimensions are invalid for international request")
			return fiber.NewError(fiber.StatusBadRequest, "Package dimensions must be greater than 0")
		}

		h.logger.Debugf("All dimension values are valid")

		// Log successful validation with proper nil check
		if request.CountryCode != nil {
			h.logger.Debugf("International request validation passed for country code: %s", *request.CountryCode)
		} else {
			h.logger.Debugf("International request validation passed (country code not resolved)")
		}
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
