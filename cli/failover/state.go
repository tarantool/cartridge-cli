package failover

import (
	"fmt"
	"reflect"
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
	if provider, found := resultMap["state_provider"]; found {
		if provider == "tarantool" {
			resultMap["state_provider"] = "stateboard"
			resultMap["stateboard_params"] = resultMap["tarantool_params"]
			delete(resultMap, "tarantool_params")
		}
	}

	return internalRecFailoverStatePrettyString(resultMap, 0)
}

func internalRecFailoverStatePrettyString(result interface{}, tabCnt int) string {
	acc := ""

	for fieldName, value := range result.(map[string]interface{}) {
		acc = fmt.Sprintf("%s %s• %s:", acc, strings.Repeat("\t", tabCnt), common.ColorHiMagenta.Sprintf(fieldName))

		switch value.(type) {
		case map[string]interface{}:
			acc = fmt.Sprintf("%s \n%s", acc, internalRecFailoverStatePrettyString(value, tabCnt+1))
		case int8:
			acc = fmt.Sprintf("%s %d\n", acc, value)
		case bool:
			acc = fmt.Sprintf("%s %t\n", acc, value)
		case string:
			acc = fmt.Sprintf("%s %s\n", acc, value)
		case []interface{}:
			for _, elem := range value.([]interface{}) {
				acc = fmt.Sprintf("%s %s,", acc, elem)
			}

			acc = fmt.Sprintf("%s\n", acc[:len(acc)-1])
		default:
			panic(project.InternalError("Unknown type: %s", reflect.TypeOf(value)))
		}
	}

	return acc
}
