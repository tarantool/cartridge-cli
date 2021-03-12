package project

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tarantool/cartridge-cli/cli/context"
	"github.com/tarantool/cartridge-cli/cli/templates"
)

func writeDockerfile(file *os.File, content string) {
	if err := ioutil.WriteFile(file.Name(), []byte(content), 0644); err != nil {
		panic(fmt.Errorf("Failed to write Dockerfile: %s", err))
	}
}

func TestCheckBaseDockerfile(t *testing.T) {
	t.Parallel()
	assert := assert.New(t)

	var err error

	// create tmp Dockerfile
	f, err := ioutil.TempFile("", "Dockerfile")
	if err != nil {
		log.Fatal(err)
	}
	defer os.Remove(f.Name())

	// non existing file
	err = CheckBaseDockerfile("bad-path")
	assert.EqualError(err, "open bad-path: no such file or directory")
}

func TestGetBaseLayers(t *testing.T) {
	assert := assert.New(t)

	var err error
	var layers string

	defaultLayers := "FROM centos:7"

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
	baseDockerfileContent := `FROM centos:7 # my base layers`
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
	var ctx context.Ctx

	// Tarantool Enterprise
	ctx.Tarantool.TarantoolIsEnterprise = true
	ctx.Build.BuildSDKDirname = "buildSDKDirname"

	expLayers = `### Set path for Tarantool Enterprise
COPY buildSDKDirname /usr/share/tarantool/sdk
ENV PATH="/usr/share/tarantool/sdk:${PATH}"
`

	layers, err = getInstallTarantoolLayers(&ctx)
	assert.Nil(err)
	assert.Equal(expLayers, layers)

	// Tarantool Opensource 2.1
	ctx.Tarantool.TarantoolIsEnterprise = false
	ctx.Tarantool.TarantoolVersion = "2.1.42"

	expLayers = `### Install opensource Tarantool
RUN curl -L https://tarantool.io/installer.sh | VER=2.1 bash \
    && yum -y install tarantool-devel
`

	layers, err = getInstallTarantoolLayers(&ctx)
	assert.Nil(err)
	assert.Equal(expLayers, layers)

	// Tarantool Opensource 1.10
	ctx.Tarantool.TarantoolIsEnterprise = false
	ctx.Tarantool.TarantoolVersion = "1.10.42"

	expLayers = `### Install opensource Tarantool
RUN curl -L https://tarantool.io/installer.sh | VER=1.10 bash \
    && yum -y install tarantool-devel
`

	layers, err = getInstallTarantoolLayers(&ctx)
	assert.Nil(err)
	assert.Equal(expLayers, layers)
}

func TestGetBuildImageDockerfileTemplateEnterprise(t *testing.T) {
	assert := assert.New(t)

	var err error
	var expLayers string
	var ctx context.Ctx
	var tmpl *templates.FileTemplate

	// create tmp Dockerfile
	f, err := ioutil.TempFile("", "Dockerfile")
	if err != nil {
		log.Fatal(err)
	}
	defer os.Remove(f.Name())

	// Tarantool Enterprise w/o --build-from
	ctx.Tarantool.TarantoolIsEnterprise = true
	ctx.Build.BuildSDKDirname = "buildSDKDirname"
	ctx.Build.DockerFrom = ""

	expLayers = `FROM centos:7

### Install packages required for build
RUN yum install -y git-core gcc gcc-c++ make cmake unzip

### Set path for Tarantool Enterprise
COPY buildSDKDirname /usr/share/tarantool/sdk
ENV PATH="/usr/share/tarantool/sdk:${PATH}"

### Wrap user
RUN if id -u {{ .UserID }} 2>/dev/null; then \
        USERNAME=$(id -nu {{ .UserID }}); \
    else \
        USERNAME=cartridge; \
        useradd -l -u {{ .UserID }} ${USERNAME}; \
    fi \
    && (usermod -a -G sudo ${USERNAME} 2>/dev/null || :) \
    && (usermod -a -G wheel ${USERNAME} 2>/dev/null || :) \
    && (usermod -a -G adm ${USERNAME} 2>/dev/null || :)
USER {{ .UserID }}
`

	tmpl, err = GetBuildImageDockerfileTemplate(&ctx)
	assert.Nil(err)
	assert.Equal(expLayers, tmpl.Content)

	// Tarantool Enterprise w/ --build-from
	baseDockerfileContent := `FROM centos:7
RUN yum install -y zip
`
	writeDockerfile(f, baseDockerfileContent)

	ctx.Tarantool.TarantoolIsEnterprise = true
	ctx.Build.BuildSDKDirname = "buildSDKDirname"
	ctx.Build.DockerFrom = f.Name()

	expLayers = `FROM centos:7
RUN yum install -y zip

### Install packages required for build
RUN yum install -y git-core gcc gcc-c++ make cmake unzip

### Set path for Tarantool Enterprise
COPY buildSDKDirname /usr/share/tarantool/sdk
ENV PATH="/usr/share/tarantool/sdk:${PATH}"

### Wrap user
RUN if id -u {{ .UserID }} 2>/dev/null; then \
        USERNAME=$(id -nu {{ .UserID }}); \
    else \
        USERNAME=cartridge; \
        useradd -l -u {{ .UserID }} ${USERNAME}; \
    fi \
    && (usermod -a -G sudo ${USERNAME} 2>/dev/null || :) \
    && (usermod -a -G wheel ${USERNAME} 2>/dev/null || :) \
    && (usermod -a -G adm ${USERNAME} 2>/dev/null || :)
USER {{ .UserID }}
`

	tmpl, err = GetBuildImageDockerfileTemplate(&ctx)
	assert.Nil(err)
	assert.Equal(expLayers, tmpl.Content)
}

