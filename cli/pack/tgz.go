package pack

import (
	"fmt"
	"path/filepath"

	"github.com/apex/log"

	"github.com/tarantool/cartridge-cli/cli/common"
	"github.com/tarantool/cartridge-cli/cli/context"
)

func packTgz(ctx *context.Ctx) error {
	var err error

	appDirPath := filepath.Join(ctx.Pack.PackageFilesDir, ctx.Project.Name)
	if err := initAppDir(appDirPath, ctx); err != nil {
		return err
	}

	err = common.RunFunctionWithSpinner(func() error {
		return common.WriteTgzArchive(ctx.Pack.PackageFilesDir, ctx.Pack.ResPackagePath)
	}, "Creating result TGZ archive...")
	if err != nil {
		return fmt.Errorf("Failed to create TGZ archive: %s", err)
	}

	log.Infof("Created result TGZ archive: %s", ctx.Pack.ResPackagePath)

	return nil
}
