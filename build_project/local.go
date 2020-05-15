package build

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	log "github.com/sirupsen/logrus"

	"github.com/tarantool/cartridge-cli/common"
	"github.com/tarantool/cartridge-cli/project"
)

func buildProjectLocally(projectCtx *project.ProjectCtx) error {
	checkLocalBuildRecommendedBinaries()

	// pre-build
	preBuildHookPath := filepath.Join(projectCtx.BuildDir, preBuildHookName)

	if _, err := os.Stat(preBuildHookPath); !os.IsNotExist(err) {
		log.Infof("Running `%s`", preBuildHookName)
		err = common.RunHook(preBuildHookPath, !projectCtx.Quiet)
		if err != nil {
			return fmt.Errorf("Failed to run pre-build hook: %s", err)
		}
	}

	// tarantoolctl rocks make
	log.Infof("Running `tarantoolctl rocks make`")
	rocksMakeCmd := exec.Command("tarantoolctl", "rocks", "make")
	err := common.RunCommand(rocksMakeCmd, projectCtx.BuildDir, !projectCtx.Quiet)
	if err != nil {
		return fmt.Errorf("Failed to install rocks: %s", err)
	}

	return nil
}

func checkLocalBuildRecommendedBinaries() {
	var recommendedBinaries = []string{
		"cmake",
		"make",
		"git",
		"unzip",
		"gcc",
	}

	// check recommended binaries
	for _, binary := range recommendedBinaries {
		if _, err := exec.LookPath(binary); err != nil {
			log.Warnf("%s binary is recommended to build application locally", binary)
		}
	}
}
