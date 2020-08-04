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


INSTANCE_TO_REMOVE_UUID = 'srv-1-uuid'
INSTANCE_NOT_IN_LEADERS_UUID = 'srv-not-in-leaders'
NON_EXISTENT_INSTANCE_UUID = 'non-existent-srv'
INSTANCE_FROM_NON_EXISTENT_RPL_UUID = 'srv-non-existent-rpl'
DISABLED_INSTANCE_UUID = 'srv-disabled'
EXPELLED_INSTANCE_UUID = 'srv-expelled'


SIMPLE_CONF = {
    'failover': False,
    'replicasets': {
        'rpl-1-uuid': {
            'alias': 'unnamed',
            'all_rw': False,
            'master': [INSTANCE_TO_REMOVE_UUID, 'srv-2-uuid', 'srv-3-uuid'],
            'roles': {'vshard-storage': True},
            'vshard_group': 'default',
            'weight': 1
        },
    },
    'servers': {
        INSTANCE_TO_REMOVE_UUID: {
            'disabled': False,
            'replicaset_uuid': 'rpl-1-uuid',
            'uri': 'localhost:3301',
        },
        'srv-2-uuid': {
            'disabled': False,
            'replicaset_uuid': 'rpl-1-uuid',
            'uri': 'localhost:3302',
        },
        'srv-3-uuid': {
            'disabled': False,
            'replicaset_uuid': 'rpl-1-uuid',
            'uri': 'localhost:3303',
        },
        INSTANCE_NOT_IN_LEADERS_UUID: {
            'disabled': False,
            'replicaset_uuid': 'rpl-1-uuid',
            'uri': 'localhost:3305',
        },
        DISABLED_INSTANCE_UUID: {
            'disabled': True,
            'replicaset_uuid': 'rpl-1-uuid',
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


def test_remove_uuid_does_not_exist(cartridge_cmd, tmpdir):
    data_dir = os.path.join(tmpdir, 'tmp', 'data')
    os.makedirs(data_dir)

    old_conf = copy.deepcopy(SIMPLE_CONF)

    instances = ['instance-1', 'instance-2']
    write_instance_topology_conf(data_dir, APPNAME, old_conf, instances)

    cmd = [
        cartridge_cmd, 'repair', 'remove-instance',
        '--name', APPNAME,
        '--data-dir', data_dir,
        NON_EXISTENT_INSTANCE_UUID,
    ]

    rc, output = run_command_and_get_output(cmd, cwd=tmpdir)
    assert rc == 1

    assert_for_all_instances(
        get_logs(output), APPNAME, instances, lambda line:
        "Instance %s isn't found in cluster" % NON_EXISTENT_INSTANCE_UUID in line
    )


@pytest.mark.parametrize('instance_uuid', [
    INSTANCE_TO_REMOVE_UUID,
    INSTANCE_NOT_IN_LEADERS_UUID,
    INSTANCE_FROM_NON_EXISTENT_RPL_UUID,
    DISABLED_INSTANCE_UUID,
    EXPELLED_INSTANCE_UUID,
])
def test_remove(cartridge_cmd, instance_uuid, tmpdir):
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
        cartridge_cmd, 'repair', 'remove-instance',
        '--name', APPNAME,
        '--data-dir', data_dir,
        instance_uuid,
    ]

    rc, output = run_command_and_get_output(cmd, cwd=tmpdir)
    assert rc == 0

    # check logs
    logs = get_logs(output)
    assert len(logs) == len(instances) + 1
    assert logs[0] == "Remove instance with UUID %s" % instance_uuid
    assert_for_all_instances(
        logs[1:], APPNAME, instances, lambda line:
        line.strip().endswith('OK')
    )

    # check config changes
    new_conf = copy.deepcopy(old_conf)

    if instance_uuid != EXPELLED_INSTANCE_UUID:
        replicaset_uuid = new_conf['servers'][instance_uuid]['replicaset_uuid']
        if replicaset_uuid in new_conf['replicasets']:
            new_leaders = new_conf['replicasets'][replicaset_uuid]['master']
            if instance_uuid in new_leaders:
                new_leaders.remove(instance_uuid)

    del new_conf['servers'][instance_uuid]

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


def get_instance_conf(conf, instance_uuid):
    return conf['servers'][instance_uuid]


exp_srv_diffs = {}
non_expelled_uuids = [
    INSTANCE_TO_REMOVE_UUID,
    INSTANCE_NOT_IN_LEADERS_UUID,
    INSTANCE_FROM_NON_EXISTENT_RPL_UUID,
    DISABLED_INSTANCE_UUID
]
for instance_uuid in non_expelled_uuids:
    instance_conf = get_instance_conf(SIMPLE_CONF, instance_uuid)
    disabled = 'true' if instance_uuid == DISABLED_INSTANCE_UUID else "false"

    exp_srv_diffs.update({instance_uuid: '\n'.join([
            '-  %s:' % instance_uuid,
            '-    disabled: %s' % disabled,
            '-    replicaset_uuid: %s' % instance_conf['replicaset_uuid'],
            '-    uri: %s' % instance_conf['uri'],
        ])
    })

exp_srv_diffs.update({
    EXPELLED_INSTANCE_UUID: '%s: expelled' % EXPELLED_INSTANCE_UUID,
})


@pytest.mark.parametrize('instance_uuid', [
    INSTANCE_TO_REMOVE_UUID,
    INSTANCE_NOT_IN_LEADERS_UUID,
    INSTANCE_FROM_NON_EXISTENT_RPL_UUID,
    DISABLED_INSTANCE_UUID,
    EXPELLED_INSTANCE_UUID,
])
def test_remove_dry_run(cartridge_cmd, instance_uuid, tmpdir):
    data_dir = os.path.join(tmpdir, 'tmp', 'data')
    os.makedirs(data_dir)

    old_conf = copy.deepcopy(SIMPLE_CONF)

    instances = ['instance-1', 'instance-2']
    conf_paths = write_instance_topology_conf(data_dir, APPNAME, old_conf, instances)

    cmd = [
        cartridge_cmd, 'repair', 'remove-instance',
        '--name', APPNAME,
        '--data-dir', data_dir,
        '--dry-run',
        instance_uuid,
    ]

    rc, output = run_command_and_get_output(cmd, cwd=tmpdir)
    assert rc == 0

    # check logs
    assert "Remove instance with UUID %s" % instance_uuid in output
    assert "Data directory is set to: %s" % data_dir in output

    assert all([
        'Topology config file: %s' % conf_path in output
        for conf_path in conf_paths
    ])

    exp_srv_diff = exp_srv_diffs[instance_uuid]
    assert exp_srv_diff in output

    if instance_uuid == INSTANCE_TO_REMOVE_UUID:
        exp_rpl_diff = '\n'.join([
            '   rpl-1-uuid:',
            '     alias: unnamed',
            '     all_rw: false',
            '     master:',
            '-    - %s' % instance_uuid,
            '     - srv-2-uuid',
            '     - srv-3-uuid',
        ])
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
