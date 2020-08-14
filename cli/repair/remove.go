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
			leaderIndex := common.StringsSliceElemIndex(replicasetConf.Leaders, instanceUUID)
			if leaderIndex != -1 {
				replicasetConf.Leaders = common.RemoveFromStringSlice(replicasetConf.Leaders, leaderIndex)
			}

			instanceIndex := common.StringsSliceElemIndex(replicasetConf.Instances, instanceUUID)
			if instanceIndex != -1 {
				replicasetConf.Instances = common.RemoveFromStringSlice(replicasetConf.Instances, instanceIndex)
			}

			if len(replicasetConf.Leaders) == 0 {
				if len(replicasetConf.Instances) == 0 {
					removeReplicasetFromRaw(topologyConf, replicasetUUID)
				} else {
					replicasetConf.Leaders = append(replicasetConf.Leaders, replicasetConf.Instances[0])
				}
			}

			if err := setReplicasetLeadersRaw(topologyConf, replicasetUUID, replicasetConf.Leaders); err != nil {
				return fmt.Errorf("Failed to set replicaset %s leaders: %s", replicasetUUID, err)
			}
		}
	}

	removeInstanceFromRaw(topologyConf, instanceUUID)

	return nil
}
