package project

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tarantool/cartridge-cli/src/templates"
)

func writeDockerfile(file *os.File, content string) {
	if err := ioutil.WriteFile(file.Name(), []byte(content), 0644); err != nil {
		panic(fmt.Errorf("Failed to write Dockerfile: %s", err))
	}
}

func TestCheckBaseDockerfile(t *testing.T) {
	assert := assert.New(t)

	var err error
	baseImageError := "The base image must be centos:8"

	// create tmp Dockerfile
	f, err := ioutil.TempFile("", "Dockerfile")
	if err != nil {
		log.Fatal(err)
	}
	defer os.Remove(f.Name())

	// non existing file
	err = CheckBaseDockerfile("bad-path")
	assert.EqualError(err, "open bad-path: no such file or directory")

	// OK
	writeDockerfile(f, `FROM centos:8`)
	err = CheckBaseDockerfile(f.Name())
	assert.Nil(err)

	writeDockerfile(f, `from centos:8`)
	err = CheckBaseDockerfile(f.Name())
	assert.Nil(err)

	writeDockerfile(f, `
# comment
FROM centos:8`)
	err = CheckBaseDockerfile(f.Name())
	assert.Nil(err)

	writeDockerfile(f, `FROM centos:8 # comment`)
	err = CheckBaseDockerfile(f.Name())
	assert.Nil(err)

	writeDockerfile(f, `# FROM centos:7
FROM centos:8`)
	err = CheckBaseDockerfile(f.Name())
	assert.Nil(err)

	// Error
	writeDockerfile(f, `FROM centos:7`)
	err = CheckBaseDockerfile(f.Name())
	assert.EqualError(err, baseImageError)

	writeDockerfile(f, ``)
	err = CheckBaseDockerfile(f.Name())
	assert.EqualError(err, baseImageError)

	writeDockerfile(f, `# from centos:8`)
	err = CheckBaseDockerfile(f.Name())
	assert.EqualError(err, baseImageError)

	writeDockerfile(f, `
# comment
FROM centos:7`)
	err = CheckBaseDockerfile(f.Name())
	assert.EqualError(err, baseImageError)
}

func TestGetBaseLayers(t *testing.T) {
	assert := assert.New(t)

	var err error
	var layers string

	defaultLayers := "FROM centos:8"

	// create tmp Dockerfile
	f, err := ioutil.TempFile("", "Dockerfile")
	if err != nil {
		log.Fatal(err)
	}
	defer os.Remove(f.Name())

	// default
	layers, err = getBaseLayers("", defaultLayers)
	assert.Nil(err)
	assert.Equal(defaultLayers, layers)

	// bad file
	layers, err = getBaseLayers("bad-path", defaultLayers)
	assert.EqualError(err, "Failed to read base Dockerfile: open bad-path: no such file or directory")

	// OK
	baseDockerfileContent := `FROM centos:8 # my base layers`
	writeDockerfile(f, baseDockerfileContent)
	layers, err = getBaseLayers(f.Name(), defaultLayers)

	assert.Nil(err)
	assert.Equal(baseDockerfileContent, layers)
}

func TestGetInstallTarantoolLayers(t *testing.T) {
	assert := assert.New(t)

	var err error
	var layers string
	var expLayers string
	var projectCtx ProjectCtx

	// Tarantool Enterprise
	projectCtx.TarantoolIsEnterprise = true
	projectCtx.BuildSDKDirname = "buildSDKDirname"

	expLayers = `### Set path for Tarantool Enterprise
COPY buildSDKDirname /usr/share/tarantool/sdk
ENV PATH="/usr/share/tarantool/sdk:${PATH}"
`

	layers, err = getInstallTarantoolLayers(&projectCtx)
	assert.Nil(err)
	assert.Equal(expLayers, layers)

	// Tarantool Opensource 2.1
	projectCtx.TarantoolIsEnterprise = false
	projectCtx.TarantoolVersion = "2.1.42"

	expLayers = `### Install opensource Tarantool
RUN curl -s \
        https://packagecloud.io/install/repositories/tarantool/2x/script.rpm.sh | bash \
    && yum -y install tarantool tarantool-devel
`

	layers, err = getInstallTarantoolLayers(&projectCtx)
	assert.Nil(err)
	assert.Equal(expLayers, layers)

	// Tarantool Opensource 1.10
	projectCtx.TarantoolIsEnterprise = false
	projectCtx.TarantoolVersion = "1.10.42"

	expLayers = `### Install opensource Tarantool
RUN curl -s \
        https://packagecloud.io/install/repositories/tarantool/1_10/script.rpm.sh | bash \
    && yum -y install tarantool tarantool-devel
`

	layers, err = getInstallTarantoolLayers(&projectCtx)
	assert.Nil(err)
	assert.Equal(expLayers, layers)
}

