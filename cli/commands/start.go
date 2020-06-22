package commands

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

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
	Use:   "start [INSTANCE_ID...]",
	Short: "Start application instance(s)",
	Long:  fmt.Sprintf("Start application instance(s)\n\n%s", runningCommonDoc),
	Run: func(cmd *cobra.Command, args []string) {
		err := runStartCmd(cmd, args)
		if err != nil {
			log.Fatalf(err.Error())
		}
	},
}

func runStartCmd(cmd *cobra.Command, args []string) error {
	if err := running.FillCtx(&projectCtx, args); err != nil {
		return err
	}

	if err := running.Start(&projectCtx); err != nil {
		return err
	}

	return nil
}

const (
	runningCommonDoc = `Starts instance(s) of current application

INSTANCE_ID is [APP_NAME].[INSTANCE_NAME]

If APP_NAME name isn't specified, it's described from rockspec
in the current directory

If INSTANCE_NAME isn't specified, then all instances described in
config file (see --cfg) are used

All flags default options can be override in ./.cartridge.yml config file
`

	scriptFlagDoc = `Application's entry point
Defaults to init.lua (or "script" in config)
`

	runDirFlagDoc = `Directory where pid and socket files are stored
Defaults to ./tmp/run (or "run-dir" in config)
`

	dataDirFlagDoc = `Directory to store instances data
Each instance workdir is
<data-dir>/<app-name>.<instance-name>
Defaults to ./tmp/data (or "data-dir" in config)
`

	logDirFlagDoc = `Directory to store instances logs
when running in background locally
Defaults to ./tmp/log (or "log-dir" in config)
`

	daemonizeFlagDoc = `Start in background
`

	stateboardFlagDoc = `Manage application stateboard as well as instances
Ignored if --stateboard-only is specified
`

	stateboardOnlyFlagDoc = `Manage only application stateboard
`

	cfgFlagDoc = `Cartridge instances config file
Defaults to ./instances.yml (or "cfg" in config)
`
)
