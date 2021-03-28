package commands

import (
	"github.com/apex/log"
	"github.com/spf13/cobra"

	"github.com/tarantool/cartridge-cli/cli/build"
)

func init() {
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

	rootCmd.AddCommand(buildCmd)

	// FLAGS
	configureFlags(buildCmd)

	// path to rockspec to use for build
	addSpecFlag(buildCmd)
}

func runBuildCommand(cmd *cobra.Command, args []string) error {
	var err error

	ctx.Project.Path = cmd.Flags().Arg(0)

	err = build.FillCtx(&ctx)
	if err != nil {
		return err
	}

	// build project
	err = build.Run(&ctx)
	if err != nil {
		return err
	}

	return nil
}
