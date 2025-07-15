package orchestrators

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"prayog-serviceability-service/internal/services/v2/partners/common"
	"prayog-serviceability-service/internal/services/v2/partners/factory"
	"prayog-serviceability-service/internal/shared/errors"
	"prayog-serviceability-service/internal/shared/models/v1"
	"prayog-serviceability-service/internal/shared/repositories/v1"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// ServiceabilityOrchestrator defines the interface for V2 serviceability orchestration
type ServiceabilityOrchestrator interface {
	CheckServiceability(ctx context.Context, req *models.ServiceabilityV2Request) (*models.ServiceabilityV2Response, error)
	BulkCheckServiceability(ctx context.Context, req *models.BulkServiceabilityV2Request) (*models.BulkServiceabilityV2Response, error)
}

// serviceabilityOrchestrator implements the V2 ServiceabilityOrchestrator interface
type serviceabilityOrchestrator struct {
	partnerFactory          factory.PartnerAdapterFactory
	partnerAttributeMapRepo repositories.PartnerAttributeMapRepository
	timeout                 time.Duration
	returnOnlyServiceable   bool
	logger                  *logrus.Logger
}

// NewServiceabilityOrchestrator creates a new V2 serviceability orchestrator
func NewServiceabilityOrchestrator(
	partnerFactory factory.PartnerAdapterFactory,
	partnerAttributeMapRepo repositories.PartnerAttributeMapRepository,
	timeout time.Duration,
	returnOnlyServiceable bool,
) ServiceabilityOrchestrator {
	// Initialize logger
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)

	return &serviceabilityOrchestrator{
		partnerFactory:          partnerFactory,
		partnerAttributeMapRepo: partnerAttributeMapRepo,
		timeout:                 timeout,
		returnOnlyServiceable:   returnOnlyServiceable,
		logger:                  logger,
	}
}

// CheckServiceability orchestrates the V2 serviceability check process
func (s *serviceabilityOrchestrator) CheckServiceability(ctx context.Context, req *models.ServiceabilityV2Request) (*models.ServiceabilityV2Response, error) {
	// Validate input request
	if req == nil {
		return nil, errors.ErrInvalidRequest("request cannot be nil")
	}

	if err := s.validateV2Request(req); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	// Set timeout context
	timeoutCtx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	// Get eligible partners based on parcel category filtering
	eligiblePartners, err := s.getEligiblePartners(timeoutCtx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get eligible partners: %w", err)
	}

	// Check serviceability with all eligible partners concurrently
	partnerResults := s.checkWithPartners(timeoutCtx, req, eligiblePartners)

	if len(eligiblePartners) == 0 {
		return &models.ServiceabilityV2Response{
			Success:  false,
			Partners: []models.PartnerV2Response{},
			Metadata: &models.V2ResponseMetadata{
				TotalPartners:    0,
				ServiceableCount: 0,
				Filters: models.V2Filters{
					CountryCode:    req.CountryCode,
					ParcelCategory: req.ParcelCategory,
					ProductType:    req.ProductType,
				},
			},
		}, nil
	}

	// Process results and build response
	response := s.buildV2Response(partnerResults, req)

	return response, nil
}

// BulkCheckServiceability orchestrates bulk V2 serviceability checks
func (s *serviceabilityOrchestrator) BulkCheckServiceability(ctx context.Context, req *models.BulkServiceabilityV2Request) (*models.BulkServiceabilityV2Response, error) {
	if req == nil || len(req.Requests) == 0 {
		return nil, errors.ErrInvalidRequest("bulk request cannot be nil or empty")
	}

	responses := make([]models.ServiceabilityV2Response, 0, len(req.Requests))
	errorCount := 0

	// Process each individual request
	for i, individualReq := range req.Requests {
		response, err := s.CheckServiceability(ctx, &individualReq)
		if err != nil {
			// Handle individual request failures
			errorCount++
			errorMsg := fmt.Sprintf("Request %d failed: %v", i+1, err)

			// Add error response for this request
			responses = append(responses, models.ServiceabilityV2Response{
				Success:  false,
				Partners: []models.PartnerV2Response{},
				Metadata: &models.V2ResponseMetadata{
					TotalPartners:    0,
					ServiceableCount: 0,
					Filters: models.V2Filters{
						CountryCode:    individualReq.CountryCode,
						ParcelCategory: individualReq.ParcelCategory,
						ProductType:    individualReq.ProductType,
					},
				},
				Error: &models.ErrorResponse{
					Code:    "REQUEST_FAILED",
					Message: errorMsg,
				},
			})
		} else {
			responses = append(responses, *response)
		}
	}

	bulkResponse := &models.BulkServiceabilityV2Response{
		Success: errorCount == 0,
		Data:    responses,
	}

	if errorCount > 0 {
		bulkResponse.Error = &models.ErrorResponse{
			Code:    "PARTIAL_FAILURE",
			Message: fmt.Sprintf("Processed %d requests with %d errors", len(req.Requests), errorCount),
		}
	}

	return bulkResponse, nil
}

