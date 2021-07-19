package failover

import (
	"github.com/tarantool/cartridge-cli/cli/context"
	"github.com/tarantool/cartridge-cli/cli/project"
)

func State(ctx *context.Ctx) error {
	//var err error

	if err := project.FillCtx(ctx); err != nil {
		return err
	}

	/*
		conn, err := connectToSomeJoinedInstance(ctx)
		if err != nil {
			return err
		}
	*/

	return nil
}
