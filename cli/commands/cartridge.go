package commands

import (
	"os"

	"github.com/apex/log"
	"github.com/apex/log/handlers/cli"
	"github.com/spf13/cobra"
	"github.com/tarantool/cartridge-cli/cli/context"
	"github.com/tarantool/cartridge-cli/cli/version"
)

var (
	ctx         context.Ctx
	needVersion bool
	rootCmd     = &cobra.Command{
		Use:   "cartridge",
		Short: "Tarantool Cartridge command-line interface",

		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			setLogLevel()
		},

		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 && !needVersion {
				cmd.Help()
			} else {
				printVersion(cmd)
			}
		},
	}
)

func init() {
	rootCmd.SetVersionTemplate("{{ .Version }}\n")

	rootCmd.PersistentFlags().BoolVar(&ctx.Cli.Verbose, "verbose", false, "Verbose output")
	rootCmd.PersistentFlags().BoolVar(&ctx.Cli.Quiet, "quiet", false, "Hide build commands output")
	rootCmd.PersistentFlags().BoolVar(&ctx.Cli.Debug, "debug", false, "Debug mode")
	rootCmd.Flags().BoolVarP(&needVersion, "version", "v", false, "Show version information")

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

	if ctx.Cli.Quiet {
		log.SetLevel(log.ErrorLevel)
	}
}

func printVersion(cmd *cobra.Command) {
	if err := version.PrintVersionString(projectPath, cmd.Flags().Changed("project-path"), showRocksVersion); err != nil {
		os.Exit(1)
	}
}
