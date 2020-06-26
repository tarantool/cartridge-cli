package running

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/apex/log"
	"github.com/tarantool/cartridge-cli/cli/common"
	"github.com/tarantool/cartridge-cli/cli/project"
)

const (
	rocksDir           = ".rocks"
	rocksDirMissedWarn = `Application dir does not contain ".rocks" directory. ` +
		`Make sure you ran "cartridge build" before running "cartridge start"`
)

func FillCtx(projectCtx *project.ProjectCtx, args []string) error {
	var err error

	if err := project.SetLocalRunningPaths(projectCtx); err != nil {
		return err
	}

	if projectCtx.AppDir == "" {
		projectCtx.AppDir, err = os.Getwd()
		if err != nil {
			return fmt.Errorf("Failed to get current directory: %s", err)
		}
	}

	if projectCtx.Name == "" {
		if projectCtx.Name, err = project.DetectName(projectCtx.AppDir); err != nil {
			return fmt.Errorf(
				"Failed to detect application name: %s. Please pass it explicitly via --name ",
				err,
			)
		}
	}

	projectCtx.StateboardName = project.GetStateboardName(projectCtx)

	if projectCtx.StateboardOnly {
		projectCtx.WithStateboard = true
	}

	if projectCtx.Instances, err = getInstancesFromArgs(args, projectCtx); err != nil {
		return err
	}

	if len(projectCtx.Instances) > 0 && projectCtx.StateboardOnly {
		log.Warnf("Specified instances are ignored due to stateboard-only flag")
	}

	return nil
}

func Start(projectCtx *project.ProjectCtx) error {
	var err error

	if err := common.CheckTarantoolBinaries(); err != nil {
		return fmt.Errorf("Tarantool is required to start the application")
	}

	if !projectCtx.StateboardOnly && len(projectCtx.Instances) == 0 {
		projectCtx.Instances, err = collectInstancesFromConf(projectCtx)
		if err != nil {
			return fmt.Errorf("Failed to get configured instances from conf: %s", err)
		}
	}

	processes, err := collectProcesses(projectCtx)
	if err != nil {
		return fmt.Errorf("Failed to collect instances processes: %s", err)
	}

	if len(*processes) == 0 {
		return fmt.Errorf("No instances to start")
	}

	if _, err := os.Stat(filepath.Join(projectCtx.AppDir, rocksDir)); os.IsNotExist(err) {
		log.Warn(rocksDirMissedWarn)
	} else if err != nil {
		log.Warnf("Failed to check .rocks directory: %s", err)
	}

	if err := processes.Start(projectCtx.Daemonize); err != nil {
		return err
	}

	return nil
}

func Stop(projectCtx *project.ProjectCtx) error {
	var err error

	if !projectCtx.StateboardOnly && len(projectCtx.Instances) == 0 {
		projectCtx.Instances, err = collectInstancesFromConf(projectCtx)
		if err != nil {
			return fmt.Errorf("Failed to get configured instances from conf: %s", err)
		}
	}

	processes, err := collectProcesses(projectCtx)
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

func Status(projectCtx *project.ProjectCtx) error {
	var err error

	if !projectCtx.StateboardOnly && len(projectCtx.Instances) == 0 {
		projectCtx.Instances, err = collectInstancesFromConf(projectCtx)
		if err != nil {
			return fmt.Errorf("Failed to get configured instances from conf: %s", err)
		}
	}

	processes, err := collectProcesses(projectCtx)
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
