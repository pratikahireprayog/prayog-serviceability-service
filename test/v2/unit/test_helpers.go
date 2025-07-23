package unit

// Common helper functions used across multiple test files

// stringPtr returns a pointer to the given string value
func stringPtr(s string) *string {
	return &s
}

// intPtr returns a pointer to the given int value
func intPtr(i int) *int {
	return &i
}

// float64Ptr returns a pointer to the given float64 value
func float64Ptr(f float64) *float64 {
	return &f
}

// boolPtr returns a pointer to the given bool value
func boolPtr(b bool) *bool {
	return &b
}
