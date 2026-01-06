package india_post_domestic

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"prayog-serviceability-service/internal/services/v2/partners/common"
	"prayog-serviceability-service/internal/shared/config"
	"prayog-serviceability-service/internal/shared/models/v1"

	"github.com/sirupsen/logrus"
)

// Adapter implements the PartnerAdapter interface for India Post Domestic
type Adapter struct {
	client *Client
	auth   *Authenticator
	config config.IndiaPostDomesticConfig
	logger *logrus.Logger
}

// NewAdapter creates a new India Post Domestic adapter instance
func NewAdapter(cfg config.IndiaPostDomesticConfig) *Adapter {
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)

	logger.WithFields(logrus.Fields{
		"partner":  "IndiaPostDomestic",
		"base_url": cfg.BaseURL,
		"enabled":  cfg.Enabled,
	}).Info("Creating India Post Domestic adapter")

	// Create authenticator
	auth := NewAuthenticator(cfg)

	// Create client with authenticator
	client := NewClient(cfg, auth)

	return &Adapter{
		client: client,
		auth:   auth,
		config: cfg,
		logger: logger,
	}
}

// GetAdapterType returns the adapter type
func (a *Adapter) GetAdapterType() common.AdapterType {
	return common.AdapterTypeHTTP
}

// IsEnabled returns whether the adapter is enabled
func (a *Adapter) IsEnabled() bool {
	return a.config.Enabled
}

