package unit

import (
	"errors"
	"testing"

	v1_models "prayog-serviceability-service/internal/shared/models/v1"
)

// TestDHLPackageWeightValidation tests DHL package weight validation requirements
func TestDHLPackageWeightValidation(t *testing.T) {
	tests := []struct {
		name          string
		packages      []v1_models.Package
		shouldBeValid bool
		expectedError string
		description   string
	}{
		{
			name: "Valid Single Package Weight",
			packages: []v1_models.Package{
				{
					Weight: &v1_models.Weight{Value: 5.0, Unit: "kg"},
					Dimensions: &v1_models.Dimensions{
						Length: 30, Width: 20, Height: 15, Unit: "cm",
					},
				},
			},
			shouldBeValid: true,
			expectedError: "",
			description:   "Should accept valid single package weight",
		},
		{
			name: "Multiple Packages Valid Weights",
			packages: []v1_models.Package{
				{
					Weight: &v1_models.Weight{Value: 2.5, Unit: "kg"},
					Dimensions: &v1_models.Dimensions{
						Length: 25, Width: 15, Height: 10, Unit: "cm",
					},
				},
				{
					Weight: &v1_models.Weight{Value: 3.0, Unit: "kg"},
					Dimensions: &v1_models.Dimensions{
						Length: 35, Width: 25, Height: 20, Unit: "cm",
					},
				},
				{
					Weight: &v1_models.Weight{Value: 1.5, Unit: "kg"},
					Dimensions: &v1_models.Dimensions{
						Length: 20, Width: 15, Height: 8, Unit: "cm",
					},
				},
			},
			shouldBeValid: true,
			expectedError: "",
			description:   "Should accept multiple packages with valid weights",
		},
		{
			name: "Package Weight Too Heavy",
			packages: []v1_models.Package{
				{
					Weight: &v1_models.Weight{Value: 71.0, Unit: "kg"},
					Dimensions: &v1_models.Dimensions{
						Length: 100, Width: 80, Height: 60, Unit: "cm",
					},
				},
			},
			shouldBeValid: false,
			expectedError: "package weight exceeds DHL maximum limit of 70kg",
			description:   "Should reject packages exceeding DHL weight limit",
		},
		{
			name: "Package Weight Too Light",
			packages: []v1_models.Package{
				{
					Weight: &v1_models.Weight{Value: 0.05, Unit: "kg"},
					Dimensions: &v1_models.Dimensions{
						Length: 5, Width: 5, Height: 5, Unit: "cm",
					},
				},
			},
			shouldBeValid: false,
			expectedError: "package weight below DHL minimum limit of 0.1kg",
			description:   "Should reject packages below minimum weight",
		},
		{
			name: "Missing Weight Information",
			packages: []v1_models.Package{
				{
					Dimensions: &v1_models.Dimensions{
						Length: 25, Width: 15, Height: 10, Unit: "cm",
					},
				},
			},
			shouldBeValid: false,
			expectedError: "package weight information is required",
			description:   "Should reject packages without weight information",
		},
		{
			name: "Invalid Weight Unit",
			packages: []v1_models.Package{
				{
					Weight: &v1_models.Weight{Value: 5.0, Unit: "tons"},
					Dimensions: &v1_models.Dimensions{
						Length: 30, Width: 20, Height: 15, Unit: "cm",
					},
				},
			},
			shouldBeValid: false,
			expectedError: "unsupported weight unit: tons",
			description:   "Should reject packages with unsupported weight units",
		},
		{
			name: "Zero Weight Value",
			packages: []v1_models.Package{
				{
					Weight: &v1_models.Weight{Value: 0.0, Unit: "kg"},
					Dimensions: &v1_models.Dimensions{
						Length: 20, Width: 15, Height: 10, Unit: "cm",
					},
				},
			},
			shouldBeValid: false,
			expectedError: "package weight must be greater than zero",
			description:   "Should reject packages with zero weight",
		},
		{
			name: "Negative Weight Value",
			packages: []v1_models.Package{
				{
					Weight: &v1_models.Weight{Value: -1.0, Unit: "kg"},
					Dimensions: &v1_models.Dimensions{
						Length: 20, Width: 15, Height: 10, Unit: "cm",
					},
				},
			},
			shouldBeValid: false,
			expectedError: "package weight cannot be negative",
			description:   "Should reject packages with negative weight",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Validate DHL package weights
			err := validateDHLPackageWeights(tt.packages)

			if tt.shouldBeValid {
				if err != nil {
					t.Errorf("Expected package weights to be valid for %s, got error: %v",
						tt.description, err)
				}
			} else {
				if err == nil {
					t.Errorf("Expected package weight validation error for %s, got no error",
						tt.description)
				} else if err.Error() != tt.expectedError {
					t.Errorf("Expected error '%s' for %s, got '%s'",
						tt.expectedError, tt.description, err.Error())
				}
			}
		})
	}
}

