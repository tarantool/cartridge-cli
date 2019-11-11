# Cartridge Command Line Interface

## Installation

```sh
tarantoolctl rocks install cartridge-cli
```

Optionally, you may add `.rocks/bin` to the executable path:
```sh
export PATH=$PWD/.rocks/bin/:$PATH
```

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
cartridge pack rpm myapp
```

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
