package versionkit

import (
	"fmt"
	"runtime/debug"
)

// BuildInfo holds build-time version information
type BuildInfo struct {
	Version   string
	GitCommit string
	BuildDate string
}

// GetBuildInfo returns a BuildInfo populated from runtime/debug build information
func GetBuildInfo() *BuildInfo {
	bi := &BuildInfo{}

	if info, ok := debug.ReadBuildInfo(); ok {
		// Get version
		if info.Main.Version != "" && info.Main.Version != "(devel)" {
			bi.Version = info.Main.Version
		}

		// Get commit and build date from VCS settings
		for _, setting := range info.Settings {
			switch setting.Key {
			case "vcs.revision":
				if len(setting.Value) > 7 {
					bi.GitCommit = setting.Value[:7] // Short commit hash
				} else {
					bi.GitCommit = setting.Value
				}
			case "vcs.time":
				bi.BuildDate = setting.Value
			}
		}
	}

	return bi
}

// GetBuildVersion returns the version, falling back to "dev" if not set
func (bi *BuildInfo) GetBuildVersion() string {
	if bi.Version != "" {
		return bi.Version
	}
	return "dev"
}

// GetBuildCommit returns the git commit hash, falling back to "unknown" if not set
func (bi *BuildInfo) GetBuildCommit() string {
	if bi.GitCommit != "" {
		return bi.GitCommit
	}
	return "unknown"
}

// GetBuildDate returns the build date, falling back to "unknown" if not set
func (bi *BuildInfo) GetBuildDate() string {
	if bi.BuildDate != "" {
		return bi.BuildDate
	}
	return "unknown"
}

// String returns formatted version info
func (bi *BuildInfo) String() string {
	return fmt.Sprintf("version %s (commit: %s, built: %s)",
		bi.GetBuildVersion(),
		bi.GetBuildCommit(),
		bi.GetBuildDate(),
	)
}
