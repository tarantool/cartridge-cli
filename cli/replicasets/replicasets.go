package replicasets

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/tarantool/cartridge-cli/cli/context"
	"github.com/tarantool/cartridge-cli/cli/project"
)

const (
	defaultReplicasetsFile = "replicasets.yml"
	instancesFile          = "instances.yml"
)

func FillCtx(ctx *context.Ctx) error {
	var err error

	if err := project.SetLocalRunningPaths(ctx); err != nil {
		return err
	}

	if ctx.Running.AppDir == "" {
		ctx.Running.AppDir, err = os.Getwd()
		if err != nil {
			return fmt.Errorf("Failed to get current directory: %s", err)
		}
	}

	if ctx.Running.AppDir, err = filepath.Abs(ctx.Running.AppDir); err != nil {
		return fmt.Errorf("Failed to get application directory absolute path: %s", err)
	}

	if ctx.Project.Name == "" {
		if ctx.Project.Name, err = project.DetectName(ctx.Running.AppDir); err != nil {
			return fmt.Errorf(
				"Failed to detect application name: %s. Please pass it explicitly via --name",
				err,
			)
		}
	}

	return nil
}
