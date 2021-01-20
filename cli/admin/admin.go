package admin

import (
	"fmt"

	"github.com/spf13/pflag"

	"github.com/tarantool/cartridge-cli/cli/connector"
	"github.com/tarantool/cartridge-cli/cli/context"
)

const (
	// Names of global functions which are called on the application side
	// to use admin functions
	// These functions are exposed by admin extension on init
	// See https://github.com/tarantool/cartridge-cli-extensions
	adminListFuncName = "__cartridge_admin_list"
	adminHelpFuncName = "__cartridge_admin_help"
	adminCallFuncName = "__cartridge_admin_call"
)

type ProcessAdminFuncType func(conn *connector.Conn, funcName string, flagSet *pflag.FlagSet, args []string) error

func Run(processAdminFunc ProcessAdminFuncType, ctx *context.Ctx, funcName string, flagSet *pflag.FlagSet, args []string) error {
	if err := checkCtx(ctx); err != nil {
		return err
	}

	conn, err := getAvaliableConn(ctx)
	if err != nil {
		return fmt.Errorf("Failed to connect to application instance: %s", err)
	}
	defer conn.Close()

	return processAdminFunc(conn, funcName, flagSet, args)
}

func List(conn *connector.Conn, funcName string, flagSet *pflag.FlagSet, args []string) error {
	return adminFuncList(conn)
}

func Help(conn *connector.Conn, funcName string, flagSet *pflag.FlagSet, args []string) error {
	return adminFuncHelp(conn, flagSet, funcName)
}

func Call(conn *connector.Conn, funcName string, flagSet *pflag.FlagSet, args []string) error {
	return adminFuncCall(conn, funcName, flagSet, args)
}
