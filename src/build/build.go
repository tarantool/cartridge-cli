package build

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/otiai10/copy"
	log "github.com/sirupsen/logrus"

	"github.com/tarantool/cartridge-cli/src/common"
	"github.com/tarantool/cartridge-cli/src/project"
)

const (
	preBuildHookName  = "cartridge.pre-build"
	postBuildHookName = "cartridge.post-build"
)

// Run builds project in projectCtx.BuildDir
// If projectCtx.BuildInDocker is set, application is built in docker
func Run(projectCtx *project.ProjectCtx) error {
	if err := project.CheckTarantoolBinaries(); err != nil {
		return fmt.Errorf("Tarantool binaries are required to build application")
	}

	if projectCtx.BuildID == "" {
		projectCtx.BuildID = common.RandomString(10)
	}

	// check context
	if err := checkCtx(projectCtx); err != nil {
		// TODO: format internal error
		panic(err)
	}

	if fileInfo, err := os.Stat(projectCtx.BuildDir); err != nil {
		return fmt.Errorf("Unable to build application in %s: %s", projectCtx.BuildDir, err)
	} else if !fileInfo.IsDir() {
		return fmt.Errorf("Unable to build application in %s: it's not a directory", projectCtx.BuildDir)
	}

	log.Infof("Building application in %s...", projectCtx.BuildDir)

	// check that application directory contains rockspec
	if rockspecPath, err := common.FindRockspec(projectCtx.Path); err != nil {
		return err
	} else if rockspecPath == "" {
		return fmt.Errorf("Application directory should contain rockspec")
	}

	// set projectCtx.SDKPath and init projectCtx.BuildSDKPath
	if projectCtx.BuildInDocker && projectCtx.TarantoolIsEnterprise {
		if err := setSDKPath(projectCtx); err != nil {
			return err
		}

		if err := initBuildSDKPath(projectCtx); err != nil {
			return err
		}
	}

	if projectCtx.BuildInDocker {
		if err := buildProjectInDocker(projectCtx); err != nil {
			return err
		}
	} else {
		if err := buildProjectLocally(projectCtx); err != nil {
			return err
		}
	}

	// clean up build SDK
	if projectCtx.BuildInDocker && projectCtx.TarantoolIsEnterprise {
		if err := os.RemoveAll(projectCtx.BuildSDKPath); err != nil {
			return fmt.Errorf("Failed to remove build SDK: %s", err)
		}
	}

	log.Infof("Application build succeeded")

	return nil
}

func checkCtx(projectCtx *project.ProjectCtx) error {
	if projectCtx.Path == "" {
		return fmt.Errorf("Missed project path")
	}

	if projectCtx.BuildDir == "" {
		return fmt.Errorf("Missed build directory")
	}

	return nil
}

func setSDKPath(projectCtx *project.ProjectCtx) error {
	if !common.OnlyOneIsTrue(projectCtx.SDKPath != "", projectCtx.SDKLocal) {
		return fmt.Errorf(sdkPathError)
	}

	if projectCtx.SDKLocal {
		projectCtx.SDKPath = projectCtx.TarantoolDir
	}

	return nil
}

func initBuildSDKPath(projectCtx *project.ProjectCtx) error {
	projectCtx.BuildSDKPath = filepath.Join(
		projectCtx.BuildDir,
		fmt.Sprintf("sdk-%s", projectCtx.BuildID),
	)

	if err := copy.Copy(projectCtx.SDKPath, projectCtx.BuildSDKPath); err != nil {
		return err
	}

	return nil
}

const (
	sdkPathError = `For packing in docker you should specify one of:
* --sdk-local: to use local SDK;;
* --sdk-path: path to SDK
  (can be passed in environment variable TARANTOOL_SDK_PATH).`
)
