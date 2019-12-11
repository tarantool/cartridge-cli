# Cartridge Command Line Interface

[![pipeline status](https://gitlab.com/tarantool/cartridge-cli/badges/master/pipeline.svg)](https://gitlab.com/tarantool/cartridge-cli/commits/master)

## Installation

### RPM package (CentOS, Fedora)

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

### DEB package (Debian, Ubuntu)

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

### Homebrew (MacOS)

```
brew install cartridge-cli

# Check the installation:
cartridge --version
```

### From luarocks

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

## Usage

For more details, say:
```sh
cartridge --help
```

### Applications lifecycle

Create an application from a template:

```sh
cartridge create --name myapp
```

Pack an application into a distributable, for example into an RPM package:

```sh
cartridge pack rpm ./myapp
```

### Application packing details

An application can be packed by running the `cartridge pack <type> <path>` command.

These types of distributables are supported: `rpm`, `deb`, `tgz`, `rock`, and
`docker`.

For `rmp`, `deb`, and `tgz`, we also deliver rocks modules and executables
specific for the system where the `cartridge pack` command is running.

For `docker`, the resulting image will contain rocks modules and executables
specific for the base image (`centos:8`).

#### TGZ

`cartridge pack tgz ./myapp` will create a .tgz archive containing the application
source code and rocks modules described in the application rockspec.

#### RPM and DEB

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

#### Docker

`cartridge pack docker ./myapp` will build a docker image.

The image is tagged as follows:
* `<name>:<detected_version>`: by default;
* `<name>:<version>`: if the `--version` parameter is specified;
* `<tag>`: if the `--tag` parameter is specified.

`<name>` can be specified in the `--name` parameter, otherwise it will be
auto-detected from application rockspec.

For Tarantool Enterprise, you should specify a download token using the
`--download_token` parameter or the `TARANTOOL_DOWNLOAD_TOKEN` environment
variable. It's needed to download the SDK to the resulting image.

If you want the `docker build` command to be run with custom arguments, you can
specify them using the `TARANTOOL_DOCKER_BUILD_ARGS` environment variable.
For example, `TARANTOOL_DOCKER_BUILD_ARGS='--no-cache --quiet'`

The Application code will be placed in the `/usr/share/tarantool/${app_name}`
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

### Managing instances

```
cartridge start [APP_NAME[.INSTANCE_NAME]] [options]

Options
    --script FILE       Application's entry point.
                        Defaults to TARANTOOL_SCRIPT,
                        or ./init.lua when running from the app's directory,
                        or :apps_path/:app_name/init.lua in a multi-app env.

    --apps_path PATH    Path to apps directory when running in a multi-app env.
                        Default to /usr/share/tarantool

    --run_dir DIR       Directory with pid and sock files.
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
`./.cartridge.yml` or `~/.cartridge.yml`:

```yaml
run_dir: tmp/run
cfg: cartridge.yml
apps_path: /usr/local/share/tarantool
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
    --run_dir DIR
    --cfg FILE
```

## Misc

### Running end-to-end tests

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
