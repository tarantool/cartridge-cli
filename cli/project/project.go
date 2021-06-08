package project

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime/debug"

	"github.com/tarantool/cartridge-cli/cli/common"
	"github.com/tarantool/cartridge-cli/cli/context"
	"github.com/tarantool/cartridge-cli/cli/version"
)

func FillTarantoolCtx(ctx *context.Ctx) error {
	var err error

	ctx.Tarantool.TarantoolDir, err = common.GetTarantoolDir()
	if err != nil {
		return fmt.Errorf("Failed to find Tarantool executable: %s", err)
	} else {
		ctx.Tarantool.TarantoolVersion, err = common.GetTarantoolVersion(ctx.Tarantool.TarantoolDir)
		if err != nil {
			return fmt.Errorf("Failed to get Tarantool version: %s", err)
		}

		ctx.Tarantool.TarantoolIsEnterprise, err = common.TarantoolIsEnterprise(ctx.Tarantool.TarantoolDir)
		if err != nil {
			return fmt.Errorf("Failed to check Tarantool version: %s", err)
		}
	}

	return nil
}

func GetStateboardName(ctx *context.Ctx) string {
	return fmt.Sprintf("%s-stateboard", ctx.Project.Name)
}

func SetProjectPath(ctx *context.Ctx) error {
	var err error

	if ctx.Project.Path == "" {
		ctx.Project.Path, err = os.Getwd()
		if err != nil {
			return fmt.Errorf("Failed to get current directory: %s", err)
		}
	}

	ctx.Project.Path, err = filepath.Abs(ctx.Project.Path)
	if err != nil {
		return fmt.Errorf("Failed to get absolute path for %s: %s", ctx.Project.Path, err)
	}

	return nil
}

func DetectName(path string) (string, error) {
	var err error

	if _, err := os.Stat(path); err != nil {
		return "", fmt.Errorf("Unable to use specified path: %s", err)
	}

	rockspecPath, err := common.FindRockspec(path)
	if err != nil {
		return "", err
	} else if rockspecPath == "" {
		return "", fmt.Errorf("Application directory should contain rockspec")
	}

	name, err := common.LuaReadStringVar(rockspecPath, "package")
	if err != nil {
		return "", fmt.Errorf("Failed to read `package` field from rockspec: %s", err)
	}

	return name, nil
}

func InternalError(format string, a ...interface{}) error {
	const internalErrorFmt = `Whoops! It looks like something is wrong with this version of Cartridge CLI.
Please, report a bug at https://github.com/tarantool/cartridge-cli/issues/new.
Error: %s
Version: %s
Stacktrace:
%s
`
	msg := fmt.Sprintf(format, a...)
	return fmt.Errorf(internalErrorFmt, msg, version.BuildCliVersionString(), debug.Stack())
}
