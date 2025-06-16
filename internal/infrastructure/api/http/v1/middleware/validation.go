package middleware

import (
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"

	"prayog-serviceability-service/internal/shared/dtos/v1"
)

// ValidationMiddleware represents the validation middleware with dependencies
type ValidationMiddleware struct {
	validator *validator.Validate
	logger    *logrus.Logger
}

// NewValidationMiddleware creates a new validation middleware
func NewValidationMiddleware(validator *validator.Validate, logger *logrus.Logger) *ValidationMiddleware {
	return &ValidationMiddleware{
		validator: validator,
		logger:    logger,
	}
}

// ValidateBody returns a middleware that validates request body against the provided struct type
func (vm *ValidationMiddleware) ValidateBody(target interface{}) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Parse the request body into the target struct
		if err := c.BodyParser(target); err != nil {
			vm.logger.WithError(err).Error("Failed to parse request body")
			return c.Status(fiber.StatusBadRequest).JSON(dtos.StandardErrorResponse{
				Success: false,
				Message: "Invalid request body",
				Error: dtos.ErrorInfo{
					Code:    "INVALID_REQUEST",
					Message: "Failed to parse JSON request body",
					Details: err.Error(),
				},
			})
		}

		// Validate the parsed struct
		if err := vm.validator.Struct(target); err != nil {
			vm.logger.WithError(err).Error("Request body validation failed")

			// Extract validation errors
			validationErrors := make([]fiber.Map, 0)
			if validatorErrors, ok := err.(validator.ValidationErrors); ok {
				for _, fieldError := range validatorErrors {
					validationErrors = append(validationErrors, fiber.Map{
						"field":   fieldError.Field(),
						"tag":     fieldError.Tag(),
						"value":   fieldError.Value(),
						"message": getValidationMessage(fieldError),
					})
				}
			}

			return c.Status(fiber.StatusBadRequest).JSON(dtos.StandardErrorResponse{
				Success: false,
				Message: "Invalid request body",
				Error: dtos.ErrorInfo{
					Code:    "INVALID_REQUEST",
					Message: "Request body contains invalid data",
					Details: validationErrors,
				},
			})
		}

		// Store the validated data in context for handler use
		c.Locals("validatedBody", target)
		return c.Next()
	}
}

// ValidateParams returns a middleware that validates URL parameters
func (vm *ValidationMiddleware) ValidateParams(target interface{}) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Parse URL parameters into the target struct
		if err := c.ParamsParser(target); err != nil {
			vm.logger.WithError(err).Error("Failed to parse URL parameters")
			return c.Status(fiber.StatusBadRequest).JSON(dtos.StandardErrorResponse{
				Success: false,
				Message: "Invalid request parameters",
				Error: dtos.ErrorInfo{
					Code:    "INVALID_REQUEST",
					Message: "Failed to parse URL parameters",
					Details: err.Error(),
				},
			})
		}

		// Validate the parsed struct
		if err := vm.validator.Struct(target); err != nil {
			vm.logger.WithError(err).Error("URL parameters validation failed")

			// Extract validation errors
			validationErrors := make([]fiber.Map, 0)
			if validatorErrors, ok := err.(validator.ValidationErrors); ok {
				for _, fieldError := range validatorErrors {
					validationErrors = append(validationErrors, fiber.Map{
						"field":   fieldError.Field(),
						"tag":     fieldError.Tag(),
						"value":   fieldError.Value(),
						"message": getValidationMessage(fieldError),
					})
				}
			}

			return c.Status(fiber.StatusBadRequest).JSON(dtos.StandardErrorResponse{
				Success: false,
				Message: "Invalid request parameters",
				Error: dtos.ErrorInfo{
					Code:    "INVALID_REQUEST",
					Message: "URL parameters contain invalid data",
					Details: validationErrors,
				},
			})
		}

		// Store the validated data in context for handler use
		c.Locals("validatedParams", target)
		return c.Next()
	}
}

// ValidateQuery returns a middleware that validates query parameters
func (vm *ValidationMiddleware) ValidateQuery(target interface{}) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Parse query parameters into the target struct
		if err := c.QueryParser(target); err != nil {
			vm.logger.WithError(err).Error("Failed to parse query parameters")
			return c.Status(fiber.StatusBadRequest).JSON(dtos.StandardErrorResponse{
				Success: false,
				Message: "Invalid query parameters",
				Error: dtos.ErrorInfo{
					Code:    "INVALID_REQUEST",
					Message: "Failed to parse query parameters",
					Details: err.Error(),
				},
			})
		}

		// Validate the parsed struct
		if err := vm.validator.Struct(target); err != nil {
			vm.logger.WithError(err).Error("Query parameters validation failed")

			// Extract validation errors
			validationErrors := make([]fiber.Map, 0)
			if validatorErrors, ok := err.(validator.ValidationErrors); ok {
				for _, fieldError := range validatorErrors {
					validationErrors = append(validationErrors, fiber.Map{
						"field":   fieldError.Field(),
						"tag":     fieldError.Tag(),
						"value":   fieldError.Value(),
						"message": getValidationMessage(fieldError),
					})
				}
			}

			return c.Status(fiber.StatusBadRequest).JSON(dtos.StandardErrorResponse{
				Success: false,
				Message: "Invalid query parameters",
				Error: dtos.ErrorInfo{
					Code:    "INVALID_REQUEST",
					Message: "Query parameters contain invalid data",
					Details: validationErrors,
				},
			})
		}

		// Store the validated data in context for handler use
		c.Locals("validatedQuery", target)
		return c.Next()
	}
}

// getValidationMessage returns a human-readable message for validation errors
func getValidationMessage(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "This field is required"
	case "email":
		return "Must be a valid email address"
	case "min":
		return "Value is too short"
	case "max":
		return "Value is too long"
	case "len":
		return "Invalid value"
	case "uuid4":
		return "Must be a valid UUID"
	case "alpha":
		return "Must contain only alphabetic characters"
	case "alphanum":
		return "Must contain only alphanumeric characters"
	case "numeric":
		return "Must be a number"
	case "gt":
		return "Value must be greater than " + fe.Param()
	case "gte":
		return "Value must be greater than or equal to " + fe.Param()
	case "lt":
		return "Value must be less than " + fe.Param()
	case "lte":
		return "Value must be less than or equal to " + fe.Param()
	default:
		return "Invalid value"
	}
}
