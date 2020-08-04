package repair

import (
	"fmt"

	"github.com/tarantool/cartridge-cli/cli/common"
	"github.com/tarantool/cartridge-cli/cli/context"
)

func patchConfRemoveInstance(workDir string, ctx *context.Ctx) ([]string, error) {
	return patchConf(removeInstance, workDir, ctx)
}

func removeInstance(topologyConf *TopologyConfType, ctx *context.Ctx) error {
	instanceUUID := ctx.Repair.RemoveInstanceUUID

	instanceConf, ok := topologyConf.Instances[instanceUUID]
	if !ok {
		return fmt.Errorf("Instance %s isn't found in cluster", instanceUUID)
	}

	if !instanceConf.IsExpelled {
		replicasetUUID := instanceConf.ReplicasetUUID

		replicasetConf, ok := topologyConf.Replicasets[replicasetUUID]

		if ok {
			instanceIndex := common.StringsSliceElemIndex(replicasetConf.Leaders, instanceUUID)
			if instanceIndex != -1 {
				replicasetConf.Leaders = common.RemoveFromStringSlice(replicasetConf.Leaders, instanceIndex)
			}

			if err := setReplicasetLeadersRaw(topologyConf, replicasetUUID, replicasetConf.Leaders); err != nil {
				return fmt.Errorf("Failed to set replicaset %s leaders: %s", replicasetUUID, err)
			}
		}
	}

	removeInstanceFromRaw(topologyConf, instanceUUID)

	return nil
}
