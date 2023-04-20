package replicasets

import (
	"fmt"
	"strings"

	"github.com/apex/log"
	"github.com/tarantool/cartridge-cli/cli/cluster"
	"github.com/tarantool/cartridge-cli/cli/common"
	"github.com/tarantool/cartridge-cli/cli/connector"
	"github.com/tarantool/cartridge-cli/cli/context"
	"github.com/tarantool/cartridge-cli/cli/project"
)

func Join(ctx *context.Ctx, args []string) error {
	var err error

	if ctx.Replicasets.ReplicasetName == "" {
		return fmt.Errorf("Please, specify replica set name via --replicaset flag")
	}

	if err := project.FillCtx(ctx); err != nil {
		return err
	}

	if ctx.Replicasets.JoinInstancesNames, err = common.GetInstancesFromArgs(args); err != nil {
		return err
	}

	if len(ctx.Replicasets.JoinInstancesNames) == 0 {
		return fmt.Errorf("Please, specify at least one instance name")
	}

	log.Infof(
		"Join instance(s) %s to replica set %s",
		strings.Join(ctx.Replicasets.JoinInstancesNames, ", "),
		ctx.Replicasets.ReplicasetName,
	)

	instancesConf, err := cluster.GetInstancesConf(ctx)
	if err != nil {
		return fmt.Errorf("Failed to get instances configuration: %s", err)
	}

	conn, err := connectToInstanceToJoin(instancesConf, ctx.Replicasets.JoinInstancesNames, ctx)
	if err != nil {
		return err
	}

	topologyReplicasets, err := getTopologyReplicasets(conn)
	if err != nil {
		return fmt.Errorf("Failed to get current topology replica sets: %s", err)
	}

	editReplicasetOpts, err := getJoinInstancesEditReplicasetsOpts(
		ctx.Replicasets.ReplicasetName, ctx.Replicasets.JoinInstancesNames,
		topologyReplicasets, instancesConf,
	)
	if err != nil {
		return common.ErrWrapCheckInstanceNameCommonMisprint(ctx.Replicasets.JoinInstancesNames, ctx.Project.Name,
			fmt.Errorf("Failed to get edit_topology options for joining instances: %s", err))
	}

	if _, err = editReplicaset(conn, editReplicasetOpts); err != nil {
		return fmt.Errorf("Failed to join instances: %s", err)
	}

	log.Infof(
		"Instance(s) %s have been successfully joined to replica set %s",
		strings.Join(ctx.Replicasets.JoinInstancesNames, ", "),
		ctx.Replicasets.ReplicasetName,
	)

	return nil
}

// connectToInstanceToJoin connects to some joined instance or first instance
// that should be joined.
// If we already have some replicasets, new instances should be joined via
// some joined instance socket, otherwise it would be two different clusters.
// If there is no joined instances, first instance should be joined via it's own socket.
func connectToInstanceToJoin(instancesConf *cluster.InstancesConf, joinInstancesNames []string, ctx *context.Ctx) (*connector.Conn, error) {
	// get some joined instance name
	instanceToJoinFromName, err := cluster.GetJoinedInstanceName(instancesConf, ctx)
	if err != nil {
		return nil, fmt.Errorf("Failed to find some instance joined to cluster")
	}

	// if there is no joined instances - use first specified instance
	if instanceToJoinFromName == "" {
		instanceToJoinFromName = ctx.Replicasets.JoinInstancesNames[0]
	}

	conn, err := cluster.ConnectToInstance(instanceToJoinFromName, ctx)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

func getJoinInstancesEditReplicasetsOpts(replicasetAlias string, joinInstancesNames []string,
	topologyReplicasets *TopologyReplicasets, instancesConf *cluster.InstancesConf) (*EditReplicasetOpts, error) {

	editReplicasetOpts := EditReplicasetOpts{}

	topologyReplicaset := topologyReplicasets.GetByAlias(replicasetAlias)
	if topologyReplicaset != nil {
		// replicaset with specified alias already exists
		// we have to specify it's UUID to join new instances to it
		editReplicasetOpts.ReplicasetUUID = topologyReplicaset.UUID
	} else {
		// replicaset with specified alias doesn't exist
		// we specify it's alias to create a new replicaset
		editReplicasetOpts.ReplicasetAlias = replicasetAlias
	}

	joinInstancesOpts, err := getJoinInstancesOpts(joinInstancesNames, instancesConf)
	if err != nil {
		return nil, fmt.Errorf("Failed to get join instances opts: %s", err)
	}
	editReplicasetOpts.JoinInstances = joinInstancesOpts

	return &editReplicasetOpts, nil
}

func getJoinInstancesOpts(instanceNames []string, instancesConf *cluster.InstancesConf) ([]JoinInstanceOpts, error) {
	joinInstancesOpts := make([]JoinInstanceOpts, len(instanceNames))
	for i, instanceName := range instanceNames {
		instanceConf, found := (*instancesConf)[instanceName]
		if !found {
			return nil, fmt.Errorf("Configuration for instance %s hasn't found in %s", instanceName, instancesFile)
		}

		joinInstancesOpts[i] = JoinInstanceOpts{
			URI: instanceConf.URI,
		}
	}

	return joinInstancesOpts, nil
}