// checkWithPartners checks serviceability with multiple partners concurrently
func (s *serviceabilityOrchestrator) checkWithPartners(ctx context.Context, req *models.ServiceabilityV2Request, partnerInfos []DatabasePartnerInfo) []partnerResult {
	var wg sync.WaitGroup
	results := make([]partnerResult, len(partnerInfos))

	// Check each partner concurrently
	for i, partnerInfo := range partnerInfos {
		wg.Add(1)
		go func(index int, info DatabasePartnerInfo) {
			defer wg.Done()
			result := s.checkWithPartner(ctx, req, info)
			// Store partner database info in result for later use
			result.PartnerInfo = &info
			results[index] = result
		}(i, partnerInfo)
	}

	wg.Wait()
	return results
}

// checkWithPartner checks serviceability with a single partner
func (s *serviceabilityOrchestrator) checkWithPartner(ctx context.Context, req *models.ServiceabilityV2Request, info DatabasePartnerInfo) partnerResult {
	adapter, exists := s.partnerFactory.GetAdapter(info.PartnerCode)
	if !exists {
		return partnerResult{
			PartnerCode: info.PartnerCode,
			Error:       errors.ErrPartnerNotFound(info.PartnerCode),
		}
	}

	// Check if adapter is healthy
	if !adapter.IsHealthy(ctx) {
		return partnerResult{
			PartnerCode: info.PartnerCode,
			Error:       errors.ErrPartnerUnavailable(info.PartnerCode),
		}
	}

	// Call the adapter
	result, err := adapter.CheckServiceability(ctx, req, common.PartnerInfo{PartnerID: info.PartnerID, PartnerCode: info.PartnerCode})
	if err != nil {
		return partnerResult{
			PartnerCode: info.PartnerCode, // No longer needed since adapter sets it
			Error:       err,
		}
	}

	return partnerResult{
		PartnerCode: info.PartnerCode,
		Result:      result,
	}
}

