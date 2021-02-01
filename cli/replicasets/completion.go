package replicasets

import (
	"fmt"
	"time"

	"github.com/tarantool/cartridge-cli/cli/codegen/static"
	"github.com/tarantool/cartridge-cli/cli/common"
	"github.com/tarantool/cartridge-cli/cli/connector"
	"github.com/tarantool/cartridge-cli/cli/context"
)

const (
	completionEvalTimeout = 3 * time.Second
)

func GetReplicasetRolesComp(ctx *context.Ctx) ([]string, error) {
	if ctx.Replicasets.ReplicasetName == "" {
		return nil, fmt.Errorf("Please, specify replica set name via --replicaset flag")
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
	var knownRoles []Role
	getKnownRolesBody, err := static.GetStaticFileContent(ReplicasetsLuaTemplateFS, "known_roles_body.lua")
	if err != nil {
		return nil, fmt.Errorf("Failed to get static file content: %s", err)
	}

	req := connector.EvalReq(getKnownRolesBody).SetReadTimeout(SimpleOperationTimeout)
	if err := conn.ExecTyped(req, &knownRoles); err != nil {
		return nil, fmt.Errorf("Failed to get known roles: %s", err)
	}

	roleNames := make([]string, len(knownRoles))
	for i, role := range knownRoles {
		roleNames[i] = role.Name
	}

	// get replicaset roles
	if ctx.Replicasets.ReplicasetName == "" {
		return roleNames, nil
	}

	topologyReplicaset, err := getTopologyReplicaset(conn, ctx.Replicasets.ReplicasetName)
	if err != nil {
		return roleNames, nil
	}

	replicasetRoles := topologyReplicaset.Roles

	rolesToAdd := common.GetStringSlicesDifference(roleNames, replicasetRoles)

	return rolesToAdd, nil
}
