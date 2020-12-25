from utils import run_command_and_get_output
from utils import get_log_lines


def test_call_many_args(cartridge_cmd, custom_admin_running_instances, tmpdir):
    project = custom_admin_running_instances['project']
    run_dir = project.get_run_dir()

    base_cmd = [
        cartridge_cmd, 'admin',
        '--name', project.name,
        '--run-dir', run_dir,
        'echo_user',
    ]

    # all args
    cmd = base_cmd + [
        '--username', 'Elizabeth',
        '--age', '23',
        '--loves-cakes',
    ]
    rc, output = run_command_and_get_output(cmd, cwd=tmpdir)
    assert rc == 0

    assert get_log_lines(output) == [
        '• Hi, Elizabeth!',
        '• You are 23 years old',
        '• I know that you like cakes!',
    ]

    # age missed
    # check that default number flag value (0) isn't passed
    cmd = base_cmd + [
        '--username', 'Elizabeth',
        '--loves-cakes',
    ]
    rc, output = run_command_and_get_output(cmd, cwd=tmpdir)
    assert rc == 0

    assert get_log_lines(output) == [
        "• Hi, Elizabeth!",
        "• I don't know your age",
        "• I know that you like cakes!",
    ]

    # bool flag is false
    cmd = base_cmd + [
        '--username', 'Elizabeth',
        '--loves-cakes=false',
    ]
    rc, output = run_command_and_get_output(cmd, cwd=tmpdir)
    assert rc == 0

    assert get_log_lines(output) == [
        "• Hi, Elizabeth!",
        "• I don't know your age",
        "• How can you not love cakes?",
    ]


def test_func_long_arg(cartridge_cmd, custom_admin_running_instances, tmpdir):
    project = custom_admin_running_instances['project']
    run_dir = project.get_run_dir()

    cmd = [
        cartridge_cmd, 'admin',
        '--name', project.name,
        '--run-dir', run_dir,
        'func_long_arg', '--long-arg', 'some-value',
    ]
    rc, output = run_command_and_get_output(cmd, cwd=tmpdir)
    assert rc == 0

    assert get_log_lines(output) == [
        '• func_long_arg was called with "some-value" arg',
    ]


def test_func_rets_str(cartridge_cmd, custom_admin_running_instances, tmpdir):
    project = custom_admin_running_instances['project']
    run_dir = project.get_run_dir()

    cmd = [
        cartridge_cmd, 'admin',
        '--name', project.name,
        '--run-dir', run_dir,
        'func_rets_str',
    ]
    rc, output = run_command_and_get_output(cmd, cwd=tmpdir)
    assert rc == 0

    assert get_log_lines(output) == [
        '• func_rets_str was called',
    ]


def test_func_rets_non_str(cartridge_cmd, custom_admin_running_instances, tmpdir):
    project = custom_admin_running_instances['project']
    run_dir = project.get_run_dir()

    cmd = [
        cartridge_cmd, 'admin',
        '--name', project.name,
        '--run-dir', run_dir,
        'func_rets_non_str',
    ]
    rc, output = run_command_and_get_output(cmd, cwd=tmpdir)
    assert rc == 0

    assert get_log_lines(output) == [
        '• 666',
        '• Admin function should return string or string array value',
    ]


def test_func_rets_err(cartridge_cmd, custom_admin_running_instances, tmpdir):
    project = custom_admin_running_instances['project']
    run_dir = project.get_run_dir()

    cmd = [
        cartridge_cmd, 'admin',
        '--name', project.name,
        '--run-dir', run_dir,
        'func_rets_err',
    ]
    rc, output = run_command_and_get_output(cmd, cwd=tmpdir)
    assert rc == 1

    assert get_log_lines(output) == [
        '⨯ Failed to call "func_rets_err": Some horrible error',
    ]


def test_func_raises_err(cartridge_cmd, custom_admin_running_instances, tmpdir):
    project = custom_admin_running_instances['project']
    run_dir = project.get_run_dir()

    cmd = [
        cartridge_cmd, 'admin',
        '--name', project.name,
        '--run-dir', run_dir,
        'func_raises_err',
    ]
    rc, output = run_command_and_get_output(cmd, cwd=tmpdir)
    assert rc == 1

    assert '⨯ Failed to call "func_raises_err":' in output
    assert 'Some horrible error raised' in output


def test_print(cartridge_cmd, custom_admin_running_instances, tmpdir):
    project = custom_admin_running_instances['project']
    run_dir = project.get_run_dir()

    ITERATIONS_NUM = 3

    cmd = [
        cartridge_cmd, 'admin',
        '--name', project.name,
        '--run-dir', run_dir,
        'func_print',
        '--num', str(ITERATIONS_NUM),
    ]
    rc, output = run_command_and_get_output(cmd, cwd=tmpdir)
    assert rc == 0

    iterations_output = []
    for i in range(1, ITERATIONS_NUM+1):
        iterations_output.extend([
            '• Iteration %s (printed)' % i,
            '• Iteration %s (pushed)' % i,
        ])

    assert get_log_lines(output) == iterations_output + [
        '• I am some great result',
    ]
