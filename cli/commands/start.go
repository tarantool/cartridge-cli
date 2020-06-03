package commands

import (
	"fmt"
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/tarantool/cartridge-cli/cli/project"
	"github.com/tarantool/cartridge-cli/cli/running"
)

func init() {
	rootCmd.AddCommand(startCmd)

	startCmd.Flags().StringVar(&projectCtx.Entrypoint, "script", "", scriptFlagDoc)
	startCmd.Flags().StringVar(&projectCtx.RunDir, "run-dir", "", runDirFlagDoc)
	startCmd.Flags().BoolVarP(&projectCtx.Daemonize, "daemonize", "d", false, daemonizeFlagDoc)
	startCmd.Flags().BoolVar(&projectCtx.Stateboard, "stateboard", false, stateboardFlagDoc)
}

var startCmd = &cobra.Command{
	Use:   "start [INSTANCE_NAME...] ",
	Short: "Start application instance(s)",
	Run: func(cmd *cobra.Command, args []string) {
		err := runStartCmd(cmd, args)
		if err != nil {
			log.Fatalf(err.Error())
		}
	},
}

func runStartCmd(cmd *cobra.Command, args []string) error {
	for _, arg := range cmd.Flags().Args() {
		projectCtx.Instances = append(projectCtx.Instances, arg)
	}

	curDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("Failed to get current directory: %s", err)
	}

	if projectCtx.RunDir == "" {
		projectCtx.RunDir = filepath.Join(curDir, "tmp")
	}

	if projectCtx.ConfDir == "" {
		projectCtx.ConfDir = filepath.Join(curDir, "instances.yml")
	}

	// fill context
	if err := project.FillCtx(&projectCtx); err != nil {
		return err
	}

	// start
	if err := running.Start(&projectCtx); err != nil {
		return err
	}

	return nil
}

const (
	scriptFlagDoc = `Application's entry point
Defaults to TARANTOOL_SCRIPT,
or ./init.lua when running from app's directory,
or :apps_path/:app_name/init.lua in multi-app env
`

	runDirFlagDoc = `Directory with pid and sock files
Defaults to TARANTOOL_RUN_DIR or /var/run/tarantool
`

	daemonizeFlagDoc = `Start in background
`

	stateboardFlagDoc = `Start application stateboard as well as instances
Defaults to TARANTOOL_STATEBOARD or false
Ignored if --stateboard-only is specified
`
)
