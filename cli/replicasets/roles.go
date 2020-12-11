package replicasets

import (
	"fmt"
	"sort"
	"strings"

	"github.com/adam-hanna/arrayOperations"
	"github.com/apex/log"
	"github.com/tarantool/cartridge-cli/cli/common"
	"github.com/tarantool/cartridge-cli/cli/context"
	"github.com/tarantool/cartridge-cli/cli/project"
)

type GetNewRolesListFunc func([]string, []string) []string

type Role struct {
	Name         string
	Dependencies []string
}

type Roles []*Role

func (role *Role) String() string {
	if len(role.Dependencies) == 0 {
		return role.Name
	}

	return fmt.Sprintf(
		"%s (+ %s)",
		role.Name,
		strings.Join(role.Dependencies, ", "),
	)
}

func ListRoles(ctx *context.Ctx, args []string) error {
	if err := FillCtx(ctx); err != nil {
		return err
	}

	conn, err := connectToSomeRunningInstance(ctx)
	if err != nil {
		return fmt.Errorf("Failed to connect to Tarantool instance: %s", err)
	}

	knownRolesRaw, err := common.EvalTarantoolConn(conn, getKnownRolesBody)
	if err != nil {
		return fmt.Errorf("Failed to get known roles: %s", err)
	}

	knownRolesSliceRaw, err := common.ConvertToSlice(knownRolesRaw)
	if err != nil {
		return project.InternalError("Roles received in bad format: %#v", knownRolesRaw)
	}

	knownRoles := Roles{}

	for _, roleRaw := range knownRolesSliceRaw {
		role := Role{}

		roleMapRaw, err := common.ConvertToMapWithStringKeys(roleRaw)
		if err != nil {
			return project.InternalError("Roles received in bad format: %#v", knownRolesRaw)
		}

		if err := getStringValueFromMap(roleMapRaw, "name", &role.Name); err != nil {
			return fmt.Errorf("Role received in wrong format: %s", err)
		}
		if err := getStringSliceValueFromMap(roleMapRaw, "dependencies", &role.Dependencies); err != nil {
			return fmt.Errorf("Role received in wrong format: %s", err)
		}

		knownRoles = append(knownRoles, &role)
	}

	if len(knownRoles) == 0 {
		log.Infof("No roles available")
	} else {

		log.Infof("Available roles:")
		for _, role := range knownRoles {
			log.Infof("  %s", role.String())
		}
	}

	return nil
}

func AddRoles(ctx *context.Ctx, args []string) error {
	ctx.Replicasets.RolesList = args

	log.Infof(
		"Add role(s) %s to replica set %s",
		strings.Join(ctx.Replicasets.RolesList, ", "),
		ctx.Replicasets.ReplicasetName,
	)

	if err := updateRoles(ctx, addRolesToList, ctx.Replicasets.VshardGroup); err != nil {
		return fmt.Errorf("failed to add roles to replica set: %s", err)
	}

	return nil
}

func RemoveRoles(ctx *context.Ctx, args []string) error {
	ctx.Replicasets.RolesList = args

	log.Infof(
		"Remove role(s) %s from replica set %s",
		strings.Join(ctx.Replicasets.RolesList, ", "),
		ctx.Replicasets.ReplicasetName,
	)

	if err := updateRoles(ctx, removeRolesFromList, ""); err != nil {
		return fmt.Errorf("failed to add roles to replica set: %s", err)
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

	conn, err := connectToSomeJoinedInstance(ctx)
	if err != nil {
		return nil, err
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

func updateRoles(ctx *context.Ctx, getNewRolesListFunc GetNewRolesListFunc, vshardGroup string) error {
	if ctx.Replicasets.ReplicasetName == "" {
		return fmt.Errorf("Please, specify replica set name via --replicaset flag")
	}

	if err := FillCtx(ctx); err != nil {
		return err
	}

	conn, err := connectToSomeJoinedInstance(ctx)
	if err != nil {
		return err
	}

	topologyReplicaset, err := getTopologyReplicaset(conn, ctx.Replicasets.ReplicasetName)
	if err != nil {
		return err
	}

	editReplicasetOpts, err := getUpdateRolesEditReplicasetsOpts(getNewRolesListFunc, ctx.Replicasets.RolesList, vshardGroup, topologyReplicaset)
	if err != nil {
		return fmt.Errorf("Failed to get edit_topology options for roles updating: %s", err)
	}

	newTopologyReplicaset, err := editReplicaset(conn, editReplicasetOpts)
	if err != nil {
		return fmt.Errorf("Failed to update roles list: %s", err)
	}

	if len(newTopologyReplicaset.Roles) == 0 {
		log.Infof("Now replica set %s has no roles enabled", ctx.Replicasets.ReplicasetName)
	} else {
		log.Infof(
			"Replica set %s now has these roles enabled:",
			ctx.Replicasets.ReplicasetName,
		)

		for _, role := range newTopologyReplicaset.Roles {
			if newTopologyReplicaset.VshardGroup != nil && *newTopologyReplicaset.VshardGroup != "" {
				log.Infof("  %s (%s)", role, *newTopologyReplicaset.VshardGroup)
			} else {
				log.Infof("  %s", role)
			}
		}
	}

	return nil
}

func getUpdateRolesEditReplicasetsOpts(getNewRolesListFunc GetNewRolesListFunc,
	specifiedRoles []string, vshardGroup string, topologyReplicaset *TopologyReplicaset) (*EditReplicasetOpts, error) {

	newRolesList := getNewRolesListFunc(topologyReplicaset.Roles, specifiedRoles)
	sort.Strings(newRolesList)

	editReplicasetOpts := EditReplicasetOpts{
		ReplicasetUUID: topologyReplicaset.UUID,
		Roles:          newRolesList,
	}

	if vshardGroup != "" {
		editReplicasetOpts.VshardGroup = &vshardGroup
	}

	return &editReplicasetOpts, nil
}

var (
	getKnownRolesBody = `
local cartridge_roles = require('cartridge.roles')
local known_roles = cartridge_roles.get_known_roles()

local ret = {}
for _, role_name in ipairs(known_roles) do
	local role = {
		name = role_name,
		dependencies = cartridge_roles.get_role_dependencies(role_name),
	}

	table.insert(ret, role)
end

return ret
`
)
