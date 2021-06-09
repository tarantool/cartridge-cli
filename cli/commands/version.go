package commands

import (
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
			if err := version.PrintVersionString(projectPath, cmd.Flags().Changed("project-path"), showRocksVersion); err != nil {
				log.Fatalf(err.Error())
			}
		},
	}

	rootCmd.AddCommand(versionCmd)
	versionCmd.Flags().BoolVar(&showRocksVersion, "rocks", false, needRocksUsage)
	versionCmd.Flags().StringVar(&projectPath, "project-path", ".", projectPathUsage)
}
