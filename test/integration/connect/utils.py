import subprocess

from utils import tarantool_short_version


class Command:
    def __init__(self, command, yaml_output=None, lua_output=None, exp_output=None):
        self.command = command
        if yaml_output is not None:
            self.exp_output = '---\n- %s\n...\n' % yaml_output
        elif lua_output is not None:
            self.exp_output = '%s;' % lua_output
        elif exp_output is not None:
            self.exp_output = exp_output


def get_successful_commands():
    common_commands = [
        # YAML output

        # valid commands
        Command('return 666', yaml_output='666'),
        Command('777', yaml_output='777'),
        # multiline statement
        Command('if\ntrue\nthen\nreturn 999\nend', yaml_output='999'),
        # undefined variable
        Command('kek', yaml_output='error: \'[string "return kek "]:1: variable \'\'kek\'\' is not declared\''),
        # syntax error
        Command('if+1', yaml_output='error: \'[string "if+1 "]:1: unexpected symbol near \'\'+\'\'\''),
    ]

    set_output_error_commands = [
        Command(
            '\\set output lua',
            yaml_output='error: \'[string "\\set output lua "]:1: unexpected symbol near \'\'\\\'\'\''
        ),
        Command('return 666', yaml_output='666'),

        Command(
            '\\set output yaml',
            yaml_output='error: \'[string "\\set output yaml "]:1: unexpected symbol near \'\'\\\'\'\''
        ),
        Command('return 666', yaml_output='666'),
    ]

    set_output_commands = [
        # Lua output
        Command('\\set output lua', lua_output='true'),

        # valid commands
        Command('return 666', lua_output='666'),
        Command('777', lua_output='777'),
        # multiline statement
        Command('if\ntrue\nthen\nreturn 999\nend', lua_output='999'),

        # YAML output again
        Command('\\set output yaml', yaml_output='true'),
        Command('return 666', yaml_output='666'),
    ]

    commands = common_commands

    if tarantool_short_version().startswith('1.10'):
        commands.extend(set_output_error_commands)
    else:
        commands.extend(set_output_commands)

    return commands


def run_commands_in_pipe(project, cmd, commands):
    process = subprocess.Popen(
        cmd,
        cwd=project.path,
        stderr=subprocess.STDOUT,
        stdout=subprocess.PIPE,
        stdin=subprocess.PIPE,
    )

    user_input = '\n'.join([c.command for c in commands]) + '\n'

    output, _ = process.communicate(user_input.encode('utf-8'))
    output = output.decode('utf-8')

    print(output)

    return process.returncode, output


def assert_successful_piped_commands(project, cmd, exp_connect):
    commands = get_successful_commands()

    rc, output = run_commands_in_pipe(project, cmd, commands)
    assert rc == 0

    out_lines = output.split('\n', maxsplit=1)
    connected_line, commands_output = out_lines

    assert connected_line == 'connected to %s' % exp_connect

    # now I have no idea how to read commands prompt using subprocess
    exp_output = '\n'.join(c.exp_output for c in commands)+'\n'
    assert commands_output == exp_output


def assert_exited_piped_commands(project, cmd, exp_connect):
    commands = [
        Command("os.exit(0)"),
    ]

    rc, output = run_commands_in_pipe(project, cmd, commands)
    assert rc == 1

    out_lines = output.split('\n', maxsplit=1)
    connected_line, errmsg = out_lines

    assert connected_line == 'connected to %s' % exp_connect

    errmsg = errmsg.strip()
    assert errmsg == "⨯ Connection was closed. Probably instance process isn't running anymore"


def get_push_tag_yaml_output(message):
    fmt = '''"%s"

---
- true
...
'''

    return fmt % message


def get_push_tag_lua_output(message):
    fmt = ''''"%s"'

true;'''

    return fmt % message


def assert_session_push_commands(project, cmd, exp_connect):
    commands = [
        Command("box.session.push('666')", exp_output=get_push_tag_yaml_output('666')),
    ]

    if tarantool_short_version().startswith('2'):
        commands.extend([
            Command("\\set output lua", lua_output='true'),
            Command("box.session.push('777')", exp_output=get_push_tag_lua_output('777')),
            Command("\\set output yaml", yaml_output='true'),
        ])

    rc, output = run_commands_in_pipe(project, cmd, commands)
    assert rc == 0

    out_lines = output.split('\n', maxsplit=1)
    connected_line, commands_output = out_lines

    assert connected_line == 'connected to %s' % exp_connect

    # now I have no idea how to read commands prompt using subprocess
    exp_output = '\n'.join(c.exp_output for c in commands)+'\n'
    assert commands_output == exp_output


def assert_error(project, cmd, errmsg):
    process = subprocess.Popen(
        cmd,
        cwd=project.path,
        stderr=subprocess.STDOUT,
        stdout=subprocess.PIPE,
        stdin=subprocess.PIPE,
    )

    output, _ = process.communicate()
    output = output.decode('utf-8')
    print(output)

    assert process.returncode == 1
    assert errmsg in output
