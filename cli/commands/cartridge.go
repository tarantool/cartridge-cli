package commands

import (
	"github.com/apex/log"
	"github.com/apex/log/handlers/cli"
	"github.com/spf13/cobra"

	"github.com/tarantool/cartridge-cli/cli/project"
)

var (
	projectCtx project.ProjectCtx

	rootCmd = &cobra.Command{
		Use:   "cartridge",
		Short: "Tarantool Cartridge command-line interface",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			setLogLevel()
		},
	}
)

func init() {
	rootCmd.PersistentFlags().BoolVar(&projectCtx.Verbose, "verbose", false, "Verbose output")
	rootCmd.PersistentFlags().BoolVar(&projectCtx.Quiet, "quiet", false, "Hide build commands output")
	rootCmd.PersistentFlags().BoolVar(&projectCtx.Debug, "debug", false, "Debug mode")

	initLogger()
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatalf(err.Error())
	}
}

func initLogger() {
	log.SetHandler(cli.Default)
}

func setLogLevel() {
	if projectCtx.Verbose || projectCtx.Debug {
		log.SetLevel(log.DebugLevel)
	}
}