// CheckServiceability checks if India Post Domestic can service the given request
func (a *Adapter) CheckServiceability(ctx context.Context, request *models.ServiceabilityV2Request, partnerInfo common.PartnerInfo) (*common.PartnerServiceabilityResult, error) {
	startTime := time.Now()

	partnerID := ""
	if partnerInfo.PartnerID != nil {
		partnerID = partnerInfo.PartnerID.String()
	}

	a.logger.WithFields(logrus.Fields{
		"component":    "india_post_domestic_adapter",
		"action":       "check_serviceability_start",
		"partner_code": partnerInfo.PartnerCode,
		"partner_id":   partnerID,
	}).Info("Starting India Post Domestic adapter serviceability check")

	// Validate India Post Domestic specific requirements
	if err := a.validateRequirements(request); err != nil {
		a.logger.WithFields(logrus.Fields{
			"component":    "india_post_domestic_adapter",
			"event":        "validation_failed",
			"partner_code": partnerInfo.PartnerCode,
			"partner_id":   partnerID,
			"error":        err.Error(),
		}).Warn("India Post Domestic validation failed")

		return &common.PartnerServiceabilityResult{
			PartnerID:    partnerInfo.PartnerID,
			PartnerCode:  partnerInfo.PartnerCode,
			Services:     make([]models.ServiceV2, 0),
			ResponseTime: time.Since(startTime),
			Error:        err,
			ErrorMessage: &[]string{fmt.Sprintf("India Post Domestic validation failed: %v", err)}[0],
			Metadata: map[string]interface{}{
				"reason": "India Post Domestic validation failed",
			},
		}, nil
	}

	// Determine which pincode to check
	// For India Post Domestic, we need the destination pincode
	var pincodeToCheck string
	if request.DestinationPostalCode != nil && *request.DestinationPostalCode != "" {
		pincodeToCheck = *request.DestinationPostalCode
	} else if request.PostalCode != nil && *request.PostalCode != "" {
		pincodeToCheck = *request.PostalCode
	} else {
		err := fmt.Errorf("destination postal code is required for India Post Domestic")
		return &common.PartnerServiceabilityResult{
			PartnerID:    partnerInfo.PartnerID,
			PartnerCode:  partnerInfo.PartnerCode,
			Services:     make([]models.ServiceV2, 0),
			ResponseTime: time.Since(startTime),
			Error:        err,
			ErrorMessage: &[]string{err.Error()}[0],
		}, nil
	}

	a.logger.WithFields(logrus.Fields{
		"component":    "india_post_domestic_adapter",
		"partner_code": partnerInfo.PartnerCode,
		"partner_id":   partnerID,
		"pincode":      pincodeToCheck,
	}).Info("Checking serviceability for India Post Domestic")

	// Call India Post Domestic API to search for pincode
	response, err := a.client.SearchPincode(ctx, pincodeToCheck)
	if err != nil {
		a.logger.WithFields(logrus.Fields{
			"component":    "india_post_domestic_adapter",
			"partner_code": partnerInfo.PartnerCode,
			"partner_id":   partnerID,
			"pincode":      pincodeToCheck,
			"error":        err.Error(),
		}).Error("India Post Domestic API call failed")

		return &common.PartnerServiceabilityResult{
			PartnerID:    partnerInfo.PartnerID,
			PartnerCode:  partnerInfo.PartnerCode,
			Services:     make([]models.ServiceV2, 0),
			ResponseTime: time.Since(startTime),
			Error:        err,
			ErrorMessage: &[]string{fmt.Sprintf("India Post Domestic API call failed: %v", err)}[0],
		}, nil
	}

	// Convert response to serviceability result
	result := a.convertToServiceabilityResult(response, partnerInfo, pincodeToCheck)
	
	// Call HubOps API to get route information and add to hub_details
	// This is called before returning the response as per requirements
	if request.SourcePostalCode != nil && *request.SourcePostalCode != "" {
		hubOpsResp, err := a.client.GetRouteByPincode(ctx, *request.SourcePostalCode, pincodeToCheck)
		if err != nil {
			// Check if this is a 400 error with "3PL is not available" message
			errStr := err.Error()
			if strings.Contains(errStr, "status 400") && strings.Contains(errStr, "3PL is not available") {
				a.logger.WithFields(logrus.Fields{
					"component":              "india_post_domestic_adapter",
					"partner_code":           partnerInfo.PartnerCode,
					"partner_id":             partnerID,
					"source_postal_code":     *request.SourcePostalCode,
					"destination_postal_code": pincodeToCheck,
					"error":                  err.Error(),
				}).Warn("3PL not available for destination pincode - marking as not serviceable")
				
				// Return not serviceable result
				return &common.PartnerServiceabilityResult{
					PartnerID:    partnerInfo.PartnerID,
					PartnerCode:  partnerInfo.PartnerCode,
					Services:     make([]models.ServiceV2, 0),
					ResponseTime: time.Since(startTime),
					Error:        err,
					ErrorMessage: &[]string{"3PL is not available for destination pincode"}[0],
					Metadata: map[string]interface{}{
						"reason": "3PL is not available for destination pincode",
						"pincode": pincodeToCheck,
					},
				}, nil
			}
			
			a.logger.WithFields(logrus.Fields{
				"component":              "india_post_domestic_adapter",
				"partner_code":           partnerInfo.PartnerCode,
				"partner_id":             partnerID,
				"source_postal_code":     *request.SourcePostalCode,
				"destination_postal_code": pincodeToCheck,
				"error":                  err.Error(),
			}).Warn("HubOps API call failed, continuing without hub_details")
		} else if hubOpsResp != nil {
			// Convert all keys to snake_case and add to metadata as hub_details
			converted := convertKeysToSnakeCase(hubOpsResp)
			if result.Metadata == nil {
				result.Metadata = make(map[string]interface{})
			}
			result.Metadata["hub_details"] = converted
			a.logger.WithFields(logrus.Fields{
				"component":              "india_post_domestic_adapter",
				"partner_code":           partnerInfo.PartnerCode,
				"partner_id":             partnerID,
				"source_postal_code":     *request.SourcePostalCode,
				"destination_postal_code": pincodeToCheck,
			}).Info("Successfully added hub_details to metadata (snake_case)")
		}
	}
	
	result.ResponseTime = time.Since(startTime)

	a.logger.WithFields(logrus.Fields{
		"component":      "india_post_domestic_adapter",
		"partner_code":   partnerInfo.PartnerCode,
		"partner_id":     partnerID,
		"pincode":        pincodeToCheck,
		"is_serviceable": len(result.Services) > 0,
		"offices_found":  response.ReturnedRecordsCount,
		"response_time":  result.ResponseTime,
	}).Info("India Post Domestic serviceability check completed")

	return result, nil
}

