package commands

import (
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/tarantool/cartridge-cli/src/project"
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
	rootCmd.PersistentFlags().BoolVar(&projectCtx.Quiet, "quiet", false, "Hide commands output")
	rootCmd.PersistentFlags().BoolVar(&projectCtx.Debug, "debug", false, "Debug mode")

	initLogger()
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatalf(err.Error())
	}
}

func initLogger() {
	log.SetFormatter(&log.TextFormatter{
		DisableLevelTruncation: true,
		DisableTimestamp:       true,
		PadLevelText:           true,
	})

	log.SetOutput(os.Stdout)
}

func setLogLevel() {
	if projectCtx.Verbose {
		log.SetLevel(log.DebugLevel)
	}

	if projectCtx.Debug {
		log.SetLevel(log.TraceLevel)
	}
}
