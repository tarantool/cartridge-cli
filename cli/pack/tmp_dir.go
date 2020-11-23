package pack

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/apex/log"

	"github.com/tarantool/cartridge-cli/cli/common"
	"github.com/tarantool/cartridge-cli/cli/context"
	"github.com/tarantool/cartridge-cli/cli/project"
)

const (
	packTmpDirNameFmt   = "pack-%s"
	packageFilesDirName = "package-files"
)

// tmp directory structure:
// ~/.cartridge/tmp/            <- ctx.Cli.CartridgeTmpDir (can be changed by CARTRIDGE_TEMPDIR)
//   pack-s18h29agl2/           <- ctx.Pack.TmpDir (ctx.Pack.ID is used)
//     package-files/           <- PackageFilesDir
//       usr/share/tarantool
//       ...
//     tmp-build-file           <- additional files used for building the application

func setPackTmpDir(ctx *context.Ctx) error {
	if err := project.SetCartridgeTmpDir(ctx); err != nil {
		return fmt.Errorf("Failed to detect tmp directory: %s", err)
	}

	packTmpDirName := fmt.Sprintf(packTmpDirNameFmt, ctx.Pack.ID)
	ctx.Pack.TmpDir = filepath.Join(ctx.Cli.CartridgeTmpDir, packTmpDirName)

	return nil
}

func initPackTmpDir(ctx *context.Ctx) error {
	if _, err := os.Stat(ctx.Pack.TmpDir); err == nil {
		log.Debugf("Tmp directory already exists. Cleaning it...")

		if err := common.ClearDir(ctx.Pack.TmpDir); err != nil {
			return fmt.Errorf("Failed to cleanup tmp dir: %s", err)
		}
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("Unable to use temporary directory %s: %s", ctx.Pack.TmpDir, err)
	} else if err := os.MkdirAll(ctx.Pack.TmpDir, 0755); err != nil {
		return fmt.Errorf("Failed to create temporary directory %s: %s", ctx.Pack.TmpDir, err)
	}

	ctx.Pack.PackageFilesDir = filepath.Join(ctx.Pack.TmpDir, packageFilesDirName)

	if err := os.MkdirAll(ctx.Pack.PackageFilesDir, 0755); err != nil {
		return fmt.Errorf(
			"Failed to create package files directory %s: %s",
			ctx.Pack.PackageFilesDir,
			err,
		)
	}

	return nil
}
