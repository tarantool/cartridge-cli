package pack

import (
	"path/filepath"

	log "github.com/sirupsen/logrus"

	"github.com/tarantool/cartridge-cli/src/common"
	"github.com/tarantool/cartridge-cli/src/project"
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
