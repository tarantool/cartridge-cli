package pack

import (
	"fmt"

	"github.com/tarantool/cartridge-cli/src/project"
	"github.com/tarantool/cartridge-cli/src/rpm"
)

func packRpm(projectCtx *project.ProjectCtx) error {
	// check context
	if err := checkPackRpmCtx(projectCtx); err != nil {
		// TODO: format internal error
		panic(err)
	}

	// // app dir
	// appDirPath := filepath.Join(projectCtx.PackageFilesDir, "/usr/share/tarantool/", projectCtx.Name)
	// if err := initAppDir(appDirPath, projectCtx); err != nil {
	// 	return err
	// }

	// systemd dir
	if err := initSystemdDir(projectCtx.PackageFilesDir, projectCtx); err != nil {
		return err
	}

	// tmpfiles dir
	if err := initTmpfilesDir(projectCtx.PackageFilesDir, projectCtx); err != nil {
		return err
	}

	// construct RPM file
	if err := rpm.Pack(projectCtx); err != nil {
		return err
	}

	return nil
}

func checkPackRpmCtx(projectCtx *project.ProjectCtx) error {
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
