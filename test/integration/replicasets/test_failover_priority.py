from utils import get_replicasets
from utils import get_log_lines
from utils import run_command_and_get_output

from integration.replicasets.utils import get_replicaset_by_alias
from integration.replicasets.utils import get_list_from_log_lines


def test_bad_replicaset_name(cartridge_cmd, project_with_vshard_replicasets):
    project = project_with_vshard_replicasets.project

    cmd = [
        cartridge_cmd, 'replicasets', 'set-failover-priority',
        '--replicaset', 'unknown-replicaset',
        'some-instance', 'one-more-instance',
    ]

    rc, output = run_command_and_get_output(cmd, cwd=project.path)
    assert rc == 1
    assert "Replica set unknown-replicaset isn't found in current topology" in output


def test_set_failover_priority(cartridge_cmd, project_with_vshard_replicasets):
    project = project_with_vshard_replicasets.project
    instances = project_with_vshard_replicasets.instances
    replicasets = project_with_vshard_replicasets.replicasets

    hot_storage_rpl = replicasets['hot-storage']
    hot_master = instances['hot-master']
    hot_replica = instances['hot-replica']

    admin_api_url = hot_master.get_admin_api_url()

    # set replicaset failover priority
    cmd = [
        cartridge_cmd, 'replicasets', 'set-failover-priority',
        '--replicaset', hot_storage_rpl.name,
        hot_replica.name, hot_master.name,
    ]

    rc, output = run_command_and_get_output(cmd, cwd=project.path)
    assert rc == 0

    exp_failover_priority = [hot_replica.name, hot_master.name]

    log_lines = get_log_lines(output)
    assert log_lines[:1] == [
        "• Replica set hot-storage failover priority was set to:",
    ]

    failover_priority_list = get_list_from_log_lines(log_lines[1:])
    assert failover_priority_list == exp_failover_priority

    replicasets = get_replicasets(admin_api_url)
    hot_replicaset = get_replicaset_by_alias(replicasets, hot_storage_rpl.name)
    assert hot_replicaset is not None

    servers_names = [s['alias'] for s in hot_replicaset['servers']]
    assert servers_names == exp_failover_priority

    # specify only one instance
    cmd = [
        cartridge_cmd, 'replicasets', 'set-failover-priority',
        '--replicaset', hot_storage_rpl.name,
        hot_master.name,
    ]

    rc, output = run_command_and_get_output(cmd, cwd=project.path)
    assert rc == 0

    exp_failover_priority = [hot_master.name, hot_replica.name]

    log_lines = get_log_lines(output)
    assert log_lines[:1] == [
        "• Replica set hot-storage failover priority was set to:",
    ]

    failover_priority_list = get_list_from_log_lines(log_lines[1:])
    assert failover_priority_list == exp_failover_priority

    replicasets = get_replicasets(admin_api_url)
    hot_replicaset = get_replicaset_by_alias(replicasets, hot_storage_rpl.name)
    assert hot_replicaset is not None

    servers_names = [s['alias'] for s in hot_replicaset['servers']]
    assert servers_names == exp_failover_priority
