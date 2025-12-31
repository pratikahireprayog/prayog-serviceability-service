package orchestrators

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"os"

	supplyrates "prayog-serviceability-service/internal/services/supply-rates"
	intlstrategy "prayog-serviceability-service/internal/services/v2/orchestrators/strategies"
	cargoStrategy "prayog-serviceability-service/internal/services/v2/orchestrators/strategies/cargo_strategy"
	dstrategy "prayog-serviceability-service/internal/services/v2/orchestrators/strategies/default_strategy"
	newintl "prayog-serviceability-service/internal/services/v2/orchestrators/strategies/international_strategy"
	npPickupDelivery "prayog-serviceability-service/internal/services/v2/orchestrators/strategies/np_extension_with_pickup_and_delivery_strategy"
	smnpstrategy "prayog-serviceability-service/internal/services/v2/orchestrators/strategies/smile_primary_np_extension_strategy"
	"prayog-serviceability-service/internal/services/v2/partners/common"
	"prayog-serviceability-service/internal/services/v2/partners/factory"

	services "prayog-serviceability-service/internal/services/v1/data"

	"prayog-serviceability-service/internal/infrastructure/external/partner_service"
	tenantcontext "prayog-serviceability-service/internal/shared/context"
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
	orchestratorFactory     OrchestratorFactory
	rateClient              *supplyrates.RateClient
	geolocationService 		services.GeolocationService
	partnerServiceClient    *partner_service.PartnerServiceClient

}

