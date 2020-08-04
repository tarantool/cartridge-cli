import os
import yaml
import copy
import pytest

from utils import write_instance_topology_conf
from utils import run_command_and_get_output
from utils import get_logs
from utils import assert_for_all_instances


APPNAME = 'myapp'
OTHER_APP_NAME = 'other-app'


RPL_TO_SET_MASER_UUID = 'rpl-1-uuid'
NON_EXISTENT_RPL_UUID = 'non-existent-rpl-uuid'

INSTANCE_TO_SET_MASTER_UUID = 'srv-3-uuid'
INSTANCE_NOT_IN_LEADERS_UUID = 'srv-not-in-leaders'
NON_EXISTENT_INSTANCE_UUID = 'non-existent-srv'
INSTANCE_FROM_NON_EXISTENT_RPL_UUID = 'srv-non-existent-rpl'
DISABLED_INSTANCE_UUID = 'srv-disabled'
EXPELLED_INSTANCE_UUID = 'srv-expelled'
OTHER_RPL_INSTANCE_UUID = 'SRV-FROM-OTHER-RPL'


SIMPLE_CONF = {
    'failover': False,
    'replicasets': {
        RPL_TO_SET_MASER_UUID: {
            'alias': 'unnamed',
            'all_rw': False,
            'master': ['srv-1-uuid', 'srv-2-uuid', INSTANCE_TO_SET_MASTER_UUID],
            'roles': {'vshard-storage': True},
            'vshard_group': 'default',
            'weight': 1
        },
        'rpl-2-uuid': {
            'alias': 'unnamed',
            'all_rw': False,
            'master': [OTHER_RPL_INSTANCE_UUID],
            'roles': {'vshard-storage': True},
            'vshard_group': 'default',
            'weight': 1
        },
    },
    'servers': {
        'srv-1-uuid': {
            'disabled': False,
            'replicaset_uuid': RPL_TO_SET_MASER_UUID,
            'uri': 'localhost:3301',
        },
        'srv-2-uuid': {
            'disabled': False,
            'replicaset_uuid': RPL_TO_SET_MASER_UUID,
            'uri': 'localhost:3302',
        },
        INSTANCE_TO_SET_MASTER_UUID: {
            'disabled': False,
            'replicaset_uuid': RPL_TO_SET_MASER_UUID,
            'uri': 'localhost:3303',
        },
        OTHER_RPL_INSTANCE_UUID: {
            'disabled': False,
            'replicaset_uuid': 'rpl-2-uuid',
            'uri': 'localhost:3304',
        },
        INSTANCE_NOT_IN_LEADERS_UUID: {
            'disabled': False,
            'replicaset_uuid': RPL_TO_SET_MASER_UUID,
            'uri': 'localhost:3305',
        },
        DISABLED_INSTANCE_UUID: {
            'disabled': True,
            'replicaset_uuid': RPL_TO_SET_MASER_UUID,
            'uri': 'localhost:3306',
        },
        INSTANCE_FROM_NON_EXISTENT_RPL_UUID: {
            'disabled': False,
            'replicaset_uuid': 'non-existent-rpl-uuid',
            'uri': 'localhost:3307',
        },
        EXPELLED_INSTANCE_UUID: 'expelled',
    }
}