// TestDHLPackageDimensionValidation tests DHL package dimension validation requirements
func TestDHLPackageDimensionValidation(t *testing.T) {
	tests := []struct {
		name          string
		packages      []v1_models.Package
		shouldBeValid bool
		expectedError string
		description   string
	}{
		{
			name: "Valid Package Dimensions",
			packages: []v1_models.Package{
				{
					Weight: &v1_models.Weight{Value: 5.0, Unit: "kg"},
					Dimensions: &v1_models.Dimensions{
						Length: 60, Width: 40, Height: 30, Unit: "cm",
					},
				},
			},
			shouldBeValid: true,
			expectedError: "",
			description:   "Should accept valid package dimensions",
		},
		{
			name: "Package Exceeds Length Limit",
			packages: []v1_models.Package{
				{
					Weight: &v1_models.Weight{Value: 5.0, Unit: "kg"},
					Dimensions: &v1_models.Dimensions{
						Length: 200, Width: 40, Height: 30, Unit: "cm",
					},
				},
			},
			shouldBeValid: false,
			expectedError: "package length exceeds DHL maximum limit of 120cm",
			description:   "Should reject packages exceeding length limit",
		},
		{
			name: "Package Exceeds Width Limit",
			packages: []v1_models.Package{
				{
					Weight: &v1_models.Weight{Value: 5.0, Unit: "kg"},
					Dimensions: &v1_models.Dimensions{
						Length: 60, Width: 100, Height: 30, Unit: "cm",
					},
				},
			},
			shouldBeValid: false,
			expectedError: "package width exceeds DHL maximum limit of 80cm",
			description:   "Should reject packages exceeding width limit",
		},
		{
			name: "Package Exceeds Height Limit",
			packages: []v1_models.Package{
				{
					Weight: &v1_models.Weight{Value: 5.0, Unit: "kg"},
					Dimensions: &v1_models.Dimensions{
						Length: 60, Width: 40, Height: 100, Unit: "cm",
					},
				},
			},
			shouldBeValid: false,
			expectedError: "package height exceeds DHL maximum limit of 80cm",
			description:   "Should reject packages exceeding height limit",
		},
		{
			name: "Missing Dimension Information",
			packages: []v1_models.Package{
				{
					Weight: &v1_models.Weight{Value: 5.0, Unit: "kg"},
				},
			},
			shouldBeValid: false,
			expectedError: "package dimension information is required",
			description:   "Should reject packages without dimension information",
		},
		{
			name: "Invalid Dimension Unit",
			packages: []v1_models.Package{
				{
					Weight: &v1_models.Weight{Value: 5.0, Unit: "kg"},
					Dimensions: &v1_models.Dimensions{
						Length: 60, Width: 40, Height: 30, Unit: "meters",
					},
				},
			},
			shouldBeValid: false,
			expectedError: "unsupported dimension unit: meters",
			description:   "Should reject packages with unsupported dimension units",
		},
		{
			name: "Zero Dimension Values",
			packages: []v1_models.Package{
				{
					Weight: &v1_models.Weight{Value: 5.0, Unit: "kg"},
					Dimensions: &v1_models.Dimensions{
						Length: 0, Width: 40, Height: 30, Unit: "cm",
					},
				},
			},
			shouldBeValid: false,
			expectedError: "package dimensions must be greater than zero",
			description:   "Should reject packages with zero dimension values",
		},
		{
			name: "Negative Dimension Values",
			packages: []v1_models.Package{
				{
					Weight: &v1_models.Weight{Value: 5.0, Unit: "kg"},
					Dimensions: &v1_models.Dimensions{
						Length: -10, Width: 40, Height: 30, Unit: "cm",
					},
				},
			},
			shouldBeValid: false,
			expectedError: "package dimensions cannot be negative",
			description:   "Should reject packages with negative dimension values",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Validate DHL package dimensions
			err := validateDHLPackageDimensions(tt.packages)

			if tt.shouldBeValid {
				if err != nil {
					t.Errorf("Expected package dimensions to be valid for %s, got error: %v",
						tt.description, err)
				}
			} else {
				if err == nil {
					t.Errorf("Expected package dimension validation error for %s, got no error",
						tt.description)
				} else if err.Error() != tt.expectedError {
					t.Errorf("Expected error '%s' for %s, got '%s'",
						tt.expectedError, tt.description, err.Error())
				}
			}
		})
	}
}

