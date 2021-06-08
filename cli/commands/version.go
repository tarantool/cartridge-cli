package commands

import (
	"fmt"

	"github.com/apex/log"
	"github.com/spf13/cobra"
	"github.com/tarantool/cartridge-cli/cli/version"
)

var (
	projectPath      string
	showRocksVersion bool
)

func init() {
	var versionCmd = &cobra.Command{
		Use:   "version",
		Short: "Show version information",
		Args:  cobra.MaximumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			versionString, err := version.BuildVersionString(projectPath, showRocksVersion)
			fmt.Println(versionString)

			if err != nil && showRocksVersion {
				log.Errorf("%s", err)
			} else if err != nil {
				log.Warnf("%s", err)
			}
		},
	}

	rootCmd.AddCommand(versionCmd)

	versionCmd.Flags().BoolVar(&showRocksVersion, "rocks", false, needRocksUsage)
	versionCmd.Flags().StringVar(&projectPath, "project-path", ".", projectPathUsage)
}