// buildV2Response builds the V2 response from partner results
func (s *serviceabilityOrchestrator) buildV2Response(partnerResults []partnerResult, req *models.ServiceabilityV2Request) *models.ServiceabilityV2Response {
	allPartners := make([]models.PartnerV2Response, 0)
	serviceablePartners := make([]models.PartnerV2Response, 0)
	serviceableCount := 0

	// Process each partner result
	for _, result := range partnerResults {
		if result.Error != nil {
			// Add error partner response using database info
			errorMsg := result.Error.Error()

			// Use partner info from database (should always be available)
			partnerID := ""
			if result.PartnerInfo != nil && result.PartnerInfo.PartnerID != nil {
				partnerID = result.PartnerInfo.PartnerID.String()
			} else {
				partnerID = "unknown"
			}
			partnerCode := result.PartnerInfo.PartnerCode

			errorPartnerResponse := models.PartnerV2Response{
				PartnerID:     partnerID,
				PartnerCode:   partnerCode,
				PartnerName:   "",  // No partner name in database
				Rating:        0.0, // No rating in database
				IsServiceable: false,
				Services:      []models.ServiceV2{},
				Capabilities:  make(map[string]interface{}),
				Error:         &errorMsg,
				ResponseTime:  0,
			}

			// Always add to all partners list
			allPartners = append(allPartners, errorPartnerResponse)

		} else if result.Result != nil {
			// Convert partner result to V2 response using database info
			partnerID := ""
			if result.Result != nil && result.Result.PartnerID != nil {
				partnerID = result.Result.PartnerID.String()
			} else if result.PartnerInfo != nil && result.PartnerInfo.PartnerID != nil {
				partnerID = result.PartnerInfo.PartnerID.String()
			} else {
				partnerID = "unknown"
			}
			partnerCode := result.PartnerInfo.PartnerCode

			partnerResponse := models.PartnerV2Response{
				PartnerID:     partnerID,
				PartnerCode:   partnerCode,
				PartnerName:   "",  // No partner name in database
				Rating:        0.0, // No rating in database
				IsServiceable: result.Result.IsServiceable,
				Services:      result.Result.Services,
				Capabilities:  result.Result.Capabilities,
				ResponseTime:  result.Result.ResponseTime,
			}

			// Add error if present
			if result.Result.ErrorMessage != nil {
				partnerResponse.Error = result.Result.ErrorMessage
			}

			// Always add to all partners list
			allPartners = append(allPartners, partnerResponse)

			// Add to serviceable partners only if serviceable
			if result.Result.IsServiceable {
				serviceablePartners = append(serviceablePartners, partnerResponse)
				serviceableCount++
			}
		}
	}

	// Check if we have any errors (validation errors, partner API errors, etc.)
	hasErrors := false
	var errorMessage string
	for _, result := range partnerResults {
		if result.Error != nil {
			hasErrors = true
			errorMessage = result.Error.Error()
			break
		}
		// Also check for embedded error messages in the result
		if result.Result != nil && result.Result.ErrorMessage != nil {
			hasErrors = true
			errorMessage = *result.Result.ErrorMessage
			break
		}
	}

	// Determine success and which partners to return based on returnOnlyServiceable setting
	isSuccess := serviceableCount > 0
	var partnersToReturn []models.PartnerV2Response

	if s.returnOnlyServiceable {
		// When returnOnlyServiceable=true, only return serviceable partners
		if isSuccess && !hasErrors {
			// Return only serviceable partners with success=true (no errors)
			partnersToReturn = serviceablePartners
		} else {
			// Any errors or no serviceable partners = empty array
			partnersToReturn = []models.PartnerV2Response{}
		}
	} else {
		// When returnOnlyServiceable=false, follow old logic
		if isSuccess {
			// When success=true, return only serviceable partners
			partnersToReturn = serviceablePartners
		} else {
			// When success=false, return all partners (including non-serviceable ones and errors)
			partnersToReturn = allPartners
		}
	}

	// Determine success based on returnOnlyServiceable setting
	var success bool
	if s.returnOnlyServiceable {
		// When returnOnlyServiceable=true, success only if serviceable partners AND no errors
		success = isSuccess && !hasErrors
	} else {
		// When returnOnlyServiceable=false, success based on serviceable partners only (old behavior)
		success = isSuccess
	}

	// Build response
	response := &models.ServiceabilityV2Response{
		Success:  success,
		Partners: partnersToReturn,
		Metadata: &models.V2ResponseMetadata{
			TotalPartners:    len(partnerResults),
			ServiceableCount: serviceableCount,
			Filters: models.V2Filters{
				CountryCode:    req.CountryCode,
				ParcelCategory: req.ParcelCategory,
				ProductType:    req.ProductType,
			},
		},
	}

	// Add error message when returnOnlyServiceable=true and there are errors or no serviceable partners
	if s.returnOnlyServiceable && (!isSuccess || hasErrors) && len(partnerResults) > 0 {
		if hasErrors {
			// Classify the error type and return appropriate error code
			errorCode := s.classifyErrorType(errorMessage)
			response.Error = &models.ErrorResponse{
				Code:    errorCode,
				Message: errorMessage,
			}
		} else {
			// No serviceable partners found
			response.Error = &models.ErrorResponse{
				Code:    "NO_SERVICEABLE_PARTNERS",
				Message: "No serviceable partners found for the given request",
			}
		}
	}

	return response
}

