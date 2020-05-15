package pack

import (
	"fmt"
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"
	"github.com/tarantool/cartridge-cli/common"
	"github.com/tarantool/cartridge-cli/project"
)

const (
	defaultHomeDir      = "/home"
	buildDirNameFmt     = "cartridge-build-%s"
	defaultBuildDirName = "cartridge.tmp"
	packageFilesDirName = "package-files"
)

var (
	cartridgeTmpDir string
)

func init() {
	homeDir, err := common.GetHomeDir()
	if err != nil {
		homeDir = defaultHomeDir
	}

	cartridgeTmpDir = filepath.Join(homeDir, ".cartridge/tmp")
}

func detectTmpDir(projectCtx *project.ProjectCtx) error {
	var err error

	if projectCtx.TmpDir == "" {
		// tmp dir wasn't specified
		buildDirName := fmt.Sprintf(buildDirNameFmt, projectCtx.BuildID)
		projectCtx.TmpDir = filepath.Join(cartridgeTmpDir, buildDirName)
	} else {
		// tmp dir was specified
		projectCtx.TmpDir, err = filepath.Abs(projectCtx.TmpDir)
		if err != nil {
			return fmt.Errorf(
				"Failed to get absolute path for specified temporary dir %s: %s",
				projectCtx.TmpDir,
				err,
			)
		}

		// tmp directory can't be project subdirectory
		if isSubDir, err := common.IsSubDir(projectCtx.TmpDir, projectCtx.Path); err != nil {
			return fmt.Errorf(
				"Failed to check that specified temporary dir %s is a project subdir: %s",
				projectCtx.TmpDir,
				err,
			)
		} else if isSubDir {
			return fmt.Errorf(
				"Temporary directory can't be project subdirectory, specified: %s",
				projectCtx.TmpDir,
			)
		}

		if fileInfo, err := os.Stat(projectCtx.TmpDir); err == nil {
			// directory is already exists

			// This little hack is used to prevent deletion of user files
			// from the specified build directory on cleanup.
			// Moreover, this subdirectory is definitely clean,
			// so we'll have no problems
			if !fileInfo.IsDir() {
				return fmt.Errorf(
					"Specified temporary directory is not a directory: %s",
					projectCtx.TmpDir,
				)
			}

			projectCtx.TmpDir = filepath.Join(projectCtx.TmpDir, defaultBuildDirName)

		} else if !os.IsNotExist(err) {
			return fmt.Errorf(
				"Unable to use specified temporary directory %s: %s",
				projectCtx.TmpDir,
				err,
			)
		}
	}

	return nil
}

// tmp directory structure:
// tmp-dir/
//   package-files/           <- package files
//     usr/share/tarantool/   <- build dir
//     or
//     appname/               <- build dir
//   tmp-build-file           <- additional files used for building the application
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
