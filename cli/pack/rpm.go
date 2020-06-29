package pack

import (
	"path/filepath"

	"github.com/tarantool/cartridge-cli/cli/common"

	log "github.com/sirupsen/logrus"

	"github.com/tarantool/cartridge-cli/cli/project"
	"github.com/tarantool/cartridge-cli/cli/rpm"
)

func packRpm(projectCtx *project.ProjectCtx) error {
	if err := common.CheckRequiredBinaries("cpio"); err != nil {
		return err
	}

	appDirPath := filepath.Join(projectCtx.PackageFilesDir, projectCtx.AppDir)
	if err := initAppDir(appDirPath, projectCtx); err != nil {
		return err
	}

	if err := initSystemdDir(projectCtx.PackageFilesDir, projectCtx); err != nil {
		return err
	}

	if err := initTmpfilesDir(projectCtx.PackageFilesDir, projectCtx); err != nil {
		return err
	}

	log.Infof("Create result RPM package...")

	// construct RPM file
	if err := rpm.Pack(projectCtx); err != nil {
		return err
	}

	log.Infof("Created result package: %s", projectCtx.ResPackagePath)

	return nil
}
