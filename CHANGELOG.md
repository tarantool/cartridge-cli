# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/en/1.0.0/)
and this project adheres to [Semantic Versioning](http://semver.org/spec/v2.0.0.html).

## [Unreleased]
<!-- Please update cartridge-cli/VERSION.lua with new release -->

## [1.3.2] - 2020-01-23

### Changed

- Common packing flow parameters are stored in the global `pack_state`

### Fixed

- Error on runnning `git clean` for submodules on `cartridge pack`

## [1.3.1] - 2020-01-20

### Added

- Allow to pass `--version` in format `major.minor.patch[-count][-hash]`

### Changed

- RPM header: added `PAYLOADDIGEST` and `PAYLOADDIGESTALGO` flags,
  removed `RPMVERSION`.

## [1.3.0] - 2020-01-13

### Added

- Packing to Docker image
- Check filemodes before packing
- `--from` option for `docker pack` command to specify base image Dockerfile path
- `cartridge.pre-build` and `cartridge.post-build` hooks
  to be ran before and after `rocks make`
- Deprecated build flow (`.cartridge.ignore` + `.cartridge.pre`) is supported
  for all distribution types except `docker`
- Recursively cleaning all submodules on application packing

### Changed

- `docker pack` log messages are coloured
- Pre-build, build and post-build actions are grouped in one RUN directive
  on packing to Docker image
- Update luatest to 0.4.0
- Freeze cartridge 2.0.0 in template

### Fixed

- Error on using environment variables in base Dockerfile
- Error on using COPY instruction in base Dockerfile
- Added missing environment variable `TARANTOOL_APP_NAME`
- Fix parsing options priority to match `cartridge.argparse`. Current parsing priority:
  firstly commandline options, then environment variables, then options from
  `.config.yml`, in the end default options. Options from `.config.yml` are
  overriden by corresponding to them environment variables.
- Error on rocks manifest processing

## [1.2.1] - 2019-11-25

- Fix building RPM package on CentOS 8
- Fix starting foreground apps with current Tarantool

## [1.2.0] - 2019-11-15

### Added

- luacheck in examples and templates
- `--version` option to display version
- Default cartridge-cli configuration in getting-started template
- Use current tarantool executable to start instance

### Changed

- Warnings in log are shown with yellow color
- `cartridge start` starts instances in foreground, `--foreground` is replaced with `--daemonize`

### Removed

- `plain` template

## [1.1.0] - 2019-10-24

### Added
- Start and stop all instances
- Start/stop instances defined in multiple files
- Colorized logs and prefix with instance name for multiple foreground instances
- Packing DEB

### Changed
- Disabled jit in tests until tarantool/tarantool#4476 is fixed
- Getting started app READMEs improved

### Fixed
- Luacheck warnings
- Missing setsearchroot in 1.10
- /var/run dir removal after reboot

## [1.0.0] - 2019-09-02

### Added

- Basic functionality
- Integration tests
- End-to-end tests
- Gitlab CI tests for opensource and enterprise Tarantool
- Packing RPM with Tarantool dependency for opensource
- Loading templates from .rocks
- Configuring systemd units using `cartrigde.argparse` way
- Getting started app
- Start and stop commands
- Cache downloaded sdk between ci jobs
