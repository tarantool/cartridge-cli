package pack

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

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

// Run packs application into project.PackType distributable
func Run(projectCtx *project.ProjectCtx) error {
	packer, found := packers[projectCtx.PackType]
	if !found {
		return fmt.Errorf("Unsupported distribution type: %s", projectCtx.PackType)
	}

	if _, err := os.Stat(projectCtx.Path); err != nil {
		return fmt.Errorf("Failed to use path %s: %s", projectCtx.Path, err)
	}

	checkPackRecommendedBinaries()

	projectCtx.BuildID = common.RandomString(10)

	// get and normalize version
	if err := detectVersion(projectCtx); err != nil {
		return err
	}

	curDir, err := os.Getwd()
	if err != nil {
		return err
	}
	projectCtx.ResPackagePath = filepath.Join(curDir, getPackageFullname(projectCtx))

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