// classifyErrorType determines the appropriate error code based on the error message
func (s *serviceabilityOrchestrator) classifyErrorType(errorMessage string) string {
	// Check for specific error patterns - order matters (most specific first)
	if strings.Contains(errorMessage, "POSTAL_CODE_NOT_FOUND") {
		return "POSTAL_CODE_NOT_FOUND"
	}
	if strings.Contains(errorMessage, "country code not found") {
		return "POSTAL_CODE_NOT_FOUND"
	}
	if strings.Contains(errorMessage, "postal code not found") {
		return "POSTAL_CODE_NOT_FOUND"
	}
	if strings.Contains(errorMessage, "Postal code") && strings.Contains(errorMessage, "does not exist") {
		return "POSTAL_CODE_NOT_FOUND"
	}
	if strings.Contains(errorMessage, "VALIDATION_ERROR") || strings.Contains(errorMessage, "validation failed") {
		return "VALIDATION_ERROR"
	}
	if strings.Contains(errorMessage, "DATABASE_ERROR") || strings.Contains(errorMessage, "database") {
		return "DATABASE_ERROR"
	}
	if strings.Contains(errorMessage, "timeout") || strings.Contains(errorMessage, "TIMEOUT") {
		return "TIMEOUT_ERROR"
	}
	if strings.Contains(errorMessage, "unavailable") || strings.Contains(errorMessage, "UNAVAILABLE") {
		return "SERVICE_UNAVAILABLE"
	}
	if strings.Contains(errorMessage, "invalid") || strings.Contains(errorMessage, "INVALID") {
		return "INVALID_REQUEST"
	}
	if strings.Contains(errorMessage, "not found") {
		return "NOT_FOUND"
	}

	// Default to PARTNER_ERROR for actual partner-specific errors
	return "PARTNER_ERROR"
}

// validateV2Request validates the V2 serviceability request
func (s *serviceabilityOrchestrator) validateV2Request(req *models.ServiceabilityV2Request) error {
	// Check postal code requirements
	hasSingleCode := req.PostalCode != nil && *req.PostalCode != ""
	hasSourceDestination := req.SourcePostalCode != nil && *req.SourcePostalCode != "" &&
		req.DestinationPostalCode != nil && *req.DestinationPostalCode != ""
	hasPostalCodeAsDestination := req.PostalCode != nil && *req.PostalCode != "" &&
		req.SourcePostalCode != nil && *req.SourcePostalCode != ""

	if !hasSingleCode && !hasSourceDestination && !hasPostalCodeAsDestination {
		return errors.ErrMissingRequiredField("postal_code or source_postal_code/destination_postal_code")
	}

	// Don't allow conflicting postal code configurations
	if hasSingleCode && hasSourceDestination {
		return errors.ErrInvalidRequest("provide either postal_code only or source/destination postal codes, not both")
	}

	if hasPostalCodeAsDestination && req.DestinationPostalCode != nil && *req.DestinationPostalCode != "" {
		return errors.ErrInvalidRequest("when using postal_code as destination with source_postal_code, do not provide destination_postal_code")
	}

	return nil
}