// TestDHLMultiplePackageValidation tests validation of multiple packages in a single shipment
func TestDHLMultiplePackageValidation(t *testing.T) {
	tests := []struct {
		name          string
		packages      []v1_models.Package
		shouldBeValid bool
		expectedError string
		description   string
	}{
		{
			name: "Valid Multiple Packages",
			packages: []v1_models.Package{
				{
					Weight: &v1_models.Weight{Value: 2.0, Unit: "kg"},
					Dimensions: &v1_models.Dimensions{
						Length: 30, Width: 20, Height: 15, Unit: "cm",
					},
				},
				{
					Weight: &v1_models.Weight{Value: 3.5, Unit: "kg"},
					Dimensions: &v1_models.Dimensions{
						Length: 40, Width: 30, Height: 25, Unit: "cm",
					},
				},
				{
					Weight: &v1_models.Weight{Value: 1.8, Unit: "kg"},
					Dimensions: &v1_models.Dimensions{
						Length: 25, Width: 18, Height: 12, Unit: "cm",
					},
				},
			},
			shouldBeValid: true,
			expectedError: "",
			description:   "Should accept valid multiple packages",
		},
		{
			name: "Too Many Packages",
			packages: func() []v1_models.Package {
				var packages []v1_models.Package
				for i := 0; i < 11; i++ { // DHL limit is typically 10 packages
					packages = append(packages, v1_models.Package{
						Weight: &v1_models.Weight{Value: 1.0, Unit: "kg"},
						Dimensions: &v1_models.Dimensions{
							Length: 20, Width: 15, Height: 10, Unit: "cm",
						},
					})
				}
				return packages
			}(),
			shouldBeValid: false,
			expectedError: "too many packages: DHL allows maximum 10 packages per shipment",
			description:   "Should reject shipments with too many packages",
		},
		{
			name: "Total Weight Exceeds Limit",
			packages: []v1_models.Package{
				{
					Weight: &v1_models.Weight{Value: 40.0, Unit: "kg"},
					Dimensions: &v1_models.Dimensions{
						Length: 80, Width: 60, Height: 40, Unit: "cm",
					},
				},
				{
					Weight: &v1_models.Weight{Value: 35.0, Unit: "kg"},
					Dimensions: &v1_models.Dimensions{
						Length: 75, Width: 55, Height: 35, Unit: "cm",
					},
				},
			},
			shouldBeValid: false,
			expectedError: "total shipment weight exceeds DHL maximum limit of 70kg",
			description:   "Should reject shipments where total weight exceeds limit",
		},
		{
			name:          "Empty Package List",
			packages:      []v1_models.Package{},
			shouldBeValid: false,
			expectedError: "at least one package is required",
			description:   "Should reject empty package lists",
		},
		{
			name: "Mixed Valid and Invalid Packages",
			packages: []v1_models.Package{
				{
					Weight: &v1_models.Weight{Value: 2.0, Unit: "kg"},
					Dimensions: &v1_models.Dimensions{
						Length: 30, Width: 20, Height: 15, Unit: "cm",
					},
				},
				{
					Weight: &v1_models.Weight{Value: -1.0, Unit: "kg"}, // Invalid
					Dimensions: &v1_models.Dimensions{
						Length: 40, Width: 30, Height: 25, Unit: "cm",
					},
				},
			},
			shouldBeValid: false,
			expectedError: "package weight cannot be negative",
			description:   "Should reject shipments with any invalid packages",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Validate multiple DHL packages
			err := validateDHLMultiplePackages(tt.packages)

			if tt.shouldBeValid {
				if err != nil {
					t.Errorf("Expected multiple packages to be valid for %s, got error: %v",
						tt.description, err)
				}
			} else {
				if err == nil {
					t.Errorf("Expected multiple package validation error for %s, got no error",
						tt.description)
				} else if err.Error() != tt.expectedError {
					t.Errorf("Expected error '%s' for %s, got '%s'",
						tt.expectedError, tt.description, err.Error())
				}
			}
		})
	}
}

