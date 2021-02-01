package replicasets

import (
	"fmt"

	"github.com/tarantool/cartridge-cli/cli/codegen/static"
	"github.com/tarantool/cartridge-cli/cli/connector"

	"github.com/apex/log"
	"github.com/tarantool/cartridge-cli/cli/context"
)

func ListVshardGroups(ctx *context.Ctx, args []string) error {
	conn, err := connectToSomeRunningInstance(ctx)
	if err != nil {
		return fmt.Errorf("Failed to connect to Tarantool instance: %s", err)
	}

	getKnownVshardGroupsBody, err := static.GetStaticFileContent(ReplicasetsLuaTemplateFS,
		"known_vshard_groups_body.lua")
	if err != nil {
		return fmt.Errorf("Failed to get static file content: %s", err)
	}

	req := connector.EvalReq(getKnownVshardGroupsBody)
	var knownVshardGroups []string

	if err := conn.ExecTyped(req, &knownVshardGroups); err != nil {
		return fmt.Errorf("Failed to get known vshard groups: %s", err)
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
