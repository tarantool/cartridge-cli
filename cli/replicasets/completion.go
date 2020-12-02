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

	conn, err := connectToSomeJoinedInstance(ctx)
	if err != nil {
		return nil, err
	}

	topologyReplicaset, err := getTopologyReplicaset(conn, ctx.Replicasets.ReplicasetName)
	if err != nil {
		return nil, err
	}

	return topologyReplicaset.Roles, nil
}

func GetReplicasetRolesToAddComp(ctx *context.Ctx) ([]string, error) {
	if err := FillCtx(ctx); err != nil {
		return nil, err
	}

	conn, err := connectToSomeJoinedInstance(ctx)
	if err != nil {
		return nil, err
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

	topologyReplicaset, err := getTopologyReplicaset(conn, ctx.Replicasets.ReplicasetName)
	if err != nil {
		return knownRoles, nil
	}

	replicasetRoles := topologyReplicaset.Roles

	rolesToAdd := common.GetStringSlicesDifference(knownRoles, replicasetRoles)

	return rolesToAdd, nil
}
