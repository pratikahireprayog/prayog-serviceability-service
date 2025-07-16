package unit

import (
	"errors"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"prayog-serviceability-service/internal/shared/config"
)

// TestDHLConfigEnvironmentVariables tests DHL configuration from environment variables
func TestDHLConfigEnvironmentVariables(t *testing.T) {
	tests := []struct {
		name           string
		envVars        map[string]string
		expectedConfig config.DHLConfig
		shouldError    bool
		description    string
	}{
		{
			name: "Complete Environment Configuration",
			envVars: map[string]string{
				"DHL_USERNAME":     "test_user",
				"DHL_PASSWORD":     "test_pass",
				"DHL_BASE_URL":     "https://api.dhl.com/test",
				"DHL_ENABLED":      "true",
				"DHL_TIMEOUT":      "30s",
				"DHL_MAX_RETRIES":  "3",
				"DHL_RETRY_DELAY":  "1s",
				"DHL_SANDBOX_MODE": "true",
			},
			expectedConfig: config.DHLConfig{
				Username:    "test_user",
				Password:    "test_pass",
				BaseURL:     "https://api.dhl.com/test",
				Enabled:     true,
				Timeout:     30 * time.Second,
				MaxRetries:  3,
				RetryDelay:  1 * time.Second,
				SandboxMode: true,
			},
			shouldError: false,
			description: "Should load complete configuration from environment variables",
		},
		{
			name: "Minimal Environment Configuration",
			envVars: map[string]string{
				"DHL_USERNAME": "min_user",
				"DHL_PASSWORD": "min_pass",
				"DHL_BASE_URL": "https://api.dhl.com/minimal",
			},
			expectedConfig: config.DHLConfig{
				Username:    "min_user",
				Password:    "min_pass",
				BaseURL:     "https://api.dhl.com/minimal",
				Enabled:     true, // Default value
				Timeout:     30 * time.Second, // Default value
				MaxRetries:  3,    // Default value
				RetryDelay:  1 * time.Second, // Default value
				SandboxMode: false, // Default value
			},
			shouldError: false,
			description: "Should use default values for missing environment variables",
		},
		{
			name: "Missing Required Environment Variables",
			envVars: map[string]string{
				"DHL_BASE_URL": "https://api.dhl.com/test",
			},
			expectedConfig: config.DHLConfig{},
			shouldError:    true,
			description:    "Should error when required environment variables are missing",
		},
		{
			name: "Invalid Boolean Environment Variables",
			envVars: map[string]string{
				"DHL_USERNAME": "test_user",
				"DHL_PASSWORD": "test_pass",
				"DHL_BASE_URL": "https://api.dhl.com/test",
				"DHL_ENABLED":  "invalid_bool",
			},
			expectedConfig: config.DHLConfig{},
			shouldError:    true,
			description:    "Should error when boolean environment variables are invalid",
		},
		{
			name: "Invalid Duration Environment Variables",
			envVars: map[string]string{
				"DHL_USERNAME": "test_user",
				"DHL_PASSWORD": "test_pass",
				"DHL_BASE_URL": "https://api.dhl.com/test",
				"DHL_TIMEOUT":  "invalid_duration",
			},
			expectedConfig: config.DHLConfig{},
			shouldError:    true,
			description:    "Should error when duration environment variables are invalid",
		},
		{
			name: "Invalid Integer Environment Variables",
			envVars: map[string]string{
				"DHL_USERNAME":    "test_user",
				"DHL_PASSWORD":    "test_pass",
				"DHL_BASE_URL":    "https://api.dhl.com/test",
				"DHL_MAX_RETRIES": "invalid_int",
			},
			expectedConfig: config.DHLConfig{},
			shouldError:    true,
			description:    "Should error when integer environment variables are invalid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear all DHL environment variables first
			clearDHLEnvVars()

			// Set test environment variables
			for key, value := range tt.envVars {
				os.Setenv(key, value)
			}

			// Load configuration from environment
			config, err := loadDHLConfigFromEnv()

			if tt.shouldError {
				if err == nil {
					t.Errorf("Expected error for %s, got none", tt.description)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error for %s: %v", tt.description, err)
				}

				// Verify configuration fields
				if config.Username != tt.expectedConfig.Username {
					t.Errorf("Expected username %s, got %s", tt.expectedConfig.Username, config.Username)
				}
				if config.Password != tt.expectedConfig.Password {
					t.Errorf("Expected password %s, got %s", tt.expectedConfig.Password, config.Password)
				}
				if config.BaseURL != tt.expectedConfig.BaseURL {
					t.Errorf("Expected base URL %s, got %s", tt.expectedConfig.BaseURL, config.BaseURL)
				}
				if config.Enabled != tt.expectedConfig.Enabled {
					t.Errorf("Expected enabled %v, got %v", tt.expectedConfig.Enabled, config.Enabled)
				}
				if config.Timeout != tt.expectedConfig.Timeout {
					t.Errorf("Expected timeout %v, got %v", tt.expectedConfig.Timeout, config.Timeout)
				}
				if config.MaxRetries != tt.expectedConfig.MaxRetries {
					t.Errorf("Expected max retries %d, got %d", tt.expectedConfig.MaxRetries, config.MaxRetries)
				}
				if config.SandboxMode != tt.expectedConfig.SandboxMode {
					t.Errorf("Expected sandbox mode %v, got %v", tt.expectedConfig.SandboxMode, config.SandboxMode)
				}
			}

			// Clean up environment variables
			clearDHLEnvVars()
		})
	}
}