func TestGetBuildImageDockerfileTemplateOpensource(t *testing.T) {
	assert := assert.New(t)

	var err error
	var expLayers string
	var ctx context.Ctx
	var tmpl *templates.FileTemplate

	// create tmp Dockerfile
	f, err := ioutil.TempFile("", "Dockerfile")
	if err != nil {
		log.Fatal(err)
	}
	defer os.Remove(f.Name())

	// Tarantool Opensource 1.10 w/o --build-from
	ctx.Tarantool.TarantoolIsEnterprise = false
	ctx.Tarantool.TarantoolVersion = "1.10.42"
	ctx.Build.DockerFrom = ""

	expLayers = `FROM centos:7

### Install packages required for build
RUN yum install -y git-core gcc gcc-c++ make cmake unzip

### Install opensource Tarantool
RUN curl -L https://tarantool.io/installer.sh | VER=1.10 bash \
    && yum -y install tarantool-devel

### Wrap user
RUN if id -u {{ .UserID }} 2>/dev/null; then \
        USERNAME=$(id -nu {{ .UserID }}); \
    else \
        USERNAME=cartridge; \
        useradd -l -u {{ .UserID }} ${USERNAME}; \
    fi \
    && (usermod -a -G sudo ${USERNAME} 2>/dev/null || :) \
    && (usermod -a -G wheel ${USERNAME} 2>/dev/null || :) \
    && (usermod -a -G adm ${USERNAME} 2>/dev/null || :)
USER {{ .UserID }}
`

	tmpl, err = GetBuildImageDockerfileTemplate(&ctx)
	assert.Nil(err)
	assert.Equal(expLayers, tmpl.Content)

	// Tarantool Opensource 1.10 w/ --build-from
	baseDockerfileContent := `FROM centos:7
RUN yum install -y zip
`
	writeDockerfile(f, baseDockerfileContent)

	ctx.Tarantool.TarantoolIsEnterprise = false
	ctx.Tarantool.TarantoolVersion = "1.10.42"
	ctx.Build.DockerFrom = f.Name()

	expLayers = `FROM centos:7
RUN yum install -y zip

### Install packages required for build
RUN yum install -y git-core gcc gcc-c++ make cmake unzip

### Install opensource Tarantool
RUN curl -L https://tarantool.io/installer.sh | VER=1.10 bash \
    && yum -y install tarantool-devel

### Wrap user
RUN if id -u {{ .UserID }} 2>/dev/null; then \
        USERNAME=$(id -nu {{ .UserID }}); \
    else \
        USERNAME=cartridge; \
        useradd -l -u {{ .UserID }} ${USERNAME}; \
    fi \
    && (usermod -a -G sudo ${USERNAME} 2>/dev/null || :) \
    && (usermod -a -G wheel ${USERNAME} 2>/dev/null || :) \
    && (usermod -a -G adm ${USERNAME} 2>/dev/null || :)
USER {{ .UserID }}
`

	tmpl, err = GetBuildImageDockerfileTemplate(&ctx)
	assert.Nil(err)
	assert.Equal(expLayers, tmpl.Content)
}

