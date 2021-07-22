package cluster

import (
	"fmt"
	"time"

	"github.com/tarantool/cartridge-cli/cli/common"
	"github.com/tarantool/cartridge-cli/cli/connector"
)

const (
	SimpleOperationTimeout = 10 * time.Second
)

func HealthCheckIsNeeded(conn *connector.Conn) (bool, error) {
	majorCartridgeVersion, err := common.GetMajorCartridgeVersion(conn)
	if err != nil {
		return false, fmt.Errorf("Failed to get Cartridge major version: %s", err)
	}

	return majorCartridgeVersion < 2, nil
}
