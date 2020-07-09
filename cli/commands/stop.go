package commands

import (
	"fmt"

	"github.com/apex/log"
	"github.com/spf13/cobra"

	"github.com/tarantool/cartridge-cli/cli/running"
)

func init() {
	rootCmd.AddCommand(stopCmd)

	stopCmd.Flags().StringVar(&ctx.Project.Name, "name", "", nameFlagDoc)

	stopCmd.Flags().StringVar(&ctx.Running.RunDir, "run-dir", "", runDirFlagDoc)
	stopCmd.Flags().StringVar(&ctx.Running.ConfPath, "cfg", "", cfgFlagDoc)

	stopCmd.Flags().BoolVar(&ctx.Running.WithStateboard, "stateboard", false, stateboardFlagDoc)
	stopCmd.Flags().BoolVar(&ctx.Running.StateboardOnly, "stateboard-only", false, stateboardOnlyFlagDoc)
}

var stopCmd = &cobra.Command{
	Use:   "stop [INSTANCE_NAME...]",
	Short: "Stop instance(s)",
	Long:  fmt.Sprintf("Stop instance(s)n\n%s", runningCommonDoc),
	Run: func(cmd *cobra.Command, args []string) {
		err := runStopCmd(cmd, args)
		if err != nil {
			log.Fatalf(err.Error())
		}
	},
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