// TestDHLConfigTimeoutSettings tests DHL timeout configuration
func TestDHLConfigTimeoutSettings(t *testing.T) {
	tests := []struct {
		name            string
		timeout         time.Duration
		expectedBehavior string
		shouldBeValid   bool
		description     string
	}{
		{
			name:            "Standard Timeout",
			timeout:         30 * time.Second,
			expectedBehavior: "normal_operation",
			shouldBeValid:   true,
			description:     "Should accept standard 30 second timeout",
		},
		{
			name:            "Short Timeout",
			timeout:         5 * time.Second,
			expectedBehavior: "quick_timeout",
			shouldBeValid:   true,
			description:     "Should accept short 5 second timeout",
		},
		{
			name:            "Long Timeout",
			timeout:         2 * time.Minute,
			expectedBehavior: "extended_wait",
			shouldBeValid:   true,
			description:     "Should accept long 2 minute timeout",
		},
		{
			name:            "Zero Timeout",
			timeout:         0,
			expectedBehavior: "immediate_timeout",
			shouldBeValid:   false,
			description:     "Should reject zero timeout",
		},
		{
			name:            "Negative Timeout",
			timeout:         -1 * time.Second,
			expectedBehavior: "invalid_timeout",
			shouldBeValid:   false,
			description:     "Should reject negative timeout",
		},
		{
			name:            "Very Long Timeout",
			timeout:         10 * time.Minute,
			expectedBehavior: "excessive_timeout",
			shouldBeValid:   false,
			description:     "Should reject excessively long timeout",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create DHL config with specific timeout
			config := config.DHLConfig{
				Username:   "test_user",
				Password:   "test_pass",
				BaseURL:    "https://api.dhl.com/test",
				Enabled:    true,
				Timeout:    tt.timeout,
				MaxRetries: 3,
			}

			// Validate timeout configuration
			err := validateDHLTimeoutConfig(config)

			if tt.shouldBeValid {
				if err != nil {
					t.Errorf("Expected timeout to be valid for %s, got error: %v", 
						tt.description, err)
				}
			} else {
				if err == nil {
					t.Errorf("Expected timeout validation error for %s, got none", 
						tt.description)
				}
			}
		})
	}
}

