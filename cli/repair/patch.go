package repair

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/tarantool/cartridge-cli/cli/connector"

	"github.com/tarantool/cartridge-cli/cli/common"
	"github.com/tarantool/cartridge-cli/cli/context"
	"github.com/tarantool/cartridge-cli/cli/project"
)

const (
	confapplierWishStateTimeout = 10
)

func patchConf(patchFunc PatchConfFuncType, topologyConf *TopologyConfType, ctx *context.Ctx) ([]common.ResultMessage, error) {
	var resMessages []common.ResultMessage

	currentConfContent, err := topologyConf.MarshalContent()
	if err != nil {
		return nil, fmt.Errorf("Failed to marshal current content: %s", err)
	}

	if err := patchFunc(topologyConf, ctx); err != nil {
		return nil, fmt.Errorf("Failed to patch topology config: %s", err)
	}

	newConfContent, err := topologyConf.MarshalContent()
	if err != nil {
		return nil, fmt.Errorf("Failed to get new config content: %s", err)
	}

	if ctx.Repair.DryRun || ctx.Cli.Verbose {
		configDiff, err := getDiffLines(currentConfContent, newConfContent, "", "")
		if err != nil {
			return nil, fmt.Errorf("Failed to get config difference: %s", err)
		}

		if len(configDiff) > 0 {
			resMessages = append(resMessages, common.GetInfoMessage((strings.Join(configDiff, "\n") + "\n")))
		} else {
			resMessages = append(resMessages, common.GetInfoMessage("Topology config wasn't changed"))
		}
	}

	return resMessages, nil
}

func rewriteConf(topologyConfPath string, topologyConf *TopologyConfType) ([]common.ResultMessage, error) {
	var resMessages []common.ResultMessage

	resMessages = append(resMessages, common.GetDebugMessage("Topology config file: %s", topologyConfPath))

	backupPath, err := createFileBackup(topologyConfPath)
	if err != nil {
		return nil, fmt.Errorf("Failed to create topology config backup: %s", err)
	}
	resMessages = append(resMessages, common.GetDebugMessage("Created backup file: %s", backupPath))

	newConfContent, err := topologyConf.MarshalContent()
	if err != nil {
		return nil, fmt.Errorf("Failed to get new config content: %s", err)
	}

	confFile, err := os.OpenFile(topologyConfPath, os.O_WRONLY|os.O_TRUNC, 0755)
	if err != nil {
		return nil, fmt.Errorf("Failed to open a new config: %s", err)
	}
	defer confFile.Close()

	if _, err := confFile.Write(newConfContent); err != nil {
		return nil, fmt.Errorf("Failed to write a new config: %s", err)
	}

	return resMessages, nil
}

func reloadConf(topologyConfPath string, instanceName string, ctx *context.Ctx) ([]common.ResultMessage, error) {
	var resMessages []common.ResultMessage

	consoleSock := project.GetInstanceConsoleSock(ctx, instanceName)
	resMessages = append(resMessages, common.GetDebugMessage("Instance console socket: %s", consoleSock))

	if _, err := os.Stat(consoleSock); err != nil {
		return nil, fmt.Errorf("Failed to use instanace console socket: %s", err)
	}

	conn, err := connector.Connect(consoleSock, connector.Opts{})
	if err != nil {
		return resMessages, fmt.Errorf("Failed to connect to Tarantool instance: %s", err)
	}
	defer conn.Close()

	// eval
	confPath := filepath.Dir(topologyConfPath)
	req := connector.EvalReq(reloadClusterwideConfigFuncBody, confPath, confapplierWishStateTimeout)

	var results []bool
	if err := conn.ExecTyped(req, &results); err != nil {
		return resMessages, fmt.Errorf("Failed to reload clusterwide config: %s", err)
	}

	if len(results) != 1 {
		return resMessages, fmt.Errorf("Result received in a bad format")
	}

	reloaded := results[0]

	if !reloaded {
		resMessages = append(resMessages, common.GetWarnMessage("Cluster-wide config reload was skipped"))
	}

	return resMessages, nil
}
