package commands

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/tarantool/cartridge-cli/cli/project"
	"github.com/tarantool/cartridge-cli/cli/running"
)

func init() {
	rootCmd.AddCommand(startCmd)

	startCmd.Flags().StringVar(&projectCtx.Entrypoint, "script", "", scriptFlagDoc)
	startCmd.Flags().StringVar(&projectCtx.RunDir, "run-dir", "", runDirFlagDoc)
	startCmd.Flags().StringVar(&projectCtx.DataDir, "data-dir", "", dataDirFlagDoc)
	startCmd.Flags().StringVar(&projectCtx.LogDir, "log-dir", "", logDirFlagDoc)
	startCmd.Flags().StringVar(&projectCtx.ConfPath, "cfg", "", cfgFlagDoc)

	startCmd.Flags().BoolVarP(&projectCtx.Daemonize, "daemonize", "d", false, daemonizeFlagDoc)
	startCmd.Flags().BoolVar(&projectCtx.WithStateboard, "stateboard", false, stateboardFlagDoc)
	startCmd.Flags().BoolVar(&projectCtx.StateboardOnly, "stateboard-only", false, stateboardOnlyFlagDoc)
}

var startCmd = &cobra.Command{
	Use:   "start [INSTANCE_NAME...]",
	Short: "Start application instance(s)",
	Run: func(cmd *cobra.Command, args []string) {
		err := runStartCmd(cmd, args)
		if err != nil {
			log.Fatalf(err.Error())
		}
	},
}

func runStartCmd(cmd *cobra.Command, args []string) error {
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

	if err := running.Start(&projectCtx); err != nil {
		return err
	}

	return nil
}

const (
	scriptFlagDoc = `Application's entry point
Defaults to init.lua on local start
`

	runDirFlagDoc = `Directory where pid and socket files are stored
Defaults to ./tmp/run on local start
`

	dataDirFlagDoc = `Directory to store instances data
Each instance workdir is <data-dir>/<app-name>/<instance-name>
Defaults to ./tmp/data on local start
`

	logDirFlagDoc = `Directory to store instances logs
when running in background locally
Defaults to ./tmp/log
`

	daemonizeFlagDoc = `Start in background
`

	stateboardFlagDoc = `Manage application stateboard as well as instances
Ignored if --stateboard-only is specified
`

	stateboardOnlyFlagDoc = `Manage only application stateboard
`

	cfgFlagDoc = `Cartridge instances config file
Defaults to ./instances.yml on local start
`
)