def test_set_master_bad_instance(cartridge_cmd, tmpdir):
    data_dir = os.path.join(tmpdir, 'tmp', 'data')
    os.makedirs(data_dir)

    old_conf = copy.deepcopy(SIMPLE_CONF)

    instances = ['instance-1', 'instance-2']
    write_instance_topology_conf(data_dir, APPNAME, old_conf, instances)

    # non-existent instance
    cmd = [
        cartridge_cmd, 'repair', 'set-leader',
        '--name', APPNAME,
        '--data-dir', data_dir,
        RPL_TO_SET_MASER_UUID, NON_EXISTENT_INSTANCE_UUID,
    ]

    rc, output = run_command_and_get_output(cmd, cwd=tmpdir)
    assert rc == 1

    assert_for_all_instances(
        get_logs(output), APPNAME, instances, lambda line:
        "Instance %s isn't found in cluster" % NON_EXISTENT_INSTANCE_UUID in line
    )

    # expelled instance
    cmd = [
        cartridge_cmd, 'repair', 'set-leader',
        '--name', APPNAME,
        '--data-dir', data_dir,
        RPL_TO_SET_MASER_UUID, EXPELLED_INSTANCE_UUID,
    ]

    rc, output = run_command_and_get_output(cmd, cwd=tmpdir)
    assert rc == 1

    assert_for_all_instances(
        get_logs(output), APPNAME, instances, lambda line:
        "Instance %s is expelled" % EXPELLED_INSTANCE_UUID in line
    )

    # disabled instance
    cmd = [
        cartridge_cmd, 'repair', 'set-leader',
        '--name', APPNAME,
        '--data-dir', data_dir,
        RPL_TO_SET_MASER_UUID, DISABLED_INSTANCE_UUID,
    ]

    rc, output = run_command_and_get_output(cmd, cwd=tmpdir)
    assert rc == 1

    assert_for_all_instances(
        get_logs(output), APPNAME, instances, lambda line:
        "Instance %s is disabled" % DISABLED_INSTANCE_UUID in line
    )

    # instance from other replicaset
    cmd = [
        cartridge_cmd, 'repair', 'set-leader',
        '--name', APPNAME,
        '--data-dir', data_dir,
        RPL_TO_SET_MASER_UUID, OTHER_RPL_INSTANCE_UUID,
    ]

    rc, output = run_command_and_get_output(cmd, cwd=tmpdir)
    assert rc == 1

    assert_for_all_instances(
        get_logs(output), APPNAME, instances, lambda line:
        "Instance %s doesn't belong to replicaset %s"
        % (OTHER_RPL_INSTANCE_UUID, RPL_TO_SET_MASER_UUID) in line
    )


def test_set_master_bad_replicaset(cartridge_cmd, tmpdir):
    data_dir = os.path.join(tmpdir, 'tmp', 'data')
    os.makedirs(data_dir)

    old_conf = copy.deepcopy(SIMPLE_CONF)

    instances = ['instance-1', 'instance-2']
    write_instance_topology_conf(data_dir, APPNAME, old_conf, instances)

    # non-existent replicaset
    cmd = [
        cartridge_cmd, 'repair', 'set-leader',
        '--name', APPNAME,
        '--data-dir', data_dir,
        NON_EXISTENT_RPL_UUID, INSTANCE_FROM_NON_EXISTENT_RPL_UUID,
    ]

    rc, output = run_command_and_get_output(cmd, cwd=tmpdir)
    assert rc == 1

    assert_for_all_instances(
        get_logs(output), APPNAME, instances, lambda line:
        "Replicaset %s isn't found in the cluster" % NON_EXISTENT_RPL_UUID in line
    )


