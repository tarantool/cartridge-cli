package commands

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tarantool/cartridge-cli/cli/version"
)

var (
	projectPath string
	needRocks   bool
)

func init() {
	var versionCmd = &cobra.Command{
		Use:   "version",
		Short: "Show version information",
		Long:  `Show version information`,
		Args:  cobra.MaximumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			version := version.BuildVersionString(projectPath, needRocks)
			fmt.Println(version)
		},
	}

	rootCmd.AddCommand(versionCmd)

	versionCmd.Flags().BoolVar(&needRocks, "rocks", false, needRocksUsage)
	versionCmd.Flags().StringVar(&projectPath, "project-path", ".", projectPathUsage)
}