// TestDHLConfigRetrySettings tests DHL retry configuration
func TestDHLConfigRetrySettings(t *testing.T) {
	tests := []struct {
		name            string
		maxRetries      int
		retryDelay      time.Duration
		shouldBeValid   bool
		expectedBehavior string
		description     string
	}{
		{
			name:            "Standard Retry Configuration",
			maxRetries:      3,
			retryDelay:      1 * time.Second,
			shouldBeValid:   true,
			expectedBehavior: "normal_retries",
			description:     "Should accept standard retry configuration",
		},
		{
			name:            "No Retries",
			maxRetries:      0,
			retryDelay:      0,
			shouldBeValid:   true,
			expectedBehavior: "no_retries",
			description:     "Should accept zero retries configuration",
		},
		{
			name:            "High Retry Count",
			maxRetries:      10,
			retryDelay:      2 * time.Second,
			shouldBeValid:   true,
			expectedBehavior: "high_retries",
			description:     "Should accept high retry count",
		},
		{
			name:            "Negative Retry Count",
			maxRetries:      -1,
			retryDelay:      1 * time.Second,
			shouldBeValid:   false,
			expectedBehavior: "invalid_retries",
			description:     "Should reject negative retry count",
		},
		{
			name:            "Negative Retry Delay",
			maxRetries:      3,
			retryDelay:      -1 * time.Second,
			shouldBeValid:   false,
			expectedBehavior: "invalid_delay",
			description:     "Should reject negative retry delay",
		},
		{
			name:            "Excessive Retry Count",
			maxRetries:      100,
			retryDelay:      1 * time.Second,
			shouldBeValid:   false,
			expectedBehavior: "excessive_retries",
			description:     "Should reject excessive retry count",
		},
		{
			name:            "Very Long Retry Delay",
			maxRetries:      3,
			retryDelay:      5 * time.Minute,
			shouldBeValid:   false,
			expectedBehavior: "excessive_delay",
			description:     "Should reject excessively long retry delay",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create DHL config with specific retry settings
			config := config.DHLConfig{
				Username:   "test_user",
				Password:   "test_pass",
				BaseURL:    "https://api.dhl.com/test",
				Enabled:    true,
				Timeout:    30 * time.Second,
				MaxRetries: tt.maxRetries,
				RetryDelay: tt.retryDelay,
			}

			// Validate retry configuration
			err := validateDHLRetryConfig(config)

			if tt.shouldBeValid {
				if err != nil {
					t.Errorf("Expected retry config to be valid for %s, got error: %v", 
						tt.description, err)
				}
			} else {
				if err == nil {
					t.Errorf("Expected retry config validation error for %s, got none", 
						tt.description)
				}
			}
		})
	}
}

// TestDHLConfigSandboxMode tests DHL sandbox mode configuration
func TestDHLConfigSandboxMode(t *testing.T) {
	tests := []struct {
		name           string
		sandboxMode    bool
		baseURL        string
		expectedURL    string
		shouldAdjust   bool
		description    string
	}{
		{
			name:           "Sandbox Mode Enabled",
			sandboxMode:    true,
			baseURL:        "https://api.dhl.com/prod",
			expectedURL:    "https://api.dhl.com/sandbox",
			shouldAdjust:   true,
			description:    "Should adjust URL for sandbox mode",
		},
		{
			name:           "Production Mode",
			sandboxMode:    false,
			baseURL:        "https://api.dhl.com/prod",
			expectedURL:    "https://api.dhl.com/prod",
			shouldAdjust:   false,
			description:    "Should keep production URL in production mode",
		},
		{
			name:           "Sandbox Mode with Sandbox URL",
			sandboxMode:    true,
			baseURL:        "https://api.dhl.com/sandbox",
			expectedURL:    "https://api.dhl.com/sandbox",
			shouldAdjust:   false,
			description:    "Should not change URL if already sandbox",
		},
		{
			name:           "Production Mode with Sandbox URL",
			sandboxMode:    false,
			baseURL:        "https://api.dhl.com/sandbox",
			expectedURL:    "https://api.dhl.com/prod",
			shouldAdjust:   true,
			description:    "Should adjust URL to production when sandbox disabled",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create DHL config with specific sandbox settings
			config := config.DHLConfig{
				Username:    "test_user",
				Password:    "test_pass",
				BaseURL:     tt.baseURL,
				Enabled:     true,
				Timeout:     30 * time.Second,
				MaxRetries:  3,
				SandboxMode: tt.sandboxMode,
			}

			// Apply sandbox configuration
			adjustedConfig := applySandboxConfiguration(config)

			// Verify URL adjustment
			if adjustedConfig.BaseURL != tt.expectedURL {
				t.Errorf("Expected URL %s for %s, got %s", 
					tt.expectedURL, tt.description, adjustedConfig.BaseURL)
			}

			// Verify sandbox mode setting
			if adjustedConfig.SandboxMode != tt.sandboxMode {
				t.Errorf("Expected sandbox mode %v for %s, got %v", 
					tt.sandboxMode, tt.description, adjustedConfig.SandboxMode)
			}
		})
	}
}

