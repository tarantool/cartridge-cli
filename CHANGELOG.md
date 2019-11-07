# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/en/1.0.0/)
and this project adheres to [Semantic Versioning](http://semver.org/spec/v2.0.0.html).

## [Unreleased]
<!-- Please update cartridge-cli/VERSION.lua with new release -->

### Added

- luacheck in examples and templates
- `--version` option to display version

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
