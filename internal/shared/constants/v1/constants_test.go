package constants

import (
	"net/http"
	"testing"
)

// TestHTTPStatusCodes verifies all HTTP status codes are properly defined
func TestHTTPStatusCodes(t *testing.T) {
	tests := []struct {
		name     string
		constant int
		expected int
	}{
		{"StatusOK", StatusOK, http.StatusOK},
		{"StatusCreated", StatusCreated, http.StatusCreated},
		{"StatusAccepted", StatusAccepted, http.StatusAccepted},
		{"StatusNoContent", StatusNoContent, http.StatusNoContent},
		{"StatusBadRequest", StatusBadRequest, http.StatusBadRequest},
		{"StatusUnauthorized", StatusUnauthorized, http.StatusUnauthorized},
		{"StatusForbidden", StatusForbidden, http.StatusForbidden},
		{"StatusNotFound", StatusNotFound, http.StatusNotFound},
		{"StatusMethodNotAllowed", StatusMethodNotAllowed, http.StatusMethodNotAllowed},
		{"StatusConflict", StatusConflict, http.StatusConflict},
		{"StatusUnprocessableEntity", StatusUnprocessableEntity, http.StatusUnprocessableEntity},
		{"StatusTooManyRequests", StatusTooManyRequests, http.StatusTooManyRequests},
		{"StatusInternalServerError", StatusInternalServerError, http.StatusInternalServerError},
		{"StatusBadGateway", StatusBadGateway, http.StatusBadGateway},
		{"StatusServiceUnavailable", StatusServiceUnavailable, http.StatusServiceUnavailable},
		{"StatusGatewayTimeout", StatusGatewayTimeout, http.StatusGatewayTimeout},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.constant != tt.expected {
				t.Errorf("%s = %d, expected %d", tt.name, tt.constant, tt.expected)
			}
		})
	}
}

// TestErrorCodes verifies all error codes are properly defined and non-empty
func TestErrorCodes(t *testing.T) {
	errorCodes := map[string]string{
		"ErrorCodeInvalidPostalCode":         ErrorCodeInvalidPostalCode,
		"ErrorCodePostalCodeNotFound":        ErrorCodePostalCodeNotFound,
		"ErrorCodeInvalidCountryCode":        ErrorCodeInvalidCountryCode,
		"ErrorCodeLocationNotServiceable":    ErrorCodeLocationNotServiceable,
		"ErrorCodeInvalidRequest":            ErrorCodeInvalidRequest,
		"ErrorCodeInternalServerError":       ErrorCodeInternalServerError,
		"ErrorCodeServiceUnavailable":        ErrorCodeServiceUnavailable,
		"ErrorCodePartnerServiceError":       ErrorCodePartnerServiceError,
		"ErrorCodeSpecificationServiceError": ErrorCodeSpecificationServiceError,
		"ErrorCodeDatabaseError":             ErrorCodeDatabaseError,
		"ErrorCodeValidationError":           ErrorCodeValidationError,
		"ErrorCodeTimeoutError":              ErrorCodeTimeoutError,
		"ErrorCodeAuthFailed":                ErrorCodeAuthFailed,
		"ErrorCodeInvalidToken":              ErrorCodeInvalidToken,
		"ErrorCodeExpiredToken":              ErrorCodeExpiredToken,
		"ErrorCodeMissingCredentials":        ErrorCodeMissingCredentials,
		"ErrorCodeInsufficientPermissions":   ErrorCodeInsufficientPermissions,
		"ErrorCodeInvalidPartner":            ErrorCodeInvalidPartner,
	}

	for name, code := range errorCodes {
		t.Run(name, func(t *testing.T) {
			if code == "" {
				t.Errorf("%s is empty", name)
			}
		})
	}
}

