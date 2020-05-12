package commands

import (
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/tarantool/cartridge-cli/project"

	"github.com/spf13/cobra"
)

var (
	projectCtx project.ProjectCtx
	rootCmd    = &cobra.Command{
		Use:   "cartridge",
		Short: "Tarantool Cartridge command-line interface",
	}

	verboseLogLevel bool
)

func init() {
	rootCmd.PersistentFlags().BoolVar(&verboseLogLevel, "verbose", false, "Verbose output")

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
	if verboseLogLevel {
		log.SetLevel(log.DebugLevel)
	}
}
