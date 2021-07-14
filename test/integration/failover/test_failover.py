import pytest

from integration.failover.utils import (
    get_etcd2_failover_info,
    get_eventual_failover_info,
    get_stateboard_failover_info,
)

from integration.replicasets.conftest import default_project_with_instances
from conftest import project_without_dependencies
from utils import run_command_and_get_output


@pytest.fixture(scope="function")
def project_with_topology(cartridge_cmd, default_project_with_instances, tmpdir):
    project = default_project_with_instances.project

    cmd = [cartridge_cmd, "replicasets", "setup", "--bootstrap-vshard"]
    rc, _ = run_command_and_get_output(cmd, cwd=project.path)
    assert rc == 0

    return project


def test_default_app_stateboard_failover(cartridge_cmd, project_with_topology):
    project = project_with_topology

    cmd = [cartridge_cmd, "failover", "setup"]
    rc, output = run_command_and_get_output(cmd, cwd=project.path)
    assert rc == 0
    assert "Failover configured successfully" in output

    failover_info = get_stateboard_failover_info()
    assert {
        'fencing_enabled': False,
        'failover_timeout': 20,
        'fencing_pause': 2,
        'fencing_timeout': 10,
        'tarantool_params': {
            'uri': 'localhost:4401', 'password': 'passwd'
        },
        'mode': 'stateful',
        'state_provider': 'tarantool'
    } == failover_info


def test_setup_eventual_failover(cartridge_cmd, project_with_topology):
    project = project_with_topology

    cmd = [
        cartridge_cmd, "failover", "set", "eventual", "--params",
        "{\"fencing_enabled\": true, \"failover_timeout\": 30, \"fencing_pause\": 140, \"fencing_timeout\": 15}",
    ]

    rc, output = run_command_and_get_output(cmd, cwd=project.path)
    assert rc == 0
    assert "Failover configured successfully" in output

    failover_info = get_eventual_failover_info()
    assert {
        # Because this parameter (fencing_enabled) suitable in
        # stateful mode only - and we don't check it
        'fencing_enabled': False,
        'failover_timeout': 30,
        'fencing_pause': 140,
        'fencing_timeout': 15,
        'mode': 'eventual',
    } == failover_info


def test_setup_etcd2_failover(cartridge_cmd, project_with_topology):
    project = project_with_topology

    cmd = [
        cartridge_cmd, "failover", "set", "stateful",
        "--state-provider", "etcd2",
        "--provider-params", "{\"prefix\": \"test_prefix\", \"lock_delay\": 15}",
        "--params", "{\"fencing_enabled\": true, \"failover_timeout\": 30, \"fencing_timeout\": 12}"
    ]

    rc, output = run_command_and_get_output(cmd, cwd=project.path)
    assert rc == 0
    assert "Failover configured successfully" in output

    failover_info = get_etcd2_failover_info()
    assert {
        'fencing_enabled': True,
        'failover_timeout': 30,
        'fencing_pause': 2,
        'fencing_timeout': 12,
        'mode': 'stateful',
        'state_provider': 'etcd2',
        'etcd2_params': {
            'endpoints': ['http://127.0.0.1:4001', 'http://127.0.0.1:2379'],
            'lock_delay': 15,
            'password': '',
            'prefix': 'test_prefix',
            'username': ''
        },
    } == failover_info


def test_failover_disabled_command(cartridge_cmd, project_with_topology):
    project = project_with_topology

    cmd = [
        cartridge_cmd, "failover", "disable",
    ]

    rc, output = run_command_and_get_output(cmd, cwd=project.path)
    assert rc == 0
    assert "Failover disabled successfully" in output

    failover_info = get_eventual_failover_info()["mode"]
    assert "disabled" == failover_info


def test_disable_failover_from_set_command(cartridge_cmd, project_with_topology):
    project = project_with_topology

    cmd = [
        cartridge_cmd, "failover", "set", "disabled",
        "--params", "{\"fencing_timeout\": 31, \"failover_timeout\": 31}"
    ]

    rc, output = run_command_and_get_output(cmd, cwd=project.path)
    assert rc == 0
    assert "Failover configured successfully" in output

    failover_info = get_eventual_failover_info()
    assert {
        'fencing_enabled': False,
        'failover_timeout': 31,
        'fencing_pause': 2,
        'fencing_timeout': 31,
        'mode': 'disabled',
    } == failover_info


def test_invalid_disabled_failover(cartridge_cmd, project_without_dependencies):
    project = project_without_dependencies

    cmd = [
        cartridge_cmd, "failover", "set", "disabled",
        "--provider-params", "{\"uri\": some-uri}"
    ]

    rc, output = run_command_and_get_output(cmd, cwd=project.path)
    assert rc == 1
    assert "Please, don't specify any parameters in disabled mode" in output


def test_invalid_eventual_failover_opts(cartridge_cmd):
    pass


def test_invalid_stateboard_failover_opts(cartridge_cmd):
    pass


def test_invalid_etcd2_failover_opts(cartridge_cmd):
    pass



