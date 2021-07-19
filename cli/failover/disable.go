package failover

import (
	"github.com/apex/log"
	"github.com/tarantool/cartridge-cli/cli/context"
	"github.com/tarantool/cartridge-cli/cli/project"
)

func Disable(ctx *context.Ctx) error {
	if err := project.FillCtx(ctx); err != nil {
		return err
	}

	failoverOpts, err := getFailoverOpts(ctx)
	if err != nil {
		return err
	}

	if err := failoverOpts.Manage(ctx); err != nil {
		return err
	}

	log.Infof("Failover disabled successfully")

	return nil
}
