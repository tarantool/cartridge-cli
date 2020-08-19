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
	for instanceUUID, instanceConf := range topologyConf.Instances {
		if instanceConf.AdvertiseURI == ctx.Repair.OldURI {
			if err := topologyConf.SetInstanceURI(instanceUUID, ctx.Repair.NewURI); err != nil {
				return fmt.Errorf("Failed to change instance advertise URI: %s", err)
			}

			return nil
		}
	}

	return fmt.Errorf("Instance with URI %s isn't found in the cluster", ctx.Repair.OldURI)
}
