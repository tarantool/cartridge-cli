package build

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/apex/log"

	"github.com/tarantool/cartridge-cli/cli/common"
	"github.com/tarantool/cartridge-cli/cli/context"
	"github.com/tarantool/cartridge-cli/cli/project"
)

const (
	preBuildHookName  = "cartridge.pre-build"
	postBuildHookName = "cartridge.post-build"
)

// Run builds project in ctx.Build.Dir
// If ctx.Build.InDocker is set, application is built in docker
func Run(ctx *context.Ctx) error {
	if ctx.Build.ID == "" {
		ctx.Build.ID = common.RandomString(10)
	}

	// check context
	if err := checkCtx(ctx); err != nil {
		return project.InternalError("Build context check failed: %s", err)
	}

	if fileInfo, err := os.Stat(ctx.Build.Dir); err != nil {
		return fmt.Errorf("Unable to build application in %s: %s", ctx.Build.Dir, err)
	} else if !fileInfo.IsDir() {
		return fmt.Errorf("Unable to build application in %s: it's not a directory", ctx.Build.Dir)
	}

	log.Infof("Build application in %s", ctx.Build.Dir)

	if ctx.Build.Spec == "" {
		// check that application directory contains rockspec
		if rockspecPath, err := common.FindRockspec(ctx.Project.Path); err != nil {
			return fmt.Errorf("Unable to build application: %s", err)
		} else if rockspecPath == "" {
			return fmt.Errorf("Application directory should contain rockspec")
		}
	} else {

		// check that specified rockspec is in a project root
		if absoluteRockspecPath, err := filepath.Abs(ctx.Build.Spec); err != nil {
			return fmt.Errorf("Can't build absolute path fot rockspec %s", ctx.Build.Spec)
		} else if filepath.Dir(absoluteRockspecPath) != ctx.Project.Path {
			return fmt.Errorf("Rockspec %s should be in project root", ctx.Build.Spec)
		}

		// check that specified file is valid
		if fileInfo, err := os.Stat(ctx.Build.Spec); os.IsNotExist(err) {
			return fmt.Errorf("Rockspec %s doesn't exist", ctx.Build.Spec)
		} else if err != nil {
			return fmt.Errorf("Unable to use rockspec %s: %s", ctx.Build.Spec, err)
		} else if fileInfo.IsDir() {
			return fmt.Errorf("Unable to use rockspec %s: it is a directory", ctx.Build.Spec)
		}

		ctx.Build.Spec = filepath.Base(ctx.Build.Spec)

		if common.ContainsUpperSymbols(ctx.Build.Spec) {
			return fmt.Errorf("Please name the rockspec file in lowercase")
		}
	}

	if ctx.Build.InDocker {
		if err := buildProjectInDocker(ctx); err != nil {
			return err
		}
	} else {
		if err := buildProjectLocally(ctx); err != nil {
			return err
		}
	}

	log.Infof("Application was successfully built")

	return nil
}

func FillCtx(ctx *context.Ctx) error {
	if err := project.SetProjectPath(ctx); err != nil {
		return fmt.Errorf("Failed to set project path: %s", err)
	}

	ctx.Build.Dir = ctx.Project.Path

	return nil
}

func checkCtx(ctx *context.Ctx) error {
	if ctx.Build.Dir == "" {
		return fmt.Errorf("BuildDir is missed")
	}

	if ctx.Build.ID == "" {
		return fmt.Errorf("BuildID is missed")
	}

	if ctx.Build.InDocker {
		if ctx.Project.Name == "" {
			return fmt.Errorf("Name is missed")
		}

		if ctx.Cli.TmpDir == "" {
			return fmt.Errorf("TmpDir is missed")
		}

		if ctx.Tarantool.TarantoolIsEnterprise {
			if ctx.Build.SDKPath == "" {
				return fmt.Errorf("SDKPath is missed")
			}

			if ctx.Build.BuildSDKDirname == "" {
				return fmt.Errorf("BuildSDKDirname is missed")
			}
		} else {
			if ctx.Tarantool.TarantoolVersion == "" {
				return fmt.Errorf("TarantoolVersion is missed")
			}
		}
	}

	return nil
}
