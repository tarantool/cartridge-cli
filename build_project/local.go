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

func buildProjectLocally(projectCtx project.ProjectCtx) error {
	// pre-build
	preBuildHookPath := filepath.Join(projectCtx.BuildDir, preBuildHook)

	if _, err := os.Stat(preBuildHookPath); !os.IsNotExist(err) {
		err = runHook(preBuildHookPath, projectCtx.BuildDir)
		if err != nil {
			return fmt.Errorf("Failed to run pre-build hook: %s", err)
		}
	}

	// tarantoolctl rocks make
	if rockspec, err := common.FindRockspec(projectCtx.BuildDir); err != nil {
		return err
	} else if rockspec != "" {
		rocksMakeCmd := exec.Command("tarantoolctl", "rocks", "make")
		rocksMakeCmd.Stdout = os.Stdout
		rocksMakeCmd.Stderr = os.Stderr
		rocksMakeCmd.Dir = projectCtx.BuildDir

		err := rocksMakeCmd.Run()
		if err != nil {
			return fmt.Errorf("Failed to install rocks: %s", err)
		}
	}

	return nil
}

func runHook(hookPath, cwd string) error {
	hookName := filepath.Base(hookPath)

	if isExec, err := common.IsExecOwner(hookPath); err != nil {
		return fmt.Errorf("Failed go check hook file %q: %s", hookName, err)
	} else if !isExec {
		return fmt.Errorf("Hook %q should be executable", hookName)
	}

	log.Infof("Running %q", hookName)

	hookCmd := exec.Command(hookPath)
	hookCmd.Stdout = os.Stdout
	hookCmd.Stderr = os.Stderr
	hookCmd.Dir = cwd

	err := hookCmd.Run()
	if err != nil {
		return fmt.Errorf("Failed to run hook %q: %s", hookName, err)
	}

	return nil
}