func TestGetBuildImageDockerfileTemplateEnterprise(t *testing.T) {
	assert := assert.New(t)

	var err error
	var expLayers string
	var projectCtx ProjectCtx
	var tmpl *templates.FileTemplate

	// create tmp Dockerfile
	f, err := ioutil.TempFile("", "Dockerfile")
	if err != nil {
		log.Fatal(err)
	}
	defer os.Remove(f.Name())

	// Tarantool Enterprise w/o --build-from
	projectCtx.TarantoolIsEnterprise = true
	projectCtx.BuildSDKDirname = "buildSDKDirname"
	projectCtx.BuildFrom = ""

	expLayers = `FROM centos:8

### Install packages required for build
RUN yum install -y git-core gcc make cmake unzip

### Set path for Tarantool Enterprise
COPY buildSDKDirname /usr/share/tarantool/sdk
ENV PATH="/usr/share/tarantool/sdk:${PATH}"

### Wrap user
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

	tmpl, err = GetBuildImageDockerfileTemplate(&projectCtx)
	assert.Nil(err)
	assert.Equal(expLayers, tmpl.Content)

	// Tarantool Enterprise w/ --build-from
	baseDockerfileContent := `FROM centos:8
RUN yum install -y zip
`
	writeDockerfile(f, baseDockerfileContent)

	projectCtx.TarantoolIsEnterprise = true
	projectCtx.BuildSDKDirname = "buildSDKDirname"
	projectCtx.BuildFrom = f.Name()

	expLayers = `FROM centos:8
RUN yum install -y zip

### Install packages required for build
RUN yum install -y git-core gcc make cmake unzip

### Set path for Tarantool Enterprise
COPY buildSDKDirname /usr/share/tarantool/sdk
ENV PATH="/usr/share/tarantool/sdk:${PATH}"

### Wrap user
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

	tmpl, err = GetBuildImageDockerfileTemplate(&projectCtx)
	assert.Nil(err)
	assert.Equal(expLayers, tmpl.Content)
}

func TestGetBuildImageDockerfileTemplateOpensource(t *testing.T) {
	assert := assert.New(t)

	var err error
	var expLayers string
	var projectCtx ProjectCtx
	var tmpl *templates.FileTemplate

	// create tmp Dockerfile
	f, err := ioutil.TempFile("", "Dockerfile")
	if err != nil {
		log.Fatal(err)
	}
	defer os.Remove(f.Name())

	// Tarantool Opensource 1.10 w/o --build-from
	projectCtx.TarantoolIsEnterprise = false
	projectCtx.TarantoolVersion = "1.10.42"
	projectCtx.BuildFrom = ""

	expLayers = `FROM centos:8

### Install packages required for build
RUN yum install -y git-core gcc make cmake unzip

### Install opensource Tarantool
RUN curl -s \
        https://packagecloud.io/install/repositories/tarantool/1_10/script.rpm.sh | bash \
    && yum -y install tarantool tarantool-devel

### Wrap user
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

	tmpl, err = GetBuildImageDockerfileTemplate(&projectCtx)
	assert.Nil(err)
	assert.Equal(expLayers, tmpl.Content)

	// Tarantool Opensource 1.10 w/ --build-from
	baseDockerfileContent := `FROM centos:8
RUN yum install -y zip
`
	writeDockerfile(f, baseDockerfileContent)

	projectCtx.TarantoolIsEnterprise = false
	projectCtx.TarantoolVersion = "1.10.42"
	projectCtx.BuildFrom = f.Name()

	expLayers = `FROM centos:8
RUN yum install -y zip

### Install packages required for build
RUN yum install -y git-core gcc make cmake unzip

### Install opensource Tarantool
RUN curl -s \
        https://packagecloud.io/install/repositories/tarantool/1_10/script.rpm.sh | bash \
    && yum -y install tarantool tarantool-devel

### Wrap user
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

	tmpl, err = GetBuildImageDockerfileTemplate(&projectCtx)
	assert.Nil(err)
	assert.Equal(expLayers, tmpl.Content)
}

