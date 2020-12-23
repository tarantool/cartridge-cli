import pytest
import subprocess

from utils import run_command_and_get_output
from utils import start_instances
from utils import get_log_lines

from project import patch_cartridge_proc_titile


@pytest.fixture(scope="function")
def default_admin_running_instances(cartridge_cmd, start_stop_cli, project_with_cartridge):
    project = project_with_cartridge

    # build project
    cmd = [
        cartridge_cmd,
        "build",
    ]
    process = subprocess.run(cmd, cwd=project.path)
    assert process.returncode == 0, "Error during building the project"

    # don't change process title
    patch_cartridge_proc_titile(project)

    # start instances
    start_instances(cartridge_cmd, start_stop_cli, project)

    return {
        'project': project,
    }


def test_default_admin_func(cartridge_cmd, default_admin_running_instances, tmpdir):
    project = default_admin_running_instances['project']
    run_dir = project.get_run_dir()

    # list
    cmd = [
        cartridge_cmd, 'admin',
        '--name', project.name,
        '--run-dir', run_dir,
        '--list'
    ]
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
