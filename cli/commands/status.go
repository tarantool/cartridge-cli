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
}

var statusCmd = &cobra.Command{
	Use: "status [INSTANCE_NAME...] ",
	Run: func(cmd *cobra.Command, args []string) {
		err := runStatusCmd(cmd, args)
		if err != nil {
			log.Fatalf(err.Error())
		}
	},
}

func runStatusCmd(cmd *cobra.Command, args []string) error {
	addedInstances := make(map[string]struct{})

	for _, instanceName := range args {
		if _, found := addedInstances[instanceName]; found {
			return fmt.Errorf("Duplicate instance name: %s", instanceName)
		}

		addedInstances[instanceName] = struct{}{}
		projectCtx.Instances = append(projectCtx.Instances, instanceName)
	}

	if err := running.SetLocalRunningPaths(&projectCtx); err != nil {
		return err
	}

	// fill context
	if err := project.FillCtx(&projectCtx); err != nil {
		return err
	}

	// start
	if err := running.Status(&projectCtx); err != nil {
		return err
	}

	return nil
}
