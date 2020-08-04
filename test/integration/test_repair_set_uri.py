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

INSTANCE_TO_CHANGE_URI_UUID = 'srv-2-uuid'
DISABLED_INSTANCE_UUID = 'srv-disabled'
EXPELLED_INSTANCE_UUID = 'srv-expelled'

NEW_URI = 'localhost:3311'

SIMPLE_CONF = {
    'failover': False,
    'replicasets': {
        'rpl-1-uuid': {
            'alias': 'unnamed',
            'all_rw': False,
            'master': ['srv-1-uuid', 'srv-2-uuid', 'srv-3-uuid'],
            'roles': {'vshard-storage': True},
            'vshard_group': 'default',
            'weight': 1
        },
    },
    'servers': {
        'srv-1-uuid': {
            'disabled': False,
            'replicaset_uuid': 'rpl-1-uuid',
            'uri': 'localhost:3301',
        },
        INSTANCE_TO_CHANGE_URI_UUID: {
            'disabled': False,
            'replicaset_uuid': 'rpl-1-uuid',
            'uri': 'localhost:3302',
        },
        DISABLED_INSTANCE_UUID: {
            'disabled': True,
            'replicaset_uuid': 'rpl-1-uuid',
            'uri': 'localhost:3303',
        },
        EXPELLED_INSTANCE_UUID: 'expelled',
    }
}


def get_uri(conf, instance_uuid):
    return conf['servers'][instance_uuid]['uri']


exp_diffs = {
    INSTANCE_TO_CHANGE_URI_UUID: '\n'.join([
        '   %s:' % INSTANCE_TO_CHANGE_URI_UUID,
        '     disabled: false',
        '     replicaset_uuid: rpl-1-uuid',
        '-    uri: %s' % get_uri(SIMPLE_CONF, INSTANCE_TO_CHANGE_URI_UUID),
        '+    uri: %s' % NEW_URI,
    ]),
    DISABLED_INSTANCE_UUID: '\n'.join([
        '   %s:' % DISABLED_INSTANCE_UUID,
        '     disabled: true',
        '     replicaset_uuid: rpl-1-uuid',
        '-    uri: %s' % get_uri(SIMPLE_CONF, DISABLED_INSTANCE_UUID),
        '+    uri: %s' % NEW_URI,
    ]),
}


def test_uri_does_not_exist(cartridge_cmd, tmpdir):
    data_dir = os.path.join(tmpdir, 'tmp', 'data')
    os.makedirs(data_dir)

    old_conf = copy.deepcopy(SIMPLE_CONF)

    NON_EXISTENT_URI = 'non-existant-uri'

    instances = ['instance-1', 'instance-2']
    write_instance_topology_conf(data_dir, APPNAME, old_conf, instances)

    cmd = [
        cartridge_cmd, 'repair', 'set-uri',
        '--name', APPNAME,
        '--data-dir', data_dir,
        NON_EXISTENT_URI, NEW_URI,
    ]

    rc, output = run_command_and_get_output(cmd, cwd=tmpdir)
    assert rc == 1

    assert_for_all_instances(
        get_logs(output), APPNAME, instances, lambda line:
        "Instance with URI %s isn't found in the cluster" % NON_EXISTENT_URI in line
    )


@pytest.mark.parametrize('instance_uuid', [
    INSTANCE_TO_CHANGE_URI_UUID,
    DISABLED_INSTANCE_UUID,
])
def test_uri(cartridge_cmd, instance_uuid, tmpdir):
    data_dir = os.path.join(tmpdir, 'tmp', 'data')
    os.makedirs(data_dir)

    old_conf = copy.deepcopy(SIMPLE_CONF)
    old_uri = get_uri(old_conf, instance_uuid)

    # create app working directories
    instances = ['instance-1', 'instance-2']
    conf_paths = write_instance_topology_conf(data_dir, APPNAME, old_conf, instances)

    # create other app working directories
    other_instances = ['other-instance-1', 'other-instance-2']
    other_app_conf_paths = write_instance_topology_conf(data_dir, OTHER_APP_NAME, old_conf, other_instances)

    cmd = [
        cartridge_cmd, 'repair', 'set-uri',
        '--name', APPNAME,
        '--data-dir', data_dir,
        old_uri, NEW_URI,
    ]

    rc, output = run_command_and_get_output(cmd, cwd=tmpdir)
    assert rc == 0

    # check logs
    logs = get_logs(output)
    assert len(logs) == len(instances) + 1
    assert logs[0] == "Update advertise URI %s -> %s" % (old_uri, NEW_URI)
    assert all([line.strip().endswith('OK') for line in logs[1:]])
    assert_for_all_instances(
        logs[1:], APPNAME, instances, lambda line: line.strip().endswith('OK'),
    )

    # check app config changes
    new_conf = copy.deepcopy(old_conf)
    new_conf['servers'][instance_uuid]['uri'] = NEW_URI

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


@pytest.mark.parametrize('instance_uuid', [
    INSTANCE_TO_CHANGE_URI_UUID,
    DISABLED_INSTANCE_UUID,
])
def test_uri_dry_run(cartridge_cmd, instance_uuid, tmpdir):
    data_dir = os.path.join(tmpdir, 'tmp', 'data')
    os.makedirs(data_dir)

    old_conf = copy.deepcopy(SIMPLE_CONF)

    old_uri = old_conf['servers'][instance_uuid]['uri']

    instances = ['instance-1', 'instance-2']
    conf_paths = write_instance_topology_conf(data_dir, APPNAME, old_conf, instances)

    cmd = [
        cartridge_cmd, 'repair', 'set-uri',
        '--name', APPNAME,
        '--data-dir', data_dir,
        '--dry-run',
        old_uri, NEW_URI,
    ]

    rc, output = run_command_and_get_output(cmd, cwd=tmpdir)
    assert rc == 0

    # check logs
    assert "Update advertise URI %s -> %s" % (old_uri, NEW_URI) in output
    assert "Data directory is set to: %s" % data_dir in output

    assert all([
        'Topology config file: %s' % conf_path in output
        for conf_path in conf_paths
    ])

    exp_diff = exp_diffs[instance_uuid]
    assert exp_diff in output

    # check config wasn't changed
    for conf_path in conf_paths:
        assert os.path.exists(conf_path)

        with open(conf_path, 'r') as f:
            conf = yaml.safe_load(f.read())
            assert conf == old_conf

        # check backup
        backup_conf_path = '%s.bak' % conf_path
        assert not os.path.exists(backup_conf_path)