// NewServiceabilityOrchestrator creates a new V2 serviceability orchestrator
func NewServiceabilityOrchestrator(
	partnerFactory factory.PartnerAdapterFactory,
	partnerAttributeMapRepo repositories.PartnerAttributeMapRepository,
	timeout time.Duration,
	returnOnlyServiceable bool,
	geolocationService services.GeolocationService,
	partnerServiceClient *partner_service.PartnerServiceClient,
) ServiceabilityOrchestrator {
	// Initialize logger
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)

    s := &serviceabilityOrchestrator{
		partnerFactory:          partnerFactory,
		partnerAttributeMapRepo: partnerAttributeMapRepo,
		timeout:                 timeout,
		returnOnlyServiceable:   returnOnlyServiceable,
		logger:                  logger,
		geolocationService:      geolocationService,
		partnerServiceClient:    partnerServiceClient,
    }
	// Initialize rate client
	s.rateClient = supplyrates.NewRateClient(logger)

    // Register strategies (pattern only). Default delegates to existing flow via executeDefault.
    registry := map[string]StrategyConstructor{
        "default": func() OrchestrationStrategy {
            // Wrapper to match DefaultStrategy signature (without rates pointers)
            executeDefaultWrapper := func(ctx context.Context, req *models.ServiceabilityV2Request) (*models.ServiceabilityV2Response, error) {
                // Create dummy pointers for rates (won't be used in this path)
                var ratesResp *supplyrates.RateQuoteResponse
                var rateErr error
                return s.executeDefault(ctx, req, &ratesResp, &rateErr)
            }
            return &dstrategy.DefaultStrategy{ExecuteFunc: executeDefaultWrapper}
        },
        // Minimal implementation uses partner factory; injected here
        "smile_primary_np_extension": func() OrchestrationStrategy {
            return &smnpstrategy.SmilePrimaryNPExtensionStrategy{PartnerFactory: partnerFactory, Logger: logger}
        },
        // Smile Primary NP extension with pickup (from templates API)
        "smile_primary_np_extension_with_pickup": func() OrchestrationStrategy {
            return &intlstrategy.InternationalWithPickupStrategy{PartnerFactory: partnerFactory, Logger: logger}
        },
        // NP extension with pickup and delivery strategy for product_type=pickup_and_delivery
        "np_extension_with_pickup_and_delivery": func() OrchestrationStrategy {
            return &npPickupDelivery.NPExtensionWithPickupAndDeliveryStrategy{PartnerFactory: partnerFactory, Logger: logger}
        },
        // New international flow using HubOps by-pincode then DHL
        "international": func() OrchestrationStrategy {
            return newintl.NewInternationalStrategy(partnerFactory, geolocationService)
        },
        // Cargo strategy for cargo/freight requests
        "cargo": func() OrchestrationStrategy {
            return &cargoStrategy.CargoStrategy{PartnerFactory: partnerFactory, Logger: logger}
        },
    }
    // Configure Journey Templates base URL from environment (JOURNEY_TEMPLATE_URL), defaulting to sandbox
    baseURL := os.Getenv("JOURNEY_TEMPLATE_URL")
    if baseURL == "" {
        baseURL = "https://sandbox-apis.prayog.io"
    }
    jtClient := NewHTTPJourneyTemplatesClient(baseURL, nil)
    s.orchestratorFactory = NewOrchestratorFactory(jtClient, "default", registry)

    return s
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

	var strat OrchestrationStrategy
	var response *models.ServiceabilityV2Response
	var err error
	var strategyCode string
	// Variables for rates API response (may be set by executeDefault in parallel)
	var ratesResponse *supplyrates.RateQuoteResponse
	var rateErr error
	// PRIORITY 1: Check if specific partners are requested - if yes, detect if they're international
	// Partners array takes priority over parcel_category
	if req != nil && req.Partners != nil && len(req.Partners) > 0 {
		hasInternationalPartners := s.hasInternationalPartners(req.Partners)

		if hasInternationalPartners {
			// Route through InternationalStrategy when international partners are detected
			strat = newintl.NewInternationalStrategy(s.partnerFactory, s.geolocationService)
			s.logger.WithFields(logrus.Fields{
				"component":          "serviceability_orchestrator",
				"strategy":           "international",
				"reason":             "international_partners_detected",
				"requested_partners": func() []string {
					codes := make([]string, len(req.Partners))
					for i, p := range req.Partners {
						codes[i] = p.Code
					}
					return codes
				}(),
			}).Info("Selected international strategy because international partners detected in request (ignoring parcel_category)")

			// Delegate to international strategy immediately (partners priority)
			response, err = strat.Execute(timeoutCtx, req)
		}
	}

	// If no strategy selected yet, continue with other priorities
	if response == nil {
		// PRIORITY 2: Decide strategy based on product_type (only if no partners or non-international partners)
		if req != nil && req.ProductType != nil {
			productType := strings.ToLower(*req.ProductType)
			switch productType {
			case "nba":
				strat = &intlstrategy.InternationalWithPickupStrategy{PartnerFactory: s.partnerFactory, Logger: s.logger}
				strategyCode = "international_with_pickup"
				s.logger.WithFields(logrus.Fields{
					"component":    "serviceability_orchestrator",
					"strategy":     "international_with_pickup",
					"product_type": productType,
				}).Info("Selected strategy based on product_type")
			case "pickup_and_delivery":
				strategyCode = "np_extension_with_pickup_and_delivery"
				strat = &npPickupDelivery.NPExtensionWithPickupAndDeliveryStrategy{PartnerFactory: s.partnerFactory, Logger: s.logger}
				s.logger.WithFields(logrus.Fields{
					"component":    "serviceability_orchestrator",
					"strategy":     "np_extension_with_pickup_and_delivery",
					"product_type": productType,
				}).Info("Selected strategy based on product_type")
			case "cargo":
				strategyCode = "cargo"
				strat = &cargoStrategy.CargoStrategy{PartnerFactory: s.partnerFactory, Logger: s.logger}
				s.logger.WithFields(logrus.Fields{
					"component":    "serviceability_orchestrator",
					"strategy":     "cargo",
					"product_type": productType,
				}).Info("Selected cargo strategy based on product_type")
			}
		}

		// PRIORITY 3: If no product_type strategy found, resolve via templates (by parcel_category)
		if strat == nil && len(req.Partners) > 0 && s.orchestratorFactory != nil {
			resolved, _ := s.orchestratorFactory.Resolve(timeoutCtx, req.ParcelCategory)
			strat = resolved
			if strat != nil {
				strategyCode = strat.Code()
				s.logger.WithFields(logrus.Fields{
					"component":       "serviceability_orchestrator",
					"strategy":        strat.Code(),
					"parcel_category": req.ParcelCategory,
				}).Info("Selected strategy based on parcel_category")
			}
		}

		// PRIORITY 4: Force new international strategy when parcel_category == "international"
		if strat == nil && len(req.Partners) == 0 && req != nil && req.ParcelCategory != nil && strings.ToLower(*req.ParcelCategory) == "international" {
			strat = newintl.NewInternationalStrategy(s.partnerFactory, s.geolocationService)
			strategyCode = "international"
			s.logger.WithFields(logrus.Fields{
				"component":       "serviceability_orchestrator",
				"strategy":        "international",
				"parcel_category": *req.ParcelCategory,
			}).Info("Selected international strategy based on parcel_category")
		}

		// PRIORITY 5: Force cargo strategy when parcel_category == "cargo"
		if strat == nil && len(req.Partners) == 0 && req != nil && req.ParcelCategory != nil && strings.ToLower(*req.ParcelCategory) == "cargo" {
			strat = &cargoStrategy.CargoStrategy{PartnerFactory: s.partnerFactory, Logger: s.logger}
			strategyCode = "cargo"
			s.logger.WithFields(logrus.Fields{
				"component":       "serviceability_orchestrator",
				"strategy":        "cargo",
				"parcel_category": *req.ParcelCategory,
			}).Info("Selected cargo strategy based on parcel_category")
		}

		if strat == nil || strat.Code() == "default" {
			strategyCode = "default"
			s.logger.WithFields(logrus.Fields{
				"component": "serviceability_orchestrator",
				"strategy":  "default",
			}).Info("Falling back to default strategy")
			response, err = s.executeDefault(timeoutCtx, req, &ratesResponse, &rateErr)
		} else {
			// Delegate to non-default strategy
			response, err = strat.Execute(timeoutCtx, req)
		}
	}

	if err != nil {
		return nil, err
	}

	// CENTRALIZED RATE FETCHING: Use rates from parallel call or fetch if not already fetched
	if response != nil && len(response.Partners) > 0 {
		s.logger.WithFields(logrus.Fields{
			"component":          "serviceability_orchestrator",
			"strategy_used":      strategyCode,
			"total_partners":     len(response.Partners),
			"rate_client_ready":  s.rateClient != nil,
			"rates_already_fetched": ratesResponse != nil,
		}).Info("Processing rates for all serviceable partners across strategy")

		// If rates were not fetched in parallel (non-default strategy), fetch them now
		if ratesResponse == nil && rateErr == nil {
			// Filter partners for rate fetching - include all partners (serviceable and non-serviceable with errors)
			// This allows us to get rates even for partners that have serviceability errors
			partnersForRates := make([]models.PartnerV2Response, 0)
			for _, partner := range response.Partners {
				// Include partner if:
				// 1. It's serviceable, OR
				// 2. It has an error but was requested (so we can still try to get rates)
				if partner.IsServiceable || partner.Error != nil {
					partnersForRates = append(partnersForRates, partner)
				}
			}
			s.logger.Info("serviceablePartners", partnersForRates);
			if len(partnersForRates) > 0 {
				ratesResponse, rateErr = s.fetchRatesForPartners(context.Background(), req, partnersForRates)
			}
		}
		
		// Process rates response (either from parallel call or sequential call)
		if rateErr != nil {
			// Rate API call failed - clear services for all partners
			s.logger.WithError(rateErr).Warn("Failed to fetch rates for partners, clearing services for all partners")
			for i := range response.Partners {
				response.Partners[i].Services = []models.ServiceV2{}
				response.Partners[i].PartnerServices = []models.ServiceV2{}
				if response.Partners[i].Metadata == nil {
					response.Partners[i].Metadata = make(map[string]interface{})
				}
				response.Partners[i].Metadata["rate_api_error"] = rateErr.Error()
				response.Partners[i].Metadata["rates_available"] = false
			}
		} else if ratesResponse != nil && ratesResponse.Success {
			// Rates were fetched successfully (either in parallel or sequentially)
			s.logger.WithFields(logrus.Fields{
				"component":        "serviceability_orchestrator",
				"rates_found":      ratesResponse.Metadata.TotalRatesFound,
				"partners_with_rates": len(ratesResponse.Data.SuccessfulResponses),
			}).Info("Successfully fetched rates for partners")

			// Attach rates to partner responses
			s.mergeRatesIntoPartners(response.Partners, ratesResponse)
			// Update metadata to indicate rates are included
			if response.Metadata != nil {
				totalRatesFound := 0
				for _, partner := range response.Partners {
					for _, service := range partner.Services {
						if service.Rate != nil {
							totalRatesFound++
						}
					}
				}
				response.Metadata.RatesIncluded = totalRatesFound > 0
			}
		} else if ratesResponse != nil && !ratesResponse.Success {
			// Rate API returned unsuccessful response - clear services for all partners
			s.logger.WithField("message", ratesResponse.Message).Warn("Rate API returned unsuccessful response, clearing services for all partners")
			for i := range response.Partners {
				response.Partners[i].Services = []models.ServiceV2{}
				response.Partners[i].PartnerServices = []models.ServiceV2{}
				if response.Partners[i].Metadata == nil {
					response.Partners[i].Metadata = make(map[string]interface{})
				}
				response.Partners[i].Metadata["rate_api_error"] = ratesResponse.Message
				response.Partners[i].Metadata["rates_available"] = false
			}
		}
	}

	return response, nil
}

