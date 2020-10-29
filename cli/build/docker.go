package build

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"github.com/apex/log"
	"github.com/otiai10/copy"

	"github.com/tarantool/cartridge-cli/cli/common"
	"github.com/tarantool/cartridge-cli/cli/context"
	"github.com/tarantool/cartridge-cli/cli/docker"
	"github.com/tarantool/cartridge-cli/cli/project"
	"github.com/tarantool/cartridge-cli/cli/templates"
)

func buildProjectInDocker(ctx *context.Ctx) error {
	var err error

	if err := docker.CheckMinServerVersion(); err != nil {
		return err
	}

	log.Debugf("Check specified base build Dockerfile")
	if ctx.Build.DockerFrom != "" {
		if err := project.CheckBaseDockerfile(ctx.Build.DockerFrom); err != nil {
			return fmt.Errorf("Invalid base build Dockerfile %s: %s", ctx.Build.DockerFrom, err)
		}
	}

	if ctx.Tarantool.IsEnterprise {
		log.Debugf("Check specified SDK")
		if err := checkSDKPath(ctx.Build.SDKPath); err != nil {
			return fmt.Errorf("Unable to use specified SDK: %s", err)
		}
	}

	if ctx.Tarantool.IsEnterprise {
		// Tarantool SDK is copied to BuildDir to be used on docker build
		// It's copied to the container by BuildSDKDirname
		// All used files should be in docker context dir (BuildDir)
		buildSDKPath := filepath.Join(
			ctx.Build.Dir,
			ctx.Build.BuildSDKDirname,
		)

		if err := copy.Copy(ctx.Build.SDKPath, buildSDKPath); err != nil {
			return err
		}
		defer project.RemoveTmpPath(buildSDKPath, ctx.Cli.Debug)
	}

	// fill build context
	userID, err := common.GetCurrentUserID()
	if err != nil {
		return fmt.Errorf("Failed to get current user ID: %s", err)
	}

	dockerBuildCtx := map[string]interface{}{
		"BuildID":          ctx.Build.ID,
		"UserID":           userID,
		"PreBuildHookName": preBuildHookName,
	}

	log.Debugf("Create build image Dockerfile")
	buildImageDockerfileName := fmt.Sprintf("Dockerfile.build.%s", ctx.Build.ID)
	dockerfileTemplate, err := project.GetBuildImageDockerfileTemplate(ctx)

	if err != nil {
		return fmt.Errorf("Failed to create build image Dockerfile: %s", err)
	}

	dockerfileTemplate.Path = buildImageDockerfileName
	if err := dockerfileTemplate.Instantiate(ctx.Build.Dir, dockerBuildCtx); err != nil {
		return fmt.Errorf("Failed to create build image Dockerfile: %s", err)
	}
	defer project.RemoveTmpPath(
		filepath.Join(ctx.Build.Dir, buildImageDockerfileName),
		ctx.Cli.Debug,
	)

	// create build image
	buildImageTag := fmt.Sprintf("%s-build", ctx.Project.Name)
	log.Infof("Building base image %s", buildImageTag)

	err = docker.BuildImage(docker.BuildOpts{
		Tag:        []string{buildImageTag},
		Dockerfile: buildImageDockerfileName,
		NoCache:    ctx.Docker.NoCache,
		CacheFrom:  ctx.Docker.CacheFrom,

		BuildDir: ctx.Build.Dir,
		TmpDir:   ctx.Cli.TmpDir,
		Quiet:    ctx.Cli.Quiet,
	})

	if err != nil {
		return fmt.Errorf("Failed to build base image: %s", err)
	}

	// create build script
	log.Debugf("Create build script")
	buildScriptName := fmt.Sprintf("build.%s.sh", ctx.Build.ID)

	buildScriptCtx := map[string]interface{}{
		"PreBuildHookName": preBuildHookName,
	}

	buildScriptTemplate := getBuildScriptTemplate(ctx)
	buildScriptTemplate.Path = buildScriptName
	if err := buildScriptTemplate.Instantiate(ctx.Build.Dir, buildScriptCtx); err != nil {
		return fmt.Errorf("Failed to create build script: %s", err)
	}
	defer project.RemoveTmpPath(
		filepath.Join(ctx.Build.Dir, buildScriptName),
		ctx.Cli.Debug,
	)

	// run build script on image
	log.Infof("Build application in %s", buildImageTag)

	err = docker.RunContainer(docker.RunOpts{
		ImageTags:  buildImageTag,
		WorkingDir: containerBuildDir,
		Cmd:        []string{fmt.Sprintf("./%s", buildScriptName)},

		Volumes: map[string]string{
			ctx.Build.Dir: containerBuildDir,
		},

		Quiet: ctx.Cli.Quiet,
		Debug: ctx.Cli.Debug,
	})

	if err != nil {
		if isErrCmdNotFound(err) {
			log.Errorf(
				"It's possible that docker volumes aren't working correctly for default build tmp directory. " +
					"Try to call `CARTRIDGE_TEMPDIR=. cartridge pack ...`",
			)
		}
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
			return fmt.Errorf("%s binary is missed", binary)
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

func getBuildScriptTemplate(ctx *context.Ctx) *templates.FileTemplate {
	template := templates.FileTemplate{
		Mode:    0755,
		Content: buildScriptContent,
	}

	return &template
}

func isErrCmdNotFound(err error) bool {
	errCmdNotFoundRegexp := regexp.MustCompile(
		`container process caused "exec: .* no such file or directory"`,
	)

	return errCmdNotFoundRegexp.MatchString(err.Error())
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
