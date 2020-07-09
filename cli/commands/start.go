package commands

import (
	"fmt"

	"github.com/apex/log"
	"github.com/spf13/cobra"

	"github.com/tarantool/cartridge-cli/cli/running"
)

func init() {
	rootCmd.AddCommand(startCmd)

	startCmd.Flags().StringVar(&ctx.Project.Name, "name", "", nameFlagDoc)

	startCmd.Flags().StringVar(&ctx.Running.Entrypoint, "script", "", scriptFlagDoc)
	startCmd.Flags().StringVar(&ctx.Running.RunDir, "run-dir", "", runDirFlagDoc)
	startCmd.Flags().StringVar(&ctx.Running.DataDir, "data-dir", "", dataDirFlagDoc)
	startCmd.Flags().StringVar(&ctx.Running.LogDir, "log-dir", "", logDirFlagDoc)
	startCmd.Flags().StringVar(&ctx.Running.ConfPath, "cfg", "", cfgFlagDoc)

	startCmd.Flags().BoolVarP(&ctx.Running.Daemonize, "daemonize", "d", false, daemonizeFlagDoc)
	startCmd.Flags().BoolVar(&ctx.Running.WithStateboard, "stateboard", false, stateboardFlagDoc)
	startCmd.Flags().BoolVar(&ctx.Running.StateboardOnly, "stateboard-only", false, stateboardOnlyFlagDoc)
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
	if err := running.FillCtx(&ctx, args); err != nil {
		return err
	}

	if err := running.Start(&ctx); err != nil {
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
It should be a relative path to the entry point
in the project directory or an absolute path.
Defaults to "init.lua" (or "script" in .cartridge.yml)
`

	runDirFlagDoc = `Directory where PID and socket files are stored
Defaults to ./tmp/run (or "run-dir" in .cartridge.yml)
`

	dataDirFlagDoc = `Directory where instances' data is stored
Each instance's working directory is
"<data-dir>/<app-name>.<instance-name>".
Defaults to ./tmp/data (or "data-dir" in .cartridge.yml)
`

	logDirFlagDoc = `Directory to store instances logs
when running in background
Defaults to ./tmp/log (or "log-dir" in .cartridge.yml)
`

	cfgFlagDoc = `Configuration file for Cartridge instances
Defaults to ./instances.yml (or "cfg" in .cartridge.yml)
`

	daemonizeFlagDoc = `Start instance(s) in background
`

	stateboardFlagDoc = `Manage application stateboard as well as instances
Ignored if "--stateboard-only" is specified
`

	stateboardOnlyFlagDoc = `Manage only application stateboard
If specified, "INSTANCE_NAME..." are ignored.
`
)
