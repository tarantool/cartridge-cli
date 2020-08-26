package repair

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/apex/log"
	"github.com/tarantool/cartridge-cli/cli/common"
	"github.com/tarantool/cartridge-cli/cli/context"
	"github.com/tarantool/cartridge-cli/cli/project"
	"github.com/tarantool/cartridge-cli/cli/templates"
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
		log.Fatalf("Failed to dial: %s", err)
	}

	defer conn.Close()

	// read greeting
	if _, err := common.ReadFromConn(conn); err != nil {
		log.Fatalf("Failed to read greeting: %s", err)
	}

	// eval
	evalFuncTmpl := `
		local ClusterwideConfig = require('cartridge.clusterwide-config')
		local cfg, err = ClusterwideConfig.load('{{ .ConfigPath }}')
		if err ~= nil then
			return nil, string.format('Failed to load new config: %s', err)
		end
		cfg:lock()

		local confapplier = require('cartridge.confapplier')

		local desired_conf_state = 'RolesConfigured'
		local state = confapplier.wish_state(desired_conf_state, {{ .WishStateTimeout }})

		if state ~= desired_conf_state then
			return nil, string.format(
				'Failed to desire %s config state. Stuck in %s',
				desired_conf_state, state
			)
		end

		local ok, err = confapplier.apply_config(cfg)

		if err ~= nil then
			return nil, string.format('Failed to apply new config: %s', err)
		end

		return true
`

	evalFunc, err := templates.GetTemplatedStr(&evalFuncTmpl, map[string]string{
		"ConfigPath":       filepath.Dir(topologyConfPath), // XXX
		"WishStateTimeout": strconv.Itoa(10),               // XXX
	})

	if err != nil {
		return resMessages, fmt.Errorf("Failed to instantiate reload config function template: %s", err)
	}

	if _, err := common.EvalTarantoolConn(conn, evalFunc); err != nil {
		return resMessages, fmt.Errorf("Failed to call reload config function: %s", err)
	}

	return resMessages, nil
}
