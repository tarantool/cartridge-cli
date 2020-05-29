package pack

import (
	log "github.com/sirupsen/logrus"

	"github.com/tarantool/cartridge-cli/src/project"
)

func packDocker(projectCtx *project.ProjectCtx) error {
	// appDirPath := filepath.Join(projectCtx.PackageFilesDir, projectCtx.Name)
	// if err := initAppDir(appDirPath, projectCtx); err != nil {
	// 	return err
	// }

	log.Infof("Result image tagged as: %s", projectCtx.ResImageFullname)

	return nil
}
