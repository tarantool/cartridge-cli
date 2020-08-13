package commands

import (
	"fmt"

	"github.com/apex/log"
	"github.com/spf13/cobra"

	"github.com/tarantool/cartridge-cli/cli/project"
	"github.com/tarantool/cartridge-cli/cli/running"
)

var (
	timeoutStr string
)

func init() {
	var startCmd = &cobra.Command{
		Use:   "start [INSTANCE-NAME...]",
		Short: "Start application instance(s)",
		Long:  fmt.Sprintf("Start application instance(s)\n\n%s", runningCommonUsage),
		Run: func(cmd *cobra.Command, args []string) {
			err := runStartCmd(cmd, args)
			if err != nil {
				log.Fatalf(err.Error())
			}
		},
	}

	rootCmd.AddCommand(startCmd)

	// FLAGS
	configureFlags(startCmd)

	// application name flag
	addNameFlag(startCmd)

	// start-specific flags
	startCmd.Flags().BoolVarP(&ctx.Running.Daemonize, "daemonize", "d", false, daemonizeUsage)
	startCmd.Flags().StringVar(&timeoutStr, "timeout", "", timeoutUsage)

	// stateboard flags
	addStateboardRunningFlags(startCmd)

	// common running paths
	addCommonRunningPathsFlags(startCmd)
	// start-specific paths
	startCmd.Flags().StringVar(&ctx.Running.DataDir, "data-dir", "", dataDirUsage)
	startCmd.Flags().StringVar(&ctx.Running.LogDir, "log-dir", "", logDirUsage)
	startCmd.Flags().StringVar(&ctx.Running.Entrypoint, "script", "", scriptUsage)

}

func runStartCmd(cmd *cobra.Command, args []string) error {
	var err error

	if err := setDefaultValue(cmd.Flags(), "timeout", defaultStartTimeout.String()); err != nil {
		return project.InternalError("Failed to set default timeout value: %s", err)
	}

	if ctx.Running.StartTimeout, err = getDuration(timeoutStr); err != nil {
		cmd.Usage()
		return fmt.Errorf(`Invalid argument %q for "--%s" flag: %s`, timeoutStr, "timeout", err)
	}

	if err := running.FillCtx(&ctx, args); err != nil {
		return err
	}

	if err := running.Start(&ctx); err != nil {
		return err
	}

	return nil
}
