package running

import (
	"fmt"

	"github.com/tarantool/cartridge-cli/cli/common"
	"github.com/tarantool/cartridge-cli/cli/project"
)

func Start(projectCtx *project.ProjectCtx) error {
	// XXX: TE --globall
	if err := common.CheckTarantoolBinaries(); err != nil {
		return fmt.Errorf("Tarantool is required to start the application")
	}

	processes := ProcessesSet{}

	if projectCtx.Stateboard {
		process := NewStateboardProcess(projectCtx)
		if err := processes.Add(process); err != nil {
			return fmt.Errorf("Failed to get process args for stateboard instance: %s", err)
		}
	}

	for _, instance := range projectCtx.Instances {
		process := NewInstanceProcess(projectCtx, instance)
		if err := processes.Add(process); err != nil {
			return fmt.Errorf("Failed to get process args for %s instance: %s", instance, err)
		}
	}

	if err := processes.Start(projectCtx.Daemonize); err != nil {
		return err
	}

	return nil
}
