package admin

import (
	"fmt"
	"strings"

	"github.com/apex/log"
	"github.com/tarantool/cartridge-cli/cli/connector"
	"github.com/tarantool/cartridge-cli/cli/project"
	"gopkg.in/yaml.v2"

	"github.com/spf13/pflag"
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
	NumberValue float64
	BoolValue   bool

	// Changed is an equivalent for pflag.Flag.Changed
	// It's true when user entered argument value
	// If this value is false, argument isn't passed to function
	// (in fact, opt.<arg-name> is nil)
	Changed bool
}

func adminFuncCall(conn *connector.Conn, funcName string, flagSet *pflag.FlagSet, args []string) error {
	funcCallOpts, err := getFuncCallOpts(conn, funcName, flagSet, args)
	if err != nil {
		return fmt.Errorf("Failed to parse function call args: %s", err)
	}

	callReq := connector.CallReq(adminCallFuncName, funcName, funcCallOpts)
	callReq.SetPushCallback(func(pushedData interface{}) {
		printMessage(pushedData)
	})

	callResData, err := conn.Exec(callReq)

	if err != nil {
		return fmt.Errorf("Failed to call %q: %s", funcName, err)
	}

	// it could be one of
	// return res
	// return nil, err
	if len(callResData) < 1 || len(callResData) > 2 {
		return fmt.Errorf("Bad data len: %d", len(callResData))
	}

	if len(callResData) == 2 {
		if funcErr := callResData[1]; funcErr != nil {
			return fmt.Errorf("Failed to call %q: %s", funcName, funcErr)
		}
	}

	callRes := callResData[0]

	printCallRes(callRes)

	return nil
}

func getFuncCallOpts(conn *connector.Conn, funcName string, flagSet *pflag.FlagSet, args []string) (map[string]interface{}, error) {
	funcCallArgsList, err := getFuncCallArgsList(conn, funcName, flagSet, args)
	if err != nil {
		return nil, err
	}

	funcCallOpts := make(map[string]interface{})

	for _, funcCallArg := range funcCallArgsList {
		if !funcCallArg.Changed {
			continue
		}

		var value interface{}

		switch funcCallArg.Type {
		case argTypeString:
			value = funcCallArg.StringValue
		case argTypeNumber:
			value = funcCallArg.NumberValue
		case argTypeBool:
			value = funcCallArg.BoolValue
		default:
			return nil, project.InternalError("Unknown argument type: %s", funcCallArg.Type)
		}

		funcCallOpts[funcCallArg.Name] = value
	}

	return funcCallOpts, nil
}

func getFuncCallArgsList(conn *connector.Conn, funcName string, flagSet *pflag.FlagSet, args []string) ([]FuncCallArg, error) {
	funcInfo, err := getFuncInfo(funcName, conn)
	if err != nil {
		return nil, getCliExtError("Failed to get function %q signature: %s", funcName, err)
	}

	conflictingFlagNames := getConflictingFlagNames(funcInfo.Args, flagSet)
	if len(conflictingFlagNames) > 0 {
		return nil, fmt.Errorf(
			"Function has arguments with names that conflict with `cartridge admin` flags: %s",
			strings.Join(conflictingFlagNames, ", "),
		)
	}

	funcCallArgs := make([]FuncCallArg, len(funcInfo.Args))

	i := 0
	for argName, argSpec := range funcInfo.Args {
		funcCallArgs[i] = FuncCallArg{
			Name: argName,
			Type: argSpec.Type,
		}

		switch argSpec.Type {
		case argTypeString:
			flagSet.StringVar(&funcCallArgs[i].StringValue, argName, "", argSpec.Usage)
		case argTypeNumber:
			flagSet.Float64Var(&funcCallArgs[i].NumberValue, argName, 0, argSpec.Usage)
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

func printMessage(pushedData interface{}) {
	msg, ok := pushedData.(string)
	if !ok {
		log.Warnf("Intermediate message should be a string, got %v", pushedData)

		msgEncoded, err := yaml.Marshal(pushedData)
		if err != nil {
			log.Errorf("Failed to encode received intermediate message: %s", err)
		}
		fmt.Printf("%s", msgEncoded)
		return
	}

	log.Info(msg)
}
