# Cartridge Command Line Interface

## Installation

```sh
tarantoolctl rocks install cartridge-cli
```

Optionally you may add `.rocks/bin` to executable path:
```sh
export PATH=$PWD/.rocks/bin/:$PATH
```

## Usage

For more details, run 
```sh
cartridge --help
```

### Applications lifecycle

Create an application from template:

```sh
cartridge create --name myapp
```

Pack an application into distributable:

```sh
cartridge pack rpm myapp
```

### Managing instances

```
cartridge start INSTANCE_NAME [options]

Options
    --script FILE   Application's entry point. Default to ./init.lua
    --run_dir DIR   Directory with pid and sock files. Default to /var/run/tarantool
    --cfg FILE      Cartridge instances config file. Default to ./instances.yml
    --foreground    Do not daemonize
```

It starts tarantool instance in background with enforced env-vars and
waits until app's main script is finished.

```
TARANTOOL_INSTANCE_NAME
TARANTOOL_CFG
TARANTOOL_PID_FILE - %run_dir%/%instance_name%.pid
TARANTOOL_CONSOLE_SOCK - %run_dir%/%instance_name%.pid
```

`cartridge.cfg()` uses `TARANTOOL_INSTANCE_NAME` to read instance's config
from file provided in `TARANTOOL_CFG`.

Default options for `cartridge` command can be overriden in `./.cartridge.yml` or `~/.cartridge.yml`:

```yaml
run_dir: tmp/run
cfg: cartrifge.yml
```

To stop running instance pass it's name to:

```
cartridge stop INSTANCE_NAME
```

## Misc

### Running end-to-end tests

```sh
vagrant up
vagrant ssh 1_10 < test/end-to-end.sh
vagrant halt
```
