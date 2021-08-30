import copy
import os

from clusterwide_conf import (assert_conf_not_changed,
                              write_instances_topology_conf)
from utils import (assert_ok_for_instances_group, get_logs,
                   run_command_and_get_output)

APPNAME = 'myapp'
OTHER_APP_NAME = 'other-app'


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
        'rpl-2-uuid': {
            'alias': 'unnamed',
            'all_rw': False,
            'master': ['srv-4-uuid'],
            'roles': {'vshard-storage': True},
            'vshard_group': 'default',
            'weight': 1
        },
    },
    'servers': {
        'srv-1-uuid': {
            'disabled': False,
            'replicaset_uuid':  'rpl-1-uuid',
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
        'srv-4-uuid': {
            'disabled': False,
            'replicaset_uuid': 'rpl-2-uuid',
            'uri': 'localhost:3304',
        },
        'srv-5-uuid': {
            'disabled': False,
            'replicaset_uuid':  'rpl-1-uuid',
            'uri': 'localhost:3305',
        },
        'srv-6-uuid': {
            'disabled': True,
            'replicaset_uuid': 'rpl-2-uuid',
            'uri': 'localhost:3306',
        },
        'srv-expelled': 'expelled',
    }
}


def test_list_topology(cartridge_cmd, tmpdir):
    data_dir = os.path.join(tmpdir, 'tmp', 'data')
    os.makedirs(data_dir)

    old_conf = copy.deepcopy(SIMPLE_CONF)

    # create app configs
    instances = ['instance-1', 'instance-2']
    conf_paths = write_instances_topology_conf(data_dir, APPNAME, old_conf, instances)

    # create other app configs
    other_instances = ['other-instance-1', 'other-instance-2']
    other_app_conf_paths = write_instances_topology_conf(data_dir, OTHER_APP_NAME, old_conf, other_instances)

    cmd = [
        cartridge_cmd, 'repair', 'list-topology',
        '--name', APPNAME,
        '--data-dir', data_dir,
    ]

    rc, output = run_command_and_get_output(cmd, cwd=tmpdir)
    assert rc == 0

    assert_conf_not_changed(conf_paths, old_conf)
    assert_conf_not_changed(other_app_conf_paths, old_conf)

    lines = output.split('\n')
    logs = get_logs('\n'.join(lines[:3]))

    assert logs[0] == "Get current topology"
    assert logs[1] == "Process application cluster-wide configurations..."
    assert_ok_for_instances_group(logs, instances)

    summary = '\n'.join(lines[3:])

    exp_summary = '''Instances
  * srv-1-uuid
    URI: localhost:3301
    replicaset: rpl-1-uuid
  * srv-2-uuid
    URI: localhost:3302
    replicaset: rpl-1-uuid
  * srv-3-uuid
    URI: localhost:3303
    replicaset: rpl-1-uuid
  * srv-4-uuid
    URI: localhost:3304
    replicaset: rpl-2-uuid
  * srv-5-uuid
    URI: localhost:3305
    replicaset: rpl-1-uuid
  * srv-6-uuid disabled
    URI: localhost:3306
    replicaset: rpl-2-uuid
  * srv-expelled expelled
Replicasets
  * rpl-1-uuid
    roles:
     * vshard-storage
    instances:
     * srv-1-uuid
     * srv-2-uuid
     * srv-3-uuid
     * srv-5-uuid
  * rpl-2-uuid
    roles:
     * vshard-storage
    instances:
     * srv-4-uuid
     * srv-6-uuid

'''

    assert summary == exp_summary
