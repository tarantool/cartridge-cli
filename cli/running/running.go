package running

import (
	"fmt"

	"github.com/tarantool/cartridge-cli/cli/common"
	"github.com/tarantool/cartridge-cli/cli/project"
)

func Start(projectCtx *project.ProjectCtx) error {
	var err error

	// XXX: TE --globall
	if err := common.CheckTarantoolBinaries(); err != nil {
		return fmt.Errorf("Tarantool is required to start the application")
	}

	if len(projectCtx.Instances) == 0 { // XXX: && !projectCtx.StateboardOnly
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
