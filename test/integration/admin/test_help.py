from utils import run_command_and_get_output
from utils import get_log_lines


def test_help_many_args(cartridge_cmd, custom_admin_running_instances, tmpdir):
    project = custom_admin_running_instances['project']
    run_dir = project.get_run_dir()

    cmd = [
        cartridge_cmd, 'admin',
        '--name', project.name,
        '--run-dir', run_dir,
        'echo_user', '--help',
    ]
    rc, output = run_command_and_get_output(cmd, cwd=tmpdir)
    assert rc == 0

    assert get_log_lines(output) == [
        '• Admin function "echo_user" usage:',
        'echo_user usage',
        'Args:',
        '--age number           age usage',
        '--loves-cakes boolean  loves_cakes usage',
        '--username string      username usage',
    ]


def test_help_no_args(cartridge_cmd, custom_admin_running_instances, tmpdir):
    project = custom_admin_running_instances['project']
    run_dir = project.get_run_dir()

    cmd = [
        cartridge_cmd, 'admin',
        '--name', project.name,
        '--run-dir', run_dir,
        'func_no_args', '--help',
    ]
    rc, output = run_command_and_get_output(cmd, cwd=tmpdir)
    assert rc == 0

    assert get_log_lines(output) == [
        '• Admin function "func_no_args" usage:',
        'func_no_args usage',
    ]


def test_help_long_func_name(cartridge_cmd, custom_admin_running_instances, tmpdir):
    project = custom_admin_running_instances['project']
    run_dir = project.get_run_dir()

    exp_output_lines = [
        '• Admin function "func.long.name" usage:',
        'func_long_name usage',
    ]

    cmd = [
        cartridge_cmd, 'admin',
        '--name', project.name,
        '--run-dir', run_dir,
        'func.long.name', '--help',
    ]
    rc, output = run_command_and_get_output(cmd, cwd=tmpdir)
    assert rc == 0
    assert get_log_lines(output) == exp_output_lines

    cmd = [
        cartridge_cmd, 'admin',
        '--name', project.name,
        '--run-dir', run_dir,
        'func', 'long', 'name', '--help',
    ]
    rc, output = run_command_and_get_output(cmd, cwd=tmpdir)
    assert rc == 0
    assert get_log_lines(output) == exp_output_lines
