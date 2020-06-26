package commands

import (
	"fmt"

	"github.com/apex/log"
	"github.com/spf13/cobra"

	"github.com/tarantool/cartridge-cli/cli/running"
)

func init() {
	rootCmd.AddCommand(startCmd)

	startCmd.Flags().StringVar(&projectCtx.Name, "name", "", nameFlagDoc)

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

Application name is described from rockspec in the current directory.

If INSTANCE_NAMEs aren't specified, then all instances described in
config file (see --cfg) are used.

Some flags default options can be override in ./.cartridge.yml config file.
`

	scriptFlagDoc = `Application's entry point
It should be a relative path to entrypoint in the
project directory, or an absolute path
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
