import pytest

from utils import run_command_and_get_output
from utils import get_log_lines
from utils import get_admin_connection_params


@pytest.mark.parametrize('connection_type', ['find-socket', 'connect', 'instance'])
def test_call_many_args(cartridge_cmd, custom_admin_running_instances, connection_type, tmpdir):
    project = custom_admin_running_instances['project']

    base_cmd = [
        cartridge_cmd, 'admin', 'echo_user',
    ]
    base_cmd.extend(get_admin_connection_params(connection_type, project))

    # all args
    cmd = base_cmd + [
        '--username', 'Elizabeth',
        '--age', '24',
        '--loves-cakes',
    ]
    rc, output = run_command_and_get_output(cmd, cwd=tmpdir)
    assert rc == 0

    assert get_log_lines(output) == [
        '• Hi, Elizabeth!',
        '• You are 24 years old',
        '• I know that you like cakes!',
    ]

    # age is float
    cmd = base_cmd + [
        '--username', 'Elizabeth',
        '--age', '23.5',
        '--loves-cakes',
    ]
    rc, output = run_command_and_get_output(cmd, cwd=tmpdir)
    assert rc == 0

    assert get_log_lines(output) == [
        '• Hi, Elizabeth!',
        '• You are 23.5 years old',
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


@pytest.mark.parametrize('connection_type', ['find-socket', 'connect'])
def test_func_rets_str(cartridge_cmd, custom_admin_running_instances, connection_type, tmpdir):
    project = custom_admin_running_instances['project']

    cmd = [
        cartridge_cmd, 'admin',
        'func_rets_str',
    ]
    cmd.extend(get_admin_connection_params(connection_type, project))

    rc, output = run_command_and_get_output(cmd, cwd=tmpdir)
    assert rc == 0

    assert get_log_lines(output) == [
        '• func_rets_str was called',
    ]


@pytest.mark.parametrize('connection_type', ['find-socket', 'connect'])
def test_func_rets_non_str(cartridge_cmd, custom_admin_running_instances, connection_type, tmpdir):
    project = custom_admin_running_instances['project']

    cmd = [
        cartridge_cmd, 'admin',
        'func_rets_non_str',
    ]
    cmd.extend(get_admin_connection_params(connection_type, project))

    rc, output = run_command_and_get_output(cmd, cwd=tmpdir)
    assert rc == 0

    assert get_log_lines(output) == [
        '• 666',
        '• Admin function should return string or string array value',
    ]


@pytest.mark.parametrize('connection_type', ['find-socket', 'connect'])
def test_func_rets_err(cartridge_cmd, custom_admin_running_instances, connection_type, tmpdir):
    project = custom_admin_running_instances['project']

    cmd = [
        cartridge_cmd, 'admin',
        'func_rets_err',
    ]
    cmd.extend(get_admin_connection_params(connection_type, project))

    rc, output = run_command_and_get_output(cmd, cwd=tmpdir)
    assert rc == 1

    assert get_log_lines(output) == [
        '⨯ Failed to call "func_rets_err": Some horrible error',
    ]


@pytest.mark.parametrize('connection_type', ['find-socket', 'connect'])
def test_func_raises_err(cartridge_cmd, custom_admin_running_instances, connection_type, tmpdir):
    project = custom_admin_running_instances['project']

    cmd = [
        cartridge_cmd, 'admin',
        'func_raises_err',
    ]
    cmd.extend(get_admin_connection_params(connection_type, project))

    rc, output = run_command_and_get_output(cmd, cwd=tmpdir)
    assert rc == 1

    assert '⨯ Failed to call "func_raises_err":' in output
    assert 'Some horrible error raised' in output


@pytest.mark.parametrize('connection_type', ['find-socket', 'connect'])
def test_print(cartridge_cmd, custom_admin_running_instances, connection_type, tmpdir):
    project = custom_admin_running_instances['project']

    ITERATIONS_NUM = 3

    cmd = [
        cartridge_cmd, 'admin',
        'func_print',
        '--num', str(ITERATIONS_NUM),
    ]
    cmd.extend(get_admin_connection_params(connection_type, project))

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
