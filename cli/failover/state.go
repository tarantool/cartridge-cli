package failover

import (
	"fmt"
	"reflect"
	"sort"
	"strings"

	"github.com/apex/log"
	"github.com/tarantool/cartridge-cli/cli/common"
	"github.com/tarantool/cartridge-cli/cli/connector"
	"github.com/tarantool/cartridge-cli/cli/context"
	"github.com/tarantool/cartridge-cli/cli/project"
	"github.com/tarantool/cartridge-cli/cli/replicasets"
)

func State(ctx *context.Ctx) error {
	if err := project.FillCtx(ctx); err != nil {
		return err
	}

	conn, err := replicasets.ConnectToSomeRunningInstance(ctx)
	if err != nil {
		return fmt.Errorf("Failed to connect to some instance: %s", err)
	}

	var result []map[string]interface{}
	if err := conn.ExecTyped(connector.EvalReq(getFailoverParamsBody), &result); err != nil {
		return fmt.Errorf("Failed to configure failover: %s", err)
	}

	log.Infof("Current failover state: ")

	print(getFailoverStatePrettyString(result[0]))

	return nil
}

func getFailoverStatePrettyString(resultMap map[string]interface{}) string {
	if _, found := resultMap["tarantool_params"]; found {
		resultMap["stateboard_params"] = resultMap["tarantool_params"]
		delete(resultMap, "tarantool_params")
	}

	switch resultMap["state_provider"] {
	case "tarantool":
		resultMap["state_provider"] = "stateboard"
		delete(resultMap, "etcd2_params")
	case "etcd2":
		delete(resultMap, "stateboard_params")
	}

	if resultMap["mode"] == "eventual" || resultMap["mode"] == "disabled" {
		delete(resultMap, "state_provider")
		delete(resultMap, "etcd2_params")
		delete(resultMap, "stateboard_params")
	}

	return internalRecFailoverStatePrettyString(resultMap, 0)
}

func getSortedFailoverMapKeys(stringMap map[string]interface{}) []string {
	mapKeys := make([]string, 0, len(stringMap))
	for key := range stringMap {
		mapKeys = append(mapKeys, key)
	}

	// This sorting will allow us to achieve the following format:
	// * mode
	// * state_provider
	// * etcd2_params or stateboard_params
	// * everything else

	sort.Slice(mapKeys, func(a, b int) bool {
		if mapKeys[a] == "mode" {
			return true
		}

		if mapKeys[b] != "mode" {
			if mapKeys[a] == "state_provider" {
				return true
			}

			return strings.HasSuffix(mapKeys[a], "_params") && mapKeys[b] != "state_provider"
		}

		return false
	})

	return mapKeys
}

func internalRecFailoverStatePrettyString(result interface{}, indentCnt int) string {
	stateString := ""
	resultMap := result.(map[string]interface{})

	for _, fieldName := range getSortedFailoverMapKeys(resultMap) {
		stateString = fmt.Sprintf(
			"%s %s• %s:", stateString,
			strings.Repeat("    ", indentCnt), common.ColorCyan.Sprintf(fieldName),
		)

		value := resultMap[fieldName]

		switch value.(type) {
		case map[string]interface{}:
			stateString = fmt.Sprintf("%s \n%s", stateString, internalRecFailoverStatePrettyString(value, indentCnt+1))
		case int8:
			stateString = fmt.Sprintf("%s %d\n", stateString, value)
		case bool:
			stateString = fmt.Sprintf("%s %t\n", stateString, value)
		case string:
			switch value {
			case "disabled":
				stateString = fmt.Sprintf("%s %s\n", stateString, common.ColorRed.Sprintf(value.(string)))
			case "eventual", "stateful":
				stateString = fmt.Sprintf("%s %s\n", stateString, common.ColorGreen.Sprintf(value.(string)))
			default:
				stateString = fmt.Sprintf("%s %s\n", stateString, value)
			}
		case []interface{}:
			for _, elem := range value.([]interface{}) {
				stateString = fmt.Sprintf("%s %s,", stateString, elem)
			}

			stateString = fmt.Sprintf("%s\n", stateString[:len(stateString)-1])
		default:
			panic(project.InternalError("Unknown type: %s", reflect.TypeOf(value)))
		}
	}

	return stateString
}