// TestDHLInternationalPackageRequirements tests international-specific package requirements
func TestDHLInternationalPackageRequirements(t *testing.T) {
	tests := []struct {
		name          string
		packages      []v1_models.Package
		sourceCountry string
		destCountry   string
		shouldBeValid bool
		expectedError string
		description   string
	}{
		{
			name: "Valid International Package",
			packages: []v1_models.Package{
				{
					Weight: &v1_models.Weight{Value: 5.0, Unit: "kg"},
					Dimensions: &v1_models.Dimensions{
						Length: 50, Width: 30, Height: 20, Unit: "cm",
					},
				},
			},
			sourceCountry: "IN",
			destCountry:   "US",
			shouldBeValid: true,
			expectedError: "",
			description:   "Should accept valid international packages",
		},
		{
			name: "International Package Too Heavy",
			packages: []v1_models.Package{
				{
					Weight: &v1_models.Weight{Value: 31.0, Unit: "kg"},
					Dimensions: &v1_models.Dimensions{
						Length: 80, Width: 60, Height: 40, Unit: "cm",
					},
				},
			},
			sourceCountry: "IN",
			destCountry:   "US",
			shouldBeValid: false,
			expectedError: "international package weight exceeds limit of 30kg",
			description:   "Should reject international packages exceeding weight limit",
		},
		{
			name: "Domestic Package High Weight Allowed",
			packages: []v1_models.Package{
				{
					Weight: &v1_models.Weight{Value: 50.0, Unit: "kg"},
					Dimensions: &v1_models.Dimensions{
						Length: 100, Width: 70, Height: 50, Unit: "cm",
					},
				},
			},
			sourceCountry: "IN",
			destCountry:   "IN",
			shouldBeValid: true,
			expectedError: "",
			description:   "Should allow higher weights for domestic packages",
		},
		{
			name: "International Package Dimension Restrictions",
			packages: []v1_models.Package{
				{
					Weight: &v1_models.Weight{Value: 10.0, Unit: "kg"},
					Dimensions: &v1_models.Dimensions{
						Length: 110, Width: 40, Height: 30, Unit: "cm",
					},
				},
			},
			sourceCountry: "IN",
			destCountry:   "GB",
			shouldBeValid: false,
			expectedError: "international package length exceeds limit of 100cm",
			description:   "Should enforce stricter dimension limits for international packages",
		},
		{
			name: "Restricted Country Pair",
			packages: []v1_models.Package{
				{
					Weight: &v1_models.Weight{Value: 2.0, Unit: "kg"},
					Dimensions: &v1_models.Dimensions{
						Length: 30, Width: 20, Height: 15, Unit: "cm",
					},
				},
			},
			sourceCountry: "IN",
			destCountry:   "XX", // Fictional restricted country
			shouldBeValid: false,
			expectedError: "DHL does not support shipments to country XX",
			description:   "Should reject packages to restricted countries",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Validate international package requirements
			err := validateDHLInternationalPackageRequirements(tt.packages, tt.sourceCountry, tt.destCountry)

			if tt.shouldBeValid {
				if err != nil {
					t.Errorf("Expected international package to be valid for %s, got error: %v",
						tt.description, err)
				}
			} else {
				if err == nil {
					t.Errorf("Expected international package validation error for %s, got no error",
						tt.description)
				} else if err.Error() != tt.expectedError {
					t.Errorf("Expected error '%s' for %s, got '%s'",
						tt.expectedError, tt.description, err.Error())
				}
			}
		})
	}
}

// Helper functions for DHL package validation

func validateDHLPackageWeights(packages []v1_models.Package) error {
	for i, pkg := range packages {
		if pkg.Weight == nil {
			return errors.New("package weight information is required")
		}

		weight := pkg.Weight
		if weight.Value < 0 {
			return errors.New("package weight cannot be negative")
		}
		if weight.Value == 0 {
			return errors.New("package weight must be greater than zero")
		}
		if weight.Value < 0.1 {
			return errors.New("package weight below DHL minimum limit of 0.1kg")
		}
		if weight.Value > 70.0 {
			return errors.New("package weight exceeds DHL maximum limit of 70kg")
		}

		// Validate weight unit
		validUnits := []string{"kg", "g", "lb", "oz"}
		validUnit := false
		for _, unit := range validUnits {
			if weight.Unit == unit {
				validUnit = true
				break
			}
		}
		if !validUnit {
			return errors.New("unsupported weight unit: " + weight.Unit)
		}

		_ = i // Use the index to avoid unused variable warning
	}

	return nil
}

