from integration.replicasets.utils import get_replicaset_by_alias
from utils import get_log_lines, get_replicasets, run_command_and_get_output


def test_bad_replicaset_name(cartridge_cmd, project_with_vshard_replicasets):
    project = project_with_vshard_replicasets.project

    cmd = [
        cartridge_cmd, 'replicasets', 'set-weight',
        '--replicaset', 'unknown-replicaset',
        '123.45',
    ]

    rc, output = run_command_and_get_output(cmd, cwd=project.path)
    assert rc == 1
    assert "Replica set unknown-replicaset isn't found in current topology" in output


def test_set_weight(cartridge_cmd, project_with_vshard_replicasets):
    project = project_with_vshard_replicasets.project
    instances = project_with_vshard_replicasets.instances
    replicasets = project_with_vshard_replicasets.replicasets

    hot_storage_rpl = replicasets['hot-storage']
    hot_master = instances['hot-master']
    admin_api_url = hot_master.get_admin_api_url()

    NEW_WEIGHT = 123.45

    # set replicaset weight
    cmd = [
        cartridge_cmd, 'replicasets', 'set-weight',
        '--replicaset', hot_storage_rpl.name,
        str(NEW_WEIGHT),
    ]

    rc, output = run_command_and_get_output(cmd, cwd=project.path)
    assert rc == 0
    assert get_log_lines(output) == [
        '• Replica set %s weight is set to %s' % (hot_storage_rpl.name, NEW_WEIGHT),
    ]

    replicasets = get_replicasets(admin_api_url)
    hot_replicaset = get_replicaset_by_alias(replicasets, hot_storage_rpl.name)
    assert hot_replicaset is not None
    assert hot_replicaset['weight'] == NEW_WEIGHT