// executeDefault runs the existing default orchestration flow
// ratesResponsePtr and rateErrPtr are optional pointers to store rates API results for parallel execution
func (s *serviceabilityOrchestrator) executeDefault(ctx context.Context, req *models.ServiceabilityV2Request, ratesResponsePtr **supplyrates.RateQuoteResponse, rateErrPtr *error) (*models.ServiceabilityV2Response, error) {
    startTime := time.Now()
    // Get eligible partners based on parcel category filtering
    eligiblePartners, invalidPartners, err := s.getEligiblePartners(ctx, req)
    if err != nil {
        s.logger.WithFields(logrus.Fields{
            "component": "serviceability_orchestrator",
            "error":     err.Error(),
        }).Error("Failed to get eligible partners")
        return nil, fmt.Errorf("failed to get eligible partners: %w", err)
    }

    s.logger.WithFields(logrus.Fields{
        "component":         "serviceability_orchestrator",
        "eligible_partners": len(eligiblePartners),
        "partner_codes": func() []string {
            codes := make([]string, len(eligiblePartners))
            for i, partner := range eligiblePartners {
                codes[i] = partner.PartnerCode
            }
            return codes
        }(),
    }).Info("Found eligible partners, starting serviceability checks")

    // Prepare partners for rates API call (start in parallel with serviceability checks)
    // Note: ratesResponse and rateErr are declared in outer scope (CheckServiceability function)
    var ratesWg sync.WaitGroup
    
    // Start rates API call in parallel with partner checks if rate client is available
    if s.rateClient != nil && len(eligiblePartners) > 0 {
        ratesWg.Add(1)
        go func() {
            defer ratesWg.Done()
            // Convert eligible partners to PartnerV2Response format for rates API
            // We'll use all eligible partners - rates API can handle non-serviceable ones
            partnersForRates := make([]models.PartnerV2Response, 0, len(eligiblePartners))
            for _, partner := range eligiblePartners {
                partnerIDStr := ""
                if partner.PartnerID != nil {
                    partnerIDStr = partner.PartnerID.String()
                }
                partnersForRates = append(partnersForRates, models.PartnerV2Response{
                    PartnerID:   partnerIDStr,
                    PartnerCode: partner.PartnerCode,
                    // Assume serviceable for now - will be filtered later if needed
                    IsServiceable: true,
                })
            }
            
            s.logger.WithFields(logrus.Fields{
                "component":       "serviceability_orchestrator",
                "action":          "fetch_rates_parallel",
                "partners_count":  len(partnersForRates),
            }).Info("Starting rates API call in parallel with partner checks")
            
            resp, err := s.fetchRatesForPartners(ctx, req, partnersForRates)
            
            // Use pointers to set outer scope variables
            if ratesResponsePtr != nil {
                *ratesResponsePtr = resp
            }
            if rateErrPtr != nil {
                *rateErrPtr = err
            }
            
            if err != nil {
                s.logger.WithError(err).Warn("Rates API call completed with error (parallel execution)")
            } else if resp != nil {
                s.logger.WithFields(logrus.Fields{
                    "component":        "serviceability_orchestrator",
                    "action":           "fetch_rates_parallel",
                    "rates_found":       resp.Metadata.TotalRatesFound,
                    "partners_with_rates": len(resp.Data.SuccessfulResponses),
                }).Info("Rates API call completed successfully (parallel execution)")
            }
        }()
    }

    // Check serviceability with all eligible partners concurrently
    partnerResults := s.checkWithPartners(ctx, req, eligiblePartners)
    
    // Wait for rates API call to complete (if it was started)
    if s.rateClient != nil && len(eligiblePartners) > 0 {
        ratesWg.Wait()
    }

    if len(eligiblePartners) == 0 {
        s.logger.WithFields(logrus.Fields{
            "component": "serviceability_orchestrator",
        }).Info("No eligible partners found, returning empty response")
        
        // Calculate total requested partners (valid + invalid)
        totalRequestedPartners := len(invalidPartners)
        if len(req.Partners) > 0 {
            totalRequestedPartners = len(req.Partners)
        }
        
        emptyResponse := &models.ServiceabilityV2Response{
            Success:  false,
            Partners: []models.PartnerV2Response{},
            Metadata: &models.V2ResponseMetadata{
                TotalPartners:    totalRequestedPartners,
                ServiceableCount: 0,
                Filters: models.V2Filters{
                    CountryCode:       req.CountryCode,
                    ParcelCategory:    req.ParcelCategory,
                    ProductType:       req.ProductType,
                    RequestedPartners: req.Partners,
                },
            },
        }
        // Add invalid partners to filters if any
        if len(invalidPartners) > 0 {
            emptyResponse.Metadata.Filters.FailedPartners = invalidPartners
            emptyResponse.Metadata.PartnersFailed = len(invalidPartners)
        }
        return emptyResponse, nil
    }

    // Process results and build response
    response := s.buildV2Response(partnerResults, req)

    // Calculate total requested partners (valid + invalid)
    totalRequestedPartners := len(eligiblePartners) + len(invalidPartners)
    if len(req.Partners) > 0 {
        // If partners were explicitly requested, use that count
        totalRequestedPartners = len(req.Partners)
    }

    // Add invalid partners to metadata filters instead of partners array
    if len(invalidPartners) > 0 {
        s.logger.WithFields(logrus.Fields{
            "component":        "serviceability_orchestrator",
            "invalid_partners": invalidPartners,
        }).Info("Adding invalid partners to metadata filters")
        
        // Add invalid partners to metadata filters
        if response.Metadata != nil {
            // Ensure Filters is properly initialized
            if response.Metadata.Filters.RequestedPartners == nil {
                response.Metadata.Filters.RequestedPartners = req.Partners
            }
            response.Metadata.Filters.FailedPartners = invalidPartners
            
            // Update metadata to reflect failed partners count
            response.Metadata.PartnersFailed += len(invalidPartners)
        }
    }

    // Update TotalPartners to include all requested partners (valid + invalid)
    if response.Metadata != nil {
        response.Metadata.TotalPartners = totalRequestedPartners
    }

    s.logger.WithFields(logrus.Fields{
        "component":         "serviceability_orchestrator",
        "action":            "check_serviceability",
        "total_partners":    len(partnerResults),
		"partnerResults": partnerResults,
        "serviceable_count": len(response.Partners),
        "success":           response.Success,
        "invalid_partners":  len(invalidPartners),
    }).Info("V2 serviceability check completed")

    // Set processing time in metadata
    if response != nil && response.Metadata != nil {
        response.Metadata.ProcessingTime = time.Since(startTime)
    }
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
						CountryCode:       individualReq.CountryCode,
						ParcelCategory:    individualReq.ParcelCategory,
						ProductType:       individualReq.ProductType,
						RequestedPartners: individualReq.Partners,
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

// partnerResult represents the result of checking serviceability with a single partner
type partnerResult struct {
	PartnerCode string
	Result      *common.PartnerServiceabilityResult
	Error       error
	PartnerInfo *DatabasePartnerInfo
	LogoPath    string // Path to logo file from partner service
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

	// Fetch tenant-specific credentials if tenant_id is present in context
	tenantID, hasTenantID := tenantcontext.GetTenantID(ctx)
	var logoPath string
	s.logger.Info("aaaaaaaaaaaaaaaaa",tenantID, hasTenantID)
	if hasTenantID && s.partnerServiceClient != nil {
		credentials, logo, err := s.partnerServiceClient.GetTenantPartnerCredentials(ctx, tenantID, info.PartnerCode)
		if err != nil {
			// Log warning but continue - will fallback to default credentials
			s.logger.WithFields(logrus.Fields{
				"component":    "serviceability_orchestrator",
				"partner_code": info.PartnerCode,
				"tenant_id":    tenantID,
				"error":        err.Error(),
			}).Debug("Failed to fetch tenant partner credentials, will use default credentials")
		} else {
			logoPath = logo
			if len(credentials) > 0 {
				// Store credentials in context for adapter to use
				ctx = tenantcontext.WithPartnerCredentials(ctx, info.PartnerCode, credentials)
				s.logger.WithFields(logrus.Fields{
					"component":         "serviceability_orchestrator",
					"partner_code":      info.PartnerCode,
					"tenant_id":         tenantID,
					"credentials_count": len(credentials),
				}).Info("Using tenant-specific credentials for partner")
			} else {
				s.logger.WithFields(logrus.Fields{
					"component":    "serviceability_orchestrator",
					"partner_code": info.PartnerCode,
					"tenant_id":    tenantID,
				}).Debug("No tenant-specific credentials found (but call succeeded), will use default credentials from env")
			}
		}
	}

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
			LogoPath:    logoPath,
		}
	}

	// Handle nil result (partner not serviceable, e.g., Porter)
	if result == nil {
		s.logger.WithFields(logrus.Fields{
			"component":    "serviceability_orchestrator",
			"partner_code": info.PartnerCode,
		}).Info("Partner returned nil result (not serviceable)")
		return partnerResult{
			PartnerCode: info.PartnerCode,
			Result:      nil, // Explicitly set to nil
			LogoPath:    logoPath,
		}
	}

	// Log partner serviceability check completion details (result is guaranteed to be non-nil here)
	s.logger.WithFields(logrus.Fields{
		"component":      "serviceability_orchestrator",
		"partner_code":   info.PartnerCode,
		"has_services":   len(result.Services) > 0,
		"has_capabilities": len(result.Capabilities) > 0,
		"has_error":      result.Error != nil,
		"error_message":  result.ErrorMessage,
	}).Info("Partner serviceability check completed successfully")

	return partnerResult{
		PartnerCode: info.PartnerCode,
		Result:      result,
		LogoPath:    logoPath,
	}
}

