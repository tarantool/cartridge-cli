package replicasets

import (
	"fmt"

	"github.com/apex/log"
	"github.com/tarantool/cartridge-cli/cli/common"
	"github.com/tarantool/cartridge-cli/cli/context"
	"github.com/tarantool/cartridge-cli/cli/project"
)

func ListVshardGroups(ctx *context.Ctx, args []string) error {
	conn, err := connectToSomeRunningInstance(ctx)
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
