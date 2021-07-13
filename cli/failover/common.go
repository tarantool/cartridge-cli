package failover

import (
	"fmt"

	"github.com/apex/log"
	"github.com/fatih/structs"
	"github.com/tarantool/cartridge-cli/cli/connector"
	"github.com/tarantool/cartridge-cli/cli/context"
	"github.com/tarantool/cartridge-cli/cli/replicasets"
)

func setupFailover(ctx *context.Ctx, failoverOpts *FailoverOpts) error {
	conn, err := replicasets.ConnectToSomeRunningInstance(ctx)
	if err != nil {
		return fmt.Errorf("Failed to connect to some instance: %s", err)
	}

	if *failoverOpts.StateProvider == "stateboard" {
		*failoverOpts.StateProvider = "tarantool"
	}

	req := connector.EvalReq(setupFailoverBody, structs.Map(failoverOpts))
	ok, err := conn.Exec(req)
	if err != nil {
		return fmt.Errorf("Failed to configure failover: %s", err)
	}

	log.Warnf("%s", ok)

	return nil
}

func disableFailover(ctx *context.Ctx) error {
	conn, err := replicasets.ConnectToSomeRunningInstance(ctx)
	if err != nil {
		return fmt.Errorf("Failed to connect to some instance: %s", err)
	}

	req := connector.EvalReq(setupFailoverBody, map[string]string{"mode": "disabled"})
	ok, err := conn.Exec(req)
	if err != nil {
		return fmt.Errorf("Failed to disable failover: %s", err)
	}

	log.Warnf("%s", ok)
	return nil
}
