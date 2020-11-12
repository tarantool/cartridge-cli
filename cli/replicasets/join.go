package replicasets

import (
	"fmt"
	"strings"

	"github.com/apex/log"
	"github.com/tarantool/cartridge-cli/cli/common"
	"github.com/tarantool/cartridge-cli/cli/context"
)

func Join(ctx *context.Ctx, args []string) error {
	var err error

	if ctx.Replicasets.ReplicasetName == "" {
		return fmt.Errorf("Please, specify replicaset name via --replicaset flag")
	}

	if err := FillCtx(ctx); err != nil {
		return err
	}

	if ctx.Replicasets.JoinInstancesNames, err = common.GetInstancesFromArgs(args, ctx); err != nil {
		return err
	}

	if len(ctx.Replicasets.JoinInstancesNames) == 0 {
		return fmt.Errorf("Please, specify at least one instance name")
	}

	log.Infof(
		"Join instance(s) %s to replicaset %s",
		strings.Join(ctx.Replicasets.JoinInstancesNames, ", "),
		ctx.Replicasets.ReplicasetName,
	)

	instancesConf, err := getInstancesConf(ctx)
	if err != nil {
		return fmt.Errorf("Failed to get instances configuration: %s", err)
	}

	conn, err := getControlConn(instancesConf, ctx, ctx.Replicasets.JoinInstancesNames)
	if err != nil {
		return fmt.Errorf("Failed to connect to Tarantool instance: %s", err)
	}

	topologyReplicasets, err := getTopologyReplicasets(conn)
	if err != nil {
		return fmt.Errorf("Failed to get current topology replicasets: %s", err)
	}

	editReplicasetOpts, err := getJoinInstancesEditReplicasetsOpts(
		ctx.Replicasets.ReplicasetName, ctx.Replicasets.JoinInstancesNames,
		topologyReplicasets, instancesConf,
	)
	if err != nil {
		return fmt.Errorf("Failed to get edit_topology options for joining instances: %s", err)
	}

	if _, err = editReplicaset(conn, editReplicasetOpts); err != nil {
		return fmt.Errorf("Failed to join instances: %s", err)
	}

	log.Infof(
		"Instance(s) %s was successfully joined to replicaset %s",
		strings.Join(ctx.Replicasets.JoinInstancesNames, ", "),
		ctx.Replicasets.ReplicasetName,
	)

	return nil
}

func getJoinInstancesEditReplicasetsOpts(replicasetAlias string, joinInstancesNames []string,
	topologyReplicasets *TopologyReplicasets, instancesConf *InstancesConf) (*EditReplicasetOpts, error) {

	editReplicasetOpts := EditReplicasetOpts{}

	topologyReplicaset := topologyReplicasets.GetByAlias(replicasetAlias)
	if topologyReplicaset != nil {
		editReplicasetOpts.ReplicasetUUID = topologyReplicaset.UUID
	} else {
		editReplicasetOpts.ReplicasetAlias = replicasetAlias
	}

	joinInstancesURIs, err := getInstancesURIs(joinInstancesNames, instancesConf)
	if err != nil {
		return nil, fmt.Errorf("Failed to get URIs of a new instances: %s", err)
	}
	editReplicasetOpts.JoinInstancesURIs = *joinInstancesURIs

	return &editReplicasetOpts, nil
}

func getInstancesURIs(instanceNames []string, instancesConf *InstancesConf) (*[]string, error) {
	instancesURIs := make([]string, len(instanceNames))
	for i, instanceName := range instanceNames {
		instanceConf, found := (*instancesConf)[instanceName]
		if !found {
			return nil, fmt.Errorf("Configuration for instance %s isn't found", instanceName)
		}

		instancesURIs[i] = instanceConf.URI
	}

	return &instancesURIs, nil
}
