from integration.failover.utils import (assert_mode_and_params_state,
                                        get_etcd2_failover_info,
                                        get_eventual_failover_info,
                                        get_stateboard_failover_info)
from utils import run_command_and_get_output


def test_status_eventual(cartridge_cmd, project_with_topology_and_vshard):
    project = project_with_topology_and_vshard

    cmd = [
        cartridge_cmd, "failover", "set", "eventual",
        "--params", "{\"fencing_enabled\": true, \"failover_timeout\": 30, \"fencing_timeout\": 12}"
    ]

    rc, output = run_command_and_get_output(cmd, cwd=project.path)
    assert rc == 0
    assert "Failover configured successfully" in output

    failover_info = get_eventual_failover_info()

    cmd = [cartridge_cmd, "failover", "status"]
    rc, output = run_command_and_get_output(cmd, cwd=project.path)

    assert rc == 0
    assert_mode_and_params_state(failover_info, output)

    assert "stateboard_params" not in output
    assert "etcd2_params" not in output


def test_status_stateful_stateboard(cartridge_cmd, project_with_topology_and_vshard):
    project = project_with_topology_and_vshard

    cmd = [
        cartridge_cmd, "failover", "set", "stateful",
        "--state-provider", "stateboard",
        "--provider-params", "{\"uri\": \"localhost:1020\", \"password\": \"pass\"}",
    ]

    rc, output = run_command_and_get_output(cmd, cwd=project.path)
    assert rc == 0
    assert "Failover configured successfully" in output

    failover_info = get_stateboard_failover_info()

    cmd = [cartridge_cmd, "failover", "status"]
    rc, output = run_command_and_get_output(cmd, cwd=project.path)

    assert rc == 0
    assert_mode_and_params_state(failover_info, output)
    assert "etcd2_params" not in output

    assert "stateboard_params" in output
    assert f"uri: {failover_info['tarantool_params']['uri']}" in output
    assert f"password: {failover_info['tarantool_params']['password']}" in output


def test_status_stateful_etcd2(cartridge_cmd, project_with_topology_and_vshard):
    project = project_with_topology_and_vshard

    cmd = [
        cartridge_cmd, "failover", "set", "stateful",
        "--state-provider", "etcd2",
        "--provider-params", "{\"prefix\": \"test_prefix\", \"lock_delay\": 15}",
    ]

    rc, output = run_command_and_get_output(cmd, cwd=project.path)
    assert rc == 0
    assert "Failover configured successfully" in output

    failover_info = get_etcd2_failover_info()

    cmd = [cartridge_cmd, "failover", "status"]
    rc, output = run_command_and_get_output(cmd, cwd=project.path)

    assert rc == 0
    assert_mode_and_params_state(failover_info, output)
    assert "stateboard_params" not in output

    assert "etcd2_params" in output
    assert f"password: {failover_info['etcd2_params']['password']}" in output
    assert f"lock_delay: {failover_info['etcd2_params']['lock_delay']}" in output
    assert f"endpoints: {', '.join(failover_info['etcd2_params']['endpoints'])}" in output
    assert f"username: {failover_info['etcd2_params']['username']}" in output
    assert f"prefix: {failover_info['etcd2_params']['prefix']}" in output


def test_status_disabled(cartridge_cmd, project_with_topology_and_vshard):
    project = project_with_topology_and_vshard

    cmd = [cartridge_cmd, "failover", "set", "disabled"]
    rc, output = run_command_and_get_output(cmd, cwd=project.path)
    assert rc == 0
    assert "Failover disabled successfully" in output

    failover_info = get_eventual_failover_info()

    cmd = [cartridge_cmd, "failover", "status"]
    rc, output = run_command_and_get_output(cmd, cwd=project.path)

    assert rc == 0
    assert_mode_and_params_state(failover_info, output)

    assert "stateboard_params" not in output
    assert "etcd2_params" not in output
