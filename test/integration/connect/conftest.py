import os
import subprocess
import tarfile
import pytest

from project import Project
from project import patch_cartridge_proc_titile
from project import remove_dependency
from project import replace_project_file, remove_project_file
from project import INIT_NO_CARTRIDGE_FILEPATH, INIT_ROLES_RELOAD_ALLOWED

from utils import ProjectWithTopology, Instance, find_archive


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
def unpacked_tar_archive(cartridge_cmd, short_session_tmpdir):
    project = Project(cartridge_cmd, 'pack-prj', short_session_tmpdir, 'cartridge')

    # This is necessary, because default app config has parameter `stateboard: true`
    remove_project_file(project, '.cartridge.yml')

    # Pack project
    cmd = [
        cartridge_cmd,
        "pack", "tgz",
        project.path
    ]

    process = subprocess.run(cmd, cwd=short_session_tmpdir)
    assert process.returncode == 0, "Error during packing the project"

    # Unpack tar arhcive
    archive_path = find_archive(short_session_tmpdir, project.name, 'tar.gz')
    with tarfile.open(archive_path) as f:
        extract_dir = os.path.join(short_session_tmpdir, 'extract')
        f.extractall(extract_dir)

    # Change project path to directory which contains unpacked tar archive
    project.path = os.path.join(extract_dir, project.name)
    version = os.path.basename(archive_path)
    project.version = version[len(project.name) + 1:-len('.tar.gz')]

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
def project_setuped_with_instances(cartridge_cmd, unpacked_tar_archive, start_stop_cli, request):
    cli = start_stop_cli
    project = unpacked_tar_archive
    replace_project_file(project, 'init.lua', INIT_ROLES_RELOAD_ALLOWED)

    instances_list = [
        Instance('router', 8081, 'localhost:3301'),
        Instance('s1-master', 8082, 'localhost:3302'),
        Instance('s1-replica', 8083, 'localhost:3303'),
        Instance('s2-master', 8084, 'localhost:3304'),
        Instance('s2-replica', 8085, 'localhost:3305'),
    ]

    p = ProjectWithTopology(cli, project, instances_list=instances_list)
    request.addfinalizer(lambda: p.stop())
    p.start()

    cmd = [cartridge_cmd, "replicasets", "setup"]
    process = subprocess.run(cmd, cwd=project.path)
    assert process.returncode == 0, "Error during setup the project"

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