// buildV2Response builds the V2 response from partner results
func (s *serviceabilityOrchestrator) buildV2Response(partnerResults []partnerResult, req *models.ServiceabilityV2Request) *models.ServiceabilityV2Response {
	// serviceablePartners := make([]models.PartnerV2Response, 0)
	serviceableCount := 0
    var topLevelHubDetails interface{}

	// Collect address information
	var hubLocationInfo *models.HubLocationInfo
	var sourceCountryCode, destinationCountryCode string

	// First pass: build basic partner responses and collect serviceable partners
	partnerResponses := make([]models.PartnerV2Response, 0)
	serviceablePartnerResponses := make([]models.PartnerV2Response, 0)

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
			// Include them in the response so user can see what went wrong
			if result.PartnerInfo != nil {
				partnerID := "unknown"
				if result.PartnerInfo.PartnerID != nil {
					partnerID = result.PartnerInfo.PartnerID.String()
				}
				errorMsg := result.Error.Error()
				partnerResponse := models.PartnerV2Response{
					PartnerID:     partnerID,
					PartnerCode:   result.PartnerInfo.PartnerCode,
					PartnerName:   s.getPartnerDisplayName(result.PartnerInfo.PartnerCode),
					LogoURL:       s.getPartnerLogoURL(result.PartnerInfo.PartnerCode, result.LogoPath),
					Rating:       0.0,
					Services:      []models.ServiceV2{},
					Capabilities: make(map[string]interface{}),
					Metadata:      make(map[string]interface{}),
					Source:        "real_time",
					IsServiceable: false,
					Error:         &errorMsg,
				}
				partnerResponses = append(partnerResponses, partnerResponse)
			}

        } else if result.Result != nil {
			// Convert partner result to V2 response using database info
			partnerID := ""
			if result.Result.PartnerID != nil {
				partnerID = result.Result.PartnerID.String()
			} else if result.PartnerInfo != nil && result.PartnerInfo.PartnerID != nil {
				partnerID = result.PartnerInfo.PartnerID.String()
			} else {
				partnerID = "unknown"
			}
			partnerCode := result.PartnerInfo.PartnerCode

            // Extract hub_details from metadata if available and create clean metadata without hub_details
            var hubDetails interface{}
			cleanMetadata := make(map[string]interface{})
			
			if result.Result.Metadata != nil {
				for key, value := range result.Result.Metadata {
					if key == "hub_details" {
						hubDetails = value
					} else {
						cleanMetadata[key] = value
					}
				}
			}

			// Determine if partner is serviceable BEFORE building response
			// A partner is serviceable ONLY if it has services or meaningful capabilities
			// AND no error message (partners with errors are not serviceable)
			hasServices := len(result.Result.Services) > 0
			hasCapabilities := len(result.Result.Capabilities) > 0
			hasError := result.Result.ErrorMessage != nil

			// A partner is serviceable ONLY if it has services or meaningful capabilities
			// Metadata alone does not make a partner serviceable (it may just contain error info)
			isServiceable := hasServices || (hasCapabilities && !hasError)

			// Check if this is India Post Domestic - if so, extract hub_details for top level
			if result.PartnerInfo != nil {
				partnerCodeLower := strings.ToLower(result.PartnerInfo.PartnerCode)
				if partnerCodeLower == "india_post_domestic" {
					// Extract hub_details for top level, don't include in partner response
					if hubDetails != nil {
						topLevelHubDetails = hubDetails
					}
					hubDetails = nil // Don't include in partner response
				}
			}

			partnerResponse := models.PartnerV2Response{
				PartnerID:       partnerID,
				PartnerCode:     partnerCode,
				PartnerName:     s.getPartnerDisplayName(partnerCode),  // Get display name based on partner code
				LogoURL:         s.getPartnerLogoURL(partnerCode, result.LogoPath),       // Get logo URL based on partner code and override path
				Rating:          0.0, // No rating in database
				Services:        result.Result.Services,
				PartnerServices: result.Result.PartnerServices,
				Capabilities:    result.Result.Capabilities,
				ResponseTime:    result.Result.ResponseTime,
				Metadata:        cleanMetadata,
				HubDetails:      hubDetails,
				Source:          "real_time",
				IsServiceable:   isServiceable && !hasError,
			}

			// Add error if partner has error message
			if result.Result.ErrorMessage != nil {
				partnerResponse.Error = result.Result.ErrorMessage
			}

			// Determine if partner is Smile HubOps (or hyperlocal variant)
				isHubOps := false
				if result.PartnerInfo != nil {
					partnerCodeLower := strings.ToLower(result.PartnerInfo.PartnerCode)
					if partnerCodeLower == "smile_hubops" || partnerCodeLower == "smile_hyperlocal_hubops" {
						isHubOps = true
					}
				}

			// Always capture hub_details from HubOps even if not serviceable
			if isHubOps && hubDetails != nil && topLevelHubDetails == nil {
						topLevelHubDetails = hubDetails
					}

			// Add partner to response if serviceable, or if it has an error message (so user can see what went wrong)
			// Always include partners that were requested, even if they have errors
			if isServiceable && !hasError {
				if isHubOps {
					// Count Smile HubOps as serviceable but skip adding to partners array
					serviceableCount++
				} else {
					partnerResponses = append(partnerResponses, partnerResponse)
					serviceablePartnerResponses = append(serviceablePartnerResponses, partnerResponse)
					serviceableCount++
				}
			} else if hasError {
				// Include partners with errors in the response so user can see what went wrong
				// But don't count them as serviceable
				partnerResponses = append(partnerResponses, partnerResponse)
			}
			// Partners with no services/capabilities and no errors are excluded from the response
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
	partnersToReturn = partnerResponses

	// Calculate rate statistics
	totalRatesFound := 0
	partnersWithRates := 0
	for _, partner := range partnersToReturn {
		if len(partner.Services) > 0 {
			for _, service := range partner.Services {
				if service.Rate != nil {
					totalRatesFound++
				}
			}
			if totalRatesFound > 0 {
				partnersWithRates++
			}
		}
	}

	// Build response
	response := &models.ServiceabilityV2Response{
		Success:  isSuccess,
		Partners: partnersToReturn,
		Metadata: &models.V2ResponseMetadata{
			TotalPartners:    len(partnerResults),
			ServiceableCount: serviceableCount,
			// TotalRatesFound:  totalRatesFound,
			// PartnersWithRates: partnersWithRates,
			RatesIncluded:    totalRatesFound > 0,
			Filters: models.V2Filters{
				ParcelCategory:    req.ParcelCategory,
				RequestedPartners: req.Partners,
			},
		},
	}

    // Set top-level hub details if Smile HubOps was serviceable
    if topLevelHubDetails != nil {
        response.HubDetails = topLevelHubDetails
    }

	// Add address information if available
	s.populateAddressInformation(response, req, hubLocationInfo, sourceCountryCode, destinationCountryCode)

	// Add error message when returnOnlyServiceable=true and there are errors or no serviceable partners
	if s.returnOnlyServiceable && (!isSuccess || hasErrors) {
		// If partners array is empty, force legacy-style error
		if len(partnersToReturn) == 0 {
			response.Success = false
			response.Error = &models.ErrorResponse{
				Code:    "PARTNER_ERROR",
				Message: "not serviceable",
			}
		} else {
			// Otherwise, keep normalized/classified partner error
			msg := errorMessage
			if msg == "" {
				msg = "Delivery pincode is not serviceable"
			}
			msg = s.normalizeErrorMessage(msg)
			errorCode := s.classifyErrorType(msg)
			response.Error = &models.ErrorResponse{
				Code:    errorCode,
				Message: msg,
			}
		}
	}

	return response
}