@pytest.mark.parametrize('instance_uuid', [
    INSTANCE_TO_SET_MASTER_UUID,
    INSTANCE_NOT_IN_LEADERS_UUID,
])
def test_set_master(cartridge_cmd, instance_uuid, tmpdir):
    data_dir = os.path.join(tmpdir, 'tmp', 'data')
    os.makedirs(data_dir)

    old_conf = copy.deepcopy(SIMPLE_CONF)

    # create app working directories
    instances = ['instance-1', 'instance-2']
    conf_paths = write_instance_topology_conf(data_dir, APPNAME, old_conf, instances)

    # create other app working directories
    other_instances = ['other-instance-1', 'other-instance-2']
    other_app_conf_paths = write_instance_topology_conf(data_dir, OTHER_APP_NAME, old_conf, other_instances)

    cmd = [
        cartridge_cmd, 'repair', 'set-leader',
        '--name', APPNAME,
        '--data-dir', data_dir,
        RPL_TO_SET_MASER_UUID, instance_uuid,
    ]

    rc, output = run_command_and_get_output(cmd, cwd=tmpdir)
    assert rc == 0

    # check logs
    logs = get_logs(output)
    assert len(logs) == len(instances) + 1
    assert logs[0] == "Set %s master to %s" % (RPL_TO_SET_MASER_UUID, instance_uuid)
    assert all([line.strip().endswith('OK') for line in logs[1:]])
    assert_for_all_instances(
        logs[1:], APPNAME, instances, lambda line: line.strip().endswith('OK'),
    )

    # check app config changes
    new_conf = copy.deepcopy(old_conf)
    new_leaders = new_conf['replicasets'][RPL_TO_SET_MASER_UUID]['master']
    if instance_uuid in new_leaders:
        new_leaders.remove(instance_uuid)

    new_leaders.insert(0, instance_uuid)

    for conf_path in conf_paths:
        assert os.path.exists(conf_path)

        with open(conf_path, 'r') as f:
            conf = yaml.safe_load(f.read())
            assert conf == new_conf

        # check backup
        backup_conf_path = '%s.bak' % conf_path
        assert os.path.exists(backup_conf_path)

        with open(backup_conf_path, 'r') as f:
            conf = yaml.safe_load(f.read())
            assert conf == old_conf

    # check that other app config wasn't changed
    for conf_path in other_app_conf_paths:
        assert os.path.exists(conf_path)

        with open(conf_path, 'r') as f:
            conf = yaml.safe_load(f.read())
            assert conf == old_conf

        # check backup
        backup_conf_path = '%s.bak' % conf_path
        assert not os.path.exists(backup_conf_path)


exp_rpl_diffs = {
    INSTANCE_TO_SET_MASTER_UUID: '\n'.join([
        '   %s:' % RPL_TO_SET_MASER_UUID,
        '     alias: unnamed',
        '     all_rw: false',
        '     master:',
        '+    - %s' % INSTANCE_TO_SET_MASTER_UUID,
        '     - srv-1-uuid',
        '     - srv-2-uuid',
        '-    - %s' % INSTANCE_TO_SET_MASTER_UUID,
    ]),

    INSTANCE_NOT_IN_LEADERS_UUID: '\n'.join([
        '   %s:' % RPL_TO_SET_MASER_UUID,
        '     alias: unnamed',
        '     all_rw: false',
        '     master:',
        '+    - %s' % INSTANCE_NOT_IN_LEADERS_UUID,
        '     - srv-1-uuid',
        '     - srv-2-uuid',
        '     - %s' % INSTANCE_TO_SET_MASTER_UUID,
    ]),
}


@pytest.mark.parametrize('instance_uuid', [
    INSTANCE_TO_SET_MASTER_UUID,
    INSTANCE_NOT_IN_LEADERS_UUID,
])
def test_set_master_dry_run(cartridge_cmd, instance_uuid, tmpdir):
    data_dir = os.path.join(tmpdir, 'tmp', 'data')
    os.makedirs(data_dir)

    old_conf = copy.deepcopy(SIMPLE_CONF)

    # create app working directories
    instances = ['instance-1', 'instance-2']
    conf_paths = write_instance_topology_conf(data_dir, APPNAME, old_conf, instances)

    cmd = [
        cartridge_cmd, 'repair', 'set-leader',
        '--name', APPNAME,
        '--data-dir', data_dir,
        '--dry-run',
        RPL_TO_SET_MASER_UUID, instance_uuid,
    ]

    rc, output = run_command_and_get_output(cmd, cwd=tmpdir)
    assert rc == 0

    # check logs
    assert "Set %s master to %s" % (RPL_TO_SET_MASER_UUID, instance_uuid) in output
    assert "Data directory is set to: %s" % data_dir in output

    assert all([
        'Topology config file: %s' % conf_path in output
        for conf_path in conf_paths
    ])

    exp_rpl_diff = exp_rpl_diffs[instance_uuid]
    assert exp_rpl_diff in output

    # check config wasn't changed
    for conf_path in conf_paths:
        assert os.path.exists(conf_path)

        with open(conf_path, 'r') as f:
            conf = yaml.safe_load(f.read())
            assert conf == old_conf

        # check backup
        backup_conf_path = '%s.bak' % conf_path
        assert not os.path.exists(backup_conf_path)
