package replicasets

import (
	"fmt"
	"strings"

	"github.com/apex/log"
	"github.com/tarantool/cartridge-cli/cli/common"
	"github.com/tarantool/cartridge-cli/cli/context"
)

func SetFailoverPriority(ctx *context.Ctx, args []string) error {
	var err error

	if ctx.Replicasets.ReplicasetName == "" {
		return fmt.Errorf("Please, specify replicaset name via --replicaset flag")
	}

	if err := FillCtx(ctx); err != nil {
		return err
	}

	if ctx.Replicasets.FailoverPriorityNames, err = common.GetInstancesFromArgs(args, ctx); err != nil {
		return err
	}

	if len(ctx.Replicasets.FailoverPriorityNames) == 0 {
		return fmt.Errorf("Please, specify at least one instance name")
	}

	log.Infof(
		"Set %s failover priority to %s",
		ctx.Replicasets.ReplicasetName,
		strings.Join(ctx.Replicasets.FailoverPriorityNames, ", "),
	)

	instancesConf, err := getInstancesConf(ctx)
	if err != nil {
		return fmt.Errorf("Failed to get instances configuration: %s", err)
	}

	conn, err := getControlConn(instancesConf, ctx, nil)
	if err != nil {
		return fmt.Errorf("Failed to connect to Tarantool instance: %s", err)
	}

	topologyReplicasets, err := getTopologyReplicasets(conn)
	if err != nil {
		return fmt.Errorf("Failed to get current topology replicasets: %s", err)
	}

	topologyReplicaset := topologyReplicasets.GetByAlias(ctx.Replicasets.ReplicasetName)
	if topologyReplicaset == nil {
		return fmt.Errorf("Replicaset %s isn't found in current topology", ctx.Replicasets.ReplicasetName)
	}

	editReplicasetOpts, err := getSetFailoverPriorityEditReplicasetsOpts(
		topologyReplicaset, ctx.Replicasets.FailoverPriorityNames,
	)
	if err != nil {
		return fmt.Errorf("Failed to get edit_topology options for setting failover priority: %s", err)
	}

	newTopologyReplicaset, err := editReplicaset(conn, editReplicasetOpts)
	if err != nil {
		return fmt.Errorf("Failed to set failover priority: %s", err)
	}

	log.Infof("%s failover priority is:", ctx.Replicasets.ReplicasetName)
	for _, topologyInstance := range newTopologyReplicaset.Instances {
		log.Infof("  %s", topologyInstance.Alias)
	}

	return nil
}

func getSetFailoverPriorityEditReplicasetsOpts(topologyReplicaset *TopologyReplicaset, instanceNames []string) (*EditReplicasetOpts, error) {
	editReplicasetOpts := EditReplicasetOpts{
		ReplicasetUUID: topologyReplicaset.UUID,
	}

	failoverPriorityUUIDs, err := getTopologyInstancesUUIDs(&topologyReplicaset.Instances, instanceNames)
	if err != nil {
		return nil, fmt.Errorf("Failed to get UUIDs in failover priority: %s", err)
	}

	editReplicasetOpts.FailoverPriorityUUIDs = failoverPriorityUUIDs

	return &editReplicasetOpts, nil
}

func getTopologyInstancesUUIDs(replicasetInstances *TopologyInstances, instanceNames []string) ([]string, error) {
	failoverPriorityUUIDs := make([]string, len(instanceNames))

	instanceUUIDsByAliases := make(map[string]string)
	for _, topologyInstance := range *replicasetInstances {
		instanceUUIDsByAliases[topologyInstance.Alias] = topologyInstance.UUID
	}

	for i, instanceName := range instanceNames {
		instanceUUID, found := instanceUUIDsByAliases[instanceName]
		if !found {
			return nil, fmt.Errorf("Instance %s not found in replicaset", instanceName)
		}

		failoverPriorityUUIDs[i] = instanceUUID
	}

	return failoverPriorityUUIDs, nil
}
