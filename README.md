# Cartridge Command Line Interface

[![pipeline status](https://gitlab.com/tarantool/cartridge-cli/badges/master/pipeline.svg)](https://gitlab.com/tarantool/cartridge-cli/commits/master)

## Installation

### RPM package (CentOS, Fedora)

```
# Select Tarantool version (copy one of the lines):
TARANTOOL_VERSION=1_10
TARANTOOL_VERSION=2x
TARANTOOL_VERSION=2_2

# Setup Tarantool packages repository:
curl -s https://packagecloud.io/install/repositories/tarantool/$TARANTOOL_VERSION/script.rpm.sh | sudo bash

# Install package:
sudo yum install cartridge-cli

# Check installation:
cartridge --version
```

### DEB package (Debian, Ubuntu)

```
# Select Tarantool version (copy one of the lines):
TARANTOOL_VERSION=1_10
TARANTOOL_VERSION=2x
TARANTOOL_VERSION=2_2

# Setup Tarantool packages repository:
curl -s https://packagecloud.io/install/repositories/tarantool/$TARANTOOL_VERSION/script.deb.sh | sudo bash

# Install package:
sudo apt-get install cartridge-cli

# Check installation:
cartridge --version
```

### Homebrew (MacOS)

```
brew install cartridge-cli

# Check installation:
cartridge --version
```

### From luarocks

To install cartridge-cli to the project's folder
(installed [Tarantool](https://www.tarantool.io/download/) is required):

```sh
tarantoolctl rocks install cartridge-cli
```

Executable will be available at `.rocks/bin/cartridge`.
Optionally, you can add `.rocks/bin` to the executable path:
```sh
export PATH=$PWD/.rocks/bin/:$PATH
```

If you have both global package installed `cartridge` executable will use
project-specific version installed when running from its directory.

## Usage

For more details, run
```sh
cartridge --help
```

### Applications lifecycle

Create an application from a template:

```sh
cartridge create --name myapp
```

Pack an application into a distributable:

```sh
cartridge pack rpm ./myapp
```

### Application packing details

Application can be packed by running `cartridge pack <type> <path>` command.
Now `rpm`, `deb`, `tgz`, `rock` and `docker` types of distributables are supported.

There is one important detail about `rmp`, `deb` and `tgz` packing: for this types of packages rocks modules and executables specific for the system where `cartridge pack` command is running will be delivered.
If you use `docker` packing, the result image will contain rocks modules and executables specific for the base image (`centos:8`).

#### TGZ

`cartridge pack tgz ./myapp` will create .tgz archive contains application source code and rocks modules described in application rockspec.

#### RPM and DEB

`cartridge pack rpm|deb ./myapp` will create RPM or DEB package.

In case of opensource Tarantool package has `tarantool` dependency (version >= `<major>.<minor>` and < `<major+1>`, where `<major>.<minor>` is version of Tarantool used for application packing).
You should enable Tarantool repo to allow your package manager install this dependency correctly.

After package installation:

* application code and rocks modules described in application rockspec will be placed in `/usr/share/tarantool/<app_name>` directory (for Tarantool Enterprise this directory will contain also `tarantool` and `tarantoolctl` binaries);

* unit files for running application as a `systemd` service will be delivered in `/etc/systemd/system`;

This directories will be created:

* `/etc/tarantool/conf.d/` - directory for instances configuration;
* `/var/lib/tarantool/` - directory to store instances snapshots;
* `/var/run/tarantool/` - directory to store PID-files and console sockets.

Read the [doc](https://www.tarantool.io/en/doc/2.2/book/cartridge/cartridge_dev/#deploying-an-application) to learn more about Tarantool Cartridge application deployment.

To start the `instance-1` instance of the `myapp` service:

```bash
systemctl start myapp@instance-1
```

This instance will look up its [configuration](https://www.tarantool.io/en/doc/2.2/book/cartridge/cartridge_dev/#configuring-instances) across all sections of the YAML file(s) stored in /etc/tarantool/conf.d/*.

#### Docker

`cartridge pack docker ./myapp` will build docker image and tag it as `myapp:<version>-<patch>`.

For Tarantool Enterprise you should specify download token using `--download_token` parameter or `TARANTOOL_DOWNLOAD_TOKEN` environment variable.
It's needed to download SDK on result image.

Application code will be placed in `/usr/share/tarantool/${app_name}` directory.
Opensource Tarantool will be installed on image.

Run directory is `/var/run/tarantool/${app_name}`, workdir is `/var/lib/tarantool/${app_name}`.

To start the `instance-1` instance of the `myapp` application:

```bash
docker run -d --env-file env.instance-1 \
                --name instance-1 \
                myapp:1.0.0
```

File `env.instance-1` contains instance configuration (see the [doc](https://www.tarantool.io/en/doc/2.2/book/cartridge/cartridge_dev/#configuring-instances)):

```bash
TARANTOOL_INSTANCE_NAME=instance-1
TARANTOOL_ADVERTISE_URI=3302
TARANTOOL_CLUSTER_COOKIE=secret
TARANTOOL_HTTP_PORT=8082
```

By default, `TARANTOOL_INSTANCE_NAME` is set to `default`.

To check instance logs:

```bash
docker logs instance-1
```

It's user responsibility to set up right advertise URI (`<host>:<port>`) if containers are deployed on different machines.

If user specifies only port, cartridge will use auto-detected IP, so user have to configure docker networks to set up instances communication.

You can use docker volumes to store instance snaps and xlogs on host machine.
To start image with a new application code just stop the old container and start a new one using new image.

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

It starts a `tarantool` instance with enforced env-vars.
With the `--daemonize` option it also waits until the app's main script is finished.

```
TARANTOOL_INSTANCE_NAME
TARANTOOL_CFG
TARANTOOL_PID_FILE - %run_dir%/%instance_name%.pid
TARANTOOL_CONSOLE_SOCK - %run_dir%/%instance_name%.pid
```

`cartridge.cfg()` uses `TARANTOOL_INSTANCE_NAME` to read the instance's config
from the file provided in `TARANTOOL_CFG`.

Default options for `cartridge` command can be overriden in `./.cartridge.yml` or `~/.cartridge.yml`:

```yaml
run_dir: tmp/run
cfg: cartridge.yml
apps_path: /usr/local/share/tarantool
```

When `APP_NAME` is not provided, it is parsed from `./*.rockspec` filename.
When `INSTANCE_NAME` is not provided, `cartridge` reads `cfg` file and starts all defined instances:

```
# in application directory
cartridge start # starts all instances
cartridge start .router_1 # start single instance

# in multi-application environment
cartridge start app_1 # starts all instances of app_1
cartridge start app_1.router_1 # start single instance
```

To stop one or more running instances, use:

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