// validateRequirements validates India Post Domestic specific requirements
func (a *Adapter) validateRequirements(request *models.ServiceabilityV2Request) error {
	// India Post Domestic requires either destination postal code or generic postal code
	if (request.DestinationPostalCode == nil || *request.DestinationPostalCode == "") &&
		(request.PostalCode == nil || *request.PostalCode == "") {
		return fmt.Errorf("postal code is required for India Post Domestic serviceability")
	}

	// Validate pincode format (6 digits for Indian pincodes)
	var pincodeToCheck string
	if request.DestinationPostalCode != nil && *request.DestinationPostalCode != "" {
		pincodeToCheck = *request.DestinationPostalCode
	} else if request.PostalCode != nil {
		pincodeToCheck = *request.PostalCode
	}

	// Check if pincode is numeric and 6 digits
	if _, err := strconv.Atoi(pincodeToCheck); err != nil {
		return fmt.Errorf("invalid pincode format: must be numeric")
	}

	if len(pincodeToCheck) != 6 {
		return fmt.Errorf("invalid pincode format: must be 6 digits")
	}

	return nil
}

// convertToServiceabilityResult converts India Post Domestic response to common format
//
// CHANGES MADE (2025-01-XX):
// 1. Added service creation logic to populate Services array when offices are found
// 2. Removed offices_data from metadata to reduce response size
// 3. Added defensive fallback to ensure at least one service is always created
// 4. Added error logging if services array is empty after processing
//
// See inline comments marked with "CHANGE:" for details on reverting specific changes
func (a *Adapter) convertToServiceabilityResult(response *PincodeSearchResponse, partnerInfo common.PartnerInfo, pincode string) *common.PartnerServiceabilityResult {
	result := &common.PartnerServiceabilityResult{
		PartnerID:    partnerInfo.PartnerID,
		PartnerCode:  partnerInfo.PartnerCode,
		Services:     make([]models.ServiceV2, 0),
		Capabilities: make(map[string]interface{}),
		Metadata:     make(map[string]interface{}),
	}

	// If no offices found, pincode is not serviceable
	if response.ReturnedRecordsCount == 0 || len(response.Data) == 0 {
		result.Metadata["reason"] = "Pincode not found in India Post Domestic database"
		result.Metadata["pincode"] = pincode
		result.Metadata["is_serviceable"] = false
		return result
	}

	// India Post Domestic returned data - pincode is serviceable
	// Find delivery offices
	var deliveryOffices []PostalOffice
	var allOffices []PostalOffice
	
	for _, office := range response.Data {
		allOffices = append(allOffices, office)
		if office.DeliveryOfficeFlag && office.IsRolledOut {
			deliveryOffices = append(deliveryOffices, office)
		}
	}

	// Build capabilities from the offices found
	capabilities := map[string]interface{}{
		"pincode":              response.Data[0].Pincode,
		"state_name":           response.Data[0].StateName,
		"city_name":            response.Data[0].CityName,
		"taluk_name":           response.Data[0].TalukName,
		"total_offices":        response.ReturnedRecordsCount,
		"delivery_offices":     len(deliveryOffices),
		"has_delivery_office":  len(deliveryOffices) > 0,
		"is_rolled_out":        response.Data[0].IsRolledOut,
	}

	// Add office details
	if len(deliveryOffices) > 0 {
		capabilities["primary_delivery_office"] = deliveryOffices[0].OfficeName
		capabilities["primary_office_type"] = deliveryOffices[0].OfficeTypeCode
	} else if len(allOffices) > 0 {
		capabilities["primary_office"] = allOffices[0].OfficeName
		capabilities["primary_office_type"] = allOffices[0].OfficeTypeCode
	}

	result.Capabilities = capabilities

	// ============================================================================
	// CHANGE: Added service creation logic to populate Services array
	// REASON: Previously, Services array was empty even when offices were found,
	//         causing orchestrator to treat partner as "not serviceable"
	// DATE: 2025-01-XX
	// TO REVERT: Remove the entire service creation block (lines 243-331)
	//            and restore the old comment-only section
	// ============================================================================
	
	// Create service entries since India Post Domestic is serviceable at this pincode
	// Add at least one service to indicate serviceability
	if len(deliveryOffices) > 0 {
		// If we have delivery offices, create a service for each unique office type
		officeTypes := make(map[string]bool)
		for _, office := range deliveryOffices {
			if !officeTypes[office.OfficeTypeCode] {
				service := models.ServiceV2{
					ServiceCode: fmt.Sprintf("INDIA_POST_DOMESTIC_%s", office.OfficeTypeCode),
					ServiceName: fmt.Sprintf("India Post Domestic - %s", getOfficeTypeName(office.OfficeTypeCode)),
					TATDays:     3, // Default TAT for domestic delivery
					IsCOD:       true,
					Pickup:      true,
					Delivery:    true,
					Insurance:   true,
					ProductTypes: map[string]bool{
						"commercial":   true,
						"document":     true,
						"non_document": true,
					},
					DeliveryModes: map[string]bool{
						"standard": true,
					},
				}
				result.Services = append(result.Services, service)
				officeTypes[office.OfficeTypeCode] = true
			}
		}
		a.logger.WithFields(logrus.Fields{
			"component":        "india_post_domestic_adapter",
			"pincode":          pincode,
			"delivery_offices": len(deliveryOffices),
			"services_created": len(result.Services),
		}).Info("Created services from delivery offices")
	} else if len(allOffices) > 0 {
		// If we have offices but no delivery offices, still create a generic service
		service := models.ServiceV2{
			ServiceCode: "INDIA_POST_DOMESTIC_STANDARD",
			ServiceName: "India Post Domestic - Standard",
			TATDays:     5, // Longer TAT if no dedicated delivery office
			IsCOD:       true,
			Pickup:      true,
			Delivery:    true,
			Insurance:   true,
			ProductTypes: map[string]bool{
				"commercial":   true,
				"document":     true,
				"non_document": true,
			},
			DeliveryModes: map[string]bool{
				"standard": true,
			},
		}
		result.Services = append(result.Services, service)
		a.logger.WithFields(logrus.Fields{
			"component":        "india_post_domestic_adapter",
			"pincode":          pincode,
			"total_offices":    len(allOffices),
			"delivery_offices": 0,
			"services_created": 1,
		}).Info("Created generic service from all offices (no delivery offices)")
	}
	
	// ============================================================================
	// CHANGE: Added defensive fallback to ensure services are always created
	// REASON: Safety net in case service creation logic above fails
	// DATE: 2025-01-XX
	// TO REVERT: Remove this entire if block (lines 306-331)
	// ============================================================================
	// Ensure at least one service is created (defensive check)
	if len(result.Services) == 0 && len(allOffices) > 0 {
		// Fallback: create a default service if somehow none were created
		service := models.ServiceV2{
			ServiceCode: "INDIA_POST_DOMESTIC_STANDARD",
			ServiceName: "India Post Domestic - Standard",
			TATDays:     5,
			IsCOD:       true,
			Pickup:      true,
			Delivery:    true,
			Insurance:   true,
			ProductTypes: map[string]bool{
				"commercial":   true,
				"document":     true,
				"non_document": true,
			},
			DeliveryModes: map[string]bool{
				"standard": true,
			},
		}
		result.Services = append(result.Services, service)
		a.logger.WithFields(logrus.Fields{
			"component": "india_post_domestic_adapter",
			"pincode":   pincode,
		}).Warn("Fallback: Created default service (no services were created earlier)")
	}
	
	// ============================================================================
	// CHANGE: Removed offices_data from metadata to reduce response size
	// REASON: offices_data contains full office array which makes response too large
	//         and is unnecessary since office details are already in capabilities
	// DATE: 2025-01-XX
	// TO REVERT: Uncomment the line below:
	//            result.Metadata["offices_data"] = response.Data
	// ============================================================================
	
	// Add metadata (excluding offices_data to reduce response size)
	result.Metadata["pincode"] = pincode
	result.Metadata["is_serviceable"] = true
	// REMOVED: result.Metadata["offices_data"] = response.Data  // See comment above for revert
	result.Metadata["search_success"] = response.Success
	result.Metadata["message"] = response.Message

	// ============================================================================
	// CHANGE: Added final validation logging for debugging
	// REASON: To help identify if services array is empty after all processing
	// DATE: 2025-01-XX
	// TO REVERT: Remove this entire if block (lines 339-347)
	// ============================================================================
	// Final validation: log if services are still empty (should never happen)
	if len(result.Services) == 0 {
		a.logger.WithFields(logrus.Fields{
			"component":        "india_post_domestic_adapter",
			"pincode":          pincode,
			"total_offices":    len(allOffices),
			"delivery_offices": len(deliveryOffices),
		}).Error("CRITICAL: Services array is empty after processing - this should not happen")
	}

	return result
}

