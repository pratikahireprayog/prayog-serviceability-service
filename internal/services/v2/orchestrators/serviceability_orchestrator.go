package orchestrators

import (
	"context"
	"fmt"
	"sync"
	"time"

	"prayog-serviceability-service/internal/services/v2/partners/common"
	"prayog-serviceability-service/internal/services/v2/partners/factory"
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
		return nil, fmt.Errorf("request cannot be nil")
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
		return nil, fmt.Errorf("bulk request cannot be nil or empty")
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
			result := s.checkWithPartner(ctx, req, info.PartnerCode)
			// Store partner database info in result for later use
			result.PartnerInfo = &info
			results[index] = result
		}(i, partnerInfo)
	}

	wg.Wait()
	return results
}

// checkWithPartner checks serviceability with a single partner
func (s *serviceabilityOrchestrator) checkWithPartner(ctx context.Context, req *models.ServiceabilityV2Request, partnerCode string) partnerResult {
	adapter, exists := s.partnerFactory.GetAdapter(partnerCode)
	if !exists {
		return partnerResult{
			PartnerCode: partnerCode,
			Error:       fmt.Errorf("partner adapter not found for code: %s", partnerCode),
		}
	}

	// Check if adapter is healthy
	if !adapter.IsHealthy(ctx) {
		return partnerResult{
			PartnerCode: partnerCode,
			Error:       fmt.Errorf("partner %s is not healthy", partnerCode),
		}
	}

	// Call the adapter
	result, err := adapter.CheckServiceability(ctx, req)
	if err != nil {
		return partnerResult{
			PartnerCode: partnerCode,
			Error:       err,
		}
	}

	return partnerResult{
		PartnerCode: partnerCode,
		Result:      result,
	}
}

// buildV2Response builds the V2 response from partner results
func (s *serviceabilityOrchestrator) buildV2Response(partnerResults []partnerResult, req *models.ServiceabilityV2Request) *models.ServiceabilityV2Response {
	partners := make([]models.PartnerV2Response, 0)
	serviceableCount := 0

	// Process each partner result
	for _, result := range partnerResults {
		if result.Error != nil {
			// Add error partner response using database info
			errorMsg := result.Error.Error()

			// Use partner info from database if available
			partnerID := result.PartnerCode // Default to partner code
			if result.PartnerInfo != nil && result.PartnerInfo.PartnerID != nil {
				partnerID = result.PartnerInfo.PartnerID.String()
			}

			// Filter out error partners if configuration is enabled
			if !s.returnOnlyServiceable {
				partners = append(partners, models.PartnerV2Response{
					PartnerID:     partnerID,
					PartnerCode:   result.PartnerCode,
					PartnerName:   "",  // No partner name in database
					Rating:        0.0, // No rating in database
					IsServiceable: false,
					Services:      []models.ServiceV2{},
					Capabilities:  make(map[string]interface{}),
					Error:         &errorMsg,
					ResponseTime:  0,
				})
			}
		} else if result.Result != nil {
			// Convert partner result to V2 response using database info
			partnerID := result.Result.PartnerCode // Default to partner code
			if result.PartnerInfo != nil && result.PartnerInfo.PartnerID != nil {
				partnerID = result.PartnerInfo.PartnerID.String()
			}

			partnerResponse := models.PartnerV2Response{
				PartnerID:     partnerID,
				PartnerCode:   result.Result.PartnerCode,
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

			// Filter out non-serviceable partners if configuration is enabled
			if !s.returnOnlyServiceable || result.Result.IsServiceable {
				partners = append(partners, partnerResponse)
			}

			if result.Result.IsServiceable {
				serviceableCount++
			}
		}
	}

	// Build response
	response := &models.ServiceabilityV2Response{
		Success:  serviceableCount > 0,
		Partners: partners,
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

	return response
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
		return fmt.Errorf("must provide either: postal_code only, or both source_postal_code and destination_postal_code, or postal_code (destination) with source_postal_code")
	}

	// Don't allow conflicting postal code configurations
	if hasSingleCode && hasSourceDestination {
		return fmt.Errorf("provide either postal_code only or source/destination postal codes, not both")
	}

	if hasPostalCodeAsDestination && req.DestinationPostalCode != nil && *req.DestinationPostalCode != "" {
		return fmt.Errorf("when using postal_code as destination with source_postal_code, do not provide destination_postal_code")
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
	eligiblePartnersByCategory, err := s.partnerAttributeMapRepo.GetPartnerInfoByAttribute(ctx, *req.ParcelCategory)
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
