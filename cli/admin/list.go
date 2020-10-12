package admin

import (
	"fmt"
	"net"
	"sort"

	"github.com/apex/log"
	"github.com/tarantool/cartridge-cli/cli/common"
	"github.com/tarantool/cartridge-cli/cli/templates"
)

func adminFuncList(conn net.Conn) error {
	listResRawMap, err := getFuncListRawMap(conn)
	if err != nil {
		return fmt.Errorf("Failed to get functions list: %s", err)
	}

	if len(listResRawMap) == 0 {
		return fmt.Errorf("No available admin functions found")
	}

	funcUsages := make(NameUsages, len(listResRawMap))

	i := 0
	for funcNameRaw, funcSpecRaw := range listResRawMap {
		funcName, ok := funcNameRaw.(string)
		if !ok {
			return getCliExtError("Functions map key isn't a string: %#v", funcNameRaw)
		}

		adminFuncSpec, ok := funcSpecRaw.(map[interface{}]interface{})
		if !ok {
			return getCliExtError("Function %q spec isn't a map: %#v", funcName, funcSpecRaw)
		}

		funcUsage, err := getStrValueFromRawMap(adminFuncSpec, "usage")
		if err != nil {
			return getCliExtError("Failed to get function %q usage: %s", funcName, err)
		}

		funcUsages[i] = NameUsage{
			Name:  funcName,
			Usage: funcUsage,
		}

		i++
	}

	sort.Sort(funcUsages)
	log.Infof("Available admin functions:\n\n%s", funcUsages.Format())

	return nil
}

func getFuncListRawMap(conn net.Conn) (map[interface{}]interface{}, error) {
	adminListFuncBody, err := templates.GetTemplatedStr(&adminListFuncBodyTmpl, map[string]string{
		"AdminListFuncName": adminListFuncName,
	})

	listResRaw, err := common.EvalTarantoolConn(conn, adminListFuncBody)
	if err != nil {
		return nil, fmt.Errorf("Failed to call %s(): %s", adminListFuncName, err)
	}

	listResRawMap, err := convertToMap(listResRaw)
	if err != nil {
		return nil, getCliExtError("Failed to convert %q return value to map", adminListFuncName)
	}

	return listResRawMap, nil
}

var (
	adminListFuncBodyTmpl = `
	local func_help, err = {{ .AdminListFuncName }}('{{ .FuncName }}')
	return func_help, err
`
)
