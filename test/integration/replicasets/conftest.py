import subprocess
import pytest
import os

from utils import run_command_and_get_output
from utils import DEFAULT_CFG, DEFAULT_RPL_CFG

from project import Project
from project import patch_cartridge_proc_titile
from project import configure_vshard_groups
from project import add_custom_roles

from integration.replicasets.utils import ProjectWithTopology, Replicaset, Instance


@pytest.fixture(scope="session")
def built_project(cartridge_cmd, short_session_tmpdir):
    project = Project(cartridge_cmd, 'my-project', short_session_tmpdir, 'cartridge')

    add_custom_roles(project, [
        {'name': 'dep-role-1'},
        {'name': 'dep-role-2'},
        {'name': 'role-with-deps', 'dependencies': ['dep-role-1', 'dep-role-2']},
    ])
    configure_vshard_groups(project, ['hot', 'cold'])

    # build project
    cmd = [
        cartridge_cmd,
        "build",
    ]
    process = subprocess.run(cmd, cwd=project.path)
    assert process.returncode == 0, "Error during building the project"

    # don't change process title
    patch_cartridge_proc_titile(project)

    os.remove(os.path.join(project.path, DEFAULT_CFG))
    os.remove(os.path.join(project.path, DEFAULT_RPL_CFG))

    return project


@pytest.fixture(scope="function")
def project_with_instances(built_project, start_stop_cli, request):
    cli = start_stop_cli
    project = built_project

    router = Instance('router', 8081, 'localhost:3301')
    s1_master = Instance('s1-master', 8082, 'localhost:3302')
    s1_replica = Instance('s1-replica', 8083, 'localhost:3303')
    s1_replica_2 = Instance('s1-replica-2', 8084, 'localhost:3304')
    s2_master = Instance('s2-master', 8085, 'localhost:3305')

    p = ProjectWithTopology(
        cli,
        project,
        instances_list=[router, s1_master, s1_replica, s1_replica_2, s2_master],
    )

    request.addfinalizer(lambda: p.stop())

    p.start()
    return p


@pytest.fixture(scope="function")
def project_with_replicaset_no_roles(cartridge_cmd, built_project, start_stop_cli, request):
    cli = start_stop_cli
    project = built_project

    instance = Instance('some-instance', 8081, 'localhost:3301')

    p = ProjectWithTopology(
        cli,
        project,
        instances_list=[instance],
    )

    p.start()

    rpl = Replicaset('some-rpl', instances=[instance])

    cmd = [
        cartridge_cmd, 'replicasets', 'join',
        '--replicaset', rpl.name,
    ]
    cmd.extend([i.name for i in rpl.instances])
    rc, output = run_command_and_get_output(cmd, cwd=project.path)
    assert rc == 0

    p.set_replicasets([rpl])

    request.addfinalizer(lambda: p.stop())

    return p


@pytest.fixture(scope="function")
def project_with_vshard_replicasets(cartridge_cmd, built_project, start_stop_cli, request):
    cli = start_stop_cli
    project = built_project

    router = Instance('router', 8081, 'localhost:3301')
    hot_master = Instance('hot-master', 8082, 'localhost:3302')
    hot_replica = Instance('hot-replica', 8083, 'localhost:3303')
    cold_master = Instance('cold-master', 8084, 'localhost:3304')

    VSHARD_ROUTER_ROLE = 'vshard-router'
    VSHARD_STORAGE_ROLE = 'vshard-storage'

    p = ProjectWithTopology(
        cli,
        project,
        instances_list=[router, hot_master, hot_replica, cold_master],
    )

    p.start()

    # replicasets
    router_rpl = Replicaset('router', instances=[router])
    hot_storage_rpl = Replicaset('hot-storage', instances=[
        hot_master,
        hot_replica,
    ])
    cold_storage_rpl = Replicaset('cold-storage', instances=[
        cold_master,
    ])

    # router
    cmd = [
        cartridge_cmd, 'replicasets', 'join',
        '--replicaset', router_rpl.name,
    ]
    cmd.extend([i.name for i in router_rpl.instances])
    rc, output = run_command_and_get_output(cmd, cwd=project.path)
    assert rc == 0

    cmd = [
        cartridge_cmd, 'replicasets', 'add-roles',
        '--replicaset', router_rpl.name,
        VSHARD_ROUTER_ROLE,
    ]

    rc, output = run_command_and_get_output(cmd, cwd=project.path)
    assert rc == 0

    # hot-storage
    cmd = [
        cartridge_cmd, 'replicasets', 'join',
        '--replicaset', hot_storage_rpl.name,
    ]
    cmd.extend([i.name for i in hot_storage_rpl.instances])
    rc, output = run_command_and_get_output(cmd, cwd=project.path)
    assert rc == 0

    cmd = [
        cartridge_cmd, 'replicasets', 'add-roles',
        '--replicaset', hot_storage_rpl.name,
        '--vshard-group', 'hot',
        VSHARD_STORAGE_ROLE,
    ]

    rc, output = run_command_and_get_output(cmd, cwd=project.path)
    assert rc == 0

    # cold-storage
    cmd = [
        cartridge_cmd, 'replicasets', 'join',
        '--replicaset', cold_storage_rpl.name,
    ]
    cmd.extend([i.name for i in cold_storage_rpl.instances])
    rc, output = run_command_and_get_output(cmd, cwd=project.path)
    assert rc == 0

    cmd = [
        cartridge_cmd, 'replicasets', 'add-roles',
        '--replicaset', cold_storage_rpl.name,
        '--vshard-group', 'cold',
        VSHARD_STORAGE_ROLE,
    ]

    rc, output = run_command_and_get_output(cmd, cwd=project.path)
    assert rc == 0

    # save replicasets
    p.set_replicasets([router_rpl, hot_storage_rpl, cold_storage_rpl])

    request.addfinalizer(lambda: p.stop())

    return p


@pytest.fixture(scope="function")
def project_with_one_joined_instance(cartridge_cmd, built_project, start_stop_cli, request):
    cli = start_stop_cli
    project = built_project

    instance = Instance('some-instance', 8081, 'localhost:3301')

    p = ProjectWithTopology(
        cli,
        project,
        instances_list=[instance],
    )

    p.start()

    rpl = Replicaset('some-replicaset', instances=[instance])

    # create replicaset
    cmd = [
        cartridge_cmd, 'replicasets', 'join',
        '--replicaset', rpl.name,
    ]
    cmd.extend([i.name for i in rpl.instances])
    rc, output = run_command_and_get_output(cmd, cwd=project.path)
    assert rc == 0

    # save replicasets
    p.set_replicasets([rpl])

    request.addfinalizer(lambda: p.stop())

    return p
