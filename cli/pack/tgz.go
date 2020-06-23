package pack

import (
	"path/filepath"

	"github.com/apex/log"

	"github.com/tarantool/cartridge-cli/cli/common"
	"github.com/tarantool/cartridge-cli/cli/project"
)

func packTgz(projectCtx *project.ProjectCtx) error {
	var err error

	appDirPath := filepath.Join(projectCtx.PackageFilesDir, projectCtx.Name)
	if err := initAppDir(appDirPath, projectCtx); err != nil {
		return err
	}

	// create archive
	err = common.WriteTgzArchive(projectCtx.PackageFilesDir, projectCtx.ResPackagePath)
	if err != nil {
		return err
	}

	log.Infof("Created result package: %s", projectCtx.ResPackagePath)

	return nil
}
