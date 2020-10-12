package admin

import (
	"fmt"
	"net"

	"github.com/spf13/pflag"

	"github.com/apex/log"
	"github.com/tarantool/cartridge-cli/cli/context"
	"github.com/tarantool/cartridge-cli/cli/project"
)

const (
	// Names of global functions which are called on the application-side
	// to use admin functions
	// This functions are exposed by admin extention on init
	// See https://github.com/tarantool/cartridge-cli-extentions
	adminListFuncName = "__cartridge_admin_list"
	adminHelpFuncName = "__cartridge_admin_help"
	adminCallFuncName = "__cartridge_admin_call"
)

type ProcessAdminFuncType func(conn net.Conn, funcName string, flagSet *pflag.FlagSet, args []string) error

func Run(processAdminFunc ProcessAdminFuncType, ctx *context.Ctx, funcName string, flagSet *pflag.FlagSet, args []string) error {
	if ctx.Project.Name == "" {
		return fmt.Errorf("Please, specify application name using --name")
	}

	if err := project.SetSystemRunningPaths(ctx); err != nil {
		return fmt.Errorf("Failed to get default paths: %s", err)
	}

	log.Debugf("Run directory is set to: %s", ctx.Running.RunDir)

	conn, err := getAvaliableConn(ctx)
	if err != nil {
		return fmt.Errorf("Failed to connect to application instance socket: %s", err)
	}
	defer conn.Close()

	return processAdminFunc(conn, funcName, flagSet, args)
}

func List(conn net.Conn, funcName string, flagSet *pflag.FlagSet, args []string) error {
	return adminFuncList(conn)
}

func Help(conn net.Conn, funcName string, flagSet *pflag.FlagSet, args []string) error {
	return adminFuncHelp(conn, flagSet, funcName)
}

func Call(conn net.Conn, funcName string, flagSet *pflag.FlagSet, args []string) error {
	return adminFuncCall(conn, funcName, flagSet, args)
}
