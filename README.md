# Cartridge Command Line Interface

[![pipeline status](https://gitlab.com/tarantool/cartridge-cli/badges/master/pipeline.svg)](https://gitlab.com/tarantool/cartridge-cli/commits/master)

## Contents

* [Installation](#installation)
  * [RPM package (CentOS, Fedora, ALT Linux)](#rpm-package-centos-fedora-alt-linux)
  * [DEB package (Debian, Ubuntu)](#deb-package-debian-ubuntu))
  * [Homebrew (MacOS)](#homebrew-macos)
  * [From luarocks](#from-luarocks)
* [Quick start](#quick-start)
* [Usage](#usage)
  * [An application's lifecycle](#an-applications-lifecycle)
  * [Creating an application from template](#creating-an-application-from-template)
  * [Building an application](#building-an-application)
  * [Starting/stopping an application locally](#startingstopping-an-application-locally)
  * [Packing an application](#packing-an-application)
  * [TGZ](#tgz)
  * [RPM and DEB](#rpm-and-deb)
  * [Docker](#docker)
  * [Special files](#special-files)
* [Misc](#misc)
  * [Running end-to-end tests](#running-end-to-end-tests)

## Installation

### RPM package (CentOS, Fedora, ALT Linux)

```sh
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

Now you can create and start your first application!
Go to the [quick start](#quick-start) section and try it.

### DEB package (Debian, Ubuntu)

```sh
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

Now you can create and start your first application!
Go to the [quick start](#quick-start) section and try it.

### Homebrew (MacOS)

```sh
brew install cartridge-cli

# Check the installation:
cartridge --version
```

Now you can create and start your first application!
Go to the [quick start](#quick-start) section and try it.

### From luarocks

To install `cartridge-cli` to the application's directory
(installed [Tarantool](https://www.tarantool.io/download/) is required):

```sh
tarantoolctl rocks install cartridge-cli
```

The executable will be available at `.rocks/bin/cartridge`.
Optionally, you can add `.rocks/bin` to the executable path:
```sh
export PATH=$PWD/.rocks/bin/:$PATH
```

Now you can create and start your first application!
Go to the [next](#quick-start) section and try it.

## Quick start

To create your first application:

```sh
cartridge create --name myapp
```

Let's go inside:

```sh
cd myapp
```

Now build the application and start it:

```sh
cartridge build
cartridge start
```

That's all! You can visit http://localhost:8081 and see your application Admin Web UI:

<img width="640" alt="cartridge-ui" src="https://user-images.githubusercontent.com/11336358/75786427-52820c00-5d76-11ea-93a4-309623bda70f.png">

You can find more details in the [documentation](#usage) or start with our
[getting started guide](https://github.com/tarantool/cartridge-cli/blob/master/examples/getting-started-app/README.md).

## Usage

For more details, say:
```sh
cartridge --help
```

These commands are supported:

* `create` - create a new application from template;
* `build` - build the application for local development and testing;
* `start` - start a Tarantool instance(s);
* `stop` - stop a Tarantool instance(s);
* `pack` - pack the application into a distributable bundle.

### An application's lifecycle

In a nutshell:

1. [Create](#creating-an-application-from-template) an application
   (e.g. `myapp`) from template:

   ```sh
   cartridge create --name myapp
   cd ./myapp
   ```

2. [Build](#building-an-application) the application for local development
   and [testing](#running-end-to-end-tests):

   ```sh
   cartridge build
   ```

3. [Run](#startingstopping-an-application-locally) instances locally:

   ```sh
   cartridge start
   cartridge stop
   ```

4. [Pack](#packing-an-application) the application into a distributable
   (e.g. into an RPM package):

   ```sh
   cartridge pack rpm
   ```

### Creating an application from template

To create an application from the Cartridge template, say this in any directory:

```sh
cartridge create --name <app_name> /path/to/
```

This will create a simple Cartridge application in the `/path/to/<app_name>/`
directory with:

* one custom role with an HTTP endpoint;
* sample tests and basic test helpers;
* files required for development (like `.luacheckrc`).

If you have `git` installed, this will also set up a Git repository with the
initial commit, tag it with
[version](https://www.tarantool.io/en/doc/latest/book/cartridge/cartridge_dev/#application-versioning)
0.1.0, and add a `.gitignore` file to the project root.

Let's take a closer look at the files inside the `<app_name>/` directory:

* application files:
  * `app/roles/custom-role.lua` a sample
    [custom role](https://www.tarantool.io/en/doc/latest/book/cartridge/cartridge_dev/#cluster-roles)
    with simple HTTP API; can be enabled as `app.roles.custom`
  * `<app_name>-scm-1.rockspec` file where you can specify application
    dependencies
  * `init.lua` file which is the entry point for your application
* [special files](#special-files) (used to build and pack the application):
  * `cartridge.pre-build`
  * `cartridge.post-build`
  * `Dockerfile.build.cartridge`
  * `Dockerfile.cartridge`
* development files:
  * `deps.sh` script that resolves the dependencies from the `.rockspec` file
    and installs test dependencies (like `luatest`)
  * `instances.yml` file with instances configuration (used by `cartridge start`)
  * `.cartridge.yml` file with Cartridge configuration (used by `cartridge start`)
  * `tmp` directory for temporary files (used as a run dir, see `.cartridge.yml`)
  * `.git` file necessary for a Git repository
  * `.gitignore` file where you can specify the files for Git to ignore
  * `env.lua` file that sets common rock paths so that the application can be
    started from any directory.
* test files (with sample tests):
  ```
  test
  ├── helper
  │   ├── integration.lua
  │   └── unit.lua
  │   ├── helper.lua
  │   ├── integration
  │   │   └── api_test.lua
  │   └── unit
  │       └── sample_test.lua
  ```
* configuration files:
  * `.luacheckrc`
  * `.luacov`
  * `.editorconfig`

### Building an application

#### Building locally

To build your application locally (for local testing), say this in any directory:

```sh
cartridge build [<path>]
```

This command requires one argument -- the path to your application directory
(i.e. to the build source). The default path is `.` (the current directory).

This command runs:

1. `cartridge.pre-build` (or [DEPRECATED] `.cartridge.pre`), if the
   [pre-build file](#special-files) exists.
   This builds the application in the `path` directory.
2. `tarantoolctl rocks make`, if the [rockspec file](#special-files) exists.
   This installs all Lua rocks to the `path` directory.

During step 1 of the `cartridge build` command, `cartridge` builds the application
inside the application directory -- unlike when building the application as part
of the `cartridge pack` command, when the application is built in a temporary
[build directory](#build-directory) and no build artifacts remain in the
application directory.

During step 2 -- the key step here -- `cartridge` installs all dependencies
specified in the rockspec file (you can find this file within the application
directory created from template).

  **NOTE:** An advanced alternative would be to specify build logic in the
  rockspec as `cmake` commands, like we
  [do it](https://github.com/tarantool/cartridge/blob/master/cartridge-scm-1.rockspec#L26).
  for `cartridge`.

If your application depends on closed-source rocks, or if the build should contain
rocks from a project added as a submodule, then you need to **install** all these
dependencies before calling `tarantoolctl rocks make`.
You can do it using the file `cartridge.pre-build` in your application root
(again, you can find this file within the application directory created from template).
In this file, you can specify all rocks to build
(e.g. `tarantoolctl rocks make --chdir ./third_party/proj`).
For details, see [special files](#special-files).

As a result, in the application's `.rocks` directory you will get a fully built
application that you can start locally from the application's directory.

#### Building in Docker

By default, `cartridge build` is building an application locally.

However, if you build it in OS X, all rocks and executables in the resulting
package will be specific for OS X, so the application won't work in Linux.
To build an application in OS X and run it in Linux, call `cartridge build`
with the flag `--use-docker` and get the application built in a Docker container.

This image is created similarly to the [build image](#build-and-runtime-images)
created during `cartridge pack`.

### Starting/stopping an application locally

Now that the application is [built](#building-an-application), you can
run it locally:

```sh
cartridge start [APP_NAME[.INSTANCE_NAME]] [options]
```

The options are:

* `--script FILE` is the application's entry point. Defaults to:

  * TARANTOOL_SCRIPT,
  * or `./init.lua` when running from the app's directory,
  * or `:apps_path/:app_name/init.lua` in a multi-app environment.

* `--apps-path PATH` is the path to the application directory when running
  in a multi-app environment. Defaults to `/usr/share/tarantool`.

* `--run-dir DIR` is the directory with pid and sock files.
  Defaults to TARANTOOL_RUN_DIR or `/var/run/tarantool`.

* `--cfg FILE` is the configuration file for Cartridge instances.
  Defaults to TARANTOOL_CFG or `./instances.yml`.

* `--daemonize / -d` starts the instance in background.
  With this option, Tarantool also waits until the app's main script is finished.
  For example, this is useful if `init.lua` requires time-consuming startup from
  snapshot, and Tarantool waits for the startup to complete.
  This is also useful if the app's main script generates errors, and Tarantool
  can handle them.

The `cartridge start` command starts a `tarantool` instance with enforced
**environment variables**:

```sh
TARANTOOL_INSTANCE_NAME
TARANTOOL_CFG
TARANTOOL_PID_FILE - %run_dir%/%instance_name%.pid
TARANTOOL_CONSOLE_SOCK - %run_dir%/%instance_name%.sock
```

`cartridge.cfg()` uses `TARANTOOL_INSTANCE_NAME` to read the instance's configuration
from the file provided in `TARANTOOL_CFG`.

You can override default options for the `cartridge` command in
`./.cartridge.yml` or `~/.cartridge.yml`.

You can also override `.cartridge.yml` options
in corresponding environment variables (`TARANTOOL_*`).

Here is an example of `.cartridge.yml`:

```yaml
run_dir: tmp/run
cfg: cartridge.yml
apps_path: /usr/local/share/tarantool
script: init.lua
```

When `APP_NAME` is not provided, it is parsed from the `./*.rockspec` filename.

When `INSTANCE_NAME` is not provided, `cartridge` reads the `cfg` file and starts
all defined instances:

```sh
# in the application directory
cartridge start # starts all instances
cartridge start .router_1 # start single instance

# in a multi-application environment
cartridge start app_1 # starts all instances of app_1
cartridge start app_1.router_1 # start single instance
```

To stop one or more running instances, say:

```sh
cartridge stop [APP_NAME[.INSTANCE_NAME]] [options]
```

These options from the `start` command are supported:

* `--run-dir DIR`
* `--cfg FILE`

### Packing an application

To pack your application, say this in any directory:

```sh
cartridge pack [options] <type> [<path>]
```

where:

* `type` [REQUIRED] is the distribution type. The supported types are:
  `rpm`, `tgz`, `docker`, `deb`. See details below.

* `path` [OPTIONAL] is the path to the application directory to pack.
  Defaults to `.` (the current directory).

The options are:

* `--name`(common for all distribution types) is the application name.
  It coincides with the package name and the systemd-service name.
  The default name comes from the `package` field in the rockspec file.

* `--version` (common for all distribution types) is the application's package
  version. The expected pattern is `major.minor.patch[-count][-commit]`:
  if you specify `major.minor.patch`, it is normalized to `major.minor.patch-count`.
  The default version is determined as the result of `git describe --tags --long`.
  If the application is not a git repository, you need to set the `--version` option
  explicitly.

* `--suffix` (common for all distribution types) is the result file (or image)
  name suffix.

* `--unit-template` (used for `rpm` и `deb`) is the path to the template for
  the `systemd` unit file.

* `--instantiated-unit-template` (used for `rpm` и `deb`) is the path to the
  template for the `systemd` instantiated unit file.

* `--from` (used for `docker`) is the path to the base Dockerfile of the runtime
  image. Defaults to `Dockerfile.cartridge` in the application root.

* `--use-docker` (ignored for `docker`) forces to build the application in Docker.

* `--download-token` (common for all distribution types) is the download token
  for Tarantool Enterprise. Alternatively, you can pass the token via the
  `TARANTOOL_DOWNLOAD_TOKEN` environment variable (this variable is of lower
  priority).

* `--tag` (used for `docker`) is the tag of the Docker image that results from
  `pack docker`.

* `--build-from` (used for `docker`) is the path to the base Dockerfile of the
  build image. Defaults to `Dockerfile.build.cartridge` in the application root.

* `--sdk-local` (used for `docker`) is a flag that indicates if the SDK from
  the local machine should be installed to the image.

* `--sdk-path` (used for `docker`) is the path to the SDK to be installed to
  the image. Alternatively, you can pass the path via the `TARANTOOL_SDK_PATH`
  environment variable (this variable is of lower priority).

**Note:** For Tarantool Enterprise, you must specify one (and only one)
of the `--sdk-local` and `--sdk-path` options.

For `rpm`, `deb`, and `tgz`, we also deliver rocks modules and executables
specific for the system where the `cartridge pack` command is running.

For `docker`, the resulting runtime image will contain rocks modules
and executables specific for the base image (`centos:8`).

The result will be named as `<name>-<version>.<type>`.

Further on we dive deeper into the packaging process.

#### Build directory

The first step of the packaging process is to
[build the application](#building-an-application).

By default, application build is done in a temporary directory in
`~/.cartridge/tmp/`, so the packaging process doesn't affect the contents
of your application directory.

You can specify a custom build directory for your application in the `CARTRIDGE_BUILDDIR`
environment variable. If this directory doesn't exists, it will be created, used
for building the application, and then removed.

If you specify an existing directory in the `CARTRIDGE_BUILDDIR` environment
variable, the `CARTRIDGE_BUILDDIR/build.cartridge` repository will be used for
build and then removed. This directory will be cleaned up before building the
application.

  **NOTE:** The specified directory cannot be an application subdirectory.

#### Distribution directory

For each distribution type, a temporary directory with application source files
is created (further on we address it as *application directory*).
This includes 3 stages.

##### Stage 1. Cleaning up the application directory

On this stage, some files are filtered out of the application directory:

* First, `git clean -X -d -f` removes all untracked and
  ignored files (it works for submodules, too).
* After that, `.rocks` and `.git` directories are removed.

Files permissions are preserved, and the code files owner is set to
`root:root` in the resulting package.

  **Note**: All application files should have at least `a+r` permissions
  (`a+rx` for directories).
  Otherwise, `cartridge pack` command raises an error.

##### Stage 2. Building the application

On this stage, `cartridge` [builds](#building-an-application) the application
in the cleaned up application directory.

##### Stage 3. Cleaning up the files before packing

On this stage, `cartridge` runs `cartridge.post-build` (if it exists) to remove
junk files (like `node_modules`) generated during application build.

See an [example](#example-cartridgepost-build) in [special files](#special-files).

### TGZ

`cartridge pack tgz ./myapp` creates a .tgz archive. It contains all files from the
[distribution directory](#distribution-directory)
(i.e. the application source code and rocks modules described in the application
rockspec).

The result artifact name is `<name>-<version>[-<suffix>].tar.gz`.

### RPM and DEB

`cartridge pack rpm|deb ./myapp` creates an RPM or DEB package.

The result artifact name is `<name>-<version>[-<suffix>].{rpm,deb}`.

#### Usage example

After package installation you need to specify configuration for instances to start.

For example, if your application is named `myapp` and you want to start two instances,
place a `myapp.yml` file in `/etc/tarantool/conf.d` directory.

```yaml
myapp:
  cluster_cookie: secret-cookie

myapp.instance-1:
  http_port: 8081
  advertise_uri: localhost:3301

myapp.instance-2:
  http_port: 8082
  advertise_uri: localhost:3302
```

More details about instances configuration you can find in the
[documentation](https://www.tarantool.io/en/doc/2.2/book/cartridge/cartridge_dev/#configuring-instances).

Now, start configured instances:

```bash
systemctl start myapp@instance-1
systemctl start myapp@instance-2
```

#### Package details

The installed package name will be `<name>` no matter what the artifact name is.

It contains meta information: the package name (which is the application name),
and the package version.

If you use an opensource version of Tarantool, the package has a `tarantool`
dependency (version >= `<major>.<minor>` and < `<major+1>`, where `<major>.<minor>`
is the version of Tarantool used for packing the application).
You should enable the Tarantool repo to allow your package manager install
this dependency correctly:

* for RPM:

  ```sh
  curl -s \
          https://packagecloud.io/install/repositories/tarantool/${tarantool_repo_version}/script.rpm.sh | bash \
      && yum -y install tarantool tarantool-devel
  ```

* for DEB:

  ```sh
  curl -s \
          https://packagecloud.io/install/repositories/tarantool/${tarantool_repo_version}/script.deb.sh | bash \
      && apt-get -y install tarantool
  ```

The package contents is as follows:

* the contents of the distribution directory, placed in the
  `/usr/share/tarantool/<app_name>` directory
  (for Tarantool Enterprise, this directory also contains `tarantool` and
  `tarantoolctl` binaries);

* unit files for running the application as a `systemd` service:
  `/etc/systemd/system/${name}.service` and `/etc/systemd/system/${name}@.service`;

* the file `/usr/lib/tmpfiles.d/<name>.conf` that allows the instance to restart
  after server restart.

These directories are created:

* `/etc/tarantool/conf.d/` - directory for instances configuration;
* `/var/lib/tarantool/` - directory to store instances snapshots;
* `/var/run/tarantool/` - directory to store PID-files and console sockets.

See the [manual](https://www.tarantool.io/en/doc/2.2/book/cartridge/cartridge_dev/#deploying-an-application)
for details about deploying a Tarantool Cartridge application.

To start the `instance-1` instance of the `myapp` service, say:

```bash
systemctl start myapp@instance-1
```

This instance will look for its
[configuration](https://www.tarantool.io/en/doc/latest/book/cartridge/cartridge_dev/#configuring-instances)
across all sections of the YAML file(s) stored in `/etc/tarantool/conf.d/*`.

Use the options `--unit-template` and `--instantiated-unit-template` to customize
standard unit files.

  **NOTE:** You may need it first of all for DEB packages, if your build platform
  is different from the deployment platform. In this case, `ExecStartPre` may
  contain an incorrect path to `mkdir`. As a hotfix, we suggest customizing the
  unit files.

Example of an instantiated unit file:

```conf
[Unit]
Description=Tarantool Cartridge app ${name}@%i
After=network.target

[Service]
Type=simple
ExecStartPre=/bin/sh -c 'mkdir -p ${workdir}.default'
ExecStart=${bindir}/tarantool ${dir}/init.lua
User=tarantool
Group=tarantool

Environment=TARANTOOL_WORKDIR=${workdir}.%i
Environment=TARANTOOL_CFG=/etc/tarantool/conf.d/
Environment=TARANTOOL_PID_FILE=/var/run/tarantool/${name}.%i.pid
Environment=TARANTOOL_CONSOLE_SOCK=/var/run/tarantool/${name}.%i.control
Environment=TARANTOOL_INSTANCE_NAME=%i
```

In this file, you can use the following environment variables:

* `name` - the application name;
* `workdir` - path to the work directory (by default, `/var/lib/tarantool/<name>`);

### Docker

`cartridge pack docker ./myapp` builds a Docker image where you can start
one instance of the application.

#### Build and runtime images

In fact, two images are created during the packing process:
build image and runtime image.

First, the build image is used to perform application build.
The build stages here are exactly the same as for other distribution types:

* [Stage 1. Cleaning up the application directory](#stage-1-cleaning-up-the-application-directory)
* [Stage 2. Building the application](#stage-2-building-the-application)
  (the build is always done [in Docker](#building-in-docker))
* [Stage 3. Cleaning up the files before packaging](#stage-3-cleaning-up-the-files-before-packing)

Second, the files are copied to the resulting (runtime) image, similarly
to packing an application as an archive. This image is exactly the
result of running `cartridge pack docker`).

Both images are based on `centos:8`.

All packages required for the default  `cartridge` application build
(`git`, `gcc`, `make`, `cmake`, `unzip`) are installed on the build image.

A proper version of Tarantool is provided on the runtime image:

* For opensource, Tarantool of the same version as the one used for
  local development is installed to the image.
* For Tarantool Enterprise, the bundle with Tarantool Enterprise binaries is
  copied to the image.

If your application requires some other applications for build or runtime, you
can specify base layers for build and runtime images:

* build image: `Dockerfile.build.cartridge` (default) or `--build-from`;
* runtime image: `Dockerfile.cartridge` (default) or `--from`.

The Dockerfile of the base image should be started with the `FROM centos:8` line
(except comments).

For example, if your application requires `gcc-c++` for build and `zip` for
runtime, customize the Dockerfiles as follows:

* `Dockerfile.cartridge.build`:

  ```dockerfile
  FROM centos:8
  RUN yum install -y gcc-c++
  # Note that git, gcc, make, cmake, unzip packages
  # will be installed anyway
  ```

* `Dockerfile.cartridge`:

  ```dockerfile
  FROM centos:8
  RUN yum install -y zip
  ```

#### Runtime image tag

The runtime image is tagged as follows:

* `<name>:<detected_version>[-<suffix>]`: by default;
* `<name>:<version>[-<suffix>]`: if the `--version` parameter is specified;
* `<tag>`: if the `--tag` parameter is specified.

#### Tarantool Enterprise SDK

If you use Tarantool Enterprise, you should explicitly specify the Tarantool SDK
to be delivered on the runtime image.

If you want to use the SDK from your local machine, just pass the `--sdk-local`
flag to the `cartridge pack docker` command.

Alternatively, you can specify a local path to another SDK using the `--sdk-path`
option (or the environment variable `TARANTOOL_SDK_PATH`, which has lower priority).

#### Customizing the application build in Docker

You can pass custom arguments for the `docker build` command via the
`TARANTOOL_DOCKER_BUILD_ARGS` environment variable.
For example, `TARANTOOL_DOCKER_BUILD_ARGS='--no-cache --quiet'`

#### Using the runtime image

The application code is placed in the `/usr/share/tarantool/${app_name}`
directory. An opensource version of Tarantool is installed to the image.

The run directory is `/var/run/tarantool/${app_name}`,
the workdir is `/var/lib/tarantool/${app_name}`.

The runtime image also contains the file `/usr/lib/tmpfiles.d/<name>.conf`
that allows the instance to restart after container restart.

To start the `instance-1` instance of the `myapp` application, say:

```bash
docker run -d \
                --name instance-1 \
                -e TARANTOOL_INSTANCE_NAME=instance-1 \
                -e TARANTOOL_ADVERTISE_URI=3302 \
                -e TARANTOOL_CLUSTER_COOKIE=secret \
                -e TARANTOOL_HTTP_PORT=8082 \
                -p 127.0.0.1:8082:8082 \
                myapp:1.0.0
```

By default, `TARANTOOL_INSTANCE_NAME` is set to `default`.

To check the instance logs, say:

```bash
docker logs instance-1
```

It is the user's responsibility to set up a proper advertise URI (`<host>:<port>`)
if the containers are deployed on different machines.
The problem here is that an instance's advertise URI must be the same on all
machines, because it will be used by all the other instances to connect to this
one. For example, if you start an instance with an advertise URI set to
`localhost:3302`, and then address it as `<instance-host>:3302` from other
instances, this won't work: the other instances will be recognizing it only as
`localhost:3302`.

If you specify only a port, `cartridge` will use an auto-detected IP,
so you need to configure Docker networks to set up inter-instance communication.

You can use Docker volumes to store instance snapshots and xlogs on the
host machine. To start an image with a new application code, just stop the
old container and start a new one using the new image.

### Special files

You can put these files in your application root to control the application
packaging process (see examples below):

* `cartridge.pre-build`: a script to be run before `tarantoolctl rocks make`.
  The main purpose of this script is to build some non-standard rocks modules
  (for example, from a submodule).

* `cartridge.post-build`: a script to be run after `tarantoolctl rocks make`.
  The main purpose of this script is to remove build artifacts from result package.

* [DEPRECATED] `.cartridge.ignore`: here you can specify some files and directories to be
  excluded from the package build. See the
  [manual](https://www.tarantool.io/ru/doc/latest/book/cartridge/cartridge_dev/#using-cartridge-ignore-files)
  for details.

* [DEPRECATED] `.cartridge.pre`: a script to be run before `tarantoolctl rocks make`.
  The main purpose of this script is to build some non-standard rocks modules
  (for example, from a submodule).

**NOTES**:

* You can use any of these approaches (just take care not to mix them):

  * `cartridge.pre-build` + `cartridge.post-build`, or
  * [deprecated] `.cartridge.ignore` + `.cartridge.pre`.

* Packing to a Docker image isn't compatible with the deprecated packaging process.

#### Example: cartridge.pre-build

```bash
#!/bin/sh

# The main purpose of this script is to build some non-standard rocks modules.
# It will be ran before `tarantoolctl rocks make` on application build

tarantoolctl rocks make --chdir ./third_party/my-custom-rock-module
```

#### Example: cartridge.post-build

```bash
#!/bin/sh

# The main purpose of this script is to remove build artifacts from resulting package.
# It will be ran after `tarantoolctl rocks make` on application build.

rm -rf third_party
rm -rf node_modules
rm -rf doc
```
