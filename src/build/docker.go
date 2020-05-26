package build

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/tarantool/cartridge-cli/src/common"
	"github.com/tarantool/cartridge-cli/src/docker"
	"github.com/tarantool/cartridge-cli/src/project"
	"github.com/tarantool/cartridge-cli/src/templates"
)

type buildContext struct {
	UserID               string
	TarantoolRepoVersion string
	BuildID              string
	PreBuildHookName     string
}

func buildProjectInDocker(projectCtx *project.ProjectCtx) error {
	var err error

	// fill build context
	userID, err := common.GetCurrentUserID()
	if err != nil {
		return fmt.Errorf("Failed to get current user ID: %s", err)
	}

	ctx := buildContext{
		BuildID:              projectCtx.PackID,
		UserID:               userID,
		TarantoolRepoVersion: common.GetTarantoolRepoVersion(projectCtx.TarantoolVersion),
		PreBuildHookName:     preBuildHookName,
	}

	// create build image Dockerfile
	buildImageDockerfileName := fmt.Sprintf("Dockerfile.build.%s", projectCtx.PackID)
	dockerfileTemplate := getBuildImageDockerfileTemplate(projectCtx)
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
	log.Infof("Build base image: %s", buildImageTag)

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
	buildScriptName := fmt.Sprintf("build.%s.sh", projectCtx.PackID)

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
	log.Infof("Build application in %s", buildImageTag)
	containerBuildDir := "/opt/tarantool"

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

func getBuildImageDockerfileTemplate(projectCtx *project.ProjectCtx) *templates.FileTemplate {
	var dockerfileParts []string

	template := templates.FileTemplate{
		Mode: 0644,
	}

	dockerfileParts = append(dockerfileParts, defaultBaseLayers) // XXX
	dockerfileParts = append(dockerfileParts, installPackagesLayers)
	dockerfileParts = append(dockerfileParts, prepareLayers)
	dockerfileParts = append(dockerfileParts, installTarantoolOpensourceLayers)
	dockerfileParts = append(dockerfileParts, wrapUserLayers)

	template.Content = strings.Join(dockerfileParts, "\n")

	return &template
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
	defaultBaseLayers     = `FROM centos:8`
	installPackagesLayers = `### Install packages required for build
RUN yum install -y git-core gcc make cmake unzip
`
	prepareLayers = `# Create Tarantool user and directories
RUN groupadd -r tarantool \
    && useradd -M -N -g tarantool -r -d /var/lib/tarantool -s /sbin/nologin \
        -c "Tarantool Server" tarantool \
    &&  mkdir -p /var/lib/tarantool/ --mode 755 \
    && chown tarantool:tarantool /var/lib/tarantool \
    && mkdir -p /var/run/tarantool/ --mode 755 \
	&& chown tarantool:tarantool /var/run/tarantool
`

	installTarantoolOpensourceLayers = `### Install opensource Tarantool
RUN curl -s \
        https://packagecloud.io/install/repositories/tarantool/{{ .TarantoolRepoVersion }}/script.rpm.sh | bash \
	&& yum -y install tarantool tarantool-devel
`

	wrapUserLayers = `### Wrap user
RUN if id -u {{ .UserID }} 2>/dev/null; then \
        USERNAME=$(id -nu {{ .UserID }}); \
    else \
        USERNAME=cartridge; \
        useradd -u {{ .UserID }} ${USERNAME}; \
    fi \
    && (usermod -a -G sudo ${USERNAME} 2>/dev/null || :) \
    && (usermod -a -G wheel ${USERNAME} 2>/dev/null || :) \
    && (usermod -a -G adm ${USERNAME} 2>/dev/null || :)
USER {{ .UserID }}
`

	buildScriptContent = `#!/bin/bash
set -xe

if [ -f {{ .PreBuildHookName }} ]; then
    . {{ .PreBuildHookName }}
fi

tarantoolctl rocks make
`
)
