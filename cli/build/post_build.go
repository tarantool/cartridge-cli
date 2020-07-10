package build

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/apex/log"
	"github.com/tarantool/cartridge-cli/cli/common"
	"github.com/tarantool/cartridge-cli/cli/context"
)

func PostRun(ctx *context.Ctx) error {
	// post-build
	postBuildHookPath := filepath.Join(ctx.Build.Dir, postBuildHookName)

	if _, err := os.Stat(postBuildHookPath); err == nil {
		log.Infof("Running `%s`", postBuildHookName)
		err = common.RunHook(postBuildHookPath, !ctx.Cli.Quiet)
		if err != nil {
			return fmt.Errorf("Failed to run post-build hook: %s", err)
		}
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("Unable to use post-build hook: %s", err)
	}

	var buildHooks = []string{
		preBuildHookName,
		postBuildHookName,
	}

	for _, hook := range buildHooks {
		log.Debugf("Remove `%s`", hook)

		hookPath := filepath.Join(ctx.Build.Dir, hook)
		if err := os.RemoveAll(hookPath); err != nil {
			return fmt.Errorf("Failed to remove %s: %s", hookPath, err)
		}
	}

	return nil
}
