#!/usr/bin/python3

import os
import py
import pytest
import subprocess
import configparser
import tempfile
import shutil
import tarfile

@pytest.fixture(scope='session')
def session_tmpdir(request):
    dir = py.path.local(tempfile.mkdtemp())
    request.addfinalizer(lambda: dir.remove(rec=1))
    return str(dir)

@pytest.fixture(scope="session", params=['plain', 'cluster'])
def project_path(request, session_tmpdir):
    project_name = 'project-{}'.format(request.param)
    cmd = ["tarantoolapp", "create",
        "--name", project_name,
        "--template", request.param]
    process = subprocess.run(cmd, cwd=session_tmpdir)
    assert process.returncode == 0, \
        "Error during creating the project"
    return os.path.join(session_tmpdir, project_name)

def test_project(project_path):
    process = subprocess.run(['tarantoolctl', 'rocks', 'make'], cwd=project_path)
    assert process.returncode == 0, \
        "Error building project"

    process = subprocess.run(['tarantoolctl', 'rocks', 'test'], cwd=project_path)
    assert process.returncode == 0, \
        "Error testing project"
