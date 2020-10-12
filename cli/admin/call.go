package admin

import (
	"fmt"
	"net"
	"strings"

	"github.com/apex/log"
	"github.com/tarantool/cartridge-cli/cli/project"

	"github.com/spf13/pflag"

	"github.com/tarantool/cartridge-cli/cli/common"
	"github.com/tarantool/cartridge-cli/cli/templates"
)

const (
	argTypeString = "string"
	argTypeNumber = "number"
	argTypeBool   = "boolean"
)

type FuncCallArg struct {
	Name string
	Type string

	StringValue string
	NumberValue int
	BoolValue   bool

	// Changed is an equivalent for pflag.Flag.Changed
	// It's true when user entered argument value
	// If this value is false, argument isn't passed to function
	// (in fact, opt.<arg-name> is nil)
	Changed bool
}

func adminFuncCall(conn net.Conn, funcName string, flagSet *pflag.FlagSet, args []string) error {
	funcCallArgs, err := getFuncCallArgs(conn, funcName, flagSet, args)
	if err != nil {
		return fmt.Errorf("Failed to parse function call args: %s", err)
	}

	argsSerialized, err := serializeArgs(funcCallArgs)
	if err != nil {
		return fmt.Errorf("Failed to serialize function args: %s", err)
	}

	callFuncBody, err := templates.GetTemplatedStr(&adminCallFuncBodyTmpl, map[string]string{
		"AdminCallFuncName": adminCallFuncName,
		"FuncName":          funcName,
		"Args":              argsSerialized,
	})

	callResRaw, err := common.EvalTarantoolConn(conn, callFuncBody)
	if err != nil {
		return fmt.Errorf("Failed to call %q: %s", funcName, err)
	}

	printCallRes(callResRaw)

	return nil
}

func getFuncCallArgs(conn net.Conn, funcName string, flagSet *pflag.FlagSet, args []string) ([]FuncCallArg, error) {
	helpResRawMap, err := getFuncHelpRawMap(funcName, conn)
	if err != nil {
		return nil, getCliExtError("Failed to get function %q signature: %s", funcName, err)
	}

	argsSpec, err := getArgsSpec(helpResRawMap)
	if err != nil {
		return nil, getCliExtError("Failed to get %q arguments spec: %s", funcName, err)
	}

	conflictingFlagNames := getConflictingFlagNames(argsSpec, flagSet)
	if len(conflictingFlagNames) > 0 {
		return nil, fmt.Errorf(
			"Function has arguments with names that conflict with `cartridge admin` flags: %s",
			strings.Join(conflictingFlagNames, ", "),
		)
	}

	funcCallArgs := make([]FuncCallArg, len(argsSpec))

	i := 0
	for argName, argSpec := range argsSpec {
		funcCallArgs[i] = FuncCallArg{
			Name: argName,
			Type: argSpec.Type,
		}

		switch argSpec.Type {
		case argTypeString:
			flagSet.StringVar(&funcCallArgs[i].StringValue, argName, "", argSpec.Usage)
		case argTypeNumber:
			flagSet.IntVar(&funcCallArgs[i].NumberValue, argName, 0, argSpec.Usage)
		case argTypeBool:
			flagSet.BoolVar(&funcCallArgs[i].BoolValue, argName, false, argSpec.Usage)
		default:
			return nil, fmt.Errorf(
				"Admin function %q accepts value of unsupported type: %s "+
					"(supported types are %s)",
				funcName, argSpec.Type,
				strings.Join([]string{argTypeString, argTypeNumber, argTypeBool}, ", "),
			)
		}

		i++
	}

	flagSet.SetNormalizeFunc(func(f *pflag.FlagSet, name string) pflag.NormalizedName {
		return pflag.NormalizedName(normalizeFlagName(name))
	})

	flagSet.ParseErrorsWhitelist = pflag.ParseErrorsWhitelist{
		UnknownFlags: false,
	}
	if err := flagSet.Parse(args); err != nil {
		return nil, fmt.Errorf("Failed to parse %q function arguments: %s", funcName, err)
	}

	// set Changed for args
	for i := range funcCallArgs {
		flag := flagSet.Lookup(funcCallArgs[i].Name)
		if flag == nil {
			return nil, project.InternalError("Flag %q isn't found", funcCallArgs[i].Name)
		}

		funcCallArgs[i].Changed = flag.Changed
	}

	return funcCallArgs, nil
}

func serializeArgs(funcCallArgs []FuncCallArg) (string, error) {
	argStrings := []string{}

	for _, funcCallArg := range funcCallArgs {
		var valueStr string

		if !funcCallArg.Changed {
			continue
		}

		switch funcCallArg.Type {
		case argTypeString:
			valueStr = fmt.Sprintf("'%s'", funcCallArg.StringValue)
		case argTypeNumber:
			valueStr = fmt.Sprintf("%d", funcCallArg.NumberValue)
		case argTypeBool:
			valueStr = fmt.Sprintf("%t", funcCallArg.BoolValue)
		default:
			return "", project.InternalError("Received argument with unsupported type: %s", funcCallArg.Type)
		}

		argStrings = append(argStrings, fmt.Sprintf("%s = %s", funcCallArg.Name, valueStr))
	}

	return fmt.Sprintf(`{ %s }`, strings.Join(argStrings, ", ")), nil
}

func printCallRes(callResRaw interface{}) {
	needReturnValueWarn := false

	switch callRes := callResRaw.(type) {
	case string:
		log.Info(callRes)
	case []interface{}:
		for _, callResLineRaw := range callRes {
			switch callResLine := callResLineRaw.(type) {
			case string:
				log.Info(callResLine)
			default:
				needReturnValueWarn = true
				log.Infof("%v\n", callResLine)
			}
		}
	default:
		needReturnValueWarn = true
		log.Infof("%v\n", callRes)
	}

	if needReturnValueWarn {
		log.Warnf("Admin function should return string or string array value")
	}
}

var (
	adminCallFuncBodyTmpl = `
	local func_help, err = {{ .AdminCallFuncName }}(
		'{{ .FuncName }}',
		{{ .Args }}
	)
	return func_help, err
`
)
