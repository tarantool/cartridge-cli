package pack

import (
	"fmt"
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

	err = common.RunFunctionWithSpinner(func() error {
		return common.WriteTgzArchive(projectCtx.PackageFilesDir, projectCtx.ResPackagePath)
	}, "Creating result TGZ archive...")
	if err != nil {
		return fmt.Errorf("Failed to create TGZ archive: %s", err)
	}

	log.Infof("Created result TGZ archive: %s", projectCtx.ResPackagePath)

	return nil
}
