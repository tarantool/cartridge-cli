package pack

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/apex/log"

	"github.com/tarantool/cartridge-cli/cli/common"
	"github.com/tarantool/cartridge-cli/cli/project"
)

var (
	packers = map[string]func(*project.ProjectCtx) error{
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
func Run(projectCtx *project.ProjectCtx) error {
	if err := checkCtx(projectCtx); err != nil {
		return project.InternalError("Pack context check failed: %s", err)
	}

	// get packer function
	packer, found := packers[projectCtx.PackType]
	if !found {
		return fmt.Errorf("Unsupported distribution type: %s", projectCtx.PackType)
	}

	log.Infof("Packing %s into %s", projectCtx.Name, projectCtx.PackType)

	// All types except TGZ pack require init.lua in the project root
	// because project from TGZ can be started using `cartridge start` command
	// that has `--script` option, but all other types use `tarantool init.lua`
	// command to define application start command
	if projectCtx.PackType != TgzType {
		entrypointPath := filepath.Join(projectCtx.Path, projectCtx.Entrypoint)
		if _, err := os.Stat(entrypointPath); os.IsNotExist(err) {
			return fmt.Errorf("Application doesn't contain entrypoint script %s", projectCtx.Entrypoint)
		} else if err != nil {
			return fmt.Errorf("Can't use application entrypoint script: %s", err)
		}
	}

	projectCtx.PackID = common.RandomString(10)
	projectCtx.BuildID = projectCtx.PackID

	if projectCtx.PackType == DockerType {
		projectCtx.BuildInDocker = true
	}

	// set projectCtx.SDKPath and projectCtx.BuildSDKDirname
	if projectCtx.TarantoolIsEnterprise {
		if err := setSDKPath(projectCtx); err != nil {
			return err
		}

		projectCtx.BuildSDKDirname = fmt.Sprintf("sdk-%s", projectCtx.PackID)
	}

	// set base Dockerfiles
	if projectCtx.BuildInDocker {
		if projectCtx.BuildFrom == "" {
			// build Dockerfile
			defaultBaseBuildDockerfilePath := filepath.Join(projectCtx.Path, project.DefaultBaseBuildDockerfile)
			if _, err := os.Stat(defaultBaseBuildDockerfilePath); err == nil {
				log.Debugf("Default build Dockerfile is used: %s", defaultBaseBuildDockerfilePath)

				projectCtx.BuildFrom = defaultBaseBuildDockerfilePath
			} else if !os.IsNotExist(err) {
				return fmt.Errorf("Failed to use default build Dockerfile: %s", err)
			}
		}
		if projectCtx.From == "" {
			// runtime Dockerfile
			defaultBaseRuntimeDockerfilePath := filepath.Join(projectCtx.Path, project.DefaultBaseRuntimeDockerfile)
			if _, err := os.Stat(defaultBaseRuntimeDockerfilePath); err == nil {
				log.Debugf("Default runtime Dockerfile is used: %s", defaultBaseRuntimeDockerfilePath)

				projectCtx.From = defaultBaseRuntimeDockerfilePath
			} else if !os.IsNotExist(err) {
				return fmt.Errorf("Failed to use default runtime Dockerfile: %s", err)
			}
		}
	}

	if _, err := os.Stat(projectCtx.Path); err != nil {
		return fmt.Errorf("Bad path is specified: %s", err)
	}

	// check that user specified only --version,--suffix or --tag
	if err := checkTagVersionSuffix(projectCtx); err != nil {
		return err
	}

	// get and normalize version
	if projectCtx.PackType != DockerType || len(projectCtx.ImageTags) == 0 {
		if err := detectVersion(projectCtx); err != nil {
			return err
		}
	}

	// check if app has stateboard entrypoint
	stateboardEntrypointPath := filepath.Join(projectCtx.Path, projectCtx.StateboardEntrypoint)
	if _, err := os.Stat(stateboardEntrypointPath); err == nil {
		projectCtx.WithStateboard = true
	} else if os.IsNotExist(err) {
		projectCtx.WithStateboard = false
	} else {
		return fmt.Errorf("Failed to get stateboard entrypoint stat: %s", err)
	}

	if projectCtx.PackType != DockerType {
		// set result package path
		curDir, err := os.Getwd()
		if err != nil {
			return err
		}
		projectCtx.ResPackagePath = filepath.Join(curDir, getPackageFullname(projectCtx))
	} else {
		// set result image fullname
		projectCtx.ResImageTags = getImageTags(projectCtx)
	}

	// tmp directory
	if err := detectTmpDir(projectCtx); err != nil {
		return err
	}

	log.Infof("Temporary directory is set to %s", projectCtx.TmpDir)
	if err := initTmpDir(projectCtx); err != nil {
		return err
	}
	defer project.RemoveTmpPath(projectCtx.TmpDir, projectCtx.Debug)

	if err := packer(projectCtx); err != nil {
		return err
	}

	log.Infof("Application was successfully packed")

	return nil
}

func FillCtx(projectCtx *project.ProjectCtx) error {
	var err error

	if err := project.SetProjectPath(projectCtx); err != nil {
		return fmt.Errorf("Failed to set project path: %s", err)
	}

	if projectCtx.Name == "" {
		projectCtx.Name, err = project.DetectName(projectCtx.Path)
		if err != nil {
			return fmt.Errorf(
				"Failed to detect application name: %s. Please pass it explicitly via --name ",
				err,
			)
		}
	}

	projectCtx.StateboardName = project.GetStateboardName(projectCtx)

	if err := project.FillTarantoolCtx(projectCtx); err != nil {
		return fmt.Errorf("Failed to get Tarantool context: %s", err)
	}

	if err := project.SetSystemRunningPaths(projectCtx); err != nil {
		return err
	}

	sdkPathFromEnv := os.Getenv(sdkPathEnv)
	if projectCtx.TarantoolIsEnterprise && (projectCtx.PackType == DockerType || projectCtx.BuildInDocker) {
		if projectCtx.SDKPath == "" {
			projectCtx.SDKPath = sdkPathFromEnv
		}
		if !common.OnlyOneIsTrue(projectCtx.SDKPath != "", projectCtx.SDKLocal) {
			return fmt.Errorf(sdkPathError)
		}
	} else if sdkPathFromEnv != "" {
		log.Warnf("Specified %s is ignored", sdkPathEnv)
	}

	return nil
}

func checkCtx(projectCtx *project.ProjectCtx) error {
	if projectCtx.Name == "" {
		return fmt.Errorf("Name is missed")
	}

	if projectCtx.Path == "" {
		return fmt.Errorf("Path is missed")
	}

	if projectCtx.PackType == "" {
		return fmt.Errorf("PackType is missed")
	}

	if projectCtx.TarantoolIsEnterprise {
		if !projectCtx.BuildInDocker && projectCtx.TarantoolDir == "" {
			return fmt.Errorf("TarantoolDir is missed")
		}
	} else {
		if projectCtx.TarantoolVersion == "" {
			return fmt.Errorf("TarantoolVersion is missed")
		}
	}

	return nil
}

func setSDKPath(projectCtx *project.ProjectCtx) error {
	if !projectCtx.BuildInDocker {
		projectCtx.SDKPath = projectCtx.TarantoolDir
	} else if projectCtx.SDKLocal {
		projectCtx.SDKPath = projectCtx.TarantoolDir
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