// getPartnerLogoURL returns the logo URL for a partner code
func (s *serviceabilityOrchestrator) getPartnerLogoURL(code string, overridePath string) string {
	// Base URL for partner logos (can be configured via environment variable)
	// Should act as the S3 bucket root URL
	baseURL := os.Getenv("PARTNER_LOGO_BASE_URL")
	if baseURL == "" {
		baseURL = "" // No default S3 bucket
	}
	
	// Ensure base URL doesn't end with slash
	baseURL = strings.TrimSuffix(baseURL, "/")
	
	// VALIDATION: If we have an override path from the partner service, favor that
	if overridePath != "" {
		// If it's already a full URL, return it
		if strings.HasPrefix(overridePath, "http://") || strings.HasPrefix(overridePath, "https://") {
			return overridePath
		}
		// Otherwise join with base URL (S3 bucket)
		// Assuming overridePath is something like "partnerLogos/logo.png"
		if baseURL != "" {
			// Ensure no double slash
			overridePath = strings.TrimPrefix(overridePath, "/")
			return fmt.Sprintf("%s/%s", baseURL, overridePath)
		}
		// If no base URL, return path as is (relative)
		return overridePath
	}



	// If no logo path from partner service, return empty string as requested
	return ""
}

// fetchRatesForPartners fetches rates for all serviceable partners in one batch
func (s *serviceabilityOrchestrator) fetchRatesForPartners(
	ctx context.Context,
	req *models.ServiceabilityV2Request,
	serviceablePartners []models.PartnerV2Response,
) (*supplyrates.RateQuoteResponse, error) {

	if s.rateClient == nil {
		return nil, fmt.Errorf("rate client not initialized")
	}

	// Extract source and destination information
	srcPin := s.getString(req.SourcePostalCode)
	dstPin := s.getString(req.DestinationPostalCode)
	srcCC := s.getString(req.SourceCountryCode)
	dstCC := s.getString(req.DestinationCountryCode)
	
	// First, try to get country codes from partner metadata (e.g., Naqel provides them from database)
	// This takes priority over geolocation service for partners that have country codes in their metadata
	for _, partner := range serviceablePartners {
		if partner.Metadata != nil {
			// Check for source country code
			if srcCC == "" {
				if sourceCC, exists := partner.Metadata["source_country_code"]; exists {
					if ccStr, ok := sourceCC.(string); ok && ccStr != "" {
						srcCC = ccStr
						s.logger.WithFields(logrus.Fields{
							"component":        "serviceability_orchestrator",
							"partner":          partner.PartnerCode,
							"source_country":   srcCC,
						}).Debug("Using source country code from partner metadata")
					}
				}
			}
			// Check for destination country code
			if dstCC == "" {
				if destCC, exists := partner.Metadata["destination_country_code"]; exists {
					if ccStr, ok := destCC.(string); ok && ccStr != "" {
						dstCC = ccStr
						s.logger.WithFields(logrus.Fields{
							"component":          "serviceability_orchestrator",
							"partner":            partner.PartnerCode,
							"destination_country": dstCC,
						}).Debug("Using destination country code from partner metadata")
					}
				}
			}
			// If we found both, no need to check other partners
			if srcCC != "" && dstCC != "" {
				break
			}
		}
	}
	
	// Fall back to geolocation service if country codes still not found
	if s.geolocationService != nil {
		if srcCC == "" {
			srcCCPtr, err := s.geolocationService.GetCountryCodeByPostalCode(ctx, srcPin)
			if err != nil {
				s.logger.WithFields(logrus.Fields{
					"error":          err.Error(),
					"source_pin":     srcPin,
				}).Warn("Failed to get source country code by postal code from geolocation service")
			} else if srcCCPtr != nil {
				srcCC = *srcCCPtr
				s.logger.WithFields(logrus.Fields{
					"component":      "serviceability_orchestrator",
					"source_pin":     srcPin,
					"source_country": srcCC,
				}).Debug("Using source country code from geolocation service")
			}
		}

		if dstCC == "" {
			// Lookup destination country code by postal code
			dstCCPtr, err := s.geolocationService.GetCountryCodeByPostalCode(ctx, dstPin)
			if err != nil {
				s.logger.WithFields(logrus.Fields{
					"error":      err.Error(),
					"dest_pin":   dstPin,
				}).Warn("Failed to get destination country code by postal code from geolocation service")
			} else if dstCCPtr != nil {
				dstCC = *dstCCPtr
				s.logger.WithFields(logrus.Fields{
					"component":          "serviceability_orchestrator",
					"dest_pin":           dstPin,
					"destination_country": dstCC,
				}).Debug("Using destination country code from geolocation service")
			}
		}
	}
	
	

	// // Default to India if country codes not provided
	// if srcCC == "" {
	// 	srcCC = "IN"
	// }
	// if dstCC == "" { /// fix fix fix 
	// 	dstCC = "US"
	// }

	s.logger.WithFields(logrus.Fields{
		"component":        "serviceability_orchestrator",
		"source_pin":       srcPin,
		"source_country":   srcCC,
		"dest_pin":         dstPin,
		"dest_country":     dstCC,
		"partners_count":   len(serviceablePartners),
	}).Info("Calling rate client for partners")

	return s.rateClient.GetRatesForPartners(ctx, req, srcPin, srcCC, dstPin, dstCC, serviceablePartners)
}

