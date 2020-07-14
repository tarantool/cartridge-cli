package commands

import (
	"fmt"
	"strconv"

	"github.com/apex/log"
	"github.com/spf13/cobra"

	"github.com/tarantool/cartridge-cli/cli/project"
	"github.com/tarantool/cartridge-cli/cli/running"
)

func init() {
	var logCmd = &cobra.Command{
		Use:   "log [INSTANCE_NAME...]",
		Short: "Get logs of instance(s)",
		Long:  fmt.Sprintf("Get logs of instance(s)\n\n%s", runningCommonUsage),
		Run: func(cmd *cobra.Command, args []string) {
			err := runLogCmd(cmd, args)
			if err != nil {
				log.Fatalf(err.Error())
			}
		},
	}

	rootCmd.AddCommand(logCmd)

	// FLAGS
	configureFlags(logCmd)

	// application name flag
	addNameFlag(logCmd)

	// log-specific flags
	logCmd.Flags().BoolVarP(&ctx.Running.LogFollow, "follow", "f", false, logFollowUsage)
	logCmd.Flags().IntVarP(&ctx.Running.LogLines, "lines", "n", 0, logLinesUsage)

	// stateboard flags
	addStateboardRunningFlags(logCmd)

	// log-specific paths
	logCmd.Flags().StringVar(&ctx.Running.LogDir, "log-dir", "", logDirUsage)
	// common running paths
	addCommonRunningPathsFlags(logCmd)
}

func runLogCmd(cmd *cobra.Command, args []string) error {
	if err := setDefaultValue(cmd.Flags(), "lines", strconv.Itoa(defaultLogLines)); err != nil {
		return project.InternalError("Failed to set default lines value: %s", err)
	}

	if err := running.FillCtx(&ctx, args); err != nil {
		return err
	}

	if err := running.Log(&ctx); err != nil {
		return err
	}

	return nil
}
