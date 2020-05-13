package build

import (
	log "github.com/sirupsen/logrus"

	"github.com/tarantool/cartridge-cli/project"
)

const (
	preBuildHook  = "cartridge.pre-build"
	postBuildHook = "cartridge.post-build"
)

func BuildProject(projectCtx project.ProjectCtx) error {
	log.Infof("Building application in %s...", projectCtx.Path)

	var err error

	projectCtx.BuildDir = projectCtx.Path

	if projectCtx.BuildInDocker {
		log.Fatal("Not implemented yet")
	} else {
		err = buildProjectLocally(projectCtx)
	}

	if err != nil {
		return err
	}

	log.Infof("Application build succeeded")

	return nil
}