func (s *serviceabilityOrchestrator) mergeRatesIntoPartners(
    partners []models.PartnerV2Response,
    rates *supplyrates.RateQuoteResponse,
) {
	// Create a map for quick lookup: partnerCode -> available rates
	ratesMap := make(map[string][]newintl.RateQuote)

	// Populate the map from API response
	for _, successResp := range rates.Data.SuccessfulResponses {
		partnerCode := strings.ToLower(successResp.Partner.Code)

		// Log UrbanBolt rate response
		if partnerCode == "urbanbolt" {
			s.logger.WithFields(logrus.Fields{
				"component":        "serviceability_orchestrator",
				"partner_code":     partnerCode,
				"partner_name":     successResp.Partner.Name,
				"rates_count":      len(successResp.AvailableRates),
				"raw_response":     successResp,
			}).Info("UrbanBolt rate response received from rates API")
		}

		// Convert anonymous struct to our named struct type
		partnerRates := make([]newintl.RateQuote, 0, len(successResp.AvailableRates))
		for _, r := range successResp.AvailableRates {
			partnerRates = append(partnerRates, newintl.RateQuote{
				RateID:       r.RateID,
				Service:      r.Service,
				DeliveryDays: r.DeliveryDays,
				Description:  r.Description,
				Price: struct {
					Currency    string
					Amount      float64
					Type        string
					ServiceType string
				}{
					Currency:    r.Price.Currency,
					Amount:      r.Price.Amount,
					Type:        r.Price.Type,
					ServiceType: r.Price.ServiceType,
				},
			})
		}

		ratesMap[partnerCode] = partnerRates

		// Log UrbanBolt processed rates
		if partnerCode == "urbanbolt" {
			s.logger.WithFields(logrus.Fields{
				"component":    "serviceability_orchestrator",
				"partner_code": partnerCode,
				"rates_count":  len(partnerRates),
				"rates":        partnerRates,
			}).Info("UrbanBolt rates processed and stored in ratesMap")
		}
	}

	// Create a set of successful partner codes for quick lookup
	successfulPartnerCodes := make(map[string]bool)
	for _, successResp := range rates.Data.SuccessfulResponses {
		partnerCode := strings.ToLower(successResp.Partner.Code)
		successfulPartnerCodes[partnerCode] = true
	}

	// Merge rates into partners
	for i := range partners {
		partnerCode := strings.ToLower(partners[i].PartnerCode)
		
		// Log UrbanBolt rate merging process
		if partnerCode == "urbanbolt" {
			s.logger.WithFields(logrus.Fields{
				"component":        "serviceability_orchestrator",
				"partner_code":     partnerCode,
				"partner_id":       partners[i].PartnerID,
				"in_successful":    successfulPartnerCodes[partnerCode],
				"rates_map_exists": ratesMap[partnerCode] != nil,
			}).Info("Processing UrbanBolt rates merge")
		}
		
		// If partner is not in successful responses, clear their services
		if !successfulPartnerCodes[partnerCode] {
			if partnerCode == "urbanbolt" {
				s.logger.WithFields(logrus.Fields{
					"component":    "serviceability_orchestrator",
					"partner_code": partnerCode,
					"reason":       "Partner not in successful rate responses",
				}).Warn("UrbanBolt not found in successful rate responses")
			}
			partners[i].Services = []models.ServiceV2{}
			partners[i].PartnerServices = []models.ServiceV2{}
			if partners[i].Metadata == nil {
				partners[i].Metadata = make(map[string]interface{})
			}
			partners[i].Metadata["rates_available"] = false
			partners[i].Metadata["rate_api_error"] = "Partner not in successful rate responses"
			continue
		}

		// Partner is in successful responses, check if they have rates
		if availableRates, exists := ratesMap[partnerCode]; exists && len(availableRates) > 0 {
			if partnerCode == "urbanbolt" {
				s.logger.WithFields(logrus.Fields{
					"component":    "serviceability_orchestrator",
					"partner_code": partnerCode,
					"rates_count":  len(availableRates),
					"rates":        availableRates,
				}).Info("UrbanBolt rates found and being merged into partner response")
			}
			services := make([]models.ServiceV2, 0, len(availableRates))

			for _, rate := range availableRates {
				service := models.ServiceV2{
					ServiceCode: rate.Price.ServiceType,
					ServiceName: rate.Service,
					TATDays:     rate.DeliveryDays,
					IsCOD:       false,
					Pickup:      true,
					Delivery:    true,
					Insurance:   true,
					ProductTypes: map[string]bool{
						"commercial":   true,
						"document":     true,
						"non_document": true,
					},
					DeliveryModes: map[string]bool{
						"express":  strings.Contains(strings.ToLower(rate.Service), "express"),
						"standard": !strings.Contains(strings.ToLower(rate.Service), "express"),
					},
					Rate: &models.Rate{
						RateID:      rate.RateID,
						Description: rate.Description,
						Price: models.Price{
							Currency: rate.Price.Currency,
							Amount:   rate.Price.Amount,
							Type:     rate.Price.Type,
						},
					},
				}
				services = append(services, service)
			}

			partners[i].Services = services
			partners[i].PartnerServices = services
			
			// If partner has rates, mark as serviceable even if it had an error initially
			// Having rates means the partner can actually service the route
			previousStatus := partners[i].IsServiceable
			partners[i].IsServiceable = true
			
			if !previousStatus {
				s.logger.WithFields(logrus.Fields{
					"component":        "serviceability_orchestrator",
					"partner_code":     partnerCode,
					"rates_count":      len(availableRates),
					"services_count":   len(services),
					"previous_status":  previousStatus,
				}).Info("Partner marked as serviceable due to available rates (was previously not serviceable)")
			}
			
			if partners[i].Metadata == nil {
				partners[i].Metadata = make(map[string]interface{})
			}
			partners[i].Metadata["rates_available"] = true
			partners[i].Metadata["rates_count"] = len(availableRates)
		} else {
			// Partner is in successful responses but has no rates - clear services
			partners[i].Services = []models.ServiceV2{}
			partners[i].PartnerServices = []models.ServiceV2{}
			if partners[i].Metadata == nil {
				partners[i].Metadata = make(map[string]interface{})
			}
			partners[i].Metadata["rates_available"] = false
			partners[i].Metadata["rate_api_error"] = "No rates available for partner"
		}
	}
}

