package replicasets

import (
	"fmt"
	"strings"

	"github.com/apex/log"
	"github.com/tarantool/cartridge-cli/cli/cluster"
	"github.com/tarantool/cartridge-cli/cli/connector"
	"github.com/tarantool/cartridge-cli/cli/context"
)

func BootstrapVshard(ctx *context.Ctx, args []string) error {
	conn, err := cluster.ConnectToSomeJoinedInstance(ctx)
	if err != nil {
		return err
	}

	if err := bootstrapVshard(conn); err != nil {
		return fmt.Errorf("failed to bootstrap vshard: %s", err)
	}

	log.Infof("Bootstrap vshard task completed successfully, check the cluster status")

	return nil
}

func bootstrapVshard(conn *connector.Conn) error {
	req := connector.EvalReq(bootstrapVshardBody)

	if _, err := conn.Exec(req); err != nil {
		if strings.Contains(err.Error(), `Sharding config is empty`) {
			// XXX: see https://github.com/tarantool/cartridge/issues/1148
			log.Warnf(
				`It's possible that there is no running instances of some configured vshard groups. ` +
					`In this case existing storages are bootstrapped, but Cartridge returns an error`,
			)
		}

		return err
	}

	return nil
}