// getEligiblePartners filters partners based on request attributes (e.g., parcel category)
func (s *serviceabilityOrchestrator) getEligiblePartners(ctx context.Context, req *models.ServiceabilityV2Request) ([]DatabasePartnerInfo, error) {
	// Get all supported partners from factory
	allSupportedPartners := s.partnerFactory.GetSupportedPartners()

	// If no parcel category specified, return all supported partners as basic info
	if req.ParcelCategory == nil || *req.ParcelCategory == "" {
		partnerInfos := make([]DatabasePartnerInfo, 0, len(allSupportedPartners))
		for _, partnerCode := range allSupportedPartners {
			partnerInfos = append(partnerInfos, DatabasePartnerInfo{
				PartnerCode: partnerCode,
				PartnerID:   nil, // No database info available
			})
		}
		return partnerInfos, nil
	}

	// If partner attribute mapping repository is not available, return all supported partners
	// Use a defensive approach to handle both nil interface and typed nil pointers
	if s.partnerAttributeMapRepo == nil {
		s.logger.WithFields(logrus.Fields{
			"component":       "serviceability_orchestrator",
			"total_partners":  len(allSupportedPartners),
			"parcel_category": *req.ParcelCategory,
		}).Warn("Partner attribute mapping repository not available, returning all supported partners")

		partnerInfos := make([]DatabasePartnerInfo, 0, len(allSupportedPartners))
		for _, partnerCode := range allSupportedPartners {
			partnerInfos = append(partnerInfos, DatabasePartnerInfo{
				PartnerCode: partnerCode,
				PartnerID:   nil, // No database info available
			})
		}
		return partnerInfos, nil
	}

	// Get partners that support the specific parcel category from database
	// Use defensive programming to handle potential nil interface issues
	var eligiblePartnersByCategory []models.PartnerAttributeMap
	var err error

	// Log the start of the database query
	s.logger.WithFields(logrus.Fields{
		"component":       "serviceability_orchestrator",
		"parcel_category": *req.ParcelCategory,
		"total_partners":  len(allSupportedPartners),
	}).Info("Starting database query for partners by attribute")

	// Attempt to call the repository method with error recovery
	func() {
		defer func() {
			if r := recover(); r != nil {
				// If we panic due to nil pointer, treat as repository unavailable
				s.logger.WithFields(logrus.Fields{
					"component":       "serviceability_orchestrator",
					"parcel_category": *req.ParcelCategory,
					"error":           "repository method panic - treating as unavailable",
					"panic_info":      fmt.Sprintf("%v", r),
				}).Warn("Repository method panicked, falling back to all supported partners")
				err = fmt.Errorf("repository unavailable due to panic: %v", r)
			}
		}()

		// Add timing measurement
		start := time.Now()
		eligiblePartnersByCategory, err = s.partnerAttributeMapRepo.GetPartnerInfoByAttribute(ctx, *req.ParcelCategory)
		duration := time.Since(start)

		s.logger.WithFields(logrus.Fields{
			"component":       "serviceability_orchestrator",
			"parcel_category": *req.ParcelCategory,
			"query_duration":  duration,
			"partners_found":  len(eligiblePartnersByCategory),
		}).Info("Database query completed")
	}()
	if err != nil {
		// If error getting partners by attribute, log but continue with all partners
		// This ensures backward compatibility
		s.logger.WithFields(logrus.Fields{
			"component":       "serviceability_orchestrator",
			"parcel_category": *req.ParcelCategory,
			"error":           err.Error(),
			"total_partners":  len(allSupportedPartners),
		}).Warn("Failed to get partners by parcel category, returning all supported partners")

		partnerInfos := make([]DatabasePartnerInfo, 0, len(allSupportedPartners))
		for _, partnerCode := range allSupportedPartners {
			partnerInfos = append(partnerInfos, DatabasePartnerInfo{
				PartnerCode: partnerCode,
				PartnerID:   nil, // No database info available
			})
		}
		return partnerInfos, nil
	}

	// Filter to only include partners that are both:
	// 1. Supported by the factory (have adapters available)
	// 2. Have the required parcel category mapping
	eligiblePartners := make([]DatabasePartnerInfo, 0)
	supportedPartnerSet := make(map[string]bool)

	// Create a set of supported partners for quick lookup
	for _, partner := range allSupportedPartners {
		supportedPartnerSet[partner] = true
	}

	// Only include partners that are both supported and have the attribute mapping
	for _, mapping := range eligiblePartnersByCategory {
		if supportedPartnerSet[mapping.PartnerCode] {
			eligiblePartners = append(eligiblePartners, DatabasePartnerInfo{
				PartnerCode: mapping.PartnerCode,
				PartnerID:   mapping.PartnerID,
			})
		}
	}

	return eligiblePartners, nil
}

// partnerResult represents the result of checking serviceability with a single partner
type partnerResult struct {
	PartnerCode string
	Result      *common.PartnerServiceabilityResult
	Error       error
	PartnerInfo *DatabasePartnerInfo // Added field to store partner database info
}

// DatabasePartnerInfo holds partner information from the database
type DatabasePartnerInfo struct {
	PartnerID   *uuid.UUID `json:"partner_id"`
	PartnerCode string     `json:"partner_code"`
}
