package replicasets

import (
	"fmt"

	"github.com/tarantool/cartridge-cli/cli/common"
	"github.com/tarantool/cartridge-cli/cli/context"
	"github.com/tarantool/cartridge-cli/cli/project"
)

func GetReplicasetRolesComp(ctx *context.Ctx) ([]string, error) {
	if ctx.Replicasets.ReplicasetName == "" {
		return nil, fmt.Errorf("Please, specify replicaset name via --replicaset flag")
	}

	if err := FillCtx(ctx); err != nil {
		return nil, err
	}

	instancesConf, err := getInstancesConf(ctx)
	if err != nil {
		return nil, fmt.Errorf("Failed to get instances configuration: %s", err)
	}

	conn, err := getControlConn(instancesConf, ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("Failed to connect to Tarantool instance: %s", err)
	}

	topologyReplicasets, err := getTopologyReplicasets(conn)
	if err != nil {
		return nil, fmt.Errorf("Failed to get current topology replicasets: %s", err)
	}

	topologyReplicaset := topologyReplicasets.GetByAlias(ctx.Replicasets.ReplicasetName)
	if topologyReplicaset == nil {
		return nil, fmt.Errorf("Replicaset %s isn't found in current topology", ctx.Replicasets.ReplicasetName)
	}

	return topologyReplicaset.Roles, nil
}

func GetReplicasetRolesToAddComp(ctx *context.Ctx) ([]string, error) {
	if err := FillCtx(ctx); err != nil {
		return nil, err
	}

	instancesConf, err := getInstancesConf(ctx)
	if err != nil {
		return nil, fmt.Errorf("Failed to get instances configuration: %s", err)
	}

	conn, err := getControlConn(instancesConf, ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("Failed to connect to Tarantool instance: %s", err)
	}

	// get all known roles
	knownRolesRaw, err := common.EvalTarantoolConn(conn, getKnownRolesBody)
	if err != nil {
		return nil, fmt.Errorf("Failed to get known roles: %s", err)
	}

	knownRoles, err := common.ConvertToStringsSlice(knownRolesRaw)
	if err != nil {
		return nil, project.InternalError("Roles received in bad format: %#v", knownRolesRaw)
	}

	// get replicaset roles
	if ctx.Replicasets.ReplicasetName == "" {
		return knownRoles, nil
	}

	topologyReplicasets, err := getTopologyReplicasets(conn)
	if err != nil {
		return knownRoles, nil
	}

	topologyReplicaset := topologyReplicasets.GetByAlias(ctx.Replicasets.ReplicasetName)
	if topologyReplicaset == nil {
		return knownRoles, nil
	}

	replicasetRoles := topologyReplicaset.Roles

	rolesToAdd := common.GetStringSlicesDifference(knownRoles, replicasetRoles)

	return rolesToAdd, nil
}
