import os
import copy

from utils import write_instance_topology_conf
from utils import run_command_and_get_output


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


def test_list(cartridge_cmd, tmpdir):
    data_dir = os.path.join(tmpdir, 'tmp', 'data')
    os.makedirs(data_dir)

    old_conf = copy.deepcopy(SIMPLE_CONF)

    # create app working directories
    instances = ['instance-1']
    write_instance_topology_conf(data_dir, APPNAME, old_conf, instances)

    # create other app working directories
    other_instances = ['other-instance-1', 'other-instance-2']
    write_instance_topology_conf(data_dir, OTHER_APP_NAME, old_conf, other_instances)

    cmd = [
        cartridge_cmd, 'repair', 'list-topology',
        '--name', APPNAME,
        '--data-dir', data_dir,
    ]

    rc, output = run_command_and_get_output(cmd, cwd=tmpdir)
    assert rc == 0

    assert 'Instances' in output
    assert 'Replicasets' in output

    # check instances summary
    instances_list = output[output.find('Instances')+len('Instances')+1:output.find('Replicasets')]
    instances_list = instances_list.strip()

    if instances_list.startswith('* '):
        instances_list = instances_list[2:]

    instances_summary = set([
        instance_summary.strip()
        for instance_summary in instances_list.split('\n*') if instance_summary != ''
    ])

    exp_instances_summary = set([
        '\n'.join([
            'srv-1-uuid',
            '\tURI: localhost:3301',
            '\treplicaset: rpl-1-uuid',
        ]),
        '\n'.join([
            'srv-2-uuid',
            '\tURI: localhost:3302',
            '\treplicaset: rpl-1-uuid',
        ]),
        '\n'.join([
            'srv-3-uuid',
            '\tURI: localhost:3303',
            '\treplicaset: rpl-1-uuid',
        ]),
        '\n'.join([
            'srv-4-uuid',
            '\tURI: localhost:3304',
            '\treplicaset: rpl-2-uuid',
        ]),
        '\n'.join([
            'srv-5-uuid',
            '\tURI: localhost:3305',
            '\treplicaset: rpl-1-uuid',
        ]),
        '\n'.join([
            'srv-6-uuid disabled',
            '\tURI: localhost:3306',
            '\treplicaset: rpl-2-uuid',
        ]),
        'srv-expelled expelled',
    ])

    assert instances_summary == exp_instances_summary

    # check instances summary
    replicasets_list = output[output.find('Replicasets')+len('Replicasets')+1:]
    replicasets_list = replicasets_list.strip()

    if replicasets_list.startswith('* '):
        replicasets_list = replicasets_list[2:]

    replicasets_summary = set([
        replicaset_summary.strip()
        for replicaset_summary in replicasets_list.split('\n*') if replicaset_summary != ''
    ])

    exp_replicasets_summary = set([
        '\n'.join([
            'rpl-1-uuid',
            '\troles:',
            '\t * vshard-storage',
            '\tinstances:',
            '\t * srv-1-uuid',
            '\t * srv-2-uuid',
            '\t * srv-3-uuid',
            '\t * srv-5-uuid',
        ]),
        '\n'.join([
            'rpl-2-uuid',
            '\troles:',
            '\t * vshard-storage',
            '\tinstances:',
            '\t * srv-4-uuid',
            '\t * srv-6-uuid',
        ]),
    ])

    assert replicasets_summary == exp_replicasets_summary
