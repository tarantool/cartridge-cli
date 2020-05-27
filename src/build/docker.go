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
	buildImageDockerfileName := fmt.Sprintf("Dockerfile.build.%s", projectCtx.BuildID)
	dockerfileTemplate, err := getBuildImageDockerfileTemplate(projectCtx)

	if err != nil {
		return fmt.Errorf("Failed to create build image Dockerfile: %s", err)
	}

	dockerfileTemplate.Path = buildImageDockerfileName

	if err := dockerfileTemplate.Instantiate(projectCtx.BuildDir, ctx); err != nil {
		return fmt.Errorf("Failed to create build image Dockerfile: %s", err)
	}
	defer removePath(
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
	buildScriptName := fmt.Sprintf("build.%s.sh", projectCtx.BuildID)

	buildScriptTemplate := getBuildScriptTemplate(projectCtx)
	buildScriptTemplate.Path = buildScriptName
	if err := buildScriptTemplate.Instantiate(projectCtx.BuildDir, ctx); err != nil {
		return fmt.Errorf("Failed to create build script: %s", err)
	}

	defer removePath(
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

	// copy tarantool binaries to build dir
	if projectCtx.BuildInDocker && projectCtx.TarantoolIsEnterprise {
		for _, binary := range []string{"tarantool", "tarantoolctl"} {
			binaryPath := filepath.Join(projectCtx.BuildSDKPath, binary)
			destBinaryPath := filepath.Join(projectCtx.BuildDir, binary)
			if err := copy.Copy(binaryPath, destBinaryPath); err != nil {
				return fmt.Errorf("Failed to copy %s binary: %s", binary, err)
			}
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

func removePath(path string, debug bool) {
	if debug {
		log.Warnf("%s is not removed due to debug mode", path)
		return
	}
	if err := os.Remove(path); err != nil {
		log.Warnf("Failed to remove: %s", err)
	}
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
)
