package pack

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/apex/log"

	"github.com/tarantool/cartridge-cli/cli/common"
	"github.com/tarantool/cartridge-cli/cli/context"
	"github.com/tarantool/cartridge-cli/cli/project"
)

var (
	packers = map[string]func(*context.Ctx) error{
		TgzType:    packTgz,
		DebType:    packDeb,
		RpmType:    packRpm,
		DockerType: packDocker,
	}
)

const (
	TgzType    = "tgz"
	RpmType    = "rpm"
	DebType    = "deb"
	DockerType = "docker"
)

// Run packs application into project.PackType distributable
func Run(ctx *context.Ctx) error {
	if err := checkCtx(ctx); err != nil {
		return project.InternalError("Pack context check failed: %s", err)
	}

	if !ctx.Build.InDocker && (ctx.Pack.Type == RpmType || ctx.Pack.Type == DebType) {
		if runtime.GOOS != "linux" {
			return fmt.Errorf(
				"It's not possible to pack application into RPM or DEB on non-linux OS (%s). "+
					"Please, use --use-docker flag to pack application inside the Docker container",
				runtime.GOOS,
			)
		}
	}

	// get packer function
	packer, found := packers[ctx.Pack.Type]
	if !found {
		return fmt.Errorf("Unsupported distribution type: %s", ctx.Pack.Type)
	}

	log.Infof("Packing %s into %s", ctx.Project.Name, ctx.Pack.Type)

	// All types except TGZ pack require init.lua in the project root
	// because project from TGZ can be started using `cartridge start` command
	// that has `--script` option, but all other types use `tarantool init.lua`
	// command to define application start command
	if ctx.Pack.Type != TgzType {
		entrypointPath := filepath.Join(ctx.Project.Path, ctx.Running.Entrypoint)
		if _, err := os.Stat(entrypointPath); os.IsNotExist(err) {
			return fmt.Errorf("Application doesn't contain entrypoint script %s", ctx.Running.Entrypoint)
		} else if err != nil {
			return fmt.Errorf("Can't use application entrypoint script: %s", err)
		}
	}

	ctx.Pack.ID = common.RandomString(10)
	ctx.Build.ID = ctx.Pack.ID

	if ctx.Pack.Type == DockerType {
		ctx.Build.InDocker = true
	}

	// set ctx.Build.SDKPath and ctx.Build.BuildSDKDirname
	if ctx.Tarantool.TarantoolIsEnterprise {
		if err := setSDKPath(ctx); err != nil {
			return err
		}

		ctx.Build.BuildSDKDirname = fmt.Sprintf("sdk-%s", ctx.Pack.ID)
	}

	// set base Dockerfiles
	if ctx.Build.InDocker {
		if ctx.Build.DockerFrom == "" {
			// build Dockerfile
			defaultBaseBuildDockerfilePath := filepath.Join(ctx.Project.Path, project.DefaultBaseBuildDockerfile)
			if _, err := os.Stat(defaultBaseBuildDockerfilePath); err == nil {
				log.Debugf("Default build Dockerfile is used: %s", defaultBaseBuildDockerfilePath)

				ctx.Build.DockerFrom = defaultBaseBuildDockerfilePath
			} else if !os.IsNotExist(err) {
				return fmt.Errorf("Failed to use default build Dockerfile: %s", err)
			}
		}
		if ctx.Pack.DockerFrom == "" {
			// runtime Dockerfile
			defaultBaseRuntimeDockerfilePath := filepath.Join(ctx.Project.Path, project.DefaultBaseRuntimeDockerfile)
			if _, err := os.Stat(defaultBaseRuntimeDockerfilePath); err == nil {
				log.Debugf("Default runtime Dockerfile is used: %s", defaultBaseRuntimeDockerfilePath)

				ctx.Pack.DockerFrom = defaultBaseRuntimeDockerfilePath
			} else if !os.IsNotExist(err) {
				return fmt.Errorf("Failed to use default runtime Dockerfile: %s", err)
			}
		}
	}

	if _, err := os.Stat(ctx.Project.Path); err != nil {
		return fmt.Errorf("Bad path is specified: %s", err)
	}

	// check that user specified only --version,--suffix or --tag
	if err := checkTagVersionSuffix(ctx); err != nil {
		return err
	}

	// get and normalize version
	if ctx.Pack.Type != DockerType || len(ctx.Pack.ImageTags) == 0 {
		if err := detectVersion(ctx); err != nil {
			return err
		}
	}

	// check if app has stateboard entrypoint
	stateboardEntrypointPath := filepath.Join(ctx.Project.Path, ctx.Running.StateboardEntrypoint)
	if _, err := os.Stat(stateboardEntrypointPath); err == nil {
		ctx.Running.WithStateboard = true
	} else if os.IsNotExist(err) {
		ctx.Running.WithStateboard = false
	} else {
		return fmt.Errorf("Failed to get stateboard entrypoint stat: %s", err)
	}

	if ctx.Pack.Type != DockerType {
		// set result package path
		curDir, err := os.Getwd()
		if err != nil {
			return err
		}
		ctx.Pack.ResPackagePath = filepath.Join(curDir, getPackageFullname(ctx))
	} else {
		// set result image fullname
		ctx.Pack.ResImageTags = getImageTags(ctx)
	}

	// tmp directory
	if err := detectTmpDir(ctx); err != nil {
		return err
	}

	log.Infof("Temporary directory is set to %s", ctx.Cli.TmpDir)
	if err := initTmpDir(ctx); err != nil {
		return err
	}
	defer project.RemoveTmpPath(ctx.Cli.TmpDir, ctx.Cli.Debug)

	if err := packer(ctx); err != nil {
		return err
	}

	log.Infof("Application was successfully packed")

	return nil
}

