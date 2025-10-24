package main

import (
	"os"

	"github.com/half-ogre/go-kit/cmd/pgkit/subcmd"
	"github.com/half-ogre/go-kit/versionkit"
)

// These variables are set via -ldflags at build time
var (
	version   = ""
	gitCommit = ""
	buildDate = ""
)

func main() {
	// Set build info for the version command
	subcmd.SetBuildInfo(&versionkit.BuildInfo{
		Version:   version,
		GitCommit: gitCommit,
		BuildDate: buildDate,
	})

	if err := subcmd.Execute(); err != nil {
		os.Exit(1)
	}
}