// Initialize implements PartnerAdapter interface
func (a *Adapter) Initialize(ctx context.Context) error {
	// Test authentication
	if err := a.auth.Authenticate(ctx); err != nil {
		return fmt.Errorf("India Post Domestic authentication failed: %w", err)
	}

	a.logger.Info("India Post Domestic adapter initialized successfully")
	return nil
}

// IsHealthy implements PartnerAdapter interface
func (a *Adapter) IsHealthy(ctx context.Context) bool {
	if !a.config.Enabled {
		return true
	}

	// Check if we can authenticate
	if !a.auth.IsAuthenticated() {
		if err := a.auth.Authenticate(ctx); err != nil {
			a.logger.WithError(err).Warn("India Post Domestic health check failed: authentication error")
			return true
		}
	}

	return true
}

// GetMetrics implements PartnerAdapter interface
func (a *Adapter) GetMetrics() *common.PartnerMetrics {
	return &common.PartnerMetrics{
		PartnerCode:         "", // Will be set by orchestrator from database
		TotalRequests:       0,  // TODO: Implement actual metrics
		SuccessfulRequests:  0,
		FailedRequests:      0,
		AverageResponseTime: 0,
		HealthStatus:        "healthy",
		ErrorRate:           0.0,
	}
}

// Shutdown implements PartnerAdapter interface
func (a *Adapter) Shutdown(ctx context.Context) error {
	a.logger.Info("Shutting down India Post Domestic adapter")
	// No specific cleanup needed for HTTP client
	return nil
}

