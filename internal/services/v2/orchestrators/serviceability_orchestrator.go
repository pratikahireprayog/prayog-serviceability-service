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
	s.logger.WithFields(logrus.Fields{
		"component":        "serviceability_orchestrator",
		"action":           "check_serviceability",
		"source_postal":    req.SourcePostalCode,
		"dest_postal":      req.DestinationPostalCode,
		"parcel_category":  req.ParcelCategory,
		"country_code":     req.CountryCode,
		"product_type":     req.ProductType,
	}).Info("Starting V2 serviceability check")

	// Validate input request
	if req == nil {
		return nil, errors.ErrInvalidRequest("request cannot be nil")
	}

	if err := s.validateV2Request(req); err != nil {
		s.logger.WithFields(logrus.Fields{
			"component": "serviceability_orchestrator",
			"error":     err.Error(),
		}).Error("Request validation failed")
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	// Set timeout context
	timeoutCtx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	// Get eligible partners based on parcel category filtering
	eligiblePartners, err := s.getEligiblePartners(timeoutCtx, req)
	if err != nil {
		s.logger.WithFields(logrus.Fields{
			"component": "serviceability_orchestrator",
			"error":     err.Error(),
		}).Error("Failed to get eligible partners")
		return nil, fmt.Errorf("failed to get eligible partners: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"component":       "serviceability_orchestrator",
		"eligible_partners": len(eligiblePartners),
		"partner_codes":   func() []string {
			codes := make([]string, len(eligiblePartners))
			for i, partner := range eligiblePartners {
				codes[i] = partner.PartnerCode
			}
			return codes
		}(),
	}).Info("Found eligible partners, starting serviceability checks")

	// Check serviceability with all eligible partners concurrently
	partnerResults := s.checkWithPartners(timeoutCtx, req, eligiblePartners)

	if len(eligiblePartners) == 0 {
		s.logger.WithFields(logrus.Fields{
			"component": "serviceability_orchestrator",
		}).Info("No eligible partners found, returning empty response")
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

	s.logger.WithFields(logrus.Fields{
		"component":        "serviceability_orchestrator",
		"action":           "check_serviceability",
		"total_partners":   len(partnerResults),
		"serviceable_count": len(response.Partners),
		"success":          response.Success,
	}).Info("V2 serviceability check completed")

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
	s.logger.WithFields(logrus.Fields{
		"component":       "serviceability_orchestrator",
		"action":          "check_with_partners",
		"total_partners":  len(partnerInfos),
		"partner_codes":   func() []string {
			codes := make([]string, len(partnerInfos))
			for i, info := range partnerInfos {
				codes[i] = info.PartnerCode
			}
			return codes
		}(),
	}).Info("Starting concurrent partner serviceability checks")

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

	s.logger.WithFields(logrus.Fields{
		"component":       "serviceability_orchestrator",
		"action":          "check_with_partners",
		"total_partners":  len(partnerInfos),
		"completed":       len(results),
	}).Info("Completed concurrent partner serviceability checks")

	return results
}

// checkWithPartner checks serviceability with a single partner
func (s *serviceabilityOrchestrator) checkWithPartner(ctx context.Context, req *models.ServiceabilityV2Request, info DatabasePartnerInfo) partnerResult {
	s.logger.WithFields(logrus.Fields{
		"component":    "serviceability_orchestrator",
		"action":       "check_with_partner",
		"partner_code": info.PartnerCode,
		"partner_id":   info.PartnerID,
	}).Info("Starting partner serviceability check")

	adapter, exists := s.partnerFactory.GetAdapter(info.PartnerCode)
	if !exists {
		s.logger.WithFields(logrus.Fields{
			"component":    "serviceability_orchestrator",
			"partner_code": info.PartnerCode,
		}).Error("Partner adapter not found")
		return partnerResult{
			PartnerCode: info.PartnerCode,
			Error:       errors.ErrPartnerNotFound(info.PartnerCode),
		}
	}

	s.logger.WithFields(logrus.Fields{
		"component":    "serviceability_orchestrator",
		"partner_code": info.PartnerCode,
		"adapter_type": fmt.Sprintf("%T", adapter),
	}).Info("Partner adapter found, checking health")

	// Check if adapter is healthy
	if !adapter.IsHealthy(ctx) {
		s.logger.WithFields(logrus.Fields{
			"component":    "serviceability_orchestrator",
			"partner_code": info.PartnerCode,
		}).Error("Partner adapter is not healthy")
		return partnerResult{
			PartnerCode: info.PartnerCode,
			Error:       errors.ErrPartnerUnavailable(info.PartnerCode),
		}
	}

	s.logger.WithFields(logrus.Fields{
		"component":    "serviceability_orchestrator",
		"partner_code": info.PartnerCode,
	}).Info("Partner adapter is healthy, calling CheckServiceability")

	// Call the adapter
	result, err := adapter.CheckServiceability(ctx, req, common.PartnerInfo{PartnerID: info.PartnerID, PartnerCode: info.PartnerCode})
	if err != nil {
		s.logger.WithFields(logrus.Fields{
			"component":    "serviceability_orchestrator",
			"partner_code": info.PartnerCode,
			"error":        err.Error(),
		}).Error("Partner serviceability check failed")
		return partnerResult{
			PartnerCode: info.PartnerCode, // No longer needed since adapter sets it
			Error:       err,
		}
	}

	s.logger.WithFields(logrus.Fields{
		"component":    "serviceability_orchestrator",
		"partner_code": info.PartnerCode,
		"has_services": len(result.Services) > 0,
		"has_capabilities": len(result.Capabilities) > 0,
		"has_error":    result.Error != nil,
		"error_message": result.ErrorMessage,
	}).Info("Partner serviceability check completed successfully")

	return partnerResult{
		PartnerCode: info.PartnerCode,
		Result:      result,
	}
}

// buildV2Response builds the V2 response from partner results
func (s *serviceabilityOrchestrator) buildV2Response(partnerResults []partnerResult, req *models.ServiceabilityV2Request) *models.ServiceabilityV2Response {
	serviceablePartners := make([]models.PartnerV2Response, 0)
	serviceableCount := 0

	// Collect address information
	var hubLocationInfo *models.HubLocationInfo
	var sourceCountryCode, destinationCountryCode string
	// TODO: Remove the other/common operations from the partner specific code (DHL) and keep it out of that so that can be used for any workflows not only for the international
	//TODO: Remove the international code out from the partner specific code (DHL) and structure the code or files in such a way so that other workflows can be also writtern and can consume multiple partner as well
	// Process each partner result
	for _, result := range partnerResults {
		// Extract address information from partner metadata if available
		if result.Result != nil && result.Result.Metadata != nil {
			// Extract hub location info if available
			if hubInfo, exists := result.Result.Metadata["hub_info"]; exists {
				if hubLocationData, ok := hubInfo.(*models.HubLocationInfo); ok {
					hubLocationInfo = hubLocationData
				}
			}

			// Extract country codes if available
			if sourceCC, exists := result.Result.Metadata["source_country_code"]; exists {
				if ccStr, ok := sourceCC.(string); ok {
					sourceCountryCode = ccStr
				}
			}
			if destCC, exists := result.Result.Metadata["destination_country_code"]; exists {
				if ccStr, ok := destCC.(string); ok {
					destinationCountryCode = ccStr
				}
			}
		}

		if result.Error != nil {
			// Handle partner errors (API failures, timeouts, etc.)
			// Don't add error responses to the partners array
			// They will be excluded from the final response

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

			// Extract hub_details from metadata if available
			var hubDetails interface{}
			if result.Result.Metadata != nil {
				if hubDetailsValue, exists := result.Result.Metadata["hub_details"]; exists {
					hubDetails = hubDetailsValue
				}
			}

			partnerResponse := models.PartnerV2Response{
				PartnerID:       partnerID,
				PartnerCode:     partnerCode,
				PartnerName:     "",  // No partner name in database
				Rating:          0.0, // No rating in database
				Services:        result.Result.Services,
				PartnerServices: result.Result.PartnerServices,
				Capabilities:    result.Result.Capabilities,
				ResponseTime:    result.Result.ResponseTime,
				Metadata:        result.Result.Metadata,
				HubDetails:      hubDetails,
			}

			// Add to serviceable partners if they have services, capabilities, or metadata
			// AND no error message (partners with errors are not serviceable)
			hasServices := len(result.Result.Services) > 0
			hasCapabilities := len(result.Result.Capabilities) > 0
			hasMetadata := len(result.Result.Metadata) > 0
			hasError := result.Result.ErrorMessage != nil
			
			isServiceable := hasServices || hasCapabilities || hasMetadata
			
			if isServiceable && !hasError {
				serviceablePartners = append(serviceablePartners, partnerResponse)
				serviceableCount++
			}
			// Non-serviceable partners (with errors or no services/capabilities/metadata) are excluded from the response
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

	// Determine success and which partners to return
	isSuccess := serviceableCount > 0
	var partnersToReturn []models.PartnerV2Response

	// Only return serviceable partners (partners with services or capabilities)
	// Non-serviceable partners (with errors or no services) are excluded from the response
	partnersToReturn = serviceablePartners

	// Build response
	response := &models.ServiceabilityV2Response{
		Success:  isSuccess,
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

	// Add address information if available
	s.populateAddressInformation(response, req, hubLocationInfo, sourceCountryCode, destinationCountryCode)

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

	// Log the supported partners for debugging
	s.logger.WithFields(logrus.Fields{
		"component":           "serviceability_orchestrator",
		"supported_partners":  allSupportedPartners,
		"supported_partner_set": supportedPartnerSet,
	}).Info("Factory supported partners")

	// Only include partners that are both supported and have the attribute mapping
	for _, mapping := range eligiblePartnersByCategory {
		// Check if the database partner code is supported by the factory
		// We need to map database codes to implementation codes
		implCode := getImplementationCode(mapping.PartnerCode)
		s.logger.WithFields(logrus.Fields{
			"component":           "serviceability_orchestrator",
			"db_partner_code":     mapping.PartnerCode,
			"impl_code":           implCode,
			"is_supported":        supportedPartnerSet[implCode],
		}).Info("Checking partner support")

		if supportedPartnerSet[implCode] {
			eligiblePartners = append(eligiblePartners, DatabasePartnerInfo{
				PartnerCode: mapping.PartnerCode,
				PartnerID:   mapping.PartnerID,
			})
		}
	}

	return eligiblePartners, nil
}

// getImplementationCode returns the implementation code for a given database partner code
func getImplementationCode(dbPartnerCode string) string {
	// This mapping should match the one in the factory
	adapterImplementationMap := map[string]string{
		"smile_ecomm":       "smile_ecom",
		"smile_ecom":        "smile_ecom",
		"shipyaari":         "shipyaari",
		"smile_courier":     "smile_courier",
		"dhl":               "dhl",
		"smile_cargo":       "smile_cargo",
		"smile_hyperlocal":  "delcaper",
	}

	if implCode, exists := adapterImplementationMap[dbPartnerCode]; exists {
		return implCode
	}
	return dbPartnerCode // Return original code if no mapping exists
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

// populateAddressInformation populates address-related fields in the response
func (s *serviceabilityOrchestrator) populateAddressInformation(
	response *models.ServiceabilityV2Response,
	req *models.ServiceabilityV2Request,
	hubLocationInfo *models.HubLocationInfo,
	sourceCountryCode, destinationCountryCode string,
) {
	// Populate source address
	if req.SourcePostalCode != nil && sourceCountryCode != "" {
		response.SourceAddress = &models.AddressInfo{
			PostalCode:  *req.SourcePostalCode,
			CountryCode: sourceCountryCode,
		}
	}

	// Populate destination address
	if req.DestinationPostalCode != nil && destinationCountryCode != "" {
		response.DestinationAddress = &models.AddressInfo{
			PostalCode:  *req.DestinationPostalCode,
			CountryCode: destinationCountryCode,
		}
	}

	// Populate detailed addresses array with hub information
	if hubLocationInfo != nil && hubLocationInfo.HubContactInfo != nil {
		hubAddress := s.buildHubDetailedAddress(hubLocationInfo)
		if hubAddress != nil {
			response.Addresses = []models.DetailedAddress{*hubAddress}
		}
	}
}

// buildHubDetailedAddress builds a DetailedAddress from hub location information
func (s *serviceabilityOrchestrator) buildHubDetailedAddress(hubLocationInfo *models.HubLocationInfo) *models.DetailedAddress {
	if hubLocationInfo == nil || hubLocationInfo.HubContactInfo == nil {
		return nil
	}

	hubContact := hubLocationInfo.HubContactInfo
	
	// Build the detailed address with hub information
	address := &models.DetailedAddress{
		Type:        "INTERNATIONAL_HUB_ADDRESS",
		AddressName: "WAREHOUSE",
	}

	// Set postal code from hub info
	if hubLocationInfo.HubInfo != nil && hubLocationInfo.HubInfo.PostalCode != nil {
		address.Zip = fmt.Sprintf("%d", *hubLocationInfo.HubInfo.PostalCode)
	}

	// Set contact information with safe string handling
	if hubContact.ContactPersonName != nil {
		address.Name = *hubContact.ContactPersonName
	}
	if hubContact.ContactPersonPhone != nil {
		address.Phone = *hubContact.ContactPersonPhone
	}
	if hubContact.ContactPersonEmail != nil {
		address.Email = *hubContact.ContactPersonEmail
	}

	// Set address information with safe string handling
	if hubContact.Street != nil {
		address.Street = *hubContact.Street
	}
	if hubContact.Landmark != nil {
		address.Landmark = *hubContact.Landmark
	}
	if hubContact.City != nil {
		address.City = *hubContact.City
	}
	if hubContact.State != nil {
		address.State = *hubContact.State
	}
	if hubContact.Country != nil {
		address.Country = *hubContact.Country
	}

	// Set coordinates
	address.Latitude = hubContact.Lat
	address.Longitude = hubContact.Lng

	return address
}
