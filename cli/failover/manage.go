package failover

import (
	"fmt"
	"strings"

	"github.com/tarantool/cartridge-cli/cli/cluster"
	"github.com/tarantool/cartridge-cli/cli/connector"
	"github.com/tarantool/cartridge-cli/cli/context"
)

type FailoverOpts map[string]interface{}
type ProviderParams map[string]interface{}

func (failoverOpts FailoverOpts) Manage(ctx *context.Ctx) error {
	conn, err := cluster.ConnectToSomeRunningInstance(ctx)
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
			return fmt.Errorf(
				"Failed to configure failover: %s",
				// Cartridge may use 'tarantool_params' in error messages. It can confuse the user
				strings.Replace(funcErr.(string), "tarantool_params", "stateboard_params", -1),
			)
		}
	}

	return nil
}
