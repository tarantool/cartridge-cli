package pack

import (
	"fmt"
	"path/filepath"

	"github.com/tarantool/cartridge-cli/cli/common"
	"github.com/tarantool/cartridge-cli/cli/context"
	"github.com/tarantool/cartridge-cli/cli/rpm"

	"github.com/apex/log"
)

func packRpm(ctx *context.Ctx) error {
	var err error

	if err := common.CheckRequiredBinaries("cpio"); err != nil {
		return err
	}

	appDirPath := filepath.Join(ctx.Pack.PackageFilesDir, ctx.Running.AppDir)
	if err := initAppDir(appDirPath, ctx); err != nil {
		return err
	}

	if err := initSystemdDir(ctx.Pack.PackageFilesDir, ctx); err != nil {
		return err
	}

	if err := initTmpfilesDir(ctx.Pack.PackageFilesDir, ctx); err != nil {
		return err
	}

	err = common.RunFunctionWithSpinner(func() error {
		return rpm.Pack(ctx)
	}, "Creating result RPM package...")
	if err != nil {
		return fmt.Errorf("Failed to create RPM package: %s", err)
	}

	log.Infof("Created result RPM package: %s", ctx.Pack.ResPackagePath)

	return nil
}
