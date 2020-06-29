package build

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/otiai10/copy"
	log "github.com/sirupsen/logrus"

	"github.com/tarantool/cartridge-cli/src/common"
	"github.com/tarantool/cartridge-cli/src/docker"
	"github.com/tarantool/cartridge-cli/src/project"
	"github.com/tarantool/cartridge-cli/src/templates"
)

type buildContext struct {
	UserID           string
	BuildID          string
	PreBuildHookName string
}

func buildProjectInDocker(projectCtx *project.ProjectCtx) error {
	var err error

	if err := docker.CheckMinServerVersion(); err != nil {
		return err
	}

	if projectCtx.BuildFrom != "" {
		if err := project.CheckBaseDockerfile(projectCtx.BuildFrom); err != nil {
			return fmt.Errorf("Invalid base build Dockerfile %s: %s", projectCtx.BuildFrom, err)
		}
	}

	if projectCtx.TarantoolIsEnterprise {
		if err := checkSDKPath(projectCtx.SDKPath); err != nil {
			return fmt.Errorf("Unable to use specified SDK path: %s", err)
		}
	}

	if projectCtx.TarantoolIsEnterprise {
		// Tarantool SDK is copied to BuildDir to be used on docker build
		// It's copied to the container by BuildSDKDirname
		// All used files should be in docker context dir (BuildDir)
		buildSDKPath := filepath.Join(
			projectCtx.BuildDir,
			projectCtx.BuildSDKDirname,
		)

		if err := copy.Copy(projectCtx.SDKPath, buildSDKPath); err != nil {
			return err
		}
		defer project.RemoveTmpPath(buildSDKPath, projectCtx.Debug)
	}

	// fill build context
	userID, err := common.GetCurrentUserID()
	if err != nil {
		return fmt.Errorf("Failed to get current user ID: %s", err)
	}

	ctx := buildContext{
		BuildID:          projectCtx.BuildID,
		UserID:           userID,
		PreBuildHookName: preBuildHookName,
	}

	// create build image Dockerfile
	log.Debugf("Create build image Dockerfile")

	buildImageDockerfileName := fmt.Sprintf("Dockerfile.build.%s", projectCtx.BuildID)
	dockerfileTemplate, err := project.GetBuildImageDockerfileTemplate(projectCtx)

	if err != nil {
		return fmt.Errorf("Failed to create build image Dockerfile: %s", err)
	}

	dockerfileTemplate.Path = buildImageDockerfileName

	if err := dockerfileTemplate.Instantiate(projectCtx.BuildDir, ctx); err != nil {
		return fmt.Errorf("Failed to create build image Dockerfile: %s", err)
	}
	defer project.RemoveTmpPath(
		filepath.Join(projectCtx.BuildDir, buildImageDockerfileName),
		projectCtx.Debug,
	)

	// create build image
	buildImageTag := fmt.Sprintf("%s-build", projectCtx.Name)
	log.Infof("Building base image: %s", buildImageTag)

	err = docker.BuildImage(docker.BuildOpts{
		Tag:        buildImageTag,
		Dockerfile: buildImageDockerfileName,
		BuildDir:   projectCtx.BuildDir,
		TmpDir:     projectCtx.TmpDir,
		Quiet:      projectCtx.Quiet,
	})

	if err != nil {
		return fmt.Errorf("Failed to build base image: %s", err)
	}

	// create build script
	log.Debugf("Create build script")
	buildScriptName := fmt.Sprintf("build.%s.sh", projectCtx.BuildID)

	buildScriptTemplate := getBuildScriptTemplate(projectCtx)
	buildScriptTemplate.Path = buildScriptName
	if err := buildScriptTemplate.Instantiate(projectCtx.BuildDir, ctx); err != nil {
		return fmt.Errorf("Failed to create build script: %s", err)
	}
	defer project.RemoveTmpPath(
		filepath.Join(projectCtx.BuildDir, buildScriptName),
		projectCtx.Debug,
	)

	// run build script on image
	log.Infof("Building application in %s", buildImageTag)

	err = docker.RunContainer(docker.RunOpts{
		ImageTag:   buildImageTag,
		WorkingDir: containerBuildDir,
		Cmd:        []string{fmt.Sprintf("./%s", buildScriptName)},

		Volumes: map[string]string{
			projectCtx.BuildDir: containerBuildDir,
		},

		Quiet: projectCtx.Quiet,
		Debug: projectCtx.Debug,
	})

	if err != nil {
		return fmt.Errorf("Failed to build application: %s", err)
	}

	return nil
}

func checkSDKPath(path string) error {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return err
	}

	if !fileInfo.IsDir() {
		return fmt.Errorf("Is not a directory")
	}

	for _, binary := range []string{"tarantool", "tarantoolctl"} {
		binaryPath := filepath.Join(path, binary)
		if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
			return fmt.Errorf("Does not contain %s binary", binary)
		} else if err != nil {
			return fmt.Errorf("Unable to use %s binary: %s", binary, err)
		}

		if isExec, err := common.IsExecOwner(binaryPath); err != nil {
			return fmt.Errorf("Unable to use %s binary: %s", binary, err)
		} else if !isExec {
			return fmt.Errorf("%s binary is not executable", binary)
		}
	}

	return nil
}

func getBuildScriptTemplate(projectCtx *project.ProjectCtx) *templates.FileTemplate {
	template := templates.FileTemplate{
		Mode:    0755,
		Content: buildScriptContent,
	}

	return &template
}

const (
	containerBuildDir  = "/opt/tarantool"
	buildScriptContent = `#!/bin/bash
set -xe

if [ -f {{ .PreBuildHookName }} ]; then
    . {{ .PreBuildHookName }}
fi

tarantoolctl rocks make
`

	sdkPathError = `For packing in docker you should specify one of:
* --sdk-local: to use local SDK;;
* --sdk-path: path to SDK
  (can be passed in environment variable TARANTOOL_SDK_PATH).`
)
