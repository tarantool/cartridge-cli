package templates

import "github.com/tarantool/cartridge-cli/src/templates"

var buildFilesTemplate = templates.FileTreeTemplate{
	Dirs: []templates.DirTemplate{},
	Files: []templates.FileTemplate{
		{
			Path:    "cartridge.pre-build",
			Mode:    0755,
			Content: preBuildHookContent,
		},
		{
			Path:    "cartridge.post-build",
			Mode:    0755,
			Content: postBuildHookContent,
		},
		{
			Path:    "Dockerfile.build.cartridge",
			Mode:    0644,
			Content: buildDockerfileContent,
		},
		{
			Path:    "Dockerfile.cartridge",
			Mode:    0644,
			Content: runtimeDockerfileContent,
		},
	},
}

const (
	preBuildHookContent = `#!/bin/sh

# Simple pre-build script
# Will be ran before "tarantoolctl rocks make" on application build
# Could be useful to install non-standart rocks modules

# For example:
# tarantoolctl rocks make --chdir ./third_party/my-custom-rock-module
`
	postBuildHookContent = `#!/bin/sh

# Simple post-build script
# Will be ran after "tarantoolctl rocks make" on application packing
# Could be useful to remove some build artifacts from result package

# For example:
# rm -rf third_party
# rm -rf node_modules
# rm -rf doc
`
	buildDockerfileContent = `# Simple Dockerfile
# Used by "pack" command as a base for build image
# when --use-docker option is specified
#
# The base image must be centos:8
FROM centos:8

# Here you can install some packages required
#   for your application build
#
# RUN set -x \
#    && curl -sL https://rpm.nodesource.com/setup_10.x | bash - \
#    && yum -y install nodejs
`

	runtimeDockerfileContent = `# Simple Dockerfile
# Used by "pack docker" command as a base for runtime image
#
# The base image must be centos:8
FROM centos:8

# Here you can install some packages required
#   for your application in runtime
#
#
# For example, if you need to install some python packages,
#   you can do it this way:
#
# COPY requirements.txt /tmp
# RUN yum install -y python3-pip
# RUN pip3 install -r /tmp/requirements.txt
`
)
