package commands

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/tarantool/cartridge-cli/cli/project"
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
	Use: "stop [INSTANCE_NAME...] ",
	Run: func(cmd *cobra.Command, args []string) {
		err := runStopCmd(cmd, args)
		if err != nil {
			log.Fatalf(err.Error())
		}
	},
}

func runStopCmd(cmd *cobra.Command, args []string) error {
	var err error

	projectCtx.Instances = args

	if err := running.SetLocalRunningPaths(&projectCtx); err != nil {
		return err
	}

	if err := project.FillCtx(&projectCtx); err != nil {
		return err
	}

	if projectCtx.Instances, err = running.GetInstancesFromArgs(args, &projectCtx); err != nil {
		return err
	}

	// stop
	if err := running.Stop(&projectCtx); err != nil {
		return err
	}

	return nil
}
