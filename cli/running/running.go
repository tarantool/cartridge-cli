package running

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/apex/log"
	"github.com/tarantool/cartridge-cli/cli/common"
	"github.com/tarantool/cartridge-cli/cli/context"
	"github.com/tarantool/cartridge-cli/cli/project"
)

const (
	rocksDir           = ".rocks"
	rocksDirMissedWarn = `Application dir does not contain ".rocks" directory. ` +
		`Make sure you ran "cartridge build" before running "cartridge start"`
)

func FillCtx(ctx *context.Ctx, args []string) error {
	var err error

	if err := project.SetLocalRunningPaths(ctx); err != nil {
		return err
	}

	if ctx.Running.AppDir == "" {
		ctx.Running.AppDir, err = os.Getwd()
		if err != nil {
			return fmt.Errorf("Failed to get current directory: %s", err)
		}
	}

	if ctx.Project.Name == "" {
		if ctx.Project.Name, err = project.DetectName(ctx.Running.AppDir); err != nil {
			return fmt.Errorf(
				"Failed to detect application name: %s. Please pass it explicitly via --name",
				err,
			)
		}
	}

	ctx.Project.StateboardName = project.GetStateboardName(ctx)

	if ctx.Running.StateboardOnly {
		ctx.Running.WithStateboard = true
	}

	if ctx.Running.Instances, err = common.GetInstancesFromArgs(args, ctx.Project.Name); err != nil {
		return err
	}

	// In order not to start (stop, log, etc) the stateboard instance when user
	// start instances by name
	if len(args) > 0 && !common.StringSliceContains(args, "stateboard") && !ctx.Running.StateboardFlagIsSet {
		ctx.Running.WithStateboard = false
	}

	if len(ctx.Running.Instances) > 0 && ctx.Running.StateboardOnly {
		log.Warnf("Specified instances are ignored due to stateboard-only flag")
	}

	return nil
}

func Start(ctx *context.Ctx) error {
	var err error

	if err := common.CheckTarantoolBinaries(); err != nil {
		return fmt.Errorf("Tarantool is required to start the application")
	}

	if !ctx.Running.StateboardOnly && len(ctx.Running.Instances) == 0 {
		ctx.Running.Instances, err = CollectInstancesFromConf(ctx)
		if err != nil {
			return fmt.Errorf("Failed to get configured instances from conf: %s", err)
		}
	}

	processes, err := collectProcesses(ctx)
	if err != nil {
		return fmt.Errorf("Failed to collect instances processes: %s", err)
	}

	if len(*processes) == 0 {
		return fmt.Errorf("No instances to start")
	}

	if _, err := os.Stat(filepath.Join(ctx.Running.AppDir, rocksDir)); os.IsNotExist(err) {
		log.Warn(rocksDirMissedWarn)
	} else if err != nil {
		log.Warnf("Failed to check .rocks directory: %s", err)
	}

	if err := processes.Start(ctx.Running.Daemonize, ctx.Running.DisableLogPrefix, ctx.Running.StartTimeout); err != nil {
		return err
	}

	return nil
}

func Stop(ctx *context.Ctx) error {
	var err error

	if !ctx.Running.StateboardOnly && len(ctx.Running.Instances) == 0 {
		ctx.Running.Instances, err = CollectInstancesFromConf(ctx)
		if err != nil {
			return fmt.Errorf("Failed to get configured instances from conf: %s", err)
		}
	}

	processes, err := collectProcesses(ctx)
	if err != nil {
		return fmt.Errorf("Failed to collect instances processes: %s", err)
	}

	if len(*processes) == 0 {
		return fmt.Errorf("No instances specified")
	}

	if err := processes.Stop(ctx.Running.StopForced); err != nil {
		return err
	}

	return nil
}

func Status(ctx *context.Ctx) error {
	var err error

	if !ctx.Running.StateboardOnly && len(ctx.Running.Instances) == 0 {
		ctx.Running.Instances, err = CollectInstancesFromConf(ctx)
		if err != nil {
			return fmt.Errorf("Failed to get configured instances from conf: %s", err)
		}
	}

	processes, err := collectProcesses(ctx)
	if err != nil {
		return fmt.Errorf("Failed to collect instances processes: %s", err)
	}

	if len(*processes) == 0 {
		return fmt.Errorf("No instances specified")
	}

	if err := processes.Status(); err != nil {
		return err
	}

	return nil
}

func Log(ctx *context.Ctx) error {
	var err error

	if !ctx.Running.StateboardOnly && len(ctx.Running.Instances) == 0 {
		ctx.Running.Instances, err = CollectInstancesFromConf(ctx)
		if err != nil {
			return fmt.Errorf("Failed to get configured instances from conf: %s", err)
		}
	}

	processes, err := collectProcesses(ctx)
	if err != nil {
		return fmt.Errorf("Failed to collect instances processes: %s", err)
	}

	if len(*processes) == 0 {
		return fmt.Errorf("No instances specified")
	}

	if err := processes.Log(ctx.Running.LogFollow, ctx.Running.LogLines); err != nil {
		return err
	}

	return nil
}

func Clean(ctx *context.Ctx) error {
	var err error

	if !ctx.Running.StateboardOnly && len(ctx.Running.Instances) == 0 {
		ctx.Running.Instances, err = CollectInstancesFromConf(ctx)
		if err != nil {
			return fmt.Errorf("Failed to get configured instances from config: %s", err)
		}
	}

	processes, err := collectProcesses(ctx)
	if err != nil {
		return fmt.Errorf("Failed to collect instances processes: %s", err)
	}

	if len(*processes) == 0 {
		return fmt.Errorf("No instances specified")
	}

	if err := processes.Clean(); err != nil {
		return err
	}

	return nil
}
