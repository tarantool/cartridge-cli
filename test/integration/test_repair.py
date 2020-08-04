import os
import pytest

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
            'master': ['srv-1-uuid'],
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
    }
}


simple_args = {
    'set-uri': ['localhost:3301', 'localhost:3310'],
    'remove-instance': ['srv-1-uri'],
    'set-leader': ['rpl-1-uri', 'srv-1-uri']
}


########
# COMMON
@pytest.mark.parametrize('repair_cmd', ['set-uri', 'remove-instance', 'set-leader', 'list-topology'])
def test_repiar_bad_data_dir(cartridge_cmd, repair_cmd, tmpdir):

    args = simple_args.get(repair_cmd, [])

    # non existent path
    cmd = [
        cartridge_cmd, 'repair', repair_cmd,
        '--name', APPNAME,
        '--data-dir', 'non/existent/path',
    ]
    cmd.extend(args)

    rc, output = run_command_and_get_output(cmd, cwd=tmpdir)
    assert rc == 1
    assert "Specified data directory doesn't exist" in output

    # file instead of the directory
    filepath = os.path.join(tmpdir, 'data-dir-file')
    with open(filepath, 'w') as f:
        f.write("Hi")

    cmd = [
        cartridge_cmd, 'repair', repair_cmd,
        '--name', APPNAME,
        '--data-dir', filepath,
    ]
    cmd.extend(args)

    rc, output = run_command_and_get_output(cmd, cwd=tmpdir)
    assert rc == 1
    assert "is not a directory" in output


@pytest.mark.parametrize('repair_cmd', ['set-uri', 'remove-instance', 'set-leader', 'list-topology'])
def test_repiar_no_name(cartridge_cmd, repair_cmd, tmpdir):
    data_dir = os.path.join(tmpdir, 'tmp', 'data')
    os.makedirs(data_dir)

    args = simple_args.get(repair_cmd, [])

    cmd = [
        cartridge_cmd, 'repair', repair_cmd,
        '--data-dir', data_dir,
    ]
    cmd.extend(args)

    rc, output = run_command_and_get_output(cmd, cwd=tmpdir)
    assert rc == 1
    assert "Please, specify application name using --name" in output


@pytest.mark.parametrize('repair_cmd', ['set-uri', 'remove-instance', 'set-leader', 'list-topology'])
def test_repiar_no_workdirs(cartridge_cmd, repair_cmd, tmpdir):
    data_dir = os.path.join(tmpdir, 'tmp', 'data')
    os.makedirs(data_dir)

    args = simple_args.get(repair_cmd, [])

    cmd = [
        cartridge_cmd, 'repair', repair_cmd,
        '--name', APPNAME,
        '--data-dir', data_dir,
    ]
    cmd.extend(args)

    rc, output = run_command_and_get_output(cmd, cwd=tmpdir)
    assert rc == 1
    assert "No instance working directories found in %s" % data_dir in output

    # create other app workdirs
    instances = ['instance-1', 'instance-2']
    write_instance_topology_conf(data_dir, OTHER_APP_NAME, SIMPLE_CONF, instances)

    rc, output = run_command_and_get_output(cmd, cwd=tmpdir)
    assert rc == 1
    assert "No instance working directories found in %s" % data_dir in output
