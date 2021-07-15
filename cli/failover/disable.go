package failover

import (
	"github.com/apex/log"
	"github.com/tarantool/cartridge-cli/cli/context"
)

func Disable(ctx *context.Ctx) error {
	if err := FillCtx(ctx); err != nil {
		return err
	}

	failoverOpts := FailoverOpts{Mode: "disabled"}
	if err := failoverOpts.Manage(ctx); err != nil {
		return err
	}

	log.Infof("Failover disabled successfully")

	return nil
}
