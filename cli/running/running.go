package running

import (
	"fmt"
	"os"
	"os/exec"
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

	tarantoolExecName = "tarantool"
)

func FillCtx(ctx *context.Ctx, args []string) error {
	var err error

	if err := project.SetRunningPaths(ctx, true); err != nil {
		return err
	}

	if ctx.Project.Name == "" {
		if ctx.Project.Name, err = project.DetectName(ctx.Running.AppDir); err != nil {
			return fmt.Errorf(
				"Failed to detect application name: %s. Please pass it explicitly via --name ",
				err,
			)
		}
	}

	ctx.Project.StateboardName = project.GetStateboardName(ctx)

	if ctx.Running.StateboardOnly {
		ctx.Running.WithStateboard = true
	}

	if ctx.Running.Instances, err = getInstancesFromArgs(args, ctx); err != nil {
		return err
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
		ctx.Running.Instances, err = collectInstancesFromConf(ctx)
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

	var tarantoolExec string
	if _, err := os.Stat(filepath.Join(ctx.Running.AppDir, tarantoolExecName)); err == nil {
		tarantoolExec = filepath.Join(ctx.Running.AppDir, tarantoolExecName)
	} else if os.IsNotExist(err) {
		if tarantoolExec, err = exec.LookPath(tarantoolExecName); err != nil {
			return fmt.Errorf("Failed to find Tarantool executable: %s", err)
		}
	} else {
		return fmt.Errorf("Failed to check tarantool executable: %s", err)
	}
	log.Debugf("Tarantool executable %s is used", tarantoolExec)

	err = processes.Start(StartOpts{
		Daemonize:     ctx.Running.Daemonize,
		Timeout:       ctx.Running.StartTimeout,
		TarantoolExec: tarantoolExec,
	})
	if err != nil {
		return err
	}

	return nil
}

func Stop(ctx *context.Ctx) error {
	var err error

	if !ctx.Running.StateboardOnly && len(ctx.Running.Instances) == 0 {
		ctx.Running.Instances, err = collectInstancesFromConf(ctx)
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

	if err := processes.Stop(); err != nil {
		return err
	}

	return nil
}

func Status(ctx *context.Ctx) error {
	var err error

	if !ctx.Running.StateboardOnly && len(ctx.Running.Instances) == 0 {
		ctx.Running.Instances, err = collectInstancesFromConf(ctx)
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
		ctx.Running.Instances, err = collectInstancesFromConf(ctx)
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
