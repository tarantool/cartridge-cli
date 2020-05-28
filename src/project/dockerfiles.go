package project

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/tarantool/cartridge-cli/src/common"
	"github.com/tarantool/cartridge-cli/src/templates"
)

var (
	fromLayerRegexp *regexp.Regexp
)

func init() {
	fromLayerRegexp = regexp.MustCompile(`^from\s+centos:8$`)
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
		return nil, fmt.Errorf("Invalid base build Dockerfile %s: %s", projectCtx.BuildFrom, err)
	}

	installTarantoolLayers, err := getInstallTarantoolLayers(projectCtx)
	if err != nil {
		return nil, fmt.Errorf("Failed to get install Tarantool Dockerfile layers: %s", err)
	}

	dockerfileParts = append(dockerfileParts,
		baseLayers,
		installPackagesLayers,
		installTarantoolLayers,
		wrapUserLayers,
	)

	template.Content = strings.Join(dockerfileParts, "\n")

	return &template, nil
}

func getBaseLayers(specifiedDockerfile, defaultLayers string) (string, error) {
	var baseLayers string
	var err error

	if specifiedDockerfile == "" {
		return defaultLayers, nil
	}

	if err := checkBaseDockerfile(specifiedDockerfile); err != nil {
		return "", err
	}

	baseLayers, err = common.GetFileContent(specifiedDockerfile)
	if err != nil {
		return "", fmt.Errorf("Failed to read build Dockerfile: %s", err)
	}

	return baseLayers, nil
}

func checkBaseDockerfile(dockerfilePath string) error {
	file, err := os.Open(dockerfilePath)
	if err != nil {
		return err
	}

	scanner := common.FileLinesScanner(file)

	var fromLine string

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
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
	containerSDKPath = "/usr/share/tarantool/sdk"

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
)
