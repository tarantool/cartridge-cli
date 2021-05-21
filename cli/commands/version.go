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
		Short: "Show the Tarantool Cartridge version information",
		Long:  `Show the Tarantool Cartridge version information`,
		Args:  cobra.MaximumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(version.BuildVersionString(projectPath, needRocks))
		},
	}

	rootCmd.AddCommand(versionCmd)

	versionCmd.Flags().StringVar(&projectPath, "project-path", ".", projectPathUsage)
	versionCmd.Flags().BoolVar(&needRocks, "rocks", false, "123")
}