// TestErrorMessages verifies all error messages are properly defined and non-empty
func TestErrorMessages(t *testing.T) {
	errorMessages := map[string]string{
		"MsgInvalidPostalCode":         MsgInvalidPostalCode,
		"MsgPostalCodeNotFound":        MsgPostalCodeNotFound,
		"MsgInvalidCountryCode":        MsgInvalidCountryCode,
		"MsgLocationNotServiceable":    MsgLocationNotServiceable,
		"MsgInvalidRequest":            MsgInvalidRequest,
		"MsgInternalServerError":       MsgInternalServerError,
		"MsgServiceUnavailable":        MsgServiceUnavailable,
		"MsgPartnerServiceError":       MsgPartnerServiceError,
		"MsgSpecificationServiceError": MsgSpecificationServiceError,
		"MsgDatabaseError":             MsgDatabaseError,
		"MsgValidationError":           MsgValidationError,
		"MsgTimeoutError":              MsgTimeoutError,
		"MsgAuthFailed":                MsgAuthFailed,
		"MsgInvalidToken":              MsgInvalidToken,
		"MsgExpiredToken":              MsgExpiredToken,
		"MsgMissingCredentials":        MsgMissingCredentials,
		"MsgInsufficientPermissions":   MsgInsufficientPermissions,
		"MsgInvalidPartner":            MsgInvalidPartner,
	}

	for name, message := range errorMessages {
		t.Run(name, func(t *testing.T) {
			if message == "" {
				t.Errorf("%s is empty", name)
			}
		})
	}
}

// TestErrorCodeMessageMapping verifies error codes have corresponding messages
func TestErrorCodeMessageMapping(t *testing.T) {
	codeMessageMapping := map[string]string{
		ErrorCodeInvalidPostalCode:         MsgInvalidPostalCode,
		ErrorCodePostalCodeNotFound:        MsgPostalCodeNotFound,
		ErrorCodeInvalidCountryCode:        MsgInvalidCountryCode,
		ErrorCodeLocationNotServiceable:    MsgLocationNotServiceable,
		ErrorCodeInvalidRequest:            MsgInvalidRequest,
		ErrorCodeInternalServerError:       MsgInternalServerError,
		ErrorCodeServiceUnavailable:        MsgServiceUnavailable,
		ErrorCodePartnerServiceError:       MsgPartnerServiceError,
		ErrorCodeSpecificationServiceError: MsgSpecificationServiceError,
		ErrorCodeDatabaseError:             MsgDatabaseError,
		ErrorCodeValidationError:           MsgValidationError,
		ErrorCodeTimeoutError:              MsgTimeoutError,
		ErrorCodeAuthFailed:                MsgAuthFailed,
		ErrorCodeInvalidToken:              MsgInvalidToken,
		ErrorCodeExpiredToken:              MsgExpiredToken,
		ErrorCodeMissingCredentials:        MsgMissingCredentials,
		ErrorCodeInsufficientPermissions:   MsgInsufficientPermissions,
		ErrorCodeInvalidPartner:            MsgInvalidPartner,
	}

	for code, message := range codeMessageMapping {
		t.Run(code, func(t *testing.T) {
			if message == "" {
				t.Errorf("Error code %s has no corresponding message", code)
			}
		})
	}
}

// TestEnvironmentVariables verifies environment variable constants are properly defined
func TestEnvironmentVariables(t *testing.T) {
	envVars := map[string]string{
		"EnvServerPort":              EnvServerPort,
		"EnvLogLevel":                EnvLogLevel,
		"EnvDatabaseURL":             EnvDatabaseURL,
		"EnvRedisURL":                EnvRedisURL,
		"EnvPartnerServiceURL":       EnvPartnerServiceURL,
		"EnvSpecificationServiceURL": EnvSpecificationServiceURL,
		"EnvPrayogBaseURL":           EnvPrayogBaseURL,
		"EnvAPIKey":                  EnvAPIKey,
		"EnvEnvironment":             EnvEnvironment,
	}

	for name, envVar := range envVars {
		t.Run(name, func(t *testing.T) {
			if envVar == "" {
				t.Errorf("%s is empty", name)
			}
		})
	}
}
