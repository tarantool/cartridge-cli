package admin

import (
	"fmt"
	"net"
	"sort"
	"strings"

	"github.com/apex/log"
	"github.com/spf13/pflag"
	"github.com/tarantool/cartridge-cli/cli/common"

	"github.com/tarantool/cartridge-cli/cli/templates"
)

func adminFuncHelp(conn net.Conn, flagSet *pflag.FlagSet, funcName string) error {
	helpResRawMap, err := getFuncHelpRawMap(funcName, conn)
	if err != nil {
		return getCliExtError("Failed to get function %q signature: %s", funcName, err)
	}

	funcUsage, err := getStrValueFromRawMap(helpResRawMap, "usage")
	if err != nil {
		return getCliExtError("Failed to get %q usage: %s", funcName, err)
	}

	argsSpec, err := getArgsSpec(helpResRawMap)
	if err != nil {
		return getCliExtError("Failed to get %q arguments spec: %s", funcName, err)
	}

	funcHelpMsg, err := getFuncHelpMsg(funcName, funcUsage, argsSpec)
	if err != nil {
		return getCliExtError("Failed to get function %q usage: %s", funcName, err)
	}

	log.Infof("Admin function %q usage:\n\n%s", funcName, funcHelpMsg)

	conflictingFlagNames := getConflictingFlagNames(argsSpec, flagSet)
	if len(conflictingFlagNames) > 0 {
		log.Warnf(
			"Function has arguments with names that conflicts with `cartridge admin` flags: %s. "+
				"Calling this function will raise an error",
			strings.Join(conflictingFlagNames, ", "),
		)
	}

	return nil
}

func getFuncHelpRawMap(funcName string, conn net.Conn) (map[interface{}]interface{}, error) {
	adminHelpFuncBody, err := templates.GetTemplatedStr(&adminHelpFuncBodyTmpl, map[string]string{
		"AdminHelpFuncName": adminHelpFuncName,
		"FuncName":          funcName,
	})

	helpResRaw, err := common.EvalTarantoolConn(conn, adminHelpFuncBody)
	if err != nil {
		return nil, fmt.Errorf("Failed to call %s(%q): %s", adminHelpFuncName, funcName, err)
	}

	helpResRawMap, err := convertToMap(helpResRaw)
	if err != nil {
		return nil, fmt.Errorf("Failed to convert %q return value to map", adminHelpFuncName)
	}

	return helpResRawMap, nil
}

func getFuncHelpMsg(funcName string, funcUsage string, argsSpec ArgsSpec) (string, error) {
	argsUsages := make(NameUsages, len(argsSpec))

	i := 0
	for argName, argSpec := range argsSpec {
		prettyArgName := strings.ReplaceAll(argName, "_", "-")

		argsUsages[i] = NameUsage{
			Name:  fmt.Sprintf("  --%s %s", prettyArgName, argSpec.Type),
			Usage: argSpec.Usage,
		}

		i++
	}

	sort.Sort(argsUsages)
	argsUsageStr := argsUsages.Format()

	funcHelpMsg, err := templates.GetTemplatedStr(&funcHelpMsgTmpl, map[string]interface{}{
		"FuncUsage": funcUsage,
		"ArgsUsage": argsUsageStr,
	})

	if err != nil {
		return "", fmt.Errorf("Failed to execute function usage template: %s", err)
	}

	return funcHelpMsg, nil
}

var (
	adminHelpFuncBodyTmpl = `
	local func_help, err = {{ .AdminHelpFuncName }}('{{ .FuncName }}')
	return func_help, err
`

	funcHelpMsgTmpl = `{{ .FuncUsage }}{{ if .ArgsUsage }}

Args:
{{ .ArgsUsage }}{{ end }}`
)
