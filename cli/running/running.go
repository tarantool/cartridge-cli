package running

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/tarantool/cartridge-cli/cli/common"
	"github.com/tarantool/cartridge-cli/cli/project"
)

func FillCtx(projectCtx *project.ProjectCtx) error {
	if err := checkInstancesUniqueness(projectCtx.Instances); err != nil {
		return err
	}

	if err := setLocalRunningPaths(projectCtx); err != nil {
		return err
	}

	if projectCtx.StateboardOnly {
		projectCtx.WithStateboard = true
	}

	if len(projectCtx.Instances) > 0 && projectCtx.StateboardOnly {
		log.Warnf("Specified instances are ignored due to stateboard-only flag")
	}

	return nil
}

func Start(projectCtx *project.ProjectCtx) error {
	var err error

	// XXX: TE --globall
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
