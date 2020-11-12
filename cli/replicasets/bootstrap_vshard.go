package replicasets

import (
	"fmt"
	"net"

	"github.com/apex/log"
	"github.com/tarantool/cartridge-cli/cli/context"
	"github.com/tarantool/cartridge-cli/cli/project"

	"github.com/tarantool/cartridge-cli/cli/common"
)

func BootstrapVshard(ctx *context.Ctx, args []string) error {
	instancesConf, err := getInstancesConf(ctx)
	if err != nil {
		return fmt.Errorf("Failed to get instances configuration: %s", err)
	}

	conn, err := getControlConn(instancesConf, ctx, nil)
	if err != nil {
		return fmt.Errorf("Failed to connect to Tarantool instance: %s", err)
	}

	if err := bootstrapVshard(conn); err != nil {
		return fmt.Errorf("failed to bootstrap vshard: %s", err)
	}

	log.Infof("Vshard is bootstrapped successfully")

	return nil
}

func bootstrapVshard(conn net.Conn) error {
	bootstrappedRaw, err := common.EvalTarantoolConn(conn, bootstrapVshardBody)
	if err != nil {
		return fmt.Errorf("Failed to bootstrap vshard: %s", err)
	}

	bootstrapped, ok := bootstrappedRaw.(bool)
	if !ok {
		return project.InternalError("Expected boolean as a result of vshard bootstrapping? got %#v", bootstrappedRaw)
	}

	if !bootstrapped {
		return fmt.Errorf("Vshard is already bootstrapped")
	}

	return nil
}

var (
	bootstrapVshardBody = `
local vshard_utils = require('cartridge.vshard-utils')
if not vshard_utils.can_bootstrap() then
	return false, nil
end

local cartridge = require('cartridge')
local bootstrapped, err = cartridge.admin_bootstrap_vshard()
return bootstrapped, err
`
)
