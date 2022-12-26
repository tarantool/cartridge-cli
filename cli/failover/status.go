package failover

import (
	"fmt"
	"reflect"
	"sort"
	"strings"

	"github.com/apex/log"
	"github.com/tarantool/cartridge-cli/cli/cluster"
	"github.com/tarantool/cartridge-cli/cli/common"
	"github.com/tarantool/cartridge-cli/cli/connector"
	"github.com/tarantool/cartridge-cli/cli/context"
	"github.com/tarantool/cartridge-cli/cli/project"
)

func Status(ctx *context.Ctx) error {
	if err := project.FillCtx(ctx); err != nil {
		return err
	}

	conn, err := cluster.ConnectToSomeRunningInstance(ctx)
	if err != nil {
		return fmt.Errorf("Failed to connect to some instance: %s", err)
	}

	var result []map[string]interface{}
	if err := conn.ExecTyped(connector.EvalReq(getFailoverParamsBody), &result); err != nil {
		return fmt.Errorf("Failed to get current failover status: %s", err)
	}

	log.Infof("Current failover status: ")

	print(getFailoverStatusPrettyString(result[0]))

	return nil
}

func getFailoverStatusPrettyString(resultMap map[string]interface{}) string {
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

	return internalRecFailoverStatusPrettyString(resultMap, 0)
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

func internalRecFailoverStatusPrettyString(result interface{}, indentCnt int) string {
	status := ""
	resultMap := result.(map[string]interface{})

	for _, fieldName := range getSortedFailoverMapKeys(resultMap) {
		status = fmt.Sprintf(
			"%s %s• %s:", status,
			strings.Repeat("    ", indentCnt), common.ColorCyan.Sprintf(fieldName),
		)

		value := resultMap[fieldName]
		switch value.(type) {
		case map[string]interface{}:
			status = fmt.Sprintf("%s \n%s", status, internalRecFailoverStatusPrettyString(value, indentCnt+1))
		case int, int8, uint16:
			status = fmt.Sprintf("%s %d\n", status, value)
		case bool:
			status = fmt.Sprintf("%s %t\n", status, value)
		case string:
			switch value {
			case "disabled":
				status = fmt.Sprintf("%s %s\n", status, common.ColorRed.Sprintf(value.(string)))
			case "eventual", "stateful", "raft":
				status = fmt.Sprintf("%s %s\n", status, common.ColorGreen.Sprintf(value.(string)))
			default:
				status = fmt.Sprintf("%s %s\n", status, value)
			}
		case []interface{}:
			for _, elem := range value.([]interface{}) {
				status = fmt.Sprintf("%s %s,", status, elem)
			}

			status = fmt.Sprintf("%s\n", status[:len(status)-1])
		default:
			panic(project.InternalError("Field %s has unknown type: %s", fieldName, reflect.TypeOf(value)))
		}
	}

	return status
}
