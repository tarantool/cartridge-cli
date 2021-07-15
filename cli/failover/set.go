package failover

import (
	"encoding/json"
	"fmt"

	"github.com/apex/log"
	"github.com/tarantool/cartridge-cli/cli/context"
)

func Set(ctx *context.Ctx) error {
	if (ctx.Failover.Mode == "eventual" || ctx.Failover.Mode == "disabled") && ctx.Failover.ProviderParamsJSON != "" {
		return fmt.Errorf("Please, don't specify provider parameters when using %s mode", ctx.Failover.Mode)
	}

	if err := FillCtx(ctx); err != nil {
		return err
	}

	failoverOpts, err := getFailoverOpts(ctx)
	if err != nil {
		return err
	}

	log.Infof("Set up %s failover", failoverOpts.Mode)
	if err := failoverOpts.Manage(ctx); err != nil {
		return err
	}

	log.Infof("Failover configured successfully")

	return nil
}

func getFailoverOpts(ctx *context.Ctx) (*FailoverOpts, error) {
	failoverOpts, err := initFailoverOpts(ctx)
	if err != nil {
		return nil, err
	}

	if failoverOpts.Mode == "stateful" && failoverOpts.StateProvider != nil && ctx.Failover.ProviderParamsJSON != "" {
		var providerParams ProviderParams
		if err := json.Unmarshal([]byte(ctx.Failover.ProviderParamsJSON), &providerParams); err != nil {
			return nil, fmt.Errorf("Failed to parse provider parameters: %s", err)
		}

		if *failoverOpts.StateProvider == "stateboard" {
			failoverOpts.StateboardParams = &providerParams
		} else if *failoverOpts.StateProvider == "etcd2" {
			failoverOpts.Etcd2Params = &providerParams
		}
	}

	if err := validateFailoverOpts(failoverOpts); err != nil {
		return nil, err
	}

	return failoverOpts, nil
}

func initFailoverOpts(ctx *context.Ctx) (*FailoverOpts, error) {
	failoverOpts := FailoverOpts{
		Mode: ctx.Failover.Mode,
	}

	if ctx.Failover.ParamsJSON != "" {
		if err := json.Unmarshal([]byte(ctx.Failover.ParamsJSON), &failoverOpts); err != nil {
			return nil, fmt.Errorf("Failed to parse failover parameters: %s", err)
		}
	}

	if ctx.Failover.StateProvider == "" {
		failoverOpts.StateProvider = nil
	} else {
		failoverOpts.StateProvider = &ctx.Failover.StateProvider
	}

	return &failoverOpts, nil
}
