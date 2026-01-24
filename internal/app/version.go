package app

import "fmt"

// Build-time variables set via ldflags.
var (
	Version   = "dev"
	GitCommit = "unknown"
	GitTag    = ""
	BuildTime = "unknown"
)

// VersionInfo contains version information for the application.
type VersionInfo struct {
	Version   string
	GitCommit string
	GitTag    string
	BuildTime string
}

// GetVersionInfo returns the current version information.
func GetVersionInfo() VersionInfo {
	return VersionInfo{
		Version:   Version,
		GitCommit: GitCommit,
		GitTag:    GitTag,
		BuildTime: BuildTime,
	}
}

// FullString returns a detailed version string for logging.
func (v VersionInfo) FullString() string {
	version := v.Version
	if v.GitTag != "" {
		version = v.GitTag
	}
	return fmt.Sprintf("GoTune %s (commit: %s, built: %s)", version, v.GitCommit, v.BuildTime)
}
