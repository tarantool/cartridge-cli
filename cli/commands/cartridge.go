package commands

import (
	"github.com/apex/log"
	"github.com/apex/log/handlers/cli"
	"github.com/spf13/cobra"
	"github.com/tarantool/cartridge-cli/cli/context"
	"github.com/tarantool/cartridge-cli/cli/version"
)

var (
	ctx context.Ctx

	rootCmd = &cobra.Command{
		Use:   "cartridge",
		Short: "Tarantool Cartridge command-line interface",

		Version: version.BuildVersionString(),
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			setLogLevel()
		},
	}
)

func init() {
	rootCmd.SetVersionTemplate("{{ .Version }}\n")

	rootCmd.PersistentFlags().BoolVar(&ctx.Cli.Verbose, "verbose", false, "Verbose output")
	rootCmd.PersistentFlags().BoolVar(&ctx.Cli.Quiet, "quiet", false, "Hide build commands output")
	rootCmd.PersistentFlags().BoolVar(&ctx.Cli.Debug, "debug", false, "Debug mode")

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
	if ctx.Cli.Debug {
		ctx.Cli.Verbose = true
	}

	if ctx.Cli.Verbose {
		log.SetLevel(log.DebugLevel)
	}
}
