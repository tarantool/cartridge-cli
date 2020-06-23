package commands

import (
	"fmt"

	"github.com/apex/log"
	"github.com/spf13/cobra"

	"github.com/tarantool/cartridge-cli/cli/running"
)

func init() {
	rootCmd.AddCommand(stopCmd)

	stopCmd.Flags().StringVar(&projectCtx.RunDir, "run-dir", "", runDirFlagDoc)
	stopCmd.Flags().StringVar(&projectCtx.ConfPath, "cfg", "", cfgFlagDoc)

	stopCmd.Flags().BoolVar(&projectCtx.WithStateboard, "stateboard", false, stateboardFlagDoc)
	stopCmd.Flags().BoolVar(&projectCtx.StateboardOnly, "stateboard-only", false, stateboardOnlyFlagDoc)
}

var stopCmd = &cobra.Command{
	Use:   "stop [INSTANCE_ID...]",
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
	if err := running.FillCtx(&projectCtx, args); err != nil {
		return err
	}

	if err := running.Stop(&projectCtx); err != nil {
		return err
	}

	return nil
}