func TestGetRuntimeImageDockerfileTemplateEnterprise(t *testing.T) {
	assert := assert.New(t)

	var err error
	var expLayers string
	var ctx context.Ctx
	var tmpl *templates.FileTemplate

	// create tmp Dockerfile
	f, err := ioutil.TempFile("", "Dockerfile")
	if err != nil {
		log.Fatal(err)
	}
	defer os.Remove(f.Name())

	// Tarantool Enterprise w/o --from
	ctx.Tarantool.TarantoolIsEnterprise = true
	ctx.Build.BuildSDKDirname = "buildSDKDirname"
	ctx.Pack.DockerFrom = ""

	expLayers = `FROM centos:7

### Create Tarantool user and directories
RUN groupadd -r tarantool \
    && useradd -M -N -l -g tarantool -r -d /var/lib/tarantool -s /sbin/nologin \
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
ENV TARANTOOL_WORKDIR={{ .WorkDir }}
ENV TARANTOOL_PID_FILE={{ .PidFile }}
ENV TARANTOOL_CONSOLE_SOCK={{ .ConsoleSock }}

### Copy application code
COPY . {{ .AppDir }}

### Set PATH
ENV PATH="{{ .AppDir }}:${PATH}"

### Runtime command
CMD TARANTOOL_WORKDIR=$TARANTOOL_WORKDIR \
	TARANTOOL_PID_FILE=$TARANTOOL_PID_FILE \
	TARANTOOL_CONSOLE_SOCK=$TARANTOOL_CONSOLE_SOCK \
	tarantool {{ .AppEntrypointPath }}
`

	tmpl, err = GetRuntimeImageDockerfileTemplate(&ctx)
	assert.Nil(err)
	assert.Equal(expLayers, tmpl.Content)

	// Tarantool Enterprise w/ --from
	baseDockerfileContent := `FROM centos:7
RUN yum install -y zip
`
	writeDockerfile(f, baseDockerfileContent)

	ctx.Tarantool.TarantoolIsEnterprise = true
	ctx.Build.BuildSDKDirname = "buildSDKDirname"
	ctx.Pack.DockerFrom = f.Name()

	expLayers = `FROM centos:7
RUN yum install -y zip

### Create Tarantool user and directories
RUN groupadd -r tarantool \
    && useradd -M -N -l -g tarantool -r -d /var/lib/tarantool -s /sbin/nologin \
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
ENV TARANTOOL_WORKDIR={{ .WorkDir }}
ENV TARANTOOL_PID_FILE={{ .PidFile }}
ENV TARANTOOL_CONSOLE_SOCK={{ .ConsoleSock }}

### Copy application code
COPY . {{ .AppDir }}

### Set PATH
ENV PATH="{{ .AppDir }}:${PATH}"

### Runtime command
CMD TARANTOOL_WORKDIR=$TARANTOOL_WORKDIR \
	TARANTOOL_PID_FILE=$TARANTOOL_PID_FILE \
	TARANTOOL_CONSOLE_SOCK=$TARANTOOL_CONSOLE_SOCK \
	tarantool {{ .AppEntrypointPath }}
`

	tmpl, err = GetRuntimeImageDockerfileTemplate(&ctx)
	assert.Nil(err)
	assert.Equal(expLayers, tmpl.Content)

}

func TestGetRuntimeImageDockerfileTemplateOpensource(t *testing.T) {
	assert := assert.New(t)

	var err error
	var expLayers string
	var ctx context.Ctx
	var tmpl *templates.FileTemplate

	// create tmp Dockerfile
	f, err := ioutil.TempFile("", "Dockerfile")
	if err != nil {
		log.Fatal(err)
	}
	defer os.Remove(f.Name())

	// Tarantool Opensource 1.10 w/o --from
	ctx.Tarantool.TarantoolIsEnterprise = false
	ctx.Tarantool.TarantoolVersion = "1.10.42"
	ctx.Pack.DockerFrom = ""

	expLayers = `FROM centos:7

### Install opensource Tarantool
RUN curl -L https://tarantool.io/installer.sh | VER=1.10 bash \
    && yum -y install tarantool-devel

### Prepare for runtime
RUN echo '{{ .TmpFilesConf }}' > /usr/lib/tmpfiles.d/{{ .Name }}.conf \
    && chmod 644 /usr/lib/tmpfiles.d/{{ .Name }}.conf

USER tarantool:tarantool
ENV TARANTOOL_INSTANCE_NAME=default
ENV TARANTOOL_WORKDIR={{ .WorkDir }}
ENV TARANTOOL_PID_FILE={{ .PidFile }}
ENV TARANTOOL_CONSOLE_SOCK={{ .ConsoleSock }}

### Copy application code
COPY . {{ .AppDir }}

### Runtime command
CMD TARANTOOL_WORKDIR=$TARANTOOL_WORKDIR \
	TARANTOOL_PID_FILE=$TARANTOOL_PID_FILE \
	TARANTOOL_CONSOLE_SOCK=$TARANTOOL_CONSOLE_SOCK \
	tarantool {{ .AppEntrypointPath }}
`

	tmpl, err = GetRuntimeImageDockerfileTemplate(&ctx)
	assert.Nil(err)
	assert.Equal(expLayers, tmpl.Content)

	// Tarantool Opensource 1.10 w/ --from
	baseDockerfileContent := `FROM centos:7
RUN yum install -y zip
`
	writeDockerfile(f, baseDockerfileContent)

	ctx.Tarantool.TarantoolIsEnterprise = false
	ctx.Tarantool.TarantoolVersion = "1.10.42"
	ctx.Pack.DockerFrom = f.Name()

	expLayers = `FROM centos:7
RUN yum install -y zip

### Install opensource Tarantool
RUN curl -L https://tarantool.io/installer.sh | VER=1.10 bash \
    && yum -y install tarantool-devel

### Prepare for runtime
RUN echo '{{ .TmpFilesConf }}' > /usr/lib/tmpfiles.d/{{ .Name }}.conf \
    && chmod 644 /usr/lib/tmpfiles.d/{{ .Name }}.conf

USER tarantool:tarantool
ENV TARANTOOL_INSTANCE_NAME=default

### Copy application code
COPY . {{ .AppDir }}

### Runtime command
CMD TARANTOOL_WORKDIR={{ .WorkDir }} \
	TARANTOOL_PID_FILE={{ .PidFile }} \
	TARANTOOL_CONSOLE_SOCK={{ .ConsoleSock }} \
	tarantool {{ .AppEntrypointPath }}
`
}