func FillCtx(ctx *context.Ctx) error {
	var err error

	if err := project.SetProjectPath(ctx); err != nil {
		return fmt.Errorf("Failed to set project path: %s", err)
	}

	if ctx.Project.Name == "" {
		ctx.Project.Name, err = project.DetectName(ctx.Project.Path)
		if err != nil {
			return fmt.Errorf(
				"Failed to detect application name: %s. Please pass it explicitly via --name",
				err,
			)
		}
	}

	ctx.Project.StateboardName = project.GetStateboardName(ctx)

	if err := project.FillTarantoolCtx(ctx); err != nil {
		return fmt.Errorf("Failed to get Tarantool context: %s", err)
	}

	if err := project.SetSystemRunningPaths(ctx); err != nil {
		return err
	}

	sdkPathFromEnv := os.Getenv(sdkPathEnv)
	if ctx.Tarantool.TarantoolIsEnterprise && (ctx.Pack.Type == DockerType || ctx.Build.InDocker) {
		if ctx.Build.SDKPath == "" {
			ctx.Build.SDKPath = sdkPathFromEnv
		}
		if !common.OnlyOneIsTrue(ctx.Build.SDKPath != "", ctx.Build.SDKLocal) {
			return fmt.Errorf(sdkPathError)
		}
	} else if sdkPathFromEnv != "" {
		log.Warnf("Specified %s is ignored", sdkPathEnv)
	}

	if err :=  fillPreAndPostInstallScripts(ctx); err != nil {
		return err
	}

	return nil
}

func checkCtx(ctx *context.Ctx) error {
	if ctx.Project.Name == "" {
		return fmt.Errorf("Name is missed")
	}

	if ctx.Project.Path == "" {
		return fmt.Errorf("Path is missed")
	}

	if ctx.Pack.Type == "" {
		return fmt.Errorf("PackType is missed")
	}

	if ctx.Tarantool.TarantoolIsEnterprise {
		if !ctx.Build.InDocker && ctx.Tarantool.TarantoolDir == "" {
			return fmt.Errorf("TarantoolDir is missed")
		}
	} else {
		if ctx.Tarantool.TarantoolVersion == "" {
			return fmt.Errorf("TarantoolVersion is missed")
		}
	}

	return nil
}

func setSDKPath(ctx *context.Ctx) error {
	if !ctx.Build.InDocker {
		ctx.Build.SDKPath = ctx.Tarantool.TarantoolDir
	} else if ctx.Build.SDKLocal {
		ctx.Build.SDKPath = ctx.Tarantool.TarantoolDir
	}

	return nil
}

func fillScript(filename string, packType string, outputScript *string) error {
	if packType == RpmType || packType == DebType {
		var err error

		if filename != "" {
			if _, err = os.Stat(filename); os.IsNotExist(err) {
				return fmt.Errorf("Specified script %s doesn't exists", filename)
			} else if err != nil {
				return fmt.Errorf("Impossible to use specified script %s: %s", filename, err)
			}

			*outputScript, err = common.GetFileContent(filename)
			if err != nil {
				return fmt.Errorf("Failed to get file content: %s", err)
			}
		}

		return nil
	}

	if filename != "" {
		log.Warnf("You specified flag for pre/post install script, but you are not packaging RPM or DEB. "+
			"Flag will be ignored")
	}

	return nil
}

func fillPreAndPostInstallScripts(ctx *context.Ctx) error {
	if err := fillScript(ctx.Pack.PreInstallScriptFile, ctx.Pack.Type, &ctx.Pack.PreInstallScript); err != nil {
		return fmt.Errorf("Failed to use specified pre-install script: %s", err)
	}

	if err := fillScript(ctx.Pack.PostInstallScriptFile, ctx.Pack.Type, &ctx.Pack.PostInstallScript); err != nil {
		return fmt.Errorf("Failed to use specified post-install script: %s", err)
	}

	return nil
}

const (
	sdkPathEnv   = `TARANTOOL_SDK_PATH`
	sdkPathError = `For packing in docker you should specify one of:
* --sdk-local: to use local SDK
* --sdk-path: path to SDK
	(can be passed in environment variable TARANTOOL_SDK_PATH)`
)