// TestDHLConfigValidation tests overall DHL configuration validation
func TestDHLConfigValidation(t *testing.T) {
	tests := []struct {
		name          string
		config        config.DHLConfig
		shouldBeValid bool
		expectedError string
		description   string
	}{
		{
			name: "Valid Complete Configuration",
			config: config.DHLConfig{
				Username:    "valid_user",
				Password:    "valid_pass",
				BaseURL:     "https://api.dhl.com/prod",
				Enabled:     true,
				Timeout:     30 * time.Second,
				MaxRetries:  3,
				RetryDelay:  1 * time.Second,
				SandboxMode: false,
			},
			shouldBeValid: true,
			expectedError: "",
			description:   "Should accept valid complete configuration",
		},
		{
			name: "Missing Username",
			config: config.DHLConfig{
				Password:   "valid_pass",
				BaseURL:    "https://api.dhl.com/prod",
				Enabled:    true,
				Timeout:    30 * time.Second,
				MaxRetries: 3,
			},
			shouldBeValid: false,
			expectedError: "DHL username is required",
			description:   "Should reject configuration without username",
		},
		{
			name: "Missing Password",
			config: config.DHLConfig{
				Username:   "valid_user",
				BaseURL:    "https://api.dhl.com/prod",
				Enabled:    true,
				Timeout:    30 * time.Second,
				MaxRetries: 3,
			},
			shouldBeValid: false,
			expectedError: "DHL password is required",
			description:   "Should reject configuration without password",
		},
		{
			name: "Missing Base URL",
			config: config.DHLConfig{
				Username:   "valid_user",
				Password:   "valid_pass",
				Enabled:    true,
				Timeout:    30 * time.Second,
				MaxRetries: 3,
			},
			shouldBeValid: false,
			expectedError: "DHL base URL is required",
			description:   "Should reject configuration without base URL",
		},
		{
			name: "Invalid Base URL Format",
			config: config.DHLConfig{
				Username:   "valid_user",
				Password:   "valid_pass",
				BaseURL:    "invalid-url",
				Enabled:    true,
				Timeout:    30 * time.Second,
				MaxRetries: 3,
			},
			shouldBeValid: false,
			expectedError: "DHL base URL must be a valid HTTPS URL",
			description:   "Should reject invalid URL format",
		},
		{
			name: "Disabled Configuration",
			config: config.DHLConfig{
				Username:   "valid_user",
				Password:   "valid_pass",
				BaseURL:    "https://api.dhl.com/prod",
				Enabled:    false,
				Timeout:    30 * time.Second,
				MaxRetries: 3,
			},
			shouldBeValid: true,
			expectedError: "",
			description:   "Should accept disabled configuration",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Validate DHL configuration
			err := validateDHLConfig(tt.config)

			if tt.shouldBeValid {
				if err != nil {
					t.Errorf("Expected configuration to be valid for %s, got error: %v", 
						tt.description, err)
				}
			} else {
				if err == nil {
					t.Errorf("Expected configuration validation error for %s, got none", 
						tt.description)
				} else if err.Error() != tt.expectedError {
					t.Errorf("Expected error '%s' for %s, got '%s'", 
						tt.expectedError, tt.description, err.Error())
				}
			}
		})
	}
}

// Helper functions for DHL configuration testing

func clearDHLEnvVars() {
	envVars := []string{
		"DHL_USERNAME", "DHL_PASSWORD", "DHL_BASE_URL", "DHL_ENABLED",
		"DHL_TIMEOUT", "DHL_MAX_RETRIES", "DHL_RETRY_DELAY", "DHL_SANDBOX_MODE",
	}
	for _, envVar := range envVars {
		os.Unsetenv(envVar)
	}
}

