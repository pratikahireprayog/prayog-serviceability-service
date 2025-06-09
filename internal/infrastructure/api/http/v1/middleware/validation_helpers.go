package middleware

import (
	"errors"
	"strconv"

	"github.com/google/uuid"
)

// ValidationHelpers provides common validation functions for middleware
var ValidationHelpers = struct {
	ValidateUUID       func(string) error
	ValidatePagination func(offset, limit string) error
	ValidateCode       func(string) error
}{
	ValidateUUID:       validateUUID,
	ValidatePagination: validatePagination,
	ValidateCode:       validateCode,
}

// validateUUID checks if a string is a valid UUID
func validateUUID(value string) error {
	if _, err := uuid.Parse(value); err != nil {
		return errors.New("must be a valid UUID")
	}
	return nil
}

// validatePagination validates offset and limit parameters
func validatePagination(offset, limit string) error {
	if offset != "" {
		if offsetVal, err := strconv.Atoi(offset); err != nil {
			return errors.New("offset must be a valid integer")
		} else if offsetVal < 0 {
			return errors.New("offset must be non-negative")
		}
	}

	if limit != "" {
		if limitVal, err := strconv.Atoi(limit); err != nil {
			return errors.New("limit must be a valid integer")
		} else if limitVal <= 0 {
			return errors.New("limit must be greater than 0")
		} else if limitVal > 100 {
			return errors.New("limit must not exceed 100")
		}
	}

	return nil
}

// validateCode validates entity codes (country codes, region codes, etc.)
func validateCode(code string) error {
	if len(code) < 2 {
		return errors.New("code must be at least 2 characters long")
	}
	if len(code) > 10 {
		return errors.New("code must not exceed 10 characters")
	}
	// Add more specific validation rules as needed
	return nil
}

// GetCommonParamValidators returns a map of common parameter validators
func GetCommonParamValidators() map[string]func(string) error {
	return map[string]func(string) error{
		"id":         ValidationHelpers.ValidateUUID,
		"countryId":  ValidationHelpers.ValidateUUID,
		"regionId":   ValidationHelpers.ValidateUUID,
		"districtId": ValidationHelpers.ValidateUUID,
		"cityId":     ValidationHelpers.ValidateUUID,
		"areaId":     ValidationHelpers.ValidateUUID,
		"code":       ValidationHelpers.ValidateCode,
	}
}

// PaginationQueryValidator represents query parameters for pagination
type PaginationQueryValidator struct {
	Offset int `query:"offset" validate:"min=0"`
	Limit  int `query:"limit" validate:"min=1,max=100"`
}
