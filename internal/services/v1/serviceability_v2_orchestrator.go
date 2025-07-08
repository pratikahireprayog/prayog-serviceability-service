package services

import (
	"context"
	"fmt"
	"sync"
	"time"

	"prayog-serviceability-service/internal/shared/interfaces/v1"
	"prayog-serviceability-service/internal/shared/models/v1"
	"prayog-serviceability-service/internal/shared/repositories/v1"
)

// ServiceabilityV2Orchestrator defines the interface for V2 serviceability orchestration
type ServiceabilityV2Orchestrator interface {
	CheckServiceability(ctx context.Context, req *models.ServiceabilityV2Request) (*models.ServiceabilityV2Response, error)
	BulkCheckServiceability(ctx context.Context, req *models.BulkServiceabilityV2Request) (*models.BulkServiceabilityV2Response, error)
}

// serviceabilityV2Orchestrator implements ServiceabilityV2Orchestrator
type serviceabilityV2Orchestrator struct {
	partnerAdapterFactory interfaces.PartnerAdapterFactory
	partnerAttributeRepo  repositories.PartnerAttributeMapRepository
}

// NewServiceabilityV2Orchestrator creates a new V2 serviceability orchestrator
func NewServiceabilityV2Orchestrator(
	partnerAdapterFactory interfaces.PartnerAdapterFactory,
	partnerAttributeRepo repositories.PartnerAttributeMapRepository,
) ServiceabilityV2Orchestrator {
	return &serviceabilityV2Orchestrator{
		partnerAdapterFactory: partnerAdapterFactory,
		partnerAttributeRepo:  partnerAttributeRepo,
	}
}

// CheckServiceability orchestrates the V2 serviceability check process
func (s *serviceabilityV2Orchestrator) CheckServiceability(ctx context.Context, req *models.ServiceabilityV2Request) (*models.ServiceabilityV2Response, error) {
	startTime := time.Now()

	// Validate input request
	if req == nil {
		return &models.ServiceabilityV2Response{
			Success: false,
			Error: &models.ErrorResponse{
				Code:    "INVALID_REQUEST",
				Message: "Request cannot be nil",
			},
		}, nil
	}

	if err := s.validateV2Request(req); err != nil {
		return &models.ServiceabilityV2Response{
			Success: false,
			Error: &models.ErrorResponse{
				Code:    "VALIDATION_ERROR",
				Message: err.Error(),
			},
		}, nil
	}

	// Step 1: Determine eligible partners based on parcel category
	eligiblePartners, err := s.getEligiblePartners(ctx, req)
	if err != nil {
		return &models.ServiceabilityV2Response{
			Success: false,
			Error: &models.ErrorResponse{
				Code:    "PARTNER_FILTERING_ERROR",
				Message: fmt.Sprintf("Failed to determine eligible partners: %v", err),
			},
		}, nil
	}

	if len(eligiblePartners) == 0 {
		return &models.ServiceabilityV2Response{
			Success:  true,
			Partners: []models.PartnerV2Response{},
			Metadata: &models.V2ResponseMetadata{
				TotalPartners:    0,
				ServiceableCount: 0,
				ProcessingTime:   time.Since(startTime),
				Filters: models.V2Filters{
					ParcelCategory: req.ParcelCategory,
					ProductType:    req.ProductType,
					CountryCode:    req.CountryCode,
				},
				EligiblePartners: []string{},
			},
		}, nil
	}

	// Step 2: Execute concurrent partner API calls
	partnerResults := s.executePartnerCallsConcurrently(ctx, req, eligiblePartners)

	// Step 3: Transform results to V2 response format
	partnerResponses := s.transformPartnerResults(partnerResults)

	// Step 4: Build response with metadata
	response := &models.ServiceabilityV2Response{
		Success:  true,
		Partners: partnerResponses,
		Metadata: &models.V2ResponseMetadata{
			TotalPartners:    len(eligiblePartners),
			ServiceableCount: s.countServiceablePartners(partnerResponses),
			ProcessingTime:   time.Since(startTime),
			Filters: models.V2Filters{
				ParcelCategory: req.ParcelCategory,
				ProductType:    req.ProductType,
				CountryCode:    req.CountryCode,
			},
			EligiblePartners: eligiblePartners,
		},
	}

	return response, nil
}

