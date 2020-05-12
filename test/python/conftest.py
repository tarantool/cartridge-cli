import py
import pytest
import tempfile
import docker
import os
import subprocess

from project import Project
from project import remove_dependency
from project import add_dependency_submodule
from project import use_deprecated_files
from project import remove_all_dependencies


# ########
# Fixtures
# ########
@pytest.fixture(scope='module')
def module_tmpdir(request):
    dir = py.path.local(tempfile.mkdtemp())
    request.addfinalizer(lambda: dir.remove(rec=1))
    return str(dir)


@pytest.fixture(scope="session")
def docker_client():
    client = docker.from_env()
    return client


@pytest.fixture(scope="session")
def cartridge_cmd(request):
    build_dir = py.path.local(tempfile.mkdtemp())
    request.addfinalizer(lambda: build_dir.remove(rec=1))

    cli_source_path = os.path.realpath(os.path.join(os.path.dirname(__file__), '..', '..'))
    cli_build_cmd = ['go', 'build', '-o', 'cartridge', cli_source_path]

    process = subprocess.run(cli_build_cmd, cwd=build_dir)
    assert process.returncode == 0, 'Failed to build cartridge-cli executable'

    cli_path = os.path.join(build_dir, 'cartridge')
    assert os.path.exists(cli_path), 'Executable not found in {}'.format(cli_path)

    return cli_path


# ################
# Project fixtures
# ################
# There are three main types of projects:
# * light_project ({original,deprecated}_light_project):
#   Cartridge CLI creates project with cartridge dependency by default.
#   It's known that installing cartridge rocks is a long operation,
#   so we don't want to perform it on every test.
#   These fixtures are used to decrease packing time.
#   They don't have a cartridge dependency,
#   but have dependency installed from submodule
#   (see add_dependency_submodule function for details)
#   In fact, we need to install cartridge dependency only
#   for e2e
#
# * project_with_cartridge ({original,deprecated}_project_with_cartridge):
#   This is a project with cartridge dependency installed.
#   Is used in `docker pack` tests. Test image is built once and then
#   it's used in all docker tests include e2e.
#   This project also have submodule dependency (see add_dependency_submodule)
#   to test pre and post build hooks
#
# * project_without_dependencies:
#   This is the empty project without dependencies.
#   It is used for error behavior tests and tests where
#   result package content doesn't matter
#
################
# Light projects
################
@pytest.fixture(scope="function")
def original_light_project(cartridge_cmd, tmpdir):
    project = Project(cartridge_cmd, 'light-original-project', tmpdir, 'cartridge')

    remove_dependency(project, 'cartridge')
    remove_dependency(project, 'luatest')

    add_dependency_submodule(project)

    return project


@pytest.fixture(scope="function")
def deprecated_light_project(cartridge_cmd, tmpdir):
    project = Project(cartridge_cmd, 'light-deprecated-project', tmpdir, 'cartridge')

    remove_dependency(project, 'cartridge')
    remove_dependency(project, 'luatest')

    add_dependency_submodule(project)

    use_deprecated_files(project)

    return project


@pytest.fixture(scope="function", params=['original', 'deprecated'])
def light_project(original_light_project, deprecated_light_project, request):
    if request.param == 'original':
        return original_light_project
    elif request.param == 'deprecated':
        return deprecated_light_project


#########################
# Projects with cartridge
#########################
@pytest.fixture(scope="function")
def original_project_with_cartridge(cartridge_cmd, tmpdir):
    project = Project(cartridge_cmd, 'original-project-with-cartridge', tmpdir, 'cartridge')
    remove_dependency(project, 'luatest')

    add_dependency_submodule(project)

    return project


@pytest.fixture(scope="function")
def deprecated_project_with_cartridge(cartridge_cmd, tmpdir):
    project = Project(cartridge_cmd, 'deprecated-project-with-cartridge', tmpdir, 'cartridge')
    remove_dependency(project, 'luatest')

    add_dependency_submodule(project)
    use_deprecated_files(project)

    return project


@pytest.fixture(scope="function", params=['original', 'deprecated'])
def project_with_cartridge(original_project_with_cartridge, deprecated_project_with_cartridge, request):
    if request.param == 'original':
        return original_project_with_cartridge
    elif request.param == 'deprecated':
        return deprecated_project_with_cartridge


##############################
# Project without dependencies
##############################
@pytest.fixture(scope="function")
def project_without_dependencies(cartridge_cmd, tmpdir):
    project = Project(cartridge_cmd, 'empty-project', tmpdir, 'cartridge')

    remove_all_dependencies(project)
    return project
