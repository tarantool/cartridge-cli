import os
import re

import pytest
from utils import run_command_and_get_output

simple_args = {
    'list': ['--list'],
    'help-func': ['func_no_args', '--help'],
    'call': ['func_no_args'],
}


@pytest.mark.parametrize('admin_flow', ['list', 'help-func', 'call'])
def test_bad_run_dir(cartridge_cmd, custom_admin_running_instances, admin_flow, tmpdir):
    project = custom_admin_running_instances['project']

    args = simple_args[admin_flow]

    # non-exitsent-path
    cmd = [
        cartridge_cmd, 'admin',
        '--name', project.name,
        '--run-dir', 'non/existent/path',
    ]
    cmd.extend(args)

    rc, output = run_command_and_get_output(cmd, cwd=tmpdir)
    assert rc == 1

    assert re.search(r"Run directory \S+ doesn't exist", output) is not None

    # file instead of the directory
    filepath = os.path.join(tmpdir, 'run-dir-file')
    with open(filepath, 'w') as f:
        f.write("Hi")

    cmd = [
        cartridge_cmd, 'admin',
        '--name', project.name,
        '--run-dir', filepath,
    ]
    cmd.extend(args)

    rc, output = run_command_and_get_output(cmd, cwd=tmpdir)
    assert rc == 1
    assert "is not a directory" in output

    # empty directory
    dirpath = os.path.join(tmpdir, 'empty-run-dir')
    os.makedirs(dirpath)

    cmd = [
        cartridge_cmd, 'admin',
        '--name', project.name,
        '--run-dir', dirpath,
    ]
    cmd.extend(args)

    rc, output = run_command_and_get_output(cmd, cwd=tmpdir)
    assert rc == 1

    assert re.search(r"Run directory \S+ is empty", output) is not None


@pytest.mark.parametrize('admin_flow', ['list', 'help-func', 'call'])
def test_connection_params_missed(cartridge_cmd, custom_admin_running_instances, admin_flow, tmpdir):
    project = custom_admin_running_instances['project']
    run_dir = project.get_run_dir()

    args = simple_args[admin_flow]

    cmd = [
        cartridge_cmd, 'admin',
        '--run-dir', run_dir,
    ]
    cmd.extend(args)

    rc, output = run_command_and_get_output(cmd, cwd=tmpdir)
    assert rc == 1
    assert "Please, specify one of --name, --instance or --conn" in output


@pytest.mark.parametrize('admin_flow', ['help-func', 'call'])
def test_non_existent_func(cartridge_cmd, custom_admin_running_instances, admin_flow, tmpdir):
    project = custom_admin_running_instances['project']
    run_dir = project.get_run_dir()

    simple_args = {
        'help-func': ['non_existent_func', '--help'],
        'call': ['non_existent_func'],
    }

    args = simple_args[admin_flow]

    cmd = [
        cartridge_cmd, 'admin',
        '--name', project.name,
        '--run-dir', run_dir,
    ]
    cmd.extend(args)

    rc, output = run_command_and_get_output(cmd, cwd=tmpdir)
    assert rc == 1
    assert 'Function "non_existent_func" isn\'t found' in output


@pytest.mark.parametrize('admin_flow', ['help-func', 'call'])
def test_conflicting_argnames(cartridge_cmd, custom_admin_running_instances, admin_flow, tmpdir):
    project = custom_admin_running_instances['project']
    run_dir = project.get_run_dir()

    simple_args = {
        'help-func': ['func_conflicting', '--help'],
        'call': ['func_conflicting'],
    }

    exp_return_codes = {
        'help-func': 0,
        'call': 1,
    }

    args = simple_args[admin_flow]
    exp_rc = exp_return_codes[admin_flow]

    cmd = [
        cartridge_cmd, 'admin',
        '--name', project.name,
        '--run-dir', run_dir,
    ]
    cmd.extend(args)

    rc, output = run_command_and_get_output(cmd, cwd=tmpdir)
    assert rc == exp_rc

    argnames = [
        "debug", "help", "instance", "list", "name", "quiet", "run_dir", "verbose",
    ]
    exp_err = 'Function has arguments with names that conflict with `cartridge admin` flags: %s' % ', '.join(
        ['"%s"' % argname for argname in argnames]
    )

    assert exp_err in output


@pytest.mark.parametrize('admin_flow', ['list', 'help-func', 'call'])
def test_instance_specified(cartridge_cmd, start_stop_cli, custom_admin_running_instances, admin_flow, tmpdir):
    project = custom_admin_running_instances['project']
    run_dir = project.get_run_dir()

    # stop instance-2
    INSTANCE1 = 'instance-1'
    INSTANCE2 = 'instance-2'

    start_stop_cli.stop(project, [INSTANCE2])

    args = simple_args[admin_flow]

    # don't specify any instance - instance-1 is chosen
    cmd = [
        cartridge_cmd, 'admin',
        '--name', project.name,
        '--run-dir', run_dir,
    ]
    cmd.extend(args)

    rc, output = run_command_and_get_output(cmd, cwd=tmpdir)
    assert rc == 0

    # specify --instance=instance-2 - fail
    cmd = [
        cartridge_cmd, 'admin',
        '--name', project.name,
        '--run-dir', run_dir,
        '--instance', INSTANCE2,
    ]
    cmd.extend(args)

    rc, output = run_command_and_get_output(cmd, cwd=tmpdir)
    assert rc == 1
    assert "Failed to dial" in output

    # specify --instance=instance-1 - ok
    cmd = [
        cartridge_cmd, 'admin',
        '--name', project.name,
        '--run-dir', run_dir,
        '--instance', INSTANCE1,
    ]
    cmd.extend(args)

    rc, output = run_command_and_get_output(cmd, cwd=tmpdir)
    assert rc == 0
