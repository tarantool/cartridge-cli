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
	defaultBuildDirName = "build.cartridge"
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

func detectBuildDir(projectCtx *project.ProjectCtx) error {
	var err error

	if projectCtx.BuildDir == "" {
		// build dir wasn't specified
		buildDirName := fmt.Sprintf(buildDirNameFmt, projectCtx.BuildID)
		projectCtx.BuildDir = filepath.Join(cartridgeTmpDir, buildDirName)
	} else {
		// build dir was specified
		projectCtx.BuildDir, err = filepath.Abs(projectCtx.BuildDir)
		if err != nil {
			return fmt.Errorf(
				"Failed to get absolute path for specified build dir %s: %s",
				projectCtx.BuildDir,
				err,
			)
		}

		// build directory can't be project subdirectory
		if isSubDir, err := common.IsSubDir(projectCtx.BuildDir, projectCtx.Path); err != nil {
			return fmt.Errorf(
				"Failed to check that specified build dir %s is a project subdir: %s",
				projectCtx.BuildDir,
				err,
			)
		} else if isSubDir {
			return fmt.Errorf(
				"Build directory can't be project subdirectory, specified: %s",
				projectCtx.BuildDir,
			)
		}

		if fileInfo, err := os.Stat(projectCtx.BuildDir); err == nil {
			// directory is already exists

			// This little hack is used to prevent deletion of user files
			// from the specified build directory on cleanup.
			// Moreover, this subdirectory is definitely clean,
			// so we'll have no problems
			if !fileInfo.IsDir() {
				return fmt.Errorf(
					"Specified build directory is not a directory: %s",
					projectCtx.BuildDir,
				)
			}

			projectCtx.BuildDir = filepath.Join(projectCtx.BuildDir, defaultBuildDirName)

		} else if !os.IsNotExist(err) {
			return fmt.Errorf(
				"Unable to use specified build directory %s: %s",
				projectCtx.BuildDir,
				err,
			)
		}
	}

	return nil
}

// build directory structure:
// build-dir/
//   package-files/           <- package files
//     usr/share/tarantool/
//     or
//     appname/
//   tmp-build-file           <- additional files used for building the application
func initBuildDir(projectCtx *project.ProjectCtx) error {
	if _, err := os.Stat(projectCtx.BuildDir); err == nil {
		log.Debugf("Build directory already exists. Cleaning it...")

		if err := common.ClearDir(projectCtx.BuildDir); err != nil {
			return fmt.Errorf("Failed to cleanup build dir: %s", err)
		}
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("Unable to use build directory %s: %s", projectCtx.BuildDir, err)
	} else if err := os.MkdirAll(projectCtx.BuildDir, 0755); err != nil {
		return fmt.Errorf("Failed to create build directory %s: %s", projectCtx.BuildDir, err)
	}

	projectCtx.PackageFilesDir = filepath.Join(projectCtx.BuildDir, packageFilesDirName)

	if err := os.MkdirAll(projectCtx.PackageFilesDir, 0755); err != nil {
		return fmt.Errorf(
			"Failed to create package files directory %s: %s",
			projectCtx.PackageFilesDir,
			err,
		)
	}

	return nil
}

func removeBuildDir(projectCtx *project.ProjectCtx) {
	if projectCtx.Debug {
		log.Warnf("Build directory %s is not removed due to debug mode", projectCtx.BuildDir)
		return
	}

	if err := os.RemoveAll(projectCtx.BuildDir); err != nil {
		log.Warnf("Failed to remove build directory %s: %s", projectCtx.BuildDir, err)
	} else {
		log.Infof("Build directory %s is removed", projectCtx.BuildDir)
	}
}
