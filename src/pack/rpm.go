package pack

import (
	"fmt"
	"path/filepath"

	log "github.com/sirupsen/logrus"

	"github.com/tarantool/cartridge-cli/src/project"
	"github.com/tarantool/cartridge-cli/src/rpm"
)

func packRpm(projectCtx *project.ProjectCtx) error {
	// check context
	if err := checkPackRpmCtx(projectCtx); err != nil {
		// TODO: format internal error
		panic(err)
	}

	appDirPath := filepath.Join(projectCtx.PackageFilesDir, "/usr/share/tarantool/", projectCtx.Name)
	if err := initAppDir(appDirPath, projectCtx); err != nil {
		return err
	}

	if err := initSystemdDir(projectCtx.PackageFilesDir, projectCtx); err != nil {
		return err
	}

	if err := initTmpfilesDir(projectCtx.PackageFilesDir, projectCtx); err != nil {
		return err
	}

	// construct RPM file
	if err := rpm.Pack(projectCtx); err != nil {
		return err
	}

	log.Infof("Created result package: %s", projectCtx.ResPackagePath)

	return nil
}

func checkPackRpmCtx(projectCtx *project.ProjectCtx) error {
	if projectCtx.Version == "" {
		return fmt.Errorf("Version is missed")
	}

	if projectCtx.Release == "" {
		return fmt.Errorf("Release is missed")
	}

	if projectCtx.TarantoolVersion == "" {
		return fmt.Errorf("TarantoolVersion is missed")
	}

	if projectCtx.ResPackagePath == "" {
		return fmt.Errorf("ResPackagePath is missed")
	}

	if projectCtx.TmpDir == "" {
		return fmt.Errorf("TmpDir is missed")
	}

	if projectCtx.PackageFilesDir == "" {
		return fmt.Errorf("PackageFilesDir is missed")
	}

	return nil
}
