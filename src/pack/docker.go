package pack

import (
	"fmt"
	"path/filepath"

	log "github.com/sirupsen/logrus"

	"github.com/tarantool/cartridge-cli/src/docker"
	"github.com/tarantool/cartridge-cli/src/project"
)

type runtimeContext struct {
	Name         string
	TmpFilesConf string
	AppDir       string
	Entrypoint   string
	WorkDir      string
}

func packDocker(projectCtx *project.ProjectCtx) error {
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
		Name:         projectCtx.Name,
		TmpFilesConf: tmpFilesConfContent,
		AppDir:       filepath.Join("/usr/share/tarantool", projectCtx.Name),
		Entrypoint:   project.AppEntrypointName,
		WorkDir:      filepath.Join("/var/lib/tarantool/", projectCtx.Name),
	}

	// get runtime image Dockerfile template
	log.Debugf("Create runtime image Dockerfile")

	runtimeImageDockerfileName := fmt.Sprintf("Dockerfile.%s", projectCtx.PackID)
	fmt.Printf("projectCtx.From: %s\n", projectCtx.From)
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
	log.Infof("Building result image: %s", projectCtx.ResImageFullname)

	err = docker.BuildImage(docker.BuildOpts{
		Tag:        projectCtx.ResImageFullname,
		Dockerfile: runtimeImageDockerfileName,
		BuildDir:   projectCtx.BuildDir,
		TmpDir:     projectCtx.TmpDir,
		Quiet:      projectCtx.Quiet,
	})

	if err != nil {
		return fmt.Errorf("Failed to build result image: %s", err)
	}

	log.Infof("Result image tagged as: %s", projectCtx.ResImageFullname)

	return nil
}
