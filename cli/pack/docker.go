package pack

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/apex/log"

	"github.com/tarantool/cartridge-cli/cli/context"
	"github.com/tarantool/cartridge-cli/cli/docker"
	"github.com/tarantool/cartridge-cli/cli/project"
)

func packDocker(ctx *context.Ctx) error {
	if err := docker.CheckMinServerVersion(); err != nil {
		return err
	}

	if ctx.Pack.DockerFrom != "" {
		if err := project.CheckBaseDockerfile(ctx.Pack.DockerFrom); err != nil {
			return fmt.Errorf("Invalid base runtime Dockerfile %s: %s", ctx.Pack.DockerFrom, err)
		}
	}

	// app dir
	appDirPath := filepath.Join(ctx.Pack.PackageFilesDir, ctx.Project.Name)
	if err := initAppDir(appDirPath, ctx); err != nil {
		return err
	}

	runtimeContext := map[string]interface{}{
		"Name":              ctx.Project.Name,
		"TmpFilesConf":      tmpFilesConfContent,
		"AppDir":            ctx.Running.AppDir,
		"AppEntrypointPath": project.GetAppEntrypointPath(ctx),
		"WorkDir":           project.GetInstanceWorkDir(ctx, "${TARANTOOL_INSTANCE_NAME}"),
		"PidFile":           project.GetInstancePidFile(ctx, "${TARANTOOL_INSTANCE_NAME}"),
		"ConsoleSock":       project.GetInstanceConsoleSock(ctx, "${TARANTOOL_INSTANCE_NAME}"),
	}

	// get runtime image Dockerfile template
	log.Debugf("Create runtime image Dockerfile")

	runtimeImageDockerfileName := fmt.Sprintf("Dockerfile.%s", ctx.Pack.ID)
	dockerfileTemplate, err := project.GetRuntimeImageDockerfileTemplate(ctx)

	if err != nil {
		return fmt.Errorf("Failed to create runtime image Dockerfile: %s", err)
	}

	dockerfileTemplate.Path = runtimeImageDockerfileName

	if err := dockerfileTemplate.Instantiate(ctx.Build.Dir, runtimeContext); err != nil {
		return fmt.Errorf("Failed to create build image Dockerfile: %s", err)
	}
	defer project.RemoveTmpPath(
		filepath.Join(ctx.Build.Dir, runtimeImageDockerfileName),
		ctx.Cli.Debug,
	)

	// create runtime image
	log.Infof("Build result image %s", formatImageTags(ctx.Pack.ResImageTags))

	err = docker.BuildImage(docker.BuildOpts{
		Tag:        ctx.Pack.ResImageTags,
		Dockerfile: runtimeImageDockerfileName,
		NoCache:    ctx.Docker.NoCache,
		CacheFrom:  ctx.Docker.CacheFrom,

		BuildDir:   ctx.Build.Dir,
		TmpDir:     ctx.Pack.TmpDir,
		ShowOutput: ctx.Cli.Verbose,
	})

	if err != nil {
		return fmt.Errorf("Failed to build result image: %s", err)
	}

	log.Infof("Created result image %s", formatImageTags(ctx.Pack.ResImageTags))

	return nil
}

func formatImageTags(imageTags []string) string {
	if len(imageTags) == 0 {
		return "<no tags>"
	}

	if len(imageTags) == 1 {
		return imageTags[0]
	}

	return fmt.Sprintf(
		"with tags %s",
		strings.Join(imageTags, ", "),
	)
}
