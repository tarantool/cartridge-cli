package build

import (
	"fmt"
	"os"

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

	if projectCtx.BuildInDocker {
		if err := buildProjectInDocker(projectCtx); err != nil {
			return err
		}
	} else {
		if err := buildProjectLocally(projectCtx); err != nil {
			return err
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
