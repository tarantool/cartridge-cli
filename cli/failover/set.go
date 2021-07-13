package failover

import (
	"encoding/json"
	"fmt"

	"github.com/apex/log"
	"github.com/tarantool/cartridge-cli/cli/context"
)

func Set(ctx *context.Ctx, providerParamsJSON string) error {
	if err := FillCtx(ctx); err != nil {
		return err
	}

	opts, err := getFailoverOpts(ctx, providerParamsJSON)
	if err != nil {
		return err
	}

	if err := setupFailover(ctx, opts); err != nil {
		return err
	}

	log.Infof("Failover configured successfully")

	return nil
}

func getFailoverOpts(ctx *context.Ctx, providerParamsJSON string) (*FailoverOpts, error) {
	opts := initFailoverOpts(ctx)

	if opts.Mode == "stateful" && opts.StateProvider != nil {
		var providerParams ProviderParams
		if err := json.Unmarshal([]byte(providerParamsJSON), &providerParams); err != nil {
			return nil, err
		}

		if *opts.StateProvider == "stateboard" {
			opts.StateboardParams = &providerParams
		} else if *opts.StateProvider == "etcd2" {
			opts.Etcd2Params = &providerParams
		}
	}

	if err := validateFailoverOpts(opts); err != nil {
		return nil, fmt.Errorf("Failed to validate failover options: %s", err)
	}

	return opts, nil
}

func initFailoverOpts(ctx *context.Ctx) *FailoverOpts {
	opts := FailoverOpts{
		Mode:          ctx.Failover.Mode,
		StateProvider: &ctx.Failover.StateProvider,
	}

	if ctx.Failover.FailoverTimeoutIsSet {
		opts.FailoverTimeout = &ctx.Failover.FailoverTimeout
	}

	if ctx.Failover.FencingEnabledIsSet {
		opts.FencingEnabled = &ctx.Failover.FencingEnabled
	}

	if ctx.Failover.FencingTimeoutIsSet {
		opts.FencingTimeout = &ctx.Failover.FencingTimeout
	}

	if ctx.Failover.FencingPauseIsSet {
		opts.FencingPause = &ctx.Failover.FencingPause
	}

	return &opts
}
