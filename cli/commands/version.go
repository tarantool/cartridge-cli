package commands

import (
	"github.com/apex/log"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/tarantool/cartridge-cli/cli/version"
)

var (
	projectPath       string
	showRocksVersions bool
)

func init() {
	var versionCmd = &cobra.Command{
		Use:   "version",
		Short: "Show version information",
		Args:  cobra.MaximumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			projectPathIsSet := cmd.Flags().Changed("project-path")
			if err := version.PrintVersionString(projectPath, projectPathIsSet, showRocksVersions); err != nil {
				log.Fatalf(err.Error())
			}
		},
	}

	rootCmd.AddCommand(versionCmd)
	addVersionFlags(versionCmd.Flags())
}

func addVersionFlags(flagSet *pflag.FlagSet) {
	flagSet.BoolVar(&showRocksVersions, "rocks", false, needRocksUsage)
	flagSet.StringVar(&projectPath, "project-path", ".", projectPathUsage)
}
