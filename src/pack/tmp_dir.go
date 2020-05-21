package pack

import (
	"fmt"
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"

	"github.com/tarantool/cartridge-cli/src/common"
	"github.com/tarantool/cartridge-cli/src/project"
)

const (
	defaultHomeDir      = "/home"
	tmpPackDirNameFmt   = "pack-%s"
	defaultBuildDirName = "cartridge.tmp"
	packageFilesDirName = "package-files"
)

var (
	defaultCartridgeTmpDir string
)

func init() {
	homeDir, err := common.GetHomeDir()
	if err != nil {
		homeDir = defaultHomeDir
	}

	defaultCartridgeTmpDir = filepath.Join(homeDir, ".cartridge/tmp")
}

// tmp directory structure:
// ~/.cartridge/tmp/            <- cartridgeTmpDir (can be changed by CARTRIDGE_TEMPDIR)
//   pack-s18h29agl2/           <- projectCtx.TmpDir (projectCtx.PackID is used)
//     package-files/           <- PackageFilesDir
//       usr/share/tarantool
//       ...
//     tmp-build-file           <- additional files used for building the application

func detectTmpDir(projectCtx *project.ProjectCtx) error {
	var err error

	var cartridgeTmpDir string

	if projectCtx.TmpDir == "" {
		// tmp dir wasn't specified
		cartridgeTmpDir = defaultCartridgeTmpDir
	} else {
		// tmp dir was specified
		cartridgeTmpDir, err = filepath.Abs(projectCtx.TmpDir)
		if err != nil {
			return fmt.Errorf(
				"Failed to get absolute path for specified temporary dir %s: %s",
				cartridgeTmpDir,
				err,
			)
		}

		// tmp directory can't be project subdirectory
		if isSubDir, err := common.IsSubDir(cartridgeTmpDir, projectCtx.Path); err != nil {
			return fmt.Errorf(
				"Failed to check that specified temporary dir %s is a project subdir: %s",
				cartridgeTmpDir,
				err,
			)
		} else if isSubDir {
			return fmt.Errorf(
				"Temporary directory can't be project subdirectory, specified: %s",
				cartridgeTmpDir,
			)
		}

		if fileInfo, err := os.Stat(cartridgeTmpDir); err == nil {
			// directory is already exists

			if !fileInfo.IsDir() {
				return fmt.Errorf(
					"Specified temporary directory is not a directory: %s",
					cartridgeTmpDir,
				)
			}

			// This little hack is used to prevent deletion of user files
			// from the specified tmp directory on cleanup.
			cartridgeTmpDir = filepath.Join(cartridgeTmpDir, defaultBuildDirName)

		} else if !os.IsNotExist(err) {
			return fmt.Errorf(
				"Unable to use specified temporary directory %s: %s",
				cartridgeTmpDir,
				err,
			)
		}
	}

	tmpDirName := fmt.Sprintf(tmpPackDirNameFmt, projectCtx.PackID)
	projectCtx.TmpDir = filepath.Join(cartridgeTmpDir, tmpDirName)

	return nil
}

func initTmpDir(projectCtx *project.ProjectCtx) error {
	if _, err := os.Stat(projectCtx.TmpDir); err == nil {
		log.Debugf("Tmp directory already exists. Cleaning it...")

		if err := common.ClearDir(projectCtx.TmpDir); err != nil {
			return fmt.Errorf("Failed to cleanup build dir: %s", err)
		}
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("Unable to use temporary directory %s: %s", projectCtx.TmpDir, err)
	} else if err := os.MkdirAll(projectCtx.TmpDir, 0755); err != nil {
		return fmt.Errorf("Failed to create temporary directory %s: %s", projectCtx.TmpDir, err)
	}

	projectCtx.PackageFilesDir = filepath.Join(projectCtx.TmpDir, packageFilesDirName)

	if err := os.MkdirAll(projectCtx.PackageFilesDir, 0755); err != nil {
		return fmt.Errorf(
			"Failed to create package files directory %s: %s",
			projectCtx.PackageFilesDir,
			err,
		)
	}

	return nil
}

func removeTmpDir(projectCtx *project.ProjectCtx) {
	if projectCtx.Debug {
		log.Warnf("Temporary directory %s is not removed due to debug mode", projectCtx.TmpDir)
		return
	}

	if err := os.RemoveAll(projectCtx.TmpDir); err != nil {
		log.Warnf("Failed to remove tmp directory %s: %s", projectCtx.TmpDir, err)
	} else {
		log.Infof("Temporary directory %s is removed", projectCtx.TmpDir)
	}
}
