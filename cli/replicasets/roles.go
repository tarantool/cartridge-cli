package replicasets

import (
	"fmt"
	"sort"
	"strings"

	"github.com/tarantool/cartridge-cli/cli/connector"
	"github.com/vmihailenco/msgpack/v5"

	"github.com/adam-hanna/arrayOperations"
	"github.com/apex/log"
	"github.com/tarantool/cartridge-cli/cli/common"
	"github.com/tarantool/cartridge-cli/cli/context"
)

type GetNewRolesListFunc func([]string, []string) []string

type Role struct {
	Name         string
	Dependencies []string
}

func (role *Role) DecodeMsgpack(d *msgpack.Decoder) error {
	return common.DecodeMsgpackStruct(d, role)
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

	req := connector.EvalReq(getKnownRolesBody).SetReadTimeout(SimpleOperationTimeout)
	var knownRoles []Role
	if err := conn.ExecTyped(req, &knownRoles); err != nil {
		return fmt.Errorf("Failed to get known roles: %s", err)
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
