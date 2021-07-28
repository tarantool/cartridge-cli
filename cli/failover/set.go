package failover

import (
	"encoding/json"
	"fmt"

	"github.com/apex/log"
	"github.com/tarantool/cartridge-cli/cli/context"
	"github.com/tarantool/cartridge-cli/cli/project"
)

func Set(ctx *context.Ctx) error {
	if (ctx.Failover.Mode == "eventual" || ctx.Failover.Mode == "disabled") && ctx.Failover.ProviderParamsJSON != "" {
		return fmt.Errorf("Please, don't specify provider parameters when using %s mode", ctx.Failover.Mode)
	}

	if ctx.Failover.Mode == "disabled" {
		return Disable(ctx)
	}

	if err := project.FillCtx(ctx); err != nil {
		return err
	}

	failoverOpts, err := getFailoverOpts(ctx)
	if err != nil {
		return err
	}

	log.Infof("Configure %s failover", (*failoverOpts)["mode"])

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

	if (*failoverOpts)["mode"] == "stateful" {
		if _, found := (*failoverOpts)["state_provider"]; found && ctx.Failover.ProviderParamsJSON != "" {
			var providerParams ProviderParams
			if err := json.Unmarshal([]byte(ctx.Failover.ProviderParamsJSON), &providerParams); err != nil {
				return nil, fmt.Errorf("Failed to parse provider parameters: %s", err)
			}

			if (*failoverOpts)["state_provider"] == "stateboard" {
				(*failoverOpts)["stateboard_params"] = providerParams
			} else if (*failoverOpts)["state_provider"] == "etcd2" {
				(*failoverOpts)["etcd2_params"] = providerParams
			}
		}
	}

	if err := validateSetFailoverOpts(failoverOpts); err != nil {
		return nil, err
	}

	return failoverOpts, nil
}

func initFailoverOpts(ctx *context.Ctx) (*FailoverOpts, error) {
	failoverOpts := FailoverOpts{
		"mode": ctx.Failover.Mode,
	}

	if ctx.Failover.ParamsJSON != "" {
		if err := json.Unmarshal([]byte(ctx.Failover.ParamsJSON), &failoverOpts); err != nil {
			return nil, fmt.Errorf("Failed to parse failover parameters: %s", err)
		}
	}

	if ctx.Failover.StateProvider != "" {
		failoverOpts["state_provider"] = ctx.Failover.StateProvider
	}

	return &failoverOpts, nil
}
