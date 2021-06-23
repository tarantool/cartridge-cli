import os
import subprocess
import pytest

from project import Project
from project import patch_cartridge_proc_titile
from project import remove_dependency
from project import replace_project_file, remove_project_file
from project import INIT_NO_CARTRIDGE_FILEPATH

from utils import ProjectWithTopology, Instance


@pytest.fixture(scope="session")
def built_project(cartridge_cmd, short_session_tmpdir):
    project = Project(cartridge_cmd, 'some-project', short_session_tmpdir, 'cartridge')

    # This is necessary, because default app config has parameter `stateboard: true`
    remove_project_file(project, '.cartridge.yml')

    # build project
    cmd = [
        cartridge_cmd,
        "build",
    ]
    process = subprocess.run(cmd, cwd=project.path)
    assert process.returncode == 0, "Error during building the project"

    # don't change process title
    patch_cartridge_proc_titile(project)

    os.remove(project.get_cfg_path())

    return project


@pytest.fixture(scope="session")
def built_project_no_cartridge(cartridge_cmd, short_session_tmpdir):
    project = Project(cartridge_cmd, 'project-no-cartridge', short_session_tmpdir, 'cartridge')
    remove_dependency(project, 'cartridge')

    replace_project_file(project, 'init.lua', INIT_NO_CARTRIDGE_FILEPATH)
    remove_project_file(project, '.cartridge.yml')

    # build project
    cmd = [
        cartridge_cmd,
        "build",
    ]
    process = subprocess.run(cmd, cwd=project.path)
    assert process.returncode == 0, "Error during building the project"

    os.remove(project.get_cfg_path())

    return project


@pytest.fixture(scope="function")
def project_with_instances(built_project, start_stop_cli, request):
    cli = start_stop_cli
    project = built_project

    router = Instance('router', 8081, 'localhost:3301')
    s1_master = Instance('s1-master', 8082, 'localhost:3302')
    s1_replica = Instance('s1-replica', 8083, 'localhost:3303')

    p = ProjectWithTopology(
        cli,
        project,
        instances_list=[router, s1_master, s1_replica],
    )

    request.addfinalizer(lambda: p.stop())

    p.start()
    return p


@pytest.fixture(scope="function")
def project_with_instances_no_cartridge(built_project_no_cartridge, start_stop_cli, request):
    cli = start_stop_cli
    project = built_project_no_cartridge

    router = Instance('router', 8081, 'localhost:3301')
    s1_master = Instance('s1-master', 8082, 'localhost:3302')
    s1_replica = Instance('s1-replica', 8083, 'localhost:3303')

    p = ProjectWithTopology(
        cli,
        project,
        instances_list=[router, s1_master, s1_replica],
    )

    request.addfinalizer(lambda: p.stop())

    p.start()
    return p
