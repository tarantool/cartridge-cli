package build

import (
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"

	"github.com/tarantool/cartridge-cli/common"
	"github.com/tarantool/cartridge-cli/project"
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

	if fileInfo, err := os.Stat(projectCtx.Path); err != nil {
		return fmt.Errorf("Unable to build application in %s: %s", projectCtx.Path, err)
	} else if !fileInfo.IsDir() {
		return fmt.Errorf("Unable to build application in %s: it's not a directory", projectCtx.Path)
	}

	log.Infof("Building application in %s...", projectCtx.Path)

	projectCtx.BuildDir = projectCtx.Path

	// check context
	if err := checkCtx(projectCtx); err != nil {
		// TODO: format internal error
		panic(err)
	}

	// check that application directory contains rockspec
	if rockspec, err := common.FindRockspec(projectCtx.Path); err != nil {
		return err
	} else if rockspec == "" {
		return fmt.Errorf("Application directory should contain rockspec")
	}

	var err error
	if projectCtx.BuildInDocker {
		panic("Not implemented yet")
	} else {
		err = buildProjectLocally(projectCtx)
	}
	if err != nil {
		return err
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