func validateDHLPackageDimensions(packages []v1_models.Package) error {
	for i, pkg := range packages {
		if pkg.Dimensions == nil {
			return errors.New("package dimension information is required")
		}

		dim := pkg.Dimensions
		if dim.Length < 0 || dim.Width < 0 || dim.Height < 0 {
			return errors.New("package dimensions cannot be negative")
		}
		if dim.Length == 0 || dim.Width == 0 || dim.Height == 0 {
			return errors.New("package dimensions must be greater than zero")
		}
		if dim.Length > 120 {
			return errors.New("package length exceeds DHL maximum limit of 120cm")
		}
		if dim.Width > 80 {
			return errors.New("package width exceeds DHL maximum limit of 80cm")
		}
		if dim.Height > 80 {
			return errors.New("package height exceeds DHL maximum limit of 80cm")
		}

		// Validate dimension unit
		validUnits := []string{"cm", "mm", "in", "ft"}
		validUnit := false
		for _, unit := range validUnits {
			if dim.Unit == unit {
				validUnit = true
				break
			}
		}
		if !validUnit {
			return errors.New("unsupported dimension unit: " + dim.Unit)
		}

		_ = i // Use the index to avoid unused variable warning
	}

	return nil
}

func validateDHLMultiplePackages(packages []v1_models.Package) error {
	if len(packages) == 0 {
		return errors.New("at least one package is required")
	}

	if len(packages) > 10 {
		return errors.New("too many packages: DHL allows maximum 10 packages per shipment")
	}

	// Validate individual packages first
	if err := validateDHLPackageWeights(packages); err != nil {
		return err
	}
	if err := validateDHLPackageDimensions(packages); err != nil {
		return err
	}

	// Calculate total weight
	totalWeight := 0.0
	for _, pkg := range packages {
		if pkg.Weight != nil {
			// Convert all weights to kg for comparison
			weight := pkg.Weight.Value
			switch pkg.Weight.Unit {
			case "g":
				weight = weight / 1000
			case "lb":
				weight = weight * 0.453592
			case "oz":
				weight = weight * 0.0283495
			}
			totalWeight += weight
		}
	}

	if totalWeight > 70.0 {
		return errors.New("total shipment weight exceeds DHL maximum limit of 70kg")
	}

	return nil
}

func validateDHLInternationalPackageRequirements(packages []v1_models.Package, sourceCountry, destCountry string) error {
	// Check if DHL supports the destination country
	restrictedCountries := []string{"XX", "YY", "ZZ"} // Fictional restricted countries
	for _, restricted := range restrictedCountries {
		if destCountry == restricted {
			return errors.New("DHL does not support shipments to country " + destCountry)
		}
	}

	isInternational := sourceCountry != destCountry

	for _, pkg := range packages {
		if pkg.Weight != nil {
			weight := pkg.Weight.Value
			// Convert weight to kg if needed
			switch pkg.Weight.Unit {
			case "g":
				weight = weight / 1000
			case "lb":
				weight = weight * 0.453592
			case "oz":
				weight = weight * 0.0283495
			}

			// International packages have stricter weight limits
			if isInternational && weight > 30.0 {
				return errors.New("international package weight exceeds limit of 30kg")
			}
		}

		if pkg.Dimensions != nil {
			dim := pkg.Dimensions
			// Convert dimensions to cm if needed
			length, width, height := dim.Length, dim.Width, dim.Height
			switch dim.Unit {
			case "mm":
				length, width, height = length/10, width/10, height/10
			case "in":
				length, width, height = length*2.54, width*2.54, height*2.54
			case "ft":
				length, width, height = length*30.48, width*30.48, height*30.48
			}

			// International packages have stricter dimension limits
			if isInternational {
				if length > 100 {
					return errors.New("international package length exceeds limit of 100cm")
				}
				if width > 60 {
					return errors.New("international package width exceeds limit of 60cm")
				}
				if height > 60 {
					return errors.New("international package height exceeds limit of 60cm")
				}
			}
		}
	}

	return nil
}

// Note: stringPtr function is already defined in other test files
