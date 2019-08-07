#!/usr/bin/python3

import subprocess

from utils import tarantool_enterprise_is_used


def test_project(project_path):
    process = subprocess.run(['tarantoolctl', 'rocks', 'make'], cwd=project_path)
    assert process.returncode == 0, \
        "Error building project"
    if tarantool_enterprise_is_used():
        process = subprocess.run(['tarantoolctl', 'rocks', 'test'], cwd=project_path)
        assert process.returncode == 0, \
            "Error testing project"
