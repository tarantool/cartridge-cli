package pack

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/apex/log"

	"github.com/tarantool/cartridge-cli/cli/common"
	"github.com/tarantool/cartridge-cli/cli/context"
)

const (
	defaultHomeDir      = "/home"
	tmpPackDirNameFmt   = "pack-%s"
	tmpCacheDirName     = "cache"
	defaultBuildDirName = "cartridge.tmp"
	packageFilesDirName = "package-files"
)

var (
	defaultCartridgeTmpDir string
)

func init() {
	homeDir, err := common.GetHomeDir()
	if err != nil {
		homeDir = defaultHomeDir
	}

	defaultCartridgeTmpDir = filepath.Join(homeDir, ".cartridge/tmp")
}

// tmp directory structure:
// ~/.cartridge/tmp/            <- ctx.Cli.CartridgeTmpDir (can be changed by CARTRIDGE_TEMPDIR)
//   cache/                     <- ctx.Cli.CacheDir (used for saving cache modules when packing application)
//     <project-hash>/			<- Directory containing cached project modules this hash
//   pack-s18h29agl2/           <- ctx.Cli.TmpDir (ctx.Pack.ID is used)
//     package-files/           <- PackageFilesDir
//       usr/share/tarantool
//       ...
//     tmp-build-file           <- additional files used for building the application

func detectTmpDir(ctx *context.Ctx) error {
	var err error

	if ctx.Cli.CartridgeTmpDir == "" {
		// tmp dir wasn't specified
		ctx.Cli.CartridgeTmpDir = defaultCartridgeTmpDir
	} else {
		// tmp dir was specified
		ctx.Cli.CartridgeTmpDir, err = filepath.Abs(ctx.Cli.CartridgeTmpDir)
		if err != nil {
			return fmt.Errorf(
				"Failed to get absolute path for specified temporary dir %s: %s",
				ctx.Cli.CartridgeTmpDir,
				err,
			)
		}

		if fileInfo, err := os.Stat(ctx.Cli.CartridgeTmpDir); err == nil {
			// directory is already exists

			if !fileInfo.IsDir() {
				return fmt.Errorf(
					"Specified temporary directory is not a directory: %s",
					ctx.Cli.CartridgeTmpDir,
				)
			}

			// This little hack is used to prevent deletion of user files
			// from the specified tmp directory on cleanup.
			ctx.Cli.CartridgeTmpDir = filepath.Join(ctx.Cli.CartridgeTmpDir, defaultBuildDirName)

		} else if !os.IsNotExist(err) {
			return fmt.Errorf(
				"Unable to use specified temporary directory %s: %s",
				ctx.Cli.CartridgeTmpDir,
				err,
			)
		}
	}

	tmpDirName := fmt.Sprintf(tmpPackDirNameFmt, ctx.Pack.ID)
	ctx.Cli.TmpDir = filepath.Join(ctx.Cli.CartridgeTmpDir, tmpDirName)
	ctx.Cli.CacheDir = filepath.Join(ctx.Cli.CartridgeTmpDir, tmpCacheDirName)

	return nil
}

func initTmpDir(ctx *context.Ctx) error {
	if _, err := os.Stat(ctx.Cli.TmpDir); err == nil {
		log.Debugf("Tmp directory already exists. Cleaning it...")

		if err := common.ClearDir(ctx.Cli.TmpDir); err != nil {
			return fmt.Errorf("Failed to cleanup build dir: %s", err)
		}
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("Unable to use temporary directory %s: %s", ctx.Cli.TmpDir, err)
	} else if err := os.MkdirAll(ctx.Cli.TmpDir, 0755); err != nil {
		return fmt.Errorf("Failed to create temporary directory %s: %s", ctx.Cli.TmpDir, err)
	}

	ctx.Pack.PackageFilesDir = filepath.Join(ctx.Cli.TmpDir, packageFilesDirName)

	if err := os.MkdirAll(ctx.Pack.PackageFilesDir, 0755); err != nil {
		return fmt.Errorf(
			"Failed to create package files directory %s: %s",
			ctx.Pack.PackageFilesDir,
			err,
		)
	}

	return nil
}
