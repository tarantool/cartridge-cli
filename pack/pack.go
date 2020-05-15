package pack

import (
	"fmt"
	"os/exec"

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
	packer, found := packers[projectCtx.PackType]
	if !found {
		return fmt.Errorf("Unsupported distribution type: %s", projectCtx.PackType)
	}

	checkPackRecommendedBinaries()

	projectCtx.BuildID = common.RandomString(10)

	// get and normalize version
	if err := detectVersion(projectCtx); err != nil {
		return err
	}

	// build directory
	if err := detectTmpDir(projectCtx); err != nil {
		return err
	}

	log.Infof("Tmp directory is set to: %s\n", projectCtx.TmpDir)

	if err := initTmpDir(projectCtx); err != nil {
		return err
	}

	defer removeTmpDir(projectCtx)

	// call packer
	log.Infof("Packing %s into %s", projectCtx.Name, projectCtx.PackType)

	if err := packer(projectCtx); err != nil {
		return err
	}

	log.Infof("Application succeessfully packed")

	return nil
}

func checkPackRecommendedBinaries() {
	var recommendedBinaries = []string{
		"git",
	}

	// check recommended binaries
	for _, binary := range recommendedBinaries {
		if _, err := exec.LookPath(binary); err != nil {
			log.Warnf("%s binary is recommended to pack application", binary)
		}
	}
}
