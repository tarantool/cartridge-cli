package commands

import (
	"fmt"

	"github.com/apex/log"
	"github.com/spf13/cobra"

	"github.com/tarantool/cartridge-cli/cli/running"
)

func init() {
	var stopCmd = &cobra.Command{
		Use:   "stop [INSTANCE_NAME...]",
		Short: "Stop instance(s)",
		Long:  fmt.Sprintf("Stop instance(s) (sends SIGTERM)\n%s", runningCommonUsage),
		Run: func(cmd *cobra.Command, args []string) {
			err := runStopCmd(cmd, args)
			if err != nil {
				log.Fatalf(err.Error())
			}
		},
	}

	rootCmd.AddCommand(stopCmd)

	// FLAGS
	configureFlags(stopCmd)

	// application name flag
	addNameFlag(stopCmd)

	// stateboard flags
	addStateboardRunningFlags(stopCmd)

	// common running paths
	addCommonRunningPathsFlags(stopCmd)

	// add --force flag
	stopCmd.Flags().BoolVarP(&ctx.Running.StopForced, "force", "f", false, stopForceUsage)
}

func runStopCmd(cmd *cobra.Command, args []string) error {
	if err := running.FillCtx(&ctx, args); err != nil {
		return err
	}

	if err := running.Stop(&ctx); err != nil {
		return err
	}

	return nil
}
