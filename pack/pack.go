package pack

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/tarantool/cartridge-cli/common"

	"github.com/tarantool/cartridge-cli/project"
)

var (
	packers = map[string]func(*project.ProjectCtx) error{
		tgzType: packTgz,
	}
)

const (
	tgzType = "tgz"
	rpmType = "rpm"
	debType = "deb"
)

func Run(projectCtx *project.ProjectCtx) error {
	_, found := packers[projectCtx.PackType]
	if !found {
		return fmt.Errorf("Unsupported distribution type: %s", projectCtx.PackType)
	}

	log.Infof("Packing %s into %s", projectCtx.Name, projectCtx.PackType)

	projectCtx.BuildID = common.RandomString(10)

	// get and normalize version
	if err := detectVersion(projectCtx); err != nil {
		return err
	}

	// build directory
	if err := detectBuildDir(projectCtx); err != nil {
		return err
	}

	log.Infof("Build directory is set to: %s\n", projectCtx.BuildDir)

	if err := initBuildDir(projectCtx); err != nil {
		return err
	}

	defer removeBuildDir(projectCtx)

	log.Infof("Application succeessfully packed")

	return nil
}
