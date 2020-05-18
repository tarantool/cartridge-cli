package pack

import (
	"fmt"
	"path/filepath"

	log "github.com/sirupsen/logrus"
	"github.com/tarantool/cartridge-cli/project"
)

func packDeb(projectCtx *project.ProjectCtx) error {
	// check context
	if err := checkPackDebCtx(projectCtx); err != nil {
		// TODO: format internal error
		panic(err)
	}

	// app dir
	appDirPath := filepath.Join(projectCtx.PackageFilesDir, "/usr/share/tarantool/", projectCtx.Name)
	if err := initAppDir(appDirPath, projectCtx); err != nil {
		return err
	}

	// systemd dir
	if err := initSystemdDir(projectCtx.PackageFilesDir, projectCtx); err != nil {
		return err
	}

	// tmpfiles dir
	if err := initTmpfilesDir(projectCtx.PackageFilesDir, projectCtx); err != nil {
		return err
	}

	log.Infof("Created result package: %s", projectCtx.ResPackagePath)

	return nil
}

func checkPackDebCtx(projectCtx *project.ProjectCtx) error {
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
