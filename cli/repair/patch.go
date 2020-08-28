package repair

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/tarantool/cartridge-cli/cli/common"
	"github.com/tarantool/cartridge-cli/cli/context"
	"github.com/tarantool/cartridge-cli/cli/project"
	"github.com/tarantool/cartridge-cli/cli/templates"
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

	conn, err := net.Dial("unix", consoleSock)
	if err != nil {
		return resMessages, fmt.Errorf("Failed to dial: %s", err)
	}

	defer conn.Close()

	// eval
	evalFuncTmpl := `
		local ClusterwideConfig = require('cartridge.clusterwide-config')
		local confapplier = require('cartridge.confapplier')

		local roles_configured_state = 'RolesConfigured'
		local connecting_fullmesh_state = 'ConnectingFullmesh'

		local state = confapplier.wish_state(roles_configured_state, {{ .WishStateTimeout }})

		if state == connecting_fullmesh_state then
			return nil, string.format(
				'Failed to desire %s config state. Stuck in %s. ' ..
					'Call "box.cfg({replication_connect_quorum = 0})" in instance console and try again',
				roles_configured_state, state
			)
		end

		if state ~= roles_configured_state then
			return nil, string.format(
				'Failed to desire %s config state. Stuck in %s',
				roles_configured_state, state
			)
		end

		local cfg, err = ClusterwideConfig.load('{{ .ConfigPath }}')
		if err ~= nil then
			return nil, string.format('Failed to load new config: %s', err)
		end

		cfg:lock()

		local current_uuid = box.info().uuid
		if cfg:get_readonly().topology.servers[current_uuid] == nil then
			return false
		end

		local ok, err = confapplier.apply_config(cfg)

		if err ~= nil then
			return nil, string.format('Failed to apply new config: %s', err)
		end

		return true
`

	evalFunc, err := templates.GetTemplatedStr(&evalFuncTmpl, map[string]string{
		"ConfigPath":       filepath.Dir(topologyConfPath),
		"WishStateTimeout": strconv.Itoa(confapplierWishStateTimeout),
	})

	if err != nil {
		return resMessages, fmt.Errorf("Failed to instantiate reload config function template: %s", err)
	}

	reloadedRaw, err := common.EvalTarantoolConn(conn, evalFunc)
	if err != nil {
		return resMessages, fmt.Errorf("Failed to call reload config function: %s", err)
	}

	reloaded, ok := reloadedRaw.(bool)
	if !ok {
		return nil, project.InternalError("Reload function returned non-bool value: %#v", reloadedRaw)
	}

	if !reloaded {
		resMessages = append(resMessages, common.GetWarnMessage("Cluster-wide config reload was skipped"))
	}

	return resMessages, nil
}
