# Cartridge Command Line Interface

[![pipeline status](https://gitlab.com/tarantool/cartridge-cli/badges/master/pipeline.svg)](https://gitlab.com/tarantool/cartridge-cli/commits/master)

# Installation

## RPM package (CentOS, Fedora)

```
# Select a Tarantool version (copy one of these lines):
TARANTOOL_VERSION=1_10
TARANTOOL_VERSION=2x
TARANTOOL_VERSION=2_2

# Set up the Tarantool packages repository:
curl -s https://packagecloud.io/install/repositories/tarantool/$TARANTOOL_VERSION/script.rpm.sh | sudo bash

# Install the package:
sudo yum install cartridge-cli

# Check the installation:
cartridge --version
```

## DEB package (Debian, Ubuntu)

```
# Select a Tarantool version (copy one of these lines):
TARANTOOL_VERSION=1_10
TARANTOOL_VERSION=2x
TARANTOOL_VERSION=2_2

# Set up the Tarantool packages repository:
curl -s https://packagecloud.io/install/repositories/tarantool/$TARANTOOL_VERSION/script.deb.sh | sudo bash

# Install the package:
sudo apt-get install cartridge-cli

# Check the installation:
cartridge --version
```

## Homebrew (MacOS)

```
brew install cartridge-cli

# Check the installation:
cartridge --version
```

## From luarocks

To install `cartridge-cli` to the project's folder
(installed [Tarantool](https://www.tarantool.io/download/) is required):

```sh
tarantoolctl rocks install cartridge-cli
```

The executable will be available at `.rocks/bin/cartridge`.
Optionally, you can add `.rocks/bin` to the executable path:
```sh
export PATH=$PWD/.rocks/bin/:$PATH
```

If you have both global packages installed, the `cartridge` executable will use
the project-specific version installed when running from its directory.

# Usage

For more details, say:
```sh
cartridge --help
```

These commands are supported:

* `create` - create a new app from template;
* `pack` - pack application into a distributable bundle;
* `build` - build application for local development;
* `start` - start a Tarantool instance(s);
* `stop` - stop a Tarantool instance(s).

# Applications lifecycle

Create an application from a template:

```sh
cartridge create --name myapp
```

Build an application:

```sh
cartridge build ./myapp
```

Run instances locally:

```sh
cartridge start
cartridge stop
```

Pack an application into a distributable, for example into an RPM package:

```sh
cartridge pack rpm ./myapp
```

# Building an application

You can call `cartridge build [<path>]` command to build application locally.
It can be useful for local development.

This command requires one argument - path to the application.
By default - it's `.` (current directory).

These steps will be performed on running this command:

* running `cartridge.pre-build` (or [DEPRECATED] `.cartridge.pre`);
* running `tarantoolctl rocks make`.

# Application packing details

An application can be packed by running the `cartridge pack <type> [<path>]` command.

These types of packages are supported: `rpm`, `deb`, `tgz`, `rock`, and
`docker`.

If `path` isn't specified, current directory is used by default.

For `rpm`, `deb`, and `tgz`, we also deliver rocks modules and executables
specific for the system where the `cartridge pack` command is running.

For `docker`, the resulting image will contain rocks modules and executables
specific for the base image (`centos:8`).

Common options:

* `--name`: name of the app to pack;
* `--version`: application version.

The result will be named as `<name>-<version>.<type>`.
By default, the application name is detected from the rockspec, and
the application version is detected from `git describe`.

## Build directory

By default, application build is performed in the temporarily directory in the
`~/.cartridge/tmp/`, so the
packaging process doesn't affect the contents of your application directory.

You can specify custom build directory for your project in `CARTRIDGE_BUILDDIR`
environment variable. If this directory doesn't exists, it will be created, used
for building the application and then removed.
**Note**, that specified directory can't be project subdirectory.

If you specify existent directory in `CARTRIDGE_BUILDDIR` environment variable,
`CARTRIDGE_BUILDDIR/build.cartridge` repository will be used for build and then
removed. This directory will be cleaned before building application.

## General packing flow and options

A package build comprises these steps:

### 1. Forming the distribution directory

On this stage, some files will be filtered out:
* First, `git clean -X -d -f` will be called to remove all untracked and
  ignored files.
* Then `.rocks` and `.git` directories will be removed.

*Note*: All application files should have at least `a+r` permissions
(`a+rx` for directories).
Otherwise, `cartridge pack` command raises an error.
Files permissions will be kept "as they are", and the code files owner will be
set to `root:root` in the resulting package.

### 2. Building an application

*Note*: When packing in docker, this stage is running in the container itself,
so all rocks dependencies will be installed correctly.
For other package types, this stage is running on the local machine, so the
resulting package will contain rocks modules and binaries specific for the local
OS.

* First, `cartridge.pre-build` script is run (if it's present).
* Then, `tarantoolctl rocks make` command is run to deliver all rocks dependencies specified in the rockspec.
  It will form the `.rocks` directory that will be delivered in the resulting
package.
* Finally, `cartridge.post-build` script is run (if it's present).

## Special files

You can place these files in your application root to control the application
packing flow (see [examples](#examples) below):

* `cartridge.pre-build`: a script to be run before `tarantoolctl rocks make`.
  The main purpose of this script is to build some non-standard rocks modules
  (for example, from a submodule).

* `cartridge.post-build`: a script to be run after `tarantoolctl rocks make`.
  The main purpose of this script is to remove build artifacts from result package.

* [DEPRECATED] `.cartridge.ignore`: here you can specify some files and directories to be
  excluded from the package build. See the
  [documentation](https://www.tarantool.io/ru/doc/1.10/book/cartridge/cartridge_dev/#using-cartridge-ignore-files)
  for details.

* [DEPRECATED] `.cartridge.pre`: a script to be run before `tarantoolctl rocks make`.
  The main purpose of this script is to build some non-standard rocks modules
  (for example, from a submodule).

*Note*: You can use any of these approaches (just take care not to mix them): `cartridge.pre-build` + `cartridge.post-build`  or deprecated `.cartridge.ignore` + `.cartridge.pre`.

*Note*: Packing to docker image isn't compatible with the deprecated packing flow.

### Special files examples

`cartridge.pre-build`:

```bash
#!/bin/sh

# The main purpose of this script is to build some non-standard rocks modules.
# It will be ran before `tarantoolctl rocks make` on application build

tarantoolctl rocks make --chdir ./third_party/my-custom-rock-module
```

`cartridge.post-build`:

```bash
#!/bin/sh

# The main purpose of this script is to remove build artifacts from result package.
# It will be ran after `tarantoolctl rocks make` on application build

rm -rf third_party
rm -rf node_modules
rm -rf doc
```

# Application packing type-specific details

## TGZ

`cartridge pack tgz ./myapp` will create a .tgz archive containing the application
source code and rocks modules described in the application rockspec.

## RPM and DEB

`cartridge pack rpm|deb ./myapp` will create an RPM or DEB package.

If you use an opensource version of Tarantool, the package has a `tarantool`
dependency (version >= `<major>.<minor>` and < `<major+1>`, where `<major>.<minor>`
is the version of Tarantool used for application packing).
You should enable the Tarantool repo to allow your package manager install
this dependency correctly.

After package installation:

* the application code and rocks modules described in the application rockspec
  will be placed in the `/usr/share/tarantool/<app_name>` directory
  (for Tarantool Enterprise, this directory will also contain `tarantool` and
  `tarantoolctl` binaries);

* unit files for running the application as a `systemd` service will be delivered
  in `/etc/systemd/system`.

These directories will be created:

* `/etc/tarantool/conf.d/` - directory for instances configuration;
* `/var/lib/tarantool/` - directory to store instances snapshots;
* `/var/run/tarantool/` - directory to store PID-files and console sockets.

Read the [doc](https://www.tarantool.io/en/doc/2.2/book/cartridge/cartridge_dev/#deploying-an-application)
to learn more about deploying a Tarantool Cartridge application.

To start the `instance-1` instance of the `myapp` service:

```bash
systemctl start myapp@instance-1
```

This instance will look for its
[configuration](https://www.tarantool.io/en/doc/2.2/book/cartridge/cartridge_dev/#configuring-instances)
across all sections of the YAML file(s) stored in `/etc/tarantool/conf.d/*`.

## Docker

`cartridge pack docker ./myapp` will build a docker image.

Specific options:

* `--tag` - resulting image tag;

* `--from` - path to the base dockerfile for runtime image 
  (default to `Dockerfile.cartridge` in the project root);

* `--build-from` - path to the base dockerfile for build image
  (default to `Dockerfile.build.cartridge` in the project root);

* `--sdk-local` - flag indicates that SDK from local machine should be 
  installed on the image;

* `--sdk-path` - path to SDK to be installed on the image
  (env `TARANTOOL_SDK_PATH`, has lower priority);


**Note**, that one and only one of `--sdk-local` and `--sdk-path` options should be specified
for Tarantool Enterprise.

### Image tag

The image is tagged as follows:

* `<name>:<detected_version>`: by default;
* `<name>:<version>`: if the `--version` parameter is specified;
* `<tag>`: if the `--tag` parameter is specified.

`<name>` can be specified in the `--name` parameter, otherwise it will be
auto-detected from the application rockspec.

### Tarantool Enterprise SDK

If you use Tarantool Enterprise, you should explicitly specify Tarantool SDK to
be delivered on the result image.
If you want to use SDK from your local machine, just pass `--sdk-local` flag to
`cartridge pack docker` command.
You can specify local path to the other SDK using `--sdk-path` option.
(or env `TARANTOOL_SDK_PATH`).

### Build and runtime images

In fact, two images are created - build image and runtime image.
Build image is used to perform application build.
Then, application files are delivered to the runtime image (that is exactly the
result of running `cartridge pack docker`).

Both images are created from `centos:8`.
On build image all packages required for the default  `cartridge` application build
(`git`, `gcc`, `make`, `cmake`, `unzip`) are installed.
On the runtime image opensource Tarantool is installed if required.

If your application requires some other applications for build or runtime, you
can specify base layers for build and runtime images:

* build image: `Dockerfile.build.cartridge` or `--build-from`;
* runtime image: `Dockerfile.cartridge` or `--from`.

The base image dockerfile should be started with the `FROM centos:8` line
(except comments).

For example, if your application requires `gcc-c++` for build and `zip` for
runtime:

`Dockerfile.cartridge.build`:

```dockerfile
FROM centos:8
RUN yum install -y gcc-c++
# Note, that git, gcc, make, cmake, unzip packages
# will be installed anyway
```

`Dockerfile.cartridge`:

```dockerfile
FROM centos:8
RUN yum install -y zip
```

### Building the app

If you want the `docker build` command to be run with custom arguments, you can
specify them using the `TARANTOOL_DOCKER_BUILD_ARGS` environment variable.
For example, `TARANTOOL_DOCKER_BUILD_ARGS='--no-cache --quiet'`

### Using the result image

The application code will be placed in the `/usr/share/tarantool/${app_name}`
directory. An opensource version of Tarantool will be installed to the image.

The run directory is `/var/run/tarantool/${app_name}`,
the workdir is `/var/lib/tarantool/${app_name}`.

To start the `instance-1` instance of the `myapp` application, say:

```bash
docker run -d \
                --name instance-1 \
                -e TARANTOOL_INSTANCE_NAME=instance-1 \
                -e TARANTOOL_ADVERTISE_URI=3302 \
                -e TARANTOOL_CLUSTER_COOKIE=secret \
                -e TARANTOOL_HTTP_PORT=8082 \
                myapp:1.0.0
```

By default, `TARANTOOL_INSTANCE_NAME` is set to `default`.

To check the instance logs:

```bash
docker logs instance-1
```

It is the user's responsibility to set up a proper advertise URI (`<host>:<port>`)
if the containers are deployed on different machines.

If the user specifies only a port, `cartridge` will use an auto-detected IP,
so the user needs to configure docker networks to set up inter-instance communication.

You can use docker volumes to store instance snapshots and xlogs on the host machine.
To start an image with a new application code, just stop the old container and
start a new one using the new image.

# Managing instances

```
cartridge start [APP_NAME[.INSTANCE_NAME]] [options]

Options
    --script FILE       Application's entry point.
                        Defaults to TARANTOOL_SCRIPT,
                        or ./init.lua when running from the app's directory,
                        or :apps_path/:app_name/init.lua in a multi-app env.

    --apps-path PATH    Path to apps directory when running in a multi-app env.
                        Default to /usr/share/tarantool

    --run-dir DIR       Directory with pid and sock files.
                        Defaults to TARANTOOL_RUN_DIR or /var/run/tarantool

    --cfg FILE          Cartridge instances config file.
                        Defaults to TARANTOOL_CFG or ./instances.yml

    --daemonize / -d    Start in background
```

It starts a `tarantool` instance with enforced environment variables.

With the `--daemonize` option, it also waits until the app's main script is finished.

```
TARANTOOL_INSTANCE_NAME
TARANTOOL_CFG
TARANTOOL_PID_FILE - %run_dir%/%instance_name%.pid
TARANTOOL_CONSOLE_SOCK - %run_dir%/%instance_name%.pid
```

`cartridge.cfg()` uses `TARANTOOL_INSTANCE_NAME` to read the instance's configuration
from the file provided in `TARANTOOL_CFG`.

Default options for the `cartridge` command can be overridden in
`./.cartridge.yml` or `~/.cartridge.yml`, also options from `.cartridge.yml`
can be overriden by corresponding to them environment variables `TARANTOOL_*`.

Here is an example content of `.config.yml`:

```yaml
run_dir: tmp/run
cfg: cartridge.yml
apps_path: /usr/local/share/tarantool
script: init.lua
```

When `APP_NAME` is not provided, it is parsed from the `./*.rockspec` filename.

When `INSTANCE_NAME` is not provided, `cartridge` reads the `cfg` file and starts
all defined instances:

```
# in the application directory
cartridge start # starts all instances
cartridge start .router_1 # start single instance

# in a multi-application environment
cartridge start app_1 # starts all instances of app_1
cartridge start app_1.router_1 # start single instance
```

To stop one or more running instances, say:

```
cartridge stop [APP_NAME[.INSTANCE_NAME]] [options]

These options from `start` command are supported
    --run-dir DIR
    --cfg FILE
```

# Misc

## Running end-to-end tests

```sh
vagrant up

# Centos
vagrant ssh centos < test/e2e/start-rpm.sh
vagrant ssh centos < test/e2e/test-cluster.sh
vagrant reload centos
sleep 1
vagrant ssh centos < test/e2e/test-cluster.sh
vagrant ssh centos < test/e2e/cleanup.sh

# Ubuntu
vagrant ssh ubuntu < test/e2e/start-deb.sh
vagrant ssh ubuntu < test/e2e/test-cluster.sh
vagrant reload ubuntu
sleep 1
vagrant ssh ubuntu < test/e2e/test-cluster.sh
vagrant ssh ubuntu < test/e2e/cleanup.sh

vagrant halt
```
