#!/usr/bin/python3

import os
import subprocess

__tarantool_version = None

basepath = os.path.realpath(
    os.path.join(os.path.dirname(__file__), '..', '..')
)

def tarantool_version():
    global __tarantool_version
    if __tarantool_version is None:
        __tarantool_version = subprocess.check_output(['tarantool', '-V']).decode('ascii').split('\n')[0]

    return __tarantool_version


def tarantool_enterprise_is_used():
    return tarantool_version().startswith('Tarantool Enterprise')

def create_project(module_tmpdir, project_name, template):
    cmd = [os.path.join(basepath, "cartridge"), "create",
        "--name", project_name,
        "--template", template]
    process = subprocess.run(cmd, cwd=module_tmpdir)
    assert process.returncode == 0, \
        "Error during creating the project"
    return os.path.join(module_tmpdir, project_name)
