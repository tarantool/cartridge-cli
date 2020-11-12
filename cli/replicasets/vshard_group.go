package replicasets

import (
	"fmt"

	"github.com/apex/log"
	"github.com/tarantool/cartridge-cli/cli/common"
	"github.com/tarantool/cartridge-cli/cli/context"
	"github.com/tarantool/cartridge-cli/cli/project"
)

func ListVshardGroups(ctx *context.Ctx, args []string) error {
	instancesConf, err := getInstancesConf(ctx)
	if err != nil {
		return fmt.Errorf("Failed to get instances configuration: %s", err)
	}

	conn, err := getControlConn(instancesConf, ctx, nil)
	if err != nil {
		return fmt.Errorf("Failed to connect to Tarantool instance: %s", err)
	}

	knownVshardGroupsRaw, err := common.EvalTarantoolConn(conn, getKnownVshardGroupsBody)
	if err != nil {
		return fmt.Errorf("Failed to get known vshard groups: %s", err)
	}

	knownVshardGroups, err := common.ConvertToStringsSlice(knownVshardGroupsRaw)
	if err != nil {
		return project.InternalError("Known vshard groups received in bad format: %#v", knownVshardGroupsRaw)
	}

	if len(knownVshardGroups) == 0 {
		log.Infof(
			"No vshard groups available. " +
				"It's possible that your application hasn't vshard-router role registered",
		)
	} else {
		log.Infof("Available vshard groups:")
		for _, vshardGroup := range knownVshardGroups {
			log.Infof("  %s", vshardGroup)
		}
	}

	return nil
}

func SetVshardGroup(ctx *context.Ctx, args []string) error {
	if ctx.Replicasets.ReplicasetName == "" {
		return fmt.Errorf("Please, specify replicaset name via --replicaset flag")
	}

	if err := FillCtx(ctx); err != nil {
		return err
	}

	if len(args) != 1 {
		return fmt.Errorf("Should be specified one argument - vshard group name")
	}

	vshardGroup := args[0]

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

	editReplicasetOpts, err := getSetVshardGroupEditReplicasetsOpts(vshardGroup, topologyReplicaset)
	if err != nil {
		return fmt.Errorf("Failed to get edit_topology options for setting vshard group: %s", err)
	}

	newTopologyReplicaset, err := editReplicaset(conn, editReplicasetOpts)
	if err != nil {
		return fmt.Errorf("Failed to update roles list: %s", err)
	}

	log.Infof("Replicaset %s vshard group is set to %s", ctx.Replicasets.ReplicasetName, newTopologyReplicaset.VshardGroup)

	return nil
}

func getSetVshardGroupEditReplicasetsOpts(vshardGroup string, topologyReplicaset *TopologyReplicaset) (*EditReplicasetOpts, error) {
	editReplicasetOpts := EditReplicasetOpts{
		ReplicasetUUID: topologyReplicaset.UUID,
		VshardGroup:    &vshardGroup,
	}

	return &editReplicasetOpts, nil
}

var (
	getKnownVshardGroupsBody = `
local vshard_utils = require('cartridge.vshard-utils')

local known_groups = vshard_utils.get_known_groups()

local known_groups_names = {}
for group_name in pairs(known_groups) do
	table.insert(known_groups_names, group_name)
end

return known_groups_names
`
)
