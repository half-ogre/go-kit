package subcmd

import (
	"fmt"

	"github.com/half-ogre/go-kit/versionkit"
	"github.com/spf13/cobra"
)

var buildInfo *versionkit.BuildInfo

// SetBuildInfo sets the build information for the version command
func SetBuildInfo(bi *versionkit.BuildInfo) {
	buildInfo = bi
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(cmd *cobra.Command, args []string) {
		// Start with provided build info (may be from ldflags)
		bi := buildInfo
		if bi == nil {
			bi = &versionkit.BuildInfo{}
		}

		// Merge with runtime build info for any missing values
		runtimeInfo := versionkit.GetBuildInfo()
		if bi.Version == "" && runtimeInfo.Version != "" {
			bi.Version = runtimeInfo.Version
		}
		if bi.GitCommit == "" && runtimeInfo.GitCommit != "" {
			bi.GitCommit = runtimeInfo.GitCommit
		}
		if bi.BuildDate == "" && runtimeInfo.BuildDate != "" {
			bi.BuildDate = runtimeInfo.BuildDate
		}

		fmt.Printf("pgkit %s\n", bi.String())
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
