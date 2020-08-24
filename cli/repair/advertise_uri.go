package repair

import (
	"fmt"

	"github.com/tarantool/cartridge-cli/cli/common"
	"github.com/tarantool/cartridge-cli/cli/context"
)

func patchConfAdvertiseURI(workDir string, ctx *context.Ctx) ([]common.ResultMessage, error) {
	return patchConf(patchInstanceURI, workDir, ctx)
}

func patchInstanceURI(topologyConf *TopologyConfType, ctx *context.Ctx) error {
	instanceUUID := ctx.Repair.SetURIInstanceUUID

	instanceConf, ok := topologyConf.Instances[instanceUUID]
	if !ok {
		return fmt.Errorf("Instance %s isn't found in cluster", instanceUUID)
	}

	if instanceConf.IsExpelled {
		return fmt.Errorf("Instance %s is expelled", instanceUUID)
	}

	if err := topologyConf.SetInstanceURI(instanceUUID, ctx.Repair.NewURI); err != nil {
		return fmt.Errorf("Failed to change instance advertise URI: %s", err)
	}

	return nil
}