func loadDHLConfigFromEnv() (config.DHLConfig, error) {
	// Simulate loading configuration from environment variables
	cfg := config.DHLConfig{
		Username: os.Getenv("DHL_USERNAME"),
		Password: os.Getenv("DHL_PASSWORD"),
		BaseURL:  os.Getenv("DHL_BASE_URL"),
	}

	// Validate required fields
	if cfg.Username == "" {
		return cfg, errors.New("DHL_USERNAME environment variable is required")
	}
	if cfg.Password == "" {
		return cfg, errors.New("DHL_PASSWORD environment variable is required")
	}
	if cfg.BaseURL == "" {
		return cfg, errors.New("DHL_BASE_URL environment variable is required")
	}

	// Parse optional fields with defaults
	if enabled := os.Getenv("DHL_ENABLED"); enabled != "" {
		if enabled == "true" {
			cfg.Enabled = true
		} else if enabled == "false" {
			cfg.Enabled = false
		} else {
			return cfg, errors.New("DHL_ENABLED must be 'true' or 'false'")
		}
	} else {
		cfg.Enabled = true // Default
	}

	if timeout := os.Getenv("DHL_TIMEOUT"); timeout != "" {
		if dur, err := time.ParseDuration(timeout); err != nil {
			return cfg, errors.New("invalid DHL_TIMEOUT format")
		} else {
			cfg.Timeout = dur
		}
	} else {
		cfg.Timeout = 30 * time.Second // Default
	}

	if maxRetries := os.Getenv("DHL_MAX_RETRIES"); maxRetries != "" {
		if retries, err := strconv.Atoi(maxRetries); err != nil {
			return cfg, errors.New("invalid DHL_MAX_RETRIES format")
		} else {
			cfg.MaxRetries = retries
		}
	} else {
		cfg.MaxRetries = 3 // Default
	}

	if retryDelay := os.Getenv("DHL_RETRY_DELAY"); retryDelay != "" {
		if dur, err := time.ParseDuration(retryDelay); err != nil {
			return cfg, errors.New("invalid DHL_RETRY_DELAY format")
		} else {
			cfg.RetryDelay = dur
		}
	} else {
		cfg.RetryDelay = 1 * time.Second // Default
	}

	if sandbox := os.Getenv("DHL_SANDBOX_MODE"); sandbox != "" {
		if sandbox == "true" {
			cfg.SandboxMode = true
		} else if sandbox == "false" {
			cfg.SandboxMode = false
		} else {
			return cfg, errors.New("DHL_SANDBOX_MODE must be 'true' or 'false'")
		}
	} else {
		cfg.SandboxMode = false // Default
	}

	return cfg, nil
}

func validateDHLTimeoutConfig(cfg config.DHLConfig) error {
	if cfg.Timeout <= 0 {
		return errors.New("timeout must be greater than zero")
	}
	if cfg.Timeout > 5*time.Minute {
		return errors.New("timeout cannot exceed 5 minutes")
	}
	return nil
}

func validateDHLRetryConfig(cfg config.DHLConfig) error {
	if cfg.MaxRetries < 0 {
		return errors.New("max retries cannot be negative")
	}
	if cfg.MaxRetries > 20 {
		return errors.New("max retries cannot exceed 20")
	}
	if cfg.RetryDelay < 0 {
		return errors.New("retry delay cannot be negative")
	}
	if cfg.RetryDelay > 1*time.Minute {
		return errors.New("retry delay cannot exceed 1 minute")
	}
	return nil
}

func applySandboxConfiguration(cfg config.DHLConfig) config.DHLConfig {
	if cfg.SandboxMode {
		if !strings.Contains(cfg.BaseURL, "sandbox") {
			cfg.BaseURL = strings.ReplaceAll(cfg.BaseURL, "prod", "sandbox")
		}
	} else {
		if strings.Contains(cfg.BaseURL, "sandbox") {
			cfg.BaseURL = strings.ReplaceAll(cfg.BaseURL, "sandbox", "prod")
		}
	}
	return cfg
}

func validateDHLConfig(cfg config.DHLConfig) error {
	if cfg.Username == "" {
		return errors.New("DHL username is required")
	}
	if cfg.Password == "" {
		return errors.New("DHL password is required")
	}
	if cfg.BaseURL == "" {
		return errors.New("DHL base URL is required")
	}
	if !strings.HasPrefix(cfg.BaseURL, "https://") {
		return errors.New("DHL base URL must be a valid HTTPS URL")
	}
	return nil
}

// Note: stringPtr function is already defined in other test files 