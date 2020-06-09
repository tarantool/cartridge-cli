package pack

import (
	"fmt"
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"

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
	// check context
	if err := checkCtx(projectCtx); err != nil {
		// TODO: format internal error
		panic(err)
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

	// get packer function
	packer, found := packers[projectCtx.PackType]
	if !found {
		return fmt.Errorf("Unsupported distribution type: %s", projectCtx.PackType)
	}

	if _, err := os.Stat(projectCtx.Path); err != nil {
		return fmt.Errorf("Failed to use path %s: %s", projectCtx.Path, err)
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

	log.Infof("Temporary directory is set to %s\n", projectCtx.TmpDir)
	if err := initTmpDir(projectCtx); err != nil {
		return err
	}
	defer project.RemoveTmpPath(projectCtx.TmpDir, projectCtx.Debug)

	// call packer
	log.Infof("Packing %s into %s", projectCtx.Name, projectCtx.PackType)

	if err := packer(projectCtx); err != nil {
		return err
	}

	log.Infof("Application succeessfully packed")

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
	} else {
		if !common.OnlyOneIsTrue(projectCtx.SDKPath != "", projectCtx.SDKLocal) {
			return fmt.Errorf(sdkPathError)
		}

		if projectCtx.SDKLocal {
			projectCtx.SDKPath = projectCtx.TarantoolDir
		}
	}

	return nil
}

const (
	sdkPathError = `For packing in docker you should specify one of:
	* --sdk-local: to use local SDK
	* --sdk-path: path to SDK
	  (can be passed in environment variable TARANTOOL_SDK_PATH)`
)
