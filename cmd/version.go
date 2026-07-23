package cmd

import (
	"fmt"
	"runtime/debug"

	"github.com/spf13/cobra"
)

// version is injected at release time via -ldflags "-X ...cmd.version=vX.Y.Z".
var version = "dev"

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the anvil version",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, _ []string) {
		fmt.Fprintln(cmd.OutOrStdout(), versionString())
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}

func versionString() string {
	if version != "dev" {
		return "anvil " + version
	}
	if info, ok := debug.ReadBuildInfo(); ok {
		if v := info.Main.Version; v != "" && v != "(devel)" {
			return "anvil " + v
		}
	}
	return "anvil dev"
}
