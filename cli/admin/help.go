package admin

import (
	"fmt"
	"strings"
	"time"

	"github.com/apex/log"

	"github.com/spf13/pflag"
	"github.com/tarantool/cartridge-cli/cli/connector"
)

func adminFuncHelp(conn *connector.Conn, flagSet *pflag.FlagSet, funcName string) error {
	funcInfo, err := getFuncInfo(funcName, conn)
	if err != nil {
		return getCliExtError("Failed to get function %q signature: %s", funcName, err)
	}

	log.Infof("Admin function %q usage:\n\n%s", funcName, funcInfo.Format())

	conflictingFlagNames := getConflictingFlagNames(funcInfo.Args, flagSet)
	if len(conflictingFlagNames) > 0 {
		log.Warnf(
			"Function has arguments with names that conflict with `cartridge admin` flags: %s. "+
				"Calling this function will raise an error",
			strings.Join(conflictingFlagNames, ", "),
		)
	}

	return nil
}

func getFuncInfo(funcName string, conn *connector.Conn) (*FuncInfo, error) {
	funcBody, err := getAdminFuncEvalTypedBody(adminHelpFuncName)
	if err != nil {
		return nil, err
	}

	req := connector.EvalReq(funcBody, funcName).SetReadTimeout(3 * time.Second)

	funcInfoSlice := []FuncInfo{}
	if err := conn.ExecTyped(req, &funcInfoSlice); err != nil {
		return nil, err
	}

	if len(funcInfoSlice) != 1 {
		return nil, fmt.Errorf("Function signature received in a bad format")
	}

	funcInfo := funcInfoSlice[0]
	funcInfo.Name = funcName

	return &funcInfo, nil

}
