package project

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/tarantool/cartridge-cli/cli/common"
	"github.com/tarantool/cartridge-cli/cli/templates"
)

var (
	fromLayerRegexp *regexp.Regexp
)

func init() {
	fromLayerRegexp = regexp.MustCompile(`^from\s+centos:[78]$`)
}

type opensourseCtx struct {
	TarantoolRepoVersion string
}

type enterpriseCtx struct {
	HostSDKDirname   string
	ContainerSDKPath string
}

func GetBuildImageDockerfileTemplate(projectCtx *ProjectCtx) (*templates.FileTemplate, error) {
	var dockerfileParts []string

	template := templates.FileTemplate{
		Mode: 0644,
	}

	baseLayers, err := getBaseLayers(projectCtx.BuildFrom, defaultBaseLayers)
	if err != nil {
		return nil, fmt.Errorf("Failed to get base build Dockerfile %s: %s", projectCtx.BuildFrom, err)
	}

	installTarantoolLayers, err := getInstallTarantoolLayers(projectCtx)
	if err != nil {
		return nil, fmt.Errorf("Failed to get install Tarantool Dockerfile layers: %s", err)
	}

	dockerfileParts = append(dockerfileParts,
		baseLayers,
		installBuildPackagesLayers,
		installTarantoolLayers,
		wrapUserLayers,
	)

	template.Content = strings.Join(dockerfileParts, "\n")

	return &template, nil
}

func GetRuntimeImageDockerfileTemplate(projectCtx *ProjectCtx) (*templates.FileTemplate, error) {
	var dockerfileParts []string

	template := templates.FileTemplate{
		Mode: 0644,
	}

	// FROM
	baseLayers, err := getBaseLayers(projectCtx.From, defaultBaseLayers)
	if err != nil {
		return nil, fmt.Errorf("Failed to get base runtime Dockerfile %s: %s", projectCtx.BuildFrom, err)
	}

	dockerfileParts = append(dockerfileParts, baseLayers)

	// Install Tarantool Opensource or create tarantool user for Enterprise
	if !projectCtx.TarantoolIsEnterprise {
		installTarantoolLayers, err := getInstallTarantoolLayers(projectCtx)
		if err != nil {
			return nil, fmt.Errorf("Failed to get install Tarantool Dockerfile layers: %s", err)
		}

		dockerfileParts = append(dockerfileParts, installTarantoolLayers)
	} else {
		dockerfileParts = append(dockerfileParts, createUserLayers)
	}

	// Set runtime user, env and copy application code
	dockerfileParts = append(dockerfileParts,
		prepareRuntimeLayers,
		copyAppCodeLayers,
	)

	// Set PATH for Enterprise
	if projectCtx.TarantoolIsEnterprise {
		dockerfileParts = append(dockerfileParts, setTarantoolEnterprisePath)
	}

	// CMD
	dockerfileParts = append(dockerfileParts, cmdLayer)

	template.Content = strings.Join(dockerfileParts, "\n")

	return &template, nil
}

func getBaseLayers(specifiedDockerfile, defaultLayers string) (string, error) {
	var baseLayers string
	var err error

	if specifiedDockerfile == "" {
		return defaultLayers, nil
	}

	baseLayers, err = common.GetFileContent(specifiedDockerfile)
	if err != nil {
		return "", fmt.Errorf("Failed to read base Dockerfile: %s", err)
	}

	return baseLayers, nil
}

func CheckBaseDockerfile(dockerfilePath string) error {
	file, err := os.Open(dockerfilePath)
	if err != nil {
		return err
	}

	scanner := common.FileLinesScanner(file)

	var fromLine string

	for scanner.Scan() {
		line := scanner.Text()
		line = common.TrimSince(line, "#")
		line = strings.TrimSpace(line)

		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		fromLine = line
		break

	}

	fromLine = strings.ToLower(fromLine)
	if !fromLayerRegexp.MatchString(fromLine) {
		return fmt.Errorf("The base image must be centos:8")
	}

	return nil
}

func getInstallTarantoolLayers(projectCtx *ProjectCtx) (string, error) {
	var installTarantoolLayers string
	var err error

	if projectCtx.TarantoolIsEnterprise {
		tmplStr := installTarantoolEnterpriseLayers
		installTarantoolLayers, err = templates.GetTemplatedStr(&tmplStr,
			enterpriseCtx{
				HostSDKDirname:   projectCtx.BuildSDKDirname,
				ContainerSDKPath: containerSDKPath,
			},
		)

	} else {
		tmplStr := installTarantoolOpensourceLayers
		installTarantoolLayers, err = templates.GetTemplatedStr(&tmplStr,
			opensourseCtx{
				TarantoolRepoVersion: common.GetTarantoolRepoVersion(projectCtx.TarantoolVersion),
			},
		)

		if err != nil {
			return "", err
		}
	}

	return installTarantoolLayers, nil
}

const (
	DefaultBaseBuildDockerfile   = "Dockerfile.build.cartridge"
	DefaultBaseRuntimeDockerfile = "Dockerfile.cartridge"

	containerSDKPath = "/usr/share/tarantool/sdk"

	defaultBaseLayers          = "FROM centos:8\n"
	installBuildPackagesLayers = `### Install packages required for build
RUN yum install -y git-core gcc make cmake unzip
`

	createUserLayers = `### Create Tarantool user and directories
RUN groupadd -r tarantool \
    && useradd -M -N -g tarantool -r -d /var/lib/tarantool -s /sbin/nologin \
        -c "Tarantool Server" tarantool \
    &&  mkdir -p /var/lib/tarantool/ --mode 755 \
    && chown tarantool:tarantool /var/lib/tarantool \
    && mkdir -p /var/run/tarantool/ --mode 755 \
	&& chown tarantool:tarantool /var/run/tarantool
`

	prepareRuntimeLayers = `### Prepare for runtime
RUN echo '{{ .TmpFilesConf }}' > /usr/lib/tmpfiles.d/{{ .Name }}.conf \
    && chmod 644 /usr/lib/tmpfiles.d/{{ .Name }}.conf

USER tarantool:tarantool
ENV TARANTOOL_INSTANCE_NAME=default
`

	installTarantoolOpensourceLayers = `### Install opensource Tarantool
RUN curl -s \
        https://packagecloud.io/install/repositories/tarantool/{{ .TarantoolRepoVersion }}/script.rpm.sh | bash \
    && yum -y install tarantool tarantool-devel
`

	installTarantoolEnterpriseLayers = `### Set path for Tarantool Enterprise
COPY {{ .HostSDKDirname }} {{ .ContainerSDKPath }}
ENV PATH="{{ .ContainerSDKPath }}:${PATH}"
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

	copyAppCodeLayers = `### Copy application code
COPY . {{ .AppDir }}
`

	setTarantoolEnterprisePath = `### Set PATH
ENV PATH="{{ .AppDir }}:${PATH}"
`

	cmdLayer = `### Runtime command
CMD TARANTOOL_WORKDIR={{ .WorkDir }} \
    TARANTOOL_PID_FILE={{ .PidFile }} \
    TARANTOOL_CONSOLE_SOCK={{ .ConsoleSock }} \
	tarantool {{ .AppEntrypointPath }}
`
)
