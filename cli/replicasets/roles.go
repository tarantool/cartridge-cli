package replicasets

import (
	"fmt"
	"strings"

	"github.com/adam-hanna/arrayOperations"
	"github.com/apex/log"
	"github.com/tarantool/cartridge-cli/cli/common"
	"github.com/tarantool/cartridge-cli/cli/context"
	"github.com/tarantool/cartridge-cli/cli/project"
)

func ListRoles(ctx *context.Ctx, args []string) error {
	log.Infof("Get list of roles known by cluster")

	if err := FillCtx(ctx); err != nil {
		return err
	}

	instancesConf, err := getInstancesConf(ctx)
	if err != nil {
		return fmt.Errorf("Failed to get instances configuration: %s", err)
	}

	conn, err := getControlConn(instancesConf, ctx, nil)
	if err != nil {
		return fmt.Errorf("Failed to connect to Tarantool instance: %s", err)
	}

	knownRolesRaw, err := common.EvalTarantoolConn(conn, getKnownRolesBody)
	if err != nil {
		return fmt.Errorf("Failed to get known roles: %s", err)
	}

	knownRoles, err := common.ConvertToStringsSlice(knownRolesRaw)
	if err != nil {
		return project.InternalError("Roles received in bad format: %#v", knownRolesRaw)
	}

	if len(knownRoles) == 0 {
		log.Infof("No roles available")
	} else {

		log.Infof("Available roles:")
		for _, role := range knownRoles {
			log.Infof("  %s", role)
		}
	}

	return nil
}

func AddRoles(ctx *context.Ctx, args []string) error {
	ctx.Replicasets.RolesList = args

	log.Infof(
		"Add role(s) %s to replicaset %s",
		strings.Join(ctx.Replicasets.RolesList, ", "),
		ctx.Replicasets.ReplicasetName,
	)

	if err := updateRoles(ctx, addRolesToList); err != nil {
		return fmt.Errorf("failed to add roles to replicaset: %s", err)
	}

	return nil
}

func RemoveRoles(ctx *context.Ctx, args []string) error {
	ctx.Replicasets.RolesList = args

	log.Infof(
		"Remove role(s) %s from replicaset %s",
		strings.Join(ctx.Replicasets.RolesList, ", "),
		ctx.Replicasets.ReplicasetName,
	)

	if err := updateRoles(ctx, removeRolesFromList); err != nil {
		return fmt.Errorf("failed to add roles to replicaset: %s", err)
	}

	return nil
}

func addRolesToList(currentRoles, rolesToAdd []string) []string {
	return arrayOperations.UnionString(currentRoles, rolesToAdd)
}

func removeRolesFromList(currentRoles, rolesToRemove []string) []string {
	return common.GetStringSlicesDifference(currentRoles, rolesToRemove)
}

func getKnownRoles(ctx *context.Ctx) ([]string, error) {
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

	knownRolesRaw, err := common.EvalTarantoolConn(conn, getKnownRolesBody)
	if err != nil {
		return nil, fmt.Errorf("Failed to get known roles: %s", err)
	}

	knownRoles, err := common.ConvertToStringsSlice(knownRolesRaw)
	if err != nil {
		return nil, project.InternalError("Roles received in bad format: %#v", knownRolesRaw)
	}

	return knownRoles, nil
}

func updateRoles(ctx *context.Ctx, getNewRolesListFunc func([]string, []string) []string) error {
	if ctx.Replicasets.ReplicasetName == "" {
		return fmt.Errorf("Please, specify replicaset name via --replicaset flag")
	}

	if err := FillCtx(ctx); err != nil {
		return err
	}

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

	newRolesList := getNewRolesListFunc(topologyReplicaset.Roles, ctx.Replicasets.RolesList)
	editReplicasetOpts, err := getUpdateRolesEditReplicasetsOpts(newRolesList, topologyReplicaset)
	if err != nil {
		return fmt.Errorf("Failed to get edit_topology options for roles updating: %s", err)
	}

	newTopologyReplicaset, err := editReplicaset(conn, editReplicasetOpts)
	if err != nil {
		return fmt.Errorf("Failed to update roles list: %s", err)
	}

	if len(newTopologyReplicaset.Roles) == 0 {
		log.Infof("Now replicaset %s has no roles enabled", ctx.Replicasets.ReplicasetName)
	} else {
		log.Infof(
			"Replicaset %s now has these roles enabled: ",
			ctx.Replicasets.ReplicasetName,
		)

		for _, role := range newTopologyReplicaset.Roles {
			log.Infof("  %s", role)
		}
	}

	return nil
}

func getUpdateRolesEditReplicasetsOpts(replicasetRoles []string, topologyReplicaset *TopologyReplicaset) (*EditReplicasetOpts, error) {
	editReplicasetOpts := EditReplicasetOpts{
		ReplicasetUUID: topologyReplicaset.UUID,
		Roles:          replicasetRoles,
	}

	return &editReplicasetOpts, nil
}

var (
	getKnownRolesBody = `
local cartridge_roles = require('cartridge.roles')
local known_roles = cartridge_roles.get_known_roles()
return known_roles
`
)
