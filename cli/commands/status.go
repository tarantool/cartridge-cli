package commands

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/tarantool/cartridge-cli/cli/project"
	"github.com/tarantool/cartridge-cli/cli/running"
)

func init() {
	rootCmd.AddCommand(statusCmd)

	statusCmd.Flags().StringVar(&projectCtx.RunDir, "run-dir", "", runDirFlagDoc)
	statusCmd.Flags().StringVar(&projectCtx.ConfPath, "cfg", "", cfgFlagDoc)

	statusCmd.Flags().BoolVar(&projectCtx.WithStateboard, "stateboard", false, stateboardFlagDoc)
	statusCmd.Flags().BoolVar(&projectCtx.StateboardOnly, "stateboard-only", false, stateboardOnlyFlagDoc)
}

var statusCmd = &cobra.Command{
	Use:   "status [INSTANCE_ID...]",
	Short: "Get instance(s) status",
	Long:  fmt.Sprintf("Get instance(s) status\n\n%s", runningCommonDoc),
	Run: func(cmd *cobra.Command, args []string) {
		err := runStatusCmd(cmd, args)
		if err != nil {
			log.Fatalf(err.Error())
		}
	},
}

func runStatusCmd(cmd *cobra.Command, args []string) error {
	var err error

	if err := project.FillCtx(&projectCtx); err != nil {
		return err
	}

	if err := project.SetLocalRunningPaths(&projectCtx); err != nil {
		return err
	}

	if projectCtx.Instances, err = running.GetInstancesFromArgs(args, &projectCtx); err != nil {
		return err
	}

	if err := running.Status(&projectCtx); err != nil {
		return err
	}

	return nil
}