// BulkCheckServiceability orchestrates bulk V2 serviceability checks
func (s *serviceabilityV2Orchestrator) BulkCheckServiceability(ctx context.Context, req *models.BulkServiceabilityV2Request) (*models.BulkServiceabilityV2Response, error) {
	startTime := time.Now()

	if req == nil || len(req.Requests) == 0 {
		return &models.BulkServiceabilityV2Response{
			Success: false,
			Error: &models.ErrorResponse{
				Code:    "INVALID_REQUEST",
				Message: "Bulk request cannot be nil or empty",
			},
		}, nil
	}

	responses := make([]models.ServiceabilityV2Response, 0, len(req.Requests))
	successfulCount := 0
	failedCount := 0

	// Process each individual request
	for i, individualReq := range req.Requests {
		response, err := s.CheckServiceability(ctx, &individualReq)
		if err != nil {
			// Handle individual request failures
			failedCount++
			errorMsg := fmt.Sprintf("Request %d failed: %v", i+1, err)

			responses = append(responses, models.ServiceabilityV2Response{
				Success:  false,
				Partners: []models.PartnerV2Response{},
				Error: &models.ErrorResponse{
					Code:    "REQUEST_FAILED",
					Message: errorMsg,
				},
			})
		} else {
			if response.Success {
				successfulCount++
			} else {
				failedCount++
			}
			responses = append(responses, *response)
		}
	}

	bulkResponse := &models.BulkServiceabilityV2Response{
		Success: failedCount == 0,
		Data:    responses,
		Metadata: &models.BulkV2Metadata{
			TotalRequests:   len(req.Requests),
			SuccessfulCount: successfulCount,
			FailedCount:     failedCount,
			ProcessingTime:  time.Since(startTime),
		},
	}

	if failedCount > 0 {
		bulkResponse.Error = &models.ErrorResponse{
			Code:    "PARTIAL_FAILURE",
			Message: fmt.Sprintf("Processed %d requests with %d errors", len(req.Requests), failedCount),
		}
	}

	return bulkResponse, nil
}

// getEligiblePartners determines which partners are eligible based on parcel category
func (s *serviceabilityV2Orchestrator) getEligiblePartners(ctx context.Context, req *models.ServiceabilityV2Request) ([]string, error) {
	var eligiblePartners []string

	// If no parcel category specified, use all enabled partners
	if req.ParcelCategory == nil || *req.ParcelCategory == "" {
		return s.partnerAdapterFactory.GetSupportedPartners(), nil
	}

	// Get partners that support the specified parcel category
	partnerCodes, err := s.partnerAttributeRepo.GetPartnerCodesByAttribute(ctx, *req.ParcelCategory)
	if err != nil {
		return nil, fmt.Errorf("failed to get partners by attribute: %w", err)
	}

	// Filter to only include partners that are enabled in the factory
	supportedPartners := s.partnerAdapterFactory.GetSupportedPartners()
	supportedMap := make(map[string]bool)
	for _, partner := range supportedPartners {
		supportedMap[partner] = true
	}

	for _, partnerCode := range partnerCodes {
		if supportedMap[partnerCode] {
			eligiblePartners = append(eligiblePartners, partnerCode)
		}
	}

	return eligiblePartners, nil
}

