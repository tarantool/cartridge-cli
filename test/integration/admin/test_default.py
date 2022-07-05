import pytest
from project import copy_project, patch_cartridge_proc_titile
from utils import (get_admin_connection_params, get_log_lines,
                   run_command_and_get_output, start_instances)


@pytest.fixture(scope="function")
def default_admin_running_instances(start_stop_cli, project_with_cartridge):
    project = project_with_cartridge
    copy_project("project_with_cartridge", project)

    # don't change process title
    patch_cartridge_proc_titile(project)

    # start instances
    start_instances(start_stop_cli, project)

    return {
        'project': project,
    }


@pytest.mark.parametrize('connection_type', ['find-socket', 'connect', 'instance'])
def test_default_admin_func(cartridge_cmd, default_admin_running_instances, connection_type, tmpdir):
    project = default_admin_running_instances['project']
    run_dir = project.get_run_dir()

    # list
    cmd = [
        cartridge_cmd, 'admin',
        '--list'
    ]
    cmd.extend(get_admin_connection_params(connection_type, project))

    rc, output = run_command_and_get_output(cmd, cwd=tmpdir)
    assert rc == 0

    assert get_log_lines(output) == [
        '• Available admin functions:',
        'probe  Probe instance',
    ]

    # help
    cmd = [
        cartridge_cmd, 'admin',
        '--name', project.name,
        '--run-dir', run_dir,
        '--help', 'probe',
    ]
    rc, output = run_command_and_get_output(cmd, cwd=tmpdir)
    assert rc == 0

    assert get_log_lines(output) == [
        '• Admin function "probe" usage:',
        'Probe instance',
        'Args:',
        '--uri string  Instance URI',
    ]

    # call w/ --uri localhost:3301 - OK
    cmd = [
        cartridge_cmd, 'admin',
        '--name', project.name,
        '--run-dir', run_dir,
        'probe', '--uri', 'localhost:3301',
    ]
    rc, output = run_command_and_get_output(cmd, cwd=tmpdir)
    assert rc == 0

    assert get_log_lines(output) == [
        '• Probe "localhost:3301": OK',
    ]

    # call w/ --uri localhost:3311 - fail
    cmd = [
        cartridge_cmd, 'admin',
        '--name', project.name,
        '--run-dir', run_dir,
        'probe', '--uri', 'localhost:3311',
    ]
    rc, output = run_command_and_get_output(cmd, cwd=tmpdir)
    assert rc == 1

    assert get_log_lines(output) == [
        '⨯ Failed to call "probe": Probe "localhost:3311" failed: no response',
    ]
