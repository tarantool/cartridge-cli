package replicasets

import (
	"fmt"

	"github.com/apex/log"
	"github.com/tarantool/cartridge-cli/cli/cluster"
	"github.com/tarantool/cartridge-cli/cli/common"
	"github.com/tarantool/cartridge-cli/cli/context"
	"github.com/tarantool/cartridge-cli/cli/project"
)

func SetFailoverPriority(ctx *context.Ctx, args []string) error {
	var err error

	if ctx.Replicasets.ReplicasetName == "" {
		return fmt.Errorf("Please, specify replica set name via --replicaset flag")
	}

	if err := project.FillCtx(ctx); err != nil {
		return err
	}

	if ctx.Replicasets.FailoverPriorityNames, err = common.GetInstancesFromArgs(args); err != nil {
		return err
	}

	if len(ctx.Replicasets.FailoverPriorityNames) == 0 {
		return fmt.Errorf("Please, specify at least one instance name")
	}

	conn, err := cluster.ConnectToSomeJoinedInstance(ctx)
	if err != nil {
		return err
	}

	topologyReplicaset, err := getTopologyReplicaset(conn, ctx.Replicasets.ReplicasetName)
	if err != nil {
		return err
	}

	editReplicasetOpts, err := getSetFailoverPriorityEditReplicasetOpts(
		ctx.Replicasets.FailoverPriorityNames, topologyReplicaset,
	)
	if err != nil {
		return common.ErrWrapCheckInstanceNameCommonMisprint(ctx.Replicasets.FailoverPriorityNames, ctx.Project.Name,
			fmt.Errorf("Failed to get edit_topology options for setting failover priority: %s", err))
	}

	newTopologyReplicaset, err := editReplicaset(conn, editReplicasetOpts)
	if err != nil {
		return fmt.Errorf("Failed to set failover priority: %s", err)
	}

	log.Infof("Replica set %s failover priority was set to:", ctx.Replicasets.ReplicasetName)
	for _, topologyInstance := range newTopologyReplicaset.Instances {
		log.Infof("  %s", topologyInstance.Alias)
	}

	return nil
}

func getSetFailoverPriorityEditReplicasetOpts(instanceNames []string, topologyReplicaset *TopologyReplicaset) (*EditReplicasetOpts, error) {
	editReplicasetOpts := EditReplicasetOpts{
		ReplicasetUUID: topologyReplicaset.UUID,
	}

	failoverPriorityUUIDs, err := getTopologyInstancesUUIDs(instanceNames, &topologyReplicaset.Instances)
	if err != nil {
		return nil, fmt.Errorf("Failed to get UUIDs in failover priority: %s", err)
	}

	editReplicasetOpts.FailoverPriorityUUIDs = failoverPriorityUUIDs

	return &editReplicasetOpts, nil
}

func getTopologyInstancesUUIDs(instanceNames []string, replicasetInstances *TopologyInstances) ([]string, error) {
	failoverPriorityUUIDs := make([]string, len(instanceNames))

	instanceUUIDsByAliases := make(map[string]string)
	for _, topologyInstance := range *replicasetInstances {
		instanceUUIDsByAliases[topologyInstance.Alias] = topologyInstance.UUID
	}

	for i, instanceName := range instanceNames {
		instanceUUID, found := instanceUUIDsByAliases[instanceName]
		if !found {
			return nil, fmt.Errorf("Instance %s not found in replica set", instanceName)
		}

		failoverPriorityUUIDs[i] = instanceUUID
	}

	return failoverPriorityUUIDs, nil
}