// getString safely gets string value from pointer
func (s *serviceabilityOrchestrator) getString(str *string) string {
	if str != nil {
		return *str
	}
	return ""
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

// normalizeErrorMessage maps partner-specific messages to user-facing messages
func (s *serviceabilityOrchestrator) normalizeErrorMessage(msg string) string {
    lower := strings.ToLower(msg)
    if strings.Contains(lower, "from pincode not found") || strings.Contains(lower, "to pincode not found") || strings.Contains(lower, "postal code not found") {
        return "Delivery pincode is not serviceable"
    }
    return msg
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
// Returns valid partners, invalid partner codes, and error
func (s *serviceabilityOrchestrator) getEligiblePartners(ctx context.Context, req *models.ServiceabilityV2Request) ([]DatabasePartnerInfo, []string, error) {
	s.logger.WithFields(logrus.Fields{
		"component":       "serviceability_orchestrator",
		"parcel_category": req.ParcelCategory,
		"has_partners":    len(req.Partners) > 0,
		"partners_count":  len(req.Partners),
	}).Info("🔍 [FLOW TRACE] getEligiblePartners called - starting partner filtering")
	
	// Get all supported partners from factory
	allSupportedPartners := s.partnerFactory.GetSupportedPartners()
	s.logger.WithFields(logrus.Fields{
		"component":          "serviceability_orchestrator",
		"total_supported":    len(allSupportedPartners),
		"supported_partners": allSupportedPartners,
	}).Info("📋 [FLOW TRACE] Retrieved all supported partners from factory")

	// PRIORITY 1: If specific partners are requested in request body, use ONLY those (strict validation)
	if len(req.Partners) > 0 {
		s.logger.WithFields(logrus.Fields{
			"component":         "serviceability_orchestrator",
			"requested_partners": func() []string {
				codes := make([]string, len(req.Partners))
				for i, p := range req.Partners {
					codes[i] = p.Code
				}
				return codes
			}(),
		}).Info("Using specifically requested partners from request body")

		validPartners, invalidPartners := s.filterRequestedPartners(ctx, req.Partners, allSupportedPartners)
		// Continue with valid partners even if some are invalid
		if len(validPartners) == 0 && len(invalidPartners) > 0 {
			// All partners are invalid - return error
			return nil, invalidPartners, fmt.Errorf("all requested partners are invalid: %v", invalidPartners)
		}
		return validPartners, invalidPartners, nil
	}

	// PRIORITY 2: If no specific partners requested, use parcel category filtering
	// If no parcel category specified, return all supported partners as basic info
	if req.ParcelCategory == nil || *req.ParcelCategory == "" {
		partnerInfos := make([]DatabasePartnerInfo, 0, len(allSupportedPartners))
		for _, partnerCode := range allSupportedPartners {
			partnerInfos = append(partnerInfos, DatabasePartnerInfo{
				PartnerCode: partnerCode,
				PartnerID:   nil, // No database info available
			})
		}
		return partnerInfos, nil, nil
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
		return partnerInfos, nil, nil
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
		return partnerInfos, nil, nil
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
	s.logger.WithFields(logrus.Fields{
		"component":       "serviceability_orchestrator",
		"parcel_category": *req.ParcelCategory,
		"db_partners":     len(eligiblePartnersByCategory),
		"factory_partners": len(allSupportedPartners),
	}).Info("🔄 [FLOW TRACE] Filtering partners - checking which DB partners have factory adapters")
	
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

	return eligiblePartners, nil, nil
}

// filterRequestedPartners filters partners based on requested partner codes from request body
// Returns valid partners and invalid partner codes separately, continues with valid partners
func (s *serviceabilityOrchestrator) filterRequestedPartners(ctx context.Context, requestedPartners []models.PartnerFilter, allSupportedPartners []string) ([]DatabasePartnerInfo, []string) {
	// Create a map of supported partners for quick lookup (case-insensitive)
	supportedPartnerMap := make(map[string]string) // lowercase -> actual code
	for _, partner := range allSupportedPartners {
		supportedPartnerMap[strings.ToLower(partner)] = partner
	}

	filteredPartners := make([]DatabasePartnerInfo, 0, len(requestedPartners))
	invalidPartners := make([]string, 0)

	// Process each requested partner
	for _, reqPartner := range requestedPartners {
		partnerCodeLower := strings.ToLower(reqPartner.Code)
		var matchedCode string
		var isSupported bool
		
		// First, check if the requested partner code exists directly in supported partners
		if actualCode, exists := supportedPartnerMap[partnerCodeLower]; exists {
			matchedCode = actualCode
			isSupported = true
		} else {
			// If not found directly, try to map it using getImplementationCode
			// This handles aliases like smile_ecomm -> smile_ecom
			mappedCode := getImplementationCode(reqPartner.Code)
			mappedCodeLower := strings.ToLower(mappedCode)
			
			// Check if the mapped code exists in supported partners
			if actualCode, exists := supportedPartnerMap[mappedCodeLower]; exists {
				matchedCode = actualCode
				isSupported = true
				s.logger.WithFields(logrus.Fields{
					"component":        "serviceability_orchestrator",
					"requested_code":   reqPartner.Code,
					"mapped_code":      mappedCode,
					"matched_code":     actualCode,
				}).Info("Requested partner mapped to supported implementation")
			}
		}
		
		if isSupported {
			// Use the original requested code (database code) for PartnerCode
			// This preserves the database partner code in the response
			partnerInfo := DatabasePartnerInfo{
				PartnerCode: reqPartner.Code, // Keep original requested code
				PartnerID:   nil, // Will be populated if needed from database
			}

			// If ID is provided in the request, try to use it
			if reqPartner.ID != nil && *reqPartner.ID != "" {
				// Parse the ID (assuming it's a UUID string)
				if id, err := uuid.Parse(*reqPartner.ID); err == nil {
					partnerInfo.PartnerID = &id
				}
			}

			filteredPartners = append(filteredPartners, partnerInfo)

			s.logger.WithFields(logrus.Fields{
				"component":        "serviceability_orchestrator",
				"requested_code":   reqPartner.Code,
				"matched_code":     matchedCode,
			}).Info("Requested partner is supported")
		} else {
			// Collect invalid partners to track in response
			invalidPartners = append(invalidPartners, reqPartner.Code)
			s.logger.WithFields(logrus.Fields{
				"component":          "serviceability_orchestrator",
				"requested_code":     reqPartner.Code,
				"supported_partners": allSupportedPartners,
			}).Warn("Requested partner is not supported, will skip and continue with valid partners")
		}
	}

	if len(invalidPartners) > 0 {
		s.logger.WithFields(logrus.Fields{
			"component":        "serviceability_orchestrator",
			"invalid_partners": invalidPartners,
			"valid_count":      len(filteredPartners),
			"invalid_count":    len(invalidPartners),
		}).Info("Some requested partners are invalid, continuing with valid partners")
	} else {
	s.logger.WithFields(logrus.Fields{
		"component":       "serviceability_orchestrator",
		"requested_count": len(requestedPartners),
		"filtered_count":  len(filteredPartners),
	}).Info("All requested partners are valid")
	}

	return filteredPartners, invalidPartners
}

// getImplementationCode returns the implementation code for a given database partner code
func getImplementationCode(dbPartnerCode string) string {
	// This mapping should match the one in the factory
	// Convert to lowercase for case-insensitive lookup
	normalizedCode := strings.ToLower(dbPartnerCode)
	
	adapterImplementationMap := map[string]string{
		"smile_ecomm":             "smile_ecom",
		"smile_ecom":              "smile_ecom",
		"shipyaari":               "shipyaari",
		"smile_courier":           "smile_courier",
		"dhl":                     "dhl",
		"smile_cargo":             "smile_cargo",
		"smile_hyperlocal":        "delcaper",
		"smile_hubops":            "smile_hubops",
		"porter":                  "porter",
		"india_post_international": "india_post_international",
		"aramex":                  "aramex",
		"fedex":                   "fedex",
		"shipcube":                "shipcube",
		"india_post_domestic":     "india_post_domestic", // Handles both lowercase and uppercase (via lowercase conversion)
		"naqel":                   "naqel",
		"dharmendra":             "dharmendra",
		"xpressbees":             "dharmendra",          // XpressBees maps to dharmendra
		"expressbees":            "dharmendra",          // expressbees maps to dharmendra
		"sunil_baral":             "sunil_baral",
		"urbanbolt":               "urbanbolt",
		"delhivery":               "delhivery",
	}

	if implCode, exists := adapterImplementationMap[normalizedCode]; exists {
		return implCode
	}
	return dbPartnerCode // Return original code if no mapping exists
}

// getPartnerDisplayName returns the display name for a partner code
func (s *serviceabilityOrchestrator) getPartnerDisplayName(code string) string {
	// Simple name mapping for known partners
	nameMap := map[string]string{
		"dhl":                    "DHL",
		"smile_cargo":            "Smile Cargo",
		"smile_ecomm":            "Smile Ecommerce",
		"smile_ecom":             "Smile Ecommerce",
		"shipyaari":              "Shipyaari",
		"smile_courier":          "Smile Courier",
		"smile_hubops":           "Smile HubOps",
		"porter":                 "Porter",
		"india_post_international": "India Post International",
		"india_post_domestic":     "India Post Domestic",
		"naqel":                  "Naqel",
		"aramex":                 "Aramex",
		"fedex":                  "FedEx",
		"shipcube":               "ShipCube",
		"dharmendra":             "XpressBees",
		"xpressbees":             "XpressBees",
		"expressbees":            "XpressBees",
		"XpressBees":             "XpressBees",
		"sunil_baral":            "Sunil Baral",
		"urbanbolt":              "UrbanBolt",
		"delhivery":              "Delhivery",
	}

	// Check case-insensitive first
	codeLower := strings.ToLower(code)
	if name, exists := nameMap[codeLower]; exists {
		return name
	}
	
	// Check exact match
	if name, exists := nameMap[code]; exists {
		return name
	}

	// Convert snake_case to Title Case for unknown codes
	parts := strings.Split(code, "_")
	for i, part := range parts {
		if len(part) > 0 {
			parts[i] = strings.ToUpper(part[:1]) + part[1:]
		}
	}
	return strings.Join(parts, " ")
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

// hasInternationalPartners checks if any of the requested partners are international partners
// International partners: dhl, fedex, aramex, shipcube, indiapost, naqel
func (s *serviceabilityOrchestrator) hasInternationalPartners(partners []models.PartnerFilter) bool {
	if len(partners) == 0 {
		return false
	}

	// Define international partner codes (case-insensitive matching)
	internationalPartners := map[string]bool{
		"dhl":       true,
		"fedex":     true,
		"aramex":    true,
		"shipcube":  true,
		"indiapost": true,
		"naqel":     true,
	}

	// Define international partner keywords for partial matching
	internationalKeywords := []string{"dhl", "fedex", "aramex", "shipcube", "indiapost", "naqel", "india_post_international"}

	for _, partner := range partners {
		partnerCodeLower := strings.ToLower(partner.Code)
		
		// Direct match
		if internationalPartners[partnerCodeLower] {
			return true
		}
		
		// Also check if it contains international partner keyword
		for _, keyword := range internationalKeywords {
			if strings.Contains(partnerCodeLower, keyword) {
				return true
			}
		}
	}

	return false
}
