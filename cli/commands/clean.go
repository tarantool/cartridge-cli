package commands

import (
	"fmt"

	"github.com/apex/log"
	"github.com/spf13/cobra"

	"github.com/tarantool/cartridge-cli/cli/running"
)

func init() {
	var cleanCmd = &cobra.Command{
		Use:   "clean [INSTANCE-NAME...]",
		Short: "Clean instance(s) files",
		Long:  fmt.Sprintf("Clean instance(s) files\n\n%s", runningCommonUsage),
		Run: func(cmd *cobra.Command, args []string) {
			err := runCleanCmd(cmd, args)
			if err != nil {
				log.Fatalf(err.Error())
			}
		},
	}

	rootCmd.AddCommand(cleanCmd)

	// FLAGS
	configureFlags(cleanCmd)

	// application name flag
	addNameFlag(cleanCmd)

	// stateboard flags
	addStateboardRunningFlags(cleanCmd)

	// clean-specific paths
	cleanCmd.Flags().StringVar(&ctx.Running.LogDir, "log-dir", "", logDirUsage)
	cleanCmd.Flags().StringVar(&ctx.Running.DataDir, "data-dir", "", dataDirUsage)
	// common running paths
	addCommonRunningPathsFlags(cleanCmd)
}

func runCleanCmd(cmd *cobra.Command, args []string) error {
	if err := running.FillCtx(&ctx, args); err != nil {
		return err
	}

	if err := running.Clean(&ctx); err != nil {
		return err
	}

	return nil
}
