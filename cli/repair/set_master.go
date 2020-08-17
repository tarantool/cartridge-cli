package repair

import (
	"fmt"

	"github.com/tarantool/cartridge-cli/cli/common"
	"github.com/tarantool/cartridge-cli/cli/context"
)

func patchConfSetMaster(workDir string, ctx *context.Ctx) ([]string, error) {
	return patchConf(setMaster, workDir, ctx)
}

func setMaster(topologyConf *TopologyConfType, ctx *context.Ctx) error {
	instanceUUID := ctx.Repair.SetMasterInstanceUUID
	replicasetUUID := ctx.Repair.SetMasterReplicasetUUID

	instanceConf, ok := topologyConf.Instances[instanceUUID]
	if !ok {
		return fmt.Errorf("Instance %s isn't found in cluster", instanceUUID)
	}

	// check that specified instance isn't disabled or expelled
	// and belongs to specified replicaset
	if instanceConf.IsExpelled {
		return fmt.Errorf("Instance %s is expelled", instanceUUID)
	}

	if instanceConf.IsDisabled {
		return fmt.Errorf("Instance %s is disabled", instanceUUID)
	}

	if instanceConf.ReplicasetUUID != replicasetUUID {
		return fmt.Errorf("Instance %s doesn't belong to replicaset %s", instanceUUID, replicasetUUID)
	}

	replicasetConf, ok := topologyConf.Replicasets[replicasetUUID]
	if !ok {
		return fmt.Errorf("Replicaset %s isn't found in the cluster", replicasetUUID)
	}

	instanceIndex := common.StringsSliceElemIndex(replicasetConf.Leaders, instanceUUID)
	if instanceIndex != -1 {
		replicasetConf.SetLeaders(common.RemoveFromStringSlice(replicasetConf.Leaders, instanceIndex))
	}
	replicasetConf.SetLeaders(common.InsertInStringSlice(replicasetConf.Leaders, 0, instanceUUID))

	return nil
}
