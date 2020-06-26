package pack

import (
	"fmt"
	"path/filepath"

	"github.com/tarantool/cartridge-cli/cli/common"
	"github.com/tarantool/cartridge-cli/cli/rpm"

	"github.com/apex/log"

	"github.com/tarantool/cartridge-cli/cli/project"
)

func packRpm(projectCtx *project.ProjectCtx) error {
	var err error

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

	err = common.RunFunctionWithSpinner(func() error {
		return rpm.Pack(projectCtx)
	}, "Creating result RPM package...")
	if err != nil {
		return fmt.Errorf("Failed to create RPM package: %s", err)
	}

	log.Infof("Created result RPM package: %s", projectCtx.ResPackagePath)

	return nil
}