// ============================================================================
// CHANGE: Added helper function to convert office type codes to readable names
// REASON: Needed for creating service names in the format "India Post Domestic - {OfficeType}"
// DATE: 2025-01-XX
// TO REVERT: Remove this entire function and update service creation to use
//            office type codes directly or inline the mapping
// ============================================================================
// getOfficeTypeName returns a human-readable name for office type code
func getOfficeTypeName(officeTypeCode string) string {
	switch officeTypeCode {
	case "BPO":
		return "Branch Post Office"
	case "SPO":
		return "Sub Post Office"
	case "HPO":
		return "Head Post Office"
	case "SO":
		return "Small Office"
	default:
		return "Post Office"
	}
}

// convertKeysToSnakeCase recursively converts all map keys to snake_case
func convertKeysToSnakeCase(input interface{}) interface{} {
	switch v := input.(type) {
	case map[string]interface{}:
		out := make(map[string]interface{}, len(v))
		for k, val := range v {
			out[toSnakeCase(k)] = convertKeysToSnakeCase(val)
		}
		return out
	case []interface{}:
		arr := make([]interface{}, len(v))
		for i, elem := range v {
			arr[i] = convertKeysToSnakeCase(elem)
		}
		return arr
	default:
		return input
	}
}

var snakeCaseRegex1 = regexp.MustCompile("([a-z])([A-Z])")
var snakeCaseRegex2 = regexp.MustCompile("([A-Z]+)([A-Z][a-z])")
var snakeCaseRegex3 = regexp.MustCompile("([0-9])([A-Z])") // Digit before any uppercase (e.g., "3PL" -> "3_PL", "3Hub" -> "3_Hub")

func toSnakeCase(s string) string {
	if s == "" {
		return s
	}
	// Replace spaces and hyphens with underscores first
	s = strings.ReplaceAll(s, "-", "_")
	s = strings.ReplaceAll(s, " ", "_")
	// Handle digit before uppercase (e.g., "3PL" -> "3_PL", "3Hub" -> "3_Hub")
	s = snakeCaseRegex3.ReplaceAllString(s, "${1}_${2}")
	// Handle cases like JSONURL -> json_url (uppercase sequences before camelCase)
	s = snakeCaseRegex2.ReplaceAllString(s, "${1}_${2}")
	// Handle lowercase before uppercase (camelCase -> camel_case)
	s = snakeCaseRegex1.ReplaceAllString(s, "${1}_${2}")
	s = strings.ToLower(s)
	return s
}

