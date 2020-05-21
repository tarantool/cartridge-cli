package commands

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/tarantool/cartridge-cli/src/build"
	"github.com/tarantool/cartridge-cli/src/project"
)

func init() {
	rootCmd.AddCommand(buildCmd)
}

var buildCmd = &cobra.Command{
	Use:   "build [PATH]",
	Short: "Build application for local development",
	Long:  "Build application in specified PATH (default \".\")",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		err := runBuildCommand(cmd, args)
		if err != nil {
			log.Fatalf(err.Error())
		}
	},
}

func runBuildCommand(cmd *cobra.Command, args []string) error {
	var err error

	projectCtx.Path = cmd.Flags().Arg(0)

	// fill context
	err = project.FillCtx(&projectCtx)
	if err != nil {
		return err
	}

	projectCtx.BuildDir = projectCtx.Path

	// build project
	err = build.Run(&projectCtx)
	if err != nil {
		return err
	}

	return nil
}
