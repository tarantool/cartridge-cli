from utils import run_command_and_get_output
from utils import get_log_lines


def test_list(cartridge_cmd, custom_admin_running_instances, tmpdir):
    project = custom_admin_running_instances['project']
    run_dir = project.get_run_dir()

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
        'echo_user          echo_user usage',
        'func.long.name     func_long_name usage',
        'func_conflicting   func_conflicting usage',
        'func_long_arg      func_long_arg usage',
        'func_no_args       func_no_args usage',
        'func_raises_err    func_raises_err usage',
        'func_rets_err      func_rets_err usage',
        'func_rets_non_str  func_rets_non_str usage',
        'func_rets_str      func_rets_str usage',
    ]
