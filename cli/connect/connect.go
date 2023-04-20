package connect

import (
	"fmt"

	"github.com/tarantool/cartridge-cli/cli/common"
	"github.com/tarantool/cartridge-cli/cli/context"
	"github.com/tarantool/cartridge-cli/cli/project"
	"github.com/tarantool/cartridge-cli/cli/running"
)

const (
	// see https://github.com/tarantool/tarantool/blob/b53cb2aeceedc39f356ceca30bd0087ee8de7c16/src/box/lua/console.c#L265
	tarantoolWordSeparators = "\t\r\n !\"#$%&'()*+,-/;<=>?@[\\]^`{|}~"
)

func Enter(ctx *context.Ctx, args []string) error {
	var err error

	if err := project.FillCtx(ctx); err != nil {
		return err
	}

	if ctx.Running.Instances, err = common.GetInstancesFromArgs(args); err != nil {
		return err
	}

	if len(ctx.Running.Instances) != 1 {
		return fmt.Errorf("Should be specified one instance name")
	}

	instanceName := ctx.Running.Instances[0]

	process := running.NewInstanceProcess(ctx, instanceName)
	if !process.IsRunning() {
		return common.ErrWrapCheckInstanceNameCommonMisprint([]string{instanceName}, ctx.Project.Name,
			fmt.Errorf("Instance %s is not running", instanceName))
	}

	socketPath := project.GetInstanceConsoleSock(ctx, instanceName)
	title := project.GetInstanceID(ctx, instanceName)

	connOpts := ConnOpts{
		Network: "unix",
		Address: socketPath,
	}

	if err := runConsole(&connOpts, title); err != nil {
		return fmt.Errorf("Failed to run interactive console: %s", err)
	}

	return nil
}

func Connect(ctx *context.Ctx, args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("Should be specified one connection string")
	}

	connString := args[0]

	connOpts, err := getConnOpts(connString, ctx)
	if err != nil {
		return fmt.Errorf("Failed to get connection opts: %s", err)
	}

	if err := runConsole(connOpts, ""); err != nil {
		return fmt.Errorf("Failed to run interactive console: %s", err)
	}

	return nil

}

func runConsole(connOpts *ConnOpts, title string) error {
	console, err := NewConsole(connOpts, title)
	if err != nil {
		return fmt.Errorf("Failed to create new console: %s", err)
	}
	defer console.Close()

	if err := console.Run(); err != nil {
		return fmt.Errorf("Failed to start new console: %s", err)
	}

	return nil
}
