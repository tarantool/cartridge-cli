#!/usr/bin/python3

import subprocess
import pytest

from project import Project


@pytest.fixture(scope="module")
def default_project(module_tmpdir):
    project = Project('default-project', module_tmpdir, 'cartridge')
    return project


def test_project(default_project):
    project = default_project

    process = subprocess.run(['tarantoolctl', 'rocks', 'make'], cwd=project.path)
    assert process.returncode == 0, "Error building project"

    process = subprocess.run(['./deps.sh'], cwd=project.path)
    assert process.returncode == 0, "Installing deps failed"

    process = subprocess.run(['.rocks/bin/luacheck', '.'], cwd=project.path)
    assert process.returncode == 0, "luacheck failed"

    process = subprocess.run(['.rocks/bin/luatest'], cwd=project.path)
    assert process.returncode == 0, "luatest failed"
