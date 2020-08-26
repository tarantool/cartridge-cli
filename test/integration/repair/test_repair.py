import os
import re
import pytest

from utils import run_command_and_get_output
from utils import get_logs

from clusterwide_conf import write_instances_topology_conf

APPNAME = 'myapp'
OTHER_APP_NAME = 'other-app'


simple_args = {
    'set-advertise-uri': ['localhost:3301', 'localhost:3310'],
    'remove-instance': ['srv-1-uri'],
    'set-leader': ['rpl-1-uri', 'srv-1-uri']
}


########
# COMMON
@pytest.mark.parametrize('repair_cmd', ['set-advertise-uri', 'remove-instance', 'set-leader', 'list-topology'])
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


@pytest.mark.parametrize('repair_cmd', ['set-advertise-uri', 'remove-instance', 'set-leader', 'list-topology'])
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


@pytest.mark.parametrize('repair_cmd', ['set-advertise-uri', 'remove-instance', 'set-leader', 'list-topology'])
def test_repiar_no_workdirs(cartridge_cmd, clusterwide_conf_simple, repair_cmd, tmpdir):
    data_dir = os.path.join(tmpdir, 'tmp', 'data')
    os.makedirs(data_dir)

    args = simple_args.get(repair_cmd, [])

    cmd = [
        cartridge_cmd, 'repair', repair_cmd,
        '--name', APPNAME,
        '--data-dir', data_dir,
    ]

    if repair_cmd != 'list-topology':
        cmd.append('--no-reload')

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


@pytest.mark.parametrize('repair_cmd', ['set-advertise-uri', 'remove-instance', 'set-leader'])
def test_non_bootstrapped_instance(cartridge_cmd, clusterwide_conf_simple, repair_cmd, tmpdir):
    data_dir = os.path.join(tmpdir, 'tmp', 'data')
    os.makedirs(data_dir)

    config = clusterwide_conf_simple

    if repair_cmd == 'set-advertise-uri':
        args = [config.instance_uuid, config.instance_uri]
    elif repair_cmd == 'remove-instance':
        args = [config.instance_uuid]
    elif repair_cmd == 'set-leader':
        args = [config.replicaset_uuid, config.instance_uuid]

    cmd = [
        cartridge_cmd, 'repair', repair_cmd,
        '--name', APPNAME,
        '--data-dir', data_dir,
        '--no-reload',
    ]
    cmd.extend(args)

    instances = ['instance-1', 'instance-2']

    # no cluster-wide configs

    # # create empty work dirs for instance-2
    for instance in instances:
        work_dir = os.path.join(data_dir, '%s.%s' % (APPNAME, instance))
        os.makedirs(work_dir)

    rc, output = run_command_and_get_output(cmd, cwd=tmpdir)
    assert rc == 1
    assert "No cluster-wide configs found in %s" % data_dir in output

    # write config for instance-1
    write_instances_topology_conf(data_dir, APPNAME, clusterwide_conf_simple.conf, instances[:1])

    rc, output = run_command_and_get_output(cmd, cwd=tmpdir)
    assert rc == 0

    if repair_cmd == 'set-advertise-uri':
        first_log_line = "Set %s advertise URI to %s" % (args[0], args[1])
    elif repair_cmd == 'remove-instance':
        first_log_line = "Remove instance with UUID %s" % args[0]
    elif repair_cmd == 'set-leader':
        first_log_line = "Set %s leader to %s" % (args[0], args[1])

    logs = get_logs(output)
    assert len(logs) == 6
    assert logs[0] == first_log_line
    assert logs[1] == "Process application cluster-wide configurations..."
    assert logs[2] == "%s... OK" % instances[0]
    assert logs[3] == "Write application cluster-wide configurations..."
    assert logs[4] == "Reloading cluster-wide configurations is skipped"
    assert logs[5] == "%s... OK" % instances[0]
