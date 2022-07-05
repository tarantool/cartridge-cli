import os
import shutil
import subprocess

import pytest
from project import Project


@pytest.fixture(scope="module")
def default_project(cartridge_cmd, module_tmpdir):
    project = Project(cartridge_cmd, 'default-project', module_tmpdir, 'cartridge')
    return project


def test_make_project_with_cartridge(project_with_cartridge, cartridge_cmd, tmpdir):
    project = project_with_cartridge
    dir = os.getenv("CC_TEST_PREBUILT_PROJECTS")
    if dir is None:
        print("Directory for cartridge-cli integration tests prebuilt projects is not set.\n",
              "Please set enviromental variable: CC_TEST_PREBUILT_PROJECTS")
        assert dir is not None
    path = dir + "/project_with_cartridge"

    cmd = [
        cartridge_cmd, "build", project.path
    ]

    process = subprocess.run(cmd, cwd=tmpdir)
    assert process.returncode == 0
    shutil.copytree(project.path, path)


def test_make_default_project(default_project, cartridge_cmd, tmpdir):
    project = default_project
    dir = os.getenv("CC_TEST_PREBUILT_PROJECTS")
    if dir is None:
        print("Directory for cartridge-cli integration tests prebuilt projects is not set.\n",
              "Please set enviromental variable: CC_TEST_PREBUILT_PROJECTS")
        assert dir is not None
    path = dir + "/default_project"

    cmd = [
        cartridge_cmd, "build", project.path
    ]

    process = subprocess.run(cmd, cwd=tmpdir)
    assert process.returncode == 0
    shutil.copytree(project.path, path)
