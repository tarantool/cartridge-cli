package failover

import (
	"fmt"

	"github.com/tarantool/cartridge-cli/cli/connector"
	"github.com/tarantool/cartridge-cli/cli/context"
	"github.com/tarantool/cartridge-cli/cli/replicasets"
)

/*
type FailoverOpts struct {
	Mode             string          `yaml:"mode" structs:"mode"`
	StateProvider    *string         `yaml:"state_provider,omitempty" structs:"state_provider"`
	StateboardParams *ProviderParams `yaml:"stateboard_params,omitempty" structs:"tarantool_params"`
	Etcd2Params      *ProviderParams `yaml:"etcd2_params,omitempty" structs:"etcd2_params"`

	FailoverTimeout *int  `yaml:"failover_timeout,omitempty" json:"failover_timeout" structs:"failover_timeout"`
	FencingEnabled  *bool `yaml:"fencing_enabled,omitempty" json:"fencing_enabled" structs:"fencing_enabled"`
	FencingTimeout  *int  `yaml:"fencing_timeout,omitempty" json:"fencing_timeout" structs:"fencing_timeout"`
	FencingPause    *int  `yaml:"fencing_pause,omitempty" json:"fencing_pause" structs:"fencing_pause"`
}
*/

type FailoverOpts map[string]interface{}

type ProviderParams map[string]interface{}

func (failoverOpts FailoverOpts) Manage(ctx *context.Ctx) error {
	// TODO: Change this after https://github.com/tarantool/cartridge-cli/pull/593
	conn, err := replicasets.ConnectToSomeRunningInstance(ctx)
	if err != nil {
		return fmt.Errorf("Failed to connect to some instance: %s", err)
	}

	if provider, found := failoverOpts["state_provider"]; found {
		if provider == "stateboard" {
			failoverOpts["state_provider"] = "tarantool"
		}

		failoverOpts["tarantool_params"] = failoverOpts["stateboard_params"]
		delete(failoverOpts, "stateboard_params")
	}

	result, err := conn.Exec(connector.EvalReq(manageFailoverBody, failoverOpts))
	if err != nil {
		return fmt.Errorf("Failed to configure failover: %s", err)
	}

	if len(result) == 2 {
		if funcErr := result[1]; funcErr != nil {
			return fmt.Errorf("Failed to configure failover: %s", funcErr)
		}
	}

	return nil
}
