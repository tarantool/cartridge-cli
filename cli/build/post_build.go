package build

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/apex/log"
	"github.com/tarantool/cartridge-cli/cli/common"
	"github.com/tarantool/cartridge-cli/cli/project"
)

func PostRun(projectCtx *project.ProjectCtx) error {
	// post-build
	postBuildHookPath := filepath.Join(projectCtx.BuildDir, postBuildHookName)

	if _, err := os.Stat(postBuildHookPath); !os.IsNotExist(err) {
		log.Infof("Running `%s`", postBuildHookName)
		err = common.RunHook(postBuildHookPath, !projectCtx.Quiet)
		if err != nil {
			return fmt.Errorf("Failed to run post-build hook: %s", err)
		}
	} else if err != nil {
		return fmt.Errorf("Unable to use post-build hook: %s", err)
	}

	var buildHooks = []string{
		preBuildHookName,
		postBuildHookName,
	}

	for _, hook := range buildHooks {
		log.Debugf("Remove `%s`", hook)

		hookPath := filepath.Join(projectCtx.BuildDir, hook)
		if err := os.RemoveAll(hookPath); err != nil {
			return fmt.Errorf("Failed to remove %s: %s", hookPath, err)
		}
	}

	return nil
}
