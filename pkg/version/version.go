package version

import (
	"fmt"
	"runtime/debug"
)

// Version information
var (
	// Major version component
	Major = "1"
	// Minor version component
	Minor = "0"
	// Patch version component
	Patch = "0"
	// Pre-release version component (e.g., "alpha", "beta", "rc1")
	PreRelease = ""
	// Build metadata
	BuildMetadata = ""
	// Commit SHA from git, injected at build time
	Commit = ""
	// Build time, injected at build time
	BuildTime = ""
	// Dirty indicates if the build was created with uncommitted changes
	Dirty = ""
)

// Version returns the full version string
func Version() string {
	version := fmt.Sprintf("%s.%s.%s", Major, Minor, Patch)
	if PreRelease != "" {
		version = fmt.Sprintf("%s-%s", version, PreRelease)
	}
	if BuildMetadata != "" {
		version = fmt.Sprintf("%s+%s", version, BuildMetadata)
	}
	return version
}

// VersionInfo returns detailed version information
func VersionInfo() map[string]string {
	// Try to extract version info from build info if available
	buildInfo := extractBuildInfo()

	return map[string]string{
		"version":        Version(),
		"commit":         getValueOrDefault(Commit, buildInfo["vcs.revision"]),
		"build_time":     getValueOrDefault(BuildTime, buildInfo["vcs.time"]),
		"go_version":     buildInfo["go.version"],
		"dirty":          getValueOrDefault(Dirty, ""),
		"build_metadata": BuildMetadata,
	}
}

// extractBuildInfo attempts to extract info from Go build information
func extractBuildInfo() map[string]string {
	result := make(map[string]string)

	// Get build info from runtime/debug
	if info, ok := debug.ReadBuildInfo(); ok {
		result["go.version"] = info.GoVersion

		for _, setting := range info.Settings {
			result[setting.Key] = setting.Value
		}
	}

	return result
}

// getValueOrDefault returns the first non-empty value or empty string
func getValueOrDefault(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}
