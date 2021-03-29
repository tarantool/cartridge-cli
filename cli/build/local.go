package build

import (
	"fmt"
	"strings"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/apex/log"

	"github.com/tarantool/cartridge-cli/cli/common"
	"github.com/tarantool/cartridge-cli/cli/context"
)

func buildProjectLocally(ctx *context.Ctx) error {
	if err := common.CheckTarantoolBinaries(); err != nil {
		return fmt.Errorf("Tarantool binaries are required for local build: %s", err)
	}
	common.CheckRecommendedBinaries("cmake", "make", "git", "unzip", "gcc")

	// pre-build
	preBuildHookPath := filepath.Join(ctx.Build.Dir, preBuildHookName)

	if _, err := os.Stat(preBuildHookPath); err == nil {
		log.Infof("Running `%s`", preBuildHookName)
		err = common.RunHook(preBuildHookPath, ctx.Cli.Verbose)
		if err != nil {
			return fmt.Errorf("Failed to run pre-build hook: %s", err)
		}
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("Unable to use pre-build hook: %s", err)
	}

	// tarantoolctl rocks make
	rocksMakeCmdParts := []string{"tarantoolctl", "rocks", "make"}
	if ctx.Build.Spec != "" {
		rocksMakeCmdParts = append(rocksMakeCmdParts, ctx.Build.Spec)
	}

	log.Infof("Running `%s`", strings.Join(rocksMakeCmdParts, " "))
	rocksMakeCmd := exec.Command(rocksMakeCmdParts[0], rocksMakeCmdParts[1:]...)
	err := common.RunCommand(rocksMakeCmd, ctx.Build.Dir, ctx.Cli.Verbose)
	if err != nil {
		return fmt.Errorf("Failed to install rocks: %s", err)
	}

	return nil
}
