package admin

import (
	"fmt"
	"time"

	"github.com/apex/log"
	"github.com/tarantool/cartridge-cli/cli/connector"
)

func adminFuncList(conn *connector.Conn) error {
	funcInfos, err := getListFuncInfos(conn)
	if err != nil {
		return fmt.Errorf("Failed to get functions list: %s", err)
	}

	log.Infof("Available admin functions:\n\n%s", funcInfos.FormatUsages())

	return nil
}

func getListFuncInfos(conn *connector.Conn) (*FuncInfos, error) {
	funcBody, err := getAdminFuncEvalTypedBody(adminListFuncName)
	if err != nil {
		return nil, err
	}

	req := connector.EvalReq(funcBody).SetReadTimeout(3 * time.Second)

	funcInfosSlice := []FuncInfos{}
	if err := conn.ExecTyped(req, &funcInfosSlice); err != nil {
		return nil, err
	}

	if len(funcInfosSlice) != 1 {
		return nil, fmt.Errorf("Function signature received in a bad format")
	}

	funcInfos := funcInfosSlice[0]

	return &funcInfos, nil
}
