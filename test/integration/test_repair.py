import os
import re
import pytest

from utils import run_command_and_get_output

from clusterwide_conf import write_instances_topology_conf

APPNAME = 'myapp'
OTHER_APP_NAME = 'other-app'


simple_args = {
    'set-uri': ['localhost:3301', 'localhost:3310'],
    'remove-instance': ['srv-1-uri'],
    'set-leader': ['rpl-1-uri', 'srv-1-uri']
}


########
# COMMON
@pytest.mark.parametrize('repair_cmd', ['set-uri', 'remove-instance'])
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
    assert re.search(r"Data directory \S+ doesn't exist", output) is not None

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


@pytest.mark.parametrize('repair_cmd', ['set-uri', 'remove-instance'])
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


@pytest.mark.parametrize('repair_cmd', ['set-uri', 'remove-instance'])
def test_repiar_no_workdirs(cartridge_cmd, clusterwide_conf_simple, repair_cmd, tmpdir):
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
    write_instances_topology_conf(data_dir, OTHER_APP_NAME, clusterwide_conf_simple.conf, instances)

    rc, output = run_command_and_get_output(cmd, cwd=tmpdir)
    assert rc == 1
    assert "No instance working directories found in %s" % data_dir in output
