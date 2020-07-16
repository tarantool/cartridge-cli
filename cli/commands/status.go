package commands

import (
	"fmt"

	"github.com/apex/log"
	"github.com/spf13/cobra"

	"github.com/tarantool/cartridge-cli/cli/running"
)

func init() {
	var statusCmd = &cobra.Command{
		Use:   "status [APP_NAME] [INSTANCE_NAME...]",
		Short: "Get instance(s) status",
		Long:  fmt.Sprintf("Get instance(s) status\n\n%s", runningCommonUsage),
		Run: func(cmd *cobra.Command, args []string) {
			err := runStatusCmd(cmd, args)
			if err != nil {
				log.Fatalf(err.Error())
			}
		},
	}

	rootCmd.AddCommand(statusCmd)

	// FLAGS
	configureFlags(statusCmd)

	// application name flag
	addNameFlag(statusCmd)

	// global running flag
	addGlobalRunningFlag(statusCmd)

	// stateboard flags
	addStateboardRunningFlags(statusCmd)

	// common running paths
	addCommonRunningPathsFlags(statusCmd)
}

func runStatusCmd(cmd *cobra.Command, args []string) error {
	if err := running.FillCtx(&ctx, args); err != nil {
		return err
	}

	if err := running.Status(&ctx); err != nil {
		return err
	}

	return nil
}
