package pack

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/apex/log"

	"github.com/tarantool/cartridge-cli/cli/docker"
	"github.com/tarantool/cartridge-cli/cli/project"
)

type runtimeContext struct {
	Name              string
	TmpFilesConf      string
	AppDir            string
	AppEntrypointPath string
	WorkDir           string
	PidFile           string
	ConsoleSock       string
}

func packDocker(projectCtx *project.ProjectCtx) error {
	if err := docker.CheckMinServerVersion(); err != nil {
		return err
	}

	if projectCtx.From != "" {
		if err := project.CheckBaseDockerfile(projectCtx.From); err != nil {
			return fmt.Errorf("Invalid base runtime Dockerfile %s: %s", projectCtx.From, err)
		}
	}

	// app dir
	appDirPath := filepath.Join(projectCtx.PackageFilesDir, projectCtx.Name)
	if err := initAppDir(appDirPath, projectCtx); err != nil {
		return err
	}

	ctx := runtimeContext{
		Name:              projectCtx.Name,
		TmpFilesConf:      tmpFilesConfContent,
		AppDir:            projectCtx.AppDir,
		AppEntrypointPath: project.GetAppEntrypointPath(projectCtx),
		WorkDir:           project.GetInstanceWorkDir(projectCtx, "${TARANTOOL_INSTANCE_NAME}"),
		PidFile:           project.GetInstancePidFile(projectCtx, "${TARANTOOL_INSTANCE_NAME}"),
		ConsoleSock:       project.GetInstanceConsoleSock(projectCtx, "${TARANTOOL_INSTANCE_NAME}"),
	}

	// get runtime image Dockerfile template
	log.Debugf("Create runtime image Dockerfile")

	runtimeImageDockerfileName := fmt.Sprintf("Dockerfile.%s", projectCtx.PackID)
	dockerfileTemplate, err := project.GetRuntimeImageDockerfileTemplate(projectCtx)

	if err != nil {
		return fmt.Errorf("Failed to create runtime image Dockerfile: %s", err)
	}

	dockerfileTemplate.Path = runtimeImageDockerfileName

	if err := dockerfileTemplate.Instantiate(projectCtx.BuildDir, ctx); err != nil {
		return fmt.Errorf("Failed to create build image Dockerfile: %s", err)
	}
	defer project.RemoveTmpPath(
		filepath.Join(projectCtx.BuildDir, runtimeImageDockerfileName),
		projectCtx.Debug,
	)

	// create runtime image
	log.Infof("Build result image %s", formatImageTags(projectCtx.ResImageTags))

	err = docker.BuildImage(docker.BuildOpts{
		Tag:        projectCtx.ResImageTags,
		Dockerfile: runtimeImageDockerfileName,
		NoCache:    projectCtx.DockerNoCache,
		CacheFrom:  projectCtx.DockerCacheFrom,

		BuildDir: projectCtx.BuildDir,
		TmpDir:   projectCtx.TmpDir,
		Quiet:    projectCtx.Quiet,
	})

	if err != nil {
		return fmt.Errorf("Failed to build result image: %s", err)
	}

	log.Infof("Created result image %s", formatImageTags(projectCtx.ResImageTags))

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