// executePartnerCallsConcurrently executes partner API calls concurrently
func (s *serviceabilityV2Orchestrator) executePartnerCallsConcurrently(ctx context.Context, req *models.ServiceabilityV2Request, partnerCodes []string) map[string]*interfaces.PartnerServiceabilityResult {
	results := make(map[string]*interfaces.PartnerServiceabilityResult)
	var wg sync.WaitGroup
	var mu sync.Mutex

	// Create a context with timeout for partner calls
	partnerCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	for _, partnerCode := range partnerCodes {
		wg.Add(1)

		go func(code string) {
			defer wg.Done()

			// Get adapter for this partner
			adapter, err := s.partnerAdapterFactory.CreateAdapter(code)
			if err != nil {
				mu.Lock()
				results[code] = &interfaces.PartnerServiceabilityResult{
					PartnerCode: code,
					PartnerName: code,
					Error:       err,
				}
				mu.Unlock()
				return
			}

			// Execute serviceability check
			result, err := adapter.CheckServiceability(partnerCtx, req)
			if err != nil {
				mu.Lock()
				results[code] = &interfaces.PartnerServiceabilityResult{
					PartnerCode: code,
					PartnerName: adapter.GetPartnerName(),
					Error:       err,
				}
				mu.Unlock()
				return
			}

			mu.Lock()
			results[code] = result
			mu.Unlock()
		}(partnerCode)
	}

	wg.Wait()
	return results
}

// transformPartnerResults transforms partner results to V2 response format
func (s *serviceabilityV2Orchestrator) transformPartnerResults(results map[string]*interfaces.PartnerServiceabilityResult) []models.PartnerV2Response {
	var partnerResponses []models.PartnerV2Response

	for partnerCode, result := range results {
		response := models.PartnerV2Response{
			PartnerID:     partnerCode, // Using code as ID for now
			PartnerCode:   result.PartnerCode,
			PartnerName:   result.PartnerName,
			Rating:        result.Rating,
			IsServiceable: result.IsServiceable,
			Services:      result.Services,
			Capabilities:  result.Capabilities,
			ResponseTime:  result.ResponseTime,
		}

		// Handle errors
		if result.Error != nil || result.ErrorMessage != nil {
			if result.ErrorMessage != nil {
				response.Error = result.ErrorMessage
			} else {
				errorMsg := result.Error.Error()
				response.Error = &errorMsg
			}
		}

		partnerResponses = append(partnerResponses, response)
	}

	return partnerResponses
}

// countServiceablePartners counts how many partners are serviceable
func (s *serviceabilityV2Orchestrator) countServiceablePartners(partners []models.PartnerV2Response) int {
	count := 0
	for _, partner := range partners {
		if partner.IsServiceable {
			count++
		}
	}
	return count
}

// validateV2Request validates the V2 serviceability request
func (s *serviceabilityV2Orchestrator) validateV2Request(req *models.ServiceabilityV2Request) error {
	if req.CountryCode == "" {
		return fmt.Errorf("country code is required")
	}

	if len(req.CountryCode) != 2 {
		return fmt.Errorf("country code must be 2 characters")
	}

	// Must have either a single postal code or pickup/delivery postal codes
	hasSingleCode := req.PostalCode != nil && *req.PostalCode != ""
	hasPickupDelivery := req.PickupPostalCode != nil && *req.PickupPostalCode != "" &&
		req.DeliveryPostalCode != nil && *req.DeliveryPostalCode != ""

	if !hasSingleCode && !hasPickupDelivery {
		return fmt.Errorf("either postal_code or both pickup_postal_code and delivery_postal_code must be provided")
	}

	if hasSingleCode && hasPickupDelivery {
		return fmt.Errorf("provide either postal_code or pickup/delivery postal codes, not both")
	}

	// Validate parcel category if provided
	if req.ParcelCategory != nil && *req.ParcelCategory != "" {
		validCategories := []string{"ecomm", "courier", "cargo", "international", "hyperlocal"}
		isValid := false
		for _, category := range validCategories {
			if *req.ParcelCategory == category {
				isValid = true
				break
			}
		}
		if !isValid {
			return fmt.Errorf("invalid parcel category: %s. Valid categories: %v", *req.ParcelCategory, validCategories)
		}
	}

	return nil
}