func TestGetRuntimeImageDockerfileTemplateEnterprise(t *testing.T) {
	assert := assert.New(t)

	var err error
	var expLayers string
	var projectCtx ProjectCtx
	var tmpl *templates.FileTemplate

	// create tmp Dockerfile
	f, err := ioutil.TempFile("", "Dockerfile")
	if err != nil {
		log.Fatal(err)
	}
	defer os.Remove(f.Name())

	// Tarantool Enterprise w/o --build-from
	projectCtx.TarantoolIsEnterprise = true
	projectCtx.BuildSDKDirname = "buildSDKDirname"
	projectCtx.From = ""

	expLayers = `FROM centos:8

### Create Tarantool user and directories
RUN groupadd -r tarantool \
    && useradd -M -N -g tarantool -r -d /var/lib/tarantool -s /sbin/nologin \
        -c "Tarantool Server" tarantool \
    &&  mkdir -p /var/lib/tarantool/ --mode 755 \
    && chown tarantool:tarantool /var/lib/tarantool \
    && mkdir -p /var/run/tarantool/ --mode 755 \
	&& chown tarantool:tarantool /var/run/tarantool

### Prepare for runtime
RUN echo '{{ .TmpFilesConf }}' > /usr/lib/tmpfiles.d/{{ .Name }}.conf \
    && chmod 644 /usr/lib/tmpfiles.d/{{ .Name }}.conf

USER tarantool:tarantool
ENV TARANTOOL_INSTANCE_NAME=default

### Copy application code
COPY . {{ .AppDir }}

### Set PATH
ENV PATH="{{ .AppDir }}:${PATH}"

### Runtime command
CMD TARANTOOL_WORKDIR={{ .WorkDir }}.${TARANTOOL_INSTANCE_NAME} \
    TARANTOOL_PID_FILE=/var/run/tarantool/{{ .Name }}.${TARANTOOL_INSTANCE_NAME}.pid \
    TARANTOOL_CONSOLE_SOCK=/var/run/tarantool/{{ .Name }}.${TARANTOOL_INSTANCE_NAME}.control \
	tarantool {{ .AppDir }}/{{ .Entrypoint }}
`

	tmpl, err = GetRuntimeImageDockerfileTemplate(&projectCtx)
	assert.Nil(err)
	assert.Equal(expLayers, tmpl.Content)

	// TODO: Tarantool Enterprise w/ --build-from
}

func TestGetRuntimeImageDockerfileTemplateOpensource(t *testing.T) {
	assert := assert.New(t)

	var err error
	var expLayers string
	var projectCtx ProjectCtx
	var tmpl *templates.FileTemplate

	// create tmp Dockerfile
	f, err := ioutil.TempFile("", "Dockerfile")
	if err != nil {
		log.Fatal(err)
	}
	defer os.Remove(f.Name())

	// Tarantool Opensource 1.10 w/o --from
	projectCtx.TarantoolIsEnterprise = false
	projectCtx.TarantoolVersion = "1.10.42"
	projectCtx.From = ""

	expLayers = `FROM centos:8

### Install opensource Tarantool
RUN curl -s \
        https://packagecloud.io/install/repositories/tarantool/1_10/script.rpm.sh | bash \
    && yum -y install tarantool tarantool-devel

### Prepare for runtime
RUN echo '{{ .TmpFilesConf }}' > /usr/lib/tmpfiles.d/{{ .Name }}.conf \
    && chmod 644 /usr/lib/tmpfiles.d/{{ .Name }}.conf

USER tarantool:tarantool
ENV TARANTOOL_INSTANCE_NAME=default

### Copy application code
COPY . {{ .AppDir }}

### Runtime command
CMD TARANTOOL_WORKDIR={{ .WorkDir }}.${TARANTOOL_INSTANCE_NAME} \
    TARANTOOL_PID_FILE=/var/run/tarantool/{{ .Name }}.${TARANTOOL_INSTANCE_NAME}.pid \
    TARANTOOL_CONSOLE_SOCK=/var/run/tarantool/{{ .Name }}.${TARANTOOL_INSTANCE_NAME}.control \
	tarantool {{ .AppDir }}/{{ .Entrypoint }}
`

	tmpl, err = GetRuntimeImageDockerfileTemplate(&projectCtx)
	assert.Nil(err)
	assert.Equal(expLayers, tmpl.Content)

	// TODO: Tarantool Opensource 1.10 w/ --from
}
