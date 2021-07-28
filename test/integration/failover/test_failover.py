import os
import yaml

from integration.failover.utils import (
    get_etcd2_failover_info,
    get_eventual_failover_info,
    get_stateboard_failover_info,
)

from utils import run_command_and_get_output


# Tests
def test_default_app_stateboard_failover(cartridge_cmd, project_with_topology_and_vshard):
    project = project_with_topology_and_vshard

    cmd = [cartridge_cmd, "failover", "setup"]
    rc, output = run_command_and_get_output(cmd, cwd=project.path)
    assert rc == 0
    assert "Failover configured successfully" in output

    failover_info = get_stateboard_failover_info()
    assert failover_info == {
        'fencing_enabled': False,
        'failover_timeout': 20,
        'fencing_pause': 2,
        'fencing_timeout': 10,
        'tarantool_params': {
            'uri': 'localhost:4401', 'password': 'passwd'
        },
        'mode': 'stateful',
        'state_provider': 'tarantool'
    }


def test_setup_eventual_failover(cartridge_cmd, project_with_topology_and_vshard):
    project = project_with_topology_and_vshard

    cmd = [
        cartridge_cmd, "failover", "set", "eventual", "--params",
        "{\"fencing_enabled\": true, \"failover_timeout\": 30, \"fencing_pause\": 140, \"fencing_timeout\": 15}",
    ]

    rc, output = run_command_and_get_output(cmd, cwd=project.path)
    assert rc == 0
    assert "Failover configured successfully" in output

    failover_info = get_eventual_failover_info()
    assert failover_info == {
        # Because this parameter (fencing_enabled) suitable in
        # stateful mode only - and we don't check it
        'fencing_enabled': False,
        'failover_timeout': 30,
        'fencing_pause': 140,
        'fencing_timeout': 15,
        'mode': 'eventual',
    }


def test_setup_etcd2_failover(cartridge_cmd, project_with_topology_and_vshard):
    project = project_with_topology_and_vshard

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
    assert failover_info == {
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
    }


def test_failover_disabled_command(cartridge_cmd, project_with_topology_and_vshard):
    project = project_with_topology_and_vshard

    cmd = [
        cartridge_cmd, "failover", "disable",
    ]

    rc, output = run_command_and_get_output(cmd, cwd=project.path)
    assert rc == 0
    assert "Failover disabled successfully" in output

    failover_info = get_eventual_failover_info()["mode"]
    assert failover_info == "disabled"


def test_disable_failover_from_sub_command(cartridge_cmd, project_with_topology_and_vshard):
    project = project_with_topology_and_vshard

    cmd = [
        cartridge_cmd, "failover", "set", "disabled",
        "--params", "{\"fencing_timeout\": 31, \"failover_timeout\": 31, \"fencing_pause\": 3}"
    ]

    rc, output = run_command_and_get_output(cmd, cwd=project.path)
    assert rc == 0
    assert "Failover disabled successfully" in output

    failover_info = get_eventual_failover_info()
    assert failover_info == {
        'fencing_enabled': False,
        'failover_timeout': 31,
        'fencing_pause': 3,
        'fencing_timeout': 31,
        'mode': 'disabled',
    }

    with open(os.path.join(project.path, "failover.yml"), "w") as f:
        f.write(yaml.dump({"mode": "disabled", "fencing_pause": 1}))

    cmd = [cartridge_cmd, "failover", "setup"]
    rc, output = run_command_and_get_output(cmd, cwd=project.path)
    assert rc == 0
    assert "Failover configured successfully" in output

    failover_info = get_eventual_failover_info()
    assert failover_info == {
        'fencing_enabled': False,
        'failover_timeout': 31,
        'fencing_pause': 1,
        'fencing_timeout': 31,
        'mode': 'disabled',
    }


def test_set_invalid_mode(cartridge_cmd, project_without_dependencies):
    project = project_without_dependencies

    cmd = [
        cartridge_cmd, "failover", "set", "non-exists-mode",
        "--provider-params", "{\"uri\": some-uri}"
    ]

    rc, output = run_command_and_get_output(cmd, cwd=project.path)
    assert rc == 1
    assert "Failover mode should be `stateful`, `eventual` or `disabled`" in output


def test_set_invalid_provider(cartridge_cmd, project_without_dependencies):
    project = project_without_dependencies

    cmd = [
        cartridge_cmd, "failover", "set", "stateful",
        "--state-provider", "provider150"
    ]

    rc, output = run_command_and_get_output(cmd, cwd=project.path)
    assert rc == 1
    assert "--state-provider flag should be `stateboard` or `etcd2`" in output


def test_invalid_disabled_failover_opts(cartridge_cmd, project_without_dependencies):
    project = project_without_dependencies

    cmd = [
        cartridge_cmd, "failover", "set", "disabled",
        "--provider-params", "{\"uri\": some-uri}"
    ]

    rc, output = run_command_and_get_output(cmd, cwd=project.path)
    assert rc == 1
    assert "Please, don't specify provider parameters when using disabled mode" in output


def test_invalid_eventual_failover_opts(cartridge_cmd, project_without_dependencies):
    project = project_without_dependencies
    cmd = [
        cartridge_cmd, "failover", "set", "eventual",
        "--state-provider", "stateboard"
    ]

    rc, output = run_command_and_get_output(cmd, cwd=project.path)
    assert rc == 1
    assert "Please, don't specify --state-provider flag when using eventual mode" in output

    project = project_without_dependencies
    cmd = [
        cartridge_cmd, "failover", "set", "eventual",
        "--provider-params", "{\"uri\": some-uri}"
    ]

    rc, output = run_command_and_get_output(cmd, cwd=project.path)
    assert rc == 1
    assert "Please, don't specify provider parameters when using eventual mode" in output


def test_invalid_stateful_failover_opts(cartridge_cmd, project_without_dependencies):
    project = project_without_dependencies
    cmd = [
        cartridge_cmd, "failover", "set", "stateful",
    ]

    rc, output = run_command_and_get_output(cmd, cwd=project.path)
    assert rc == 1
    assert "Please, specify --state-provider flag when using stateful mode" in output

    project = project_without_dependencies
    cmd = [
        cartridge_cmd, "failover", "set", "stateful",
        "--state-provider", "stateboard"
    ]

    rc, output = run_command_and_get_output(cmd, cwd=project.path)
    assert rc == 1
    assert "Please, specify params for stateboard state provider, using " \
        "--provider-params '{\"uri\": \"localhost:4401\", \"password\": \"passwd\"}'" in output
