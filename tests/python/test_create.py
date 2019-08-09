#!/usr/bin/python3

import subprocess
import os

from utils import tarantool_enterprise_is_used


def test_project(project_path):
    process = subprocess.run(['tarantoolctl', 'rocks', 'make'], cwd=project_path)
    assert process.returncode == 0, \
        "Error building project"
    if tarantool_enterprise_is_used():
        process = subprocess.run(['tarantoolctl', 'rocks', 'test'], cwd=project_path)
        assert process.returncode == 0, \
            "Error testing project"


def test_rocks(tmpdir):
    base_dir = os.path.realpath(
        os.path.join(os.path.dirname(__file__), '..', '..')
    )
    process = subprocess.run(['tarantoolctl', 'rocks', 'make', '--chdir', base_dir], cwd=tmpdir)
    assert process.returncode == 0, "tarantoolctl rocks make failed"

    project_name = 'test_project'
    cmd = ["tarantoolapp", "create",
        "--name", project_name,
        "--template", 'plain']
    process = subprocess.run(cmd, cwd=tmpdir, env={'PATH': '.rocks/bin'})

    project_path = os.path.join(tmpdir, project_name)
    process = subprocess.run(['tarantoolctl', 'rocks', 'make'], cwd=project_path)
    assert process.returncode == 0, \
        "Error building project"
