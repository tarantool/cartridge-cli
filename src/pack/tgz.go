package pack

import (
	"fmt"
	"path/filepath"

	log "github.com/sirupsen/logrus"

	"github.com/tarantool/cartridge-cli/src/common"
	"github.com/tarantool/cartridge-cli/src/project"
)

func packTgz(projectCtx *project.ProjectCtx) error {
	var err error

	// check context
	if err := checkPackTgzCtx(projectCtx); err != nil {
		// TODO: format internal error
		panic(err)
	}

	// app dir
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

func checkPackTgzCtx(projectCtx *project.ProjectCtx) error {
	if projectCtx.Version == "" {
		return fmt.Errorf("Missed project version")
	}

	if projectCtx.Release == "" {
		return fmt.Errorf("Missed project release")
	}

	if projectCtx.VersionRelease == "" {
		return fmt.Errorf("Missed project version with release")
	}

	if projectCtx.TarantoolVersion == "" {
		return fmt.Errorf("Missed Tarantool version")
	}

	if projectCtx.ResPackagePath == "" {
		return fmt.Errorf("Missed result package path")
	}

	return nil
}
