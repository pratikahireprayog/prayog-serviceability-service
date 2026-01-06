package config

import (
	"os"
	"strings"
)

// GetAllEnvVars returns all environment variables as a map
func GetAllEnvVars() map[string]string {
	envMap := make(map[string]string)
	
	// Get all environment variables from the system
	allEnvVars := os.Environ()
	
	for _, env := range allEnvVars {
		if len(env) > 0 {
			if idx := strings.Index(env, "="); idx > 0 {
				key := env[:idx]
				value := env[idx+1:]
				envMap[key] = value
			}
		}
	}
	
	return envMap
}

