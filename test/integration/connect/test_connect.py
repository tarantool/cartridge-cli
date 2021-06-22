import os

from integration.connect.utils import assert_successful_piped_commands
from integration.connect.utils import assert_error
from integration.connect.utils import assert_exited_piped_commands
from integration.connect.utils import assert_session_push_commands
from integration.connect.utils import run_commands_in_pipe
from integration.connect.utils import Command

from utils import DEFAULT_CLUSTER_COOKIE


def test_bad_uri(cartridge_cmd, project_with_instances):
    project = project_with_instances.project

    cmd = [
        cartridge_cmd, 'connect', 'bad-host:3301'
    ]

    errmsg = "Failed to dial: dial tcp: lookup bad-host"
    assert_error(project, cmd, errmsg)


def test_bad_socket(cartridge_cmd, project_with_instances):
    project = project_with_instances.project

    cmd = [
        cartridge_cmd, 'connect', '/bad-sock-path'
    ]

    errmsg = "Failed to dial: dial unix /bad-sock-path: connect: no such file or directory"
    assert_error(project, cmd, errmsg)


def test_uri_piped(cartridge_cmd, project_with_instances):
    project = project_with_instances.project
    instances = project_with_instances.instances

    router = instances['router']

    cmd = [
        cartridge_cmd, 'connect', router.advertise_uri,
        '--username', 'admin',
        '--password', DEFAULT_CLUSTER_COOKIE,
    ]

    assert_successful_piped_commands(project, cmd, exp_connect='%s.%s' % (project.name, router.name))


def test_socket_piped(cartridge_cmd, project_with_instances):
    project = project_with_instances.project
    instances = project_with_instances.instances

    router = instances['router']
    console_sock_path = project.get_console_sock(router.name)

    cmd = [
        cartridge_cmd, 'connect', console_sock_path,
    ]

    assert_successful_piped_commands(project, cmd, exp_connect='%s.%s' % (project.name, router.name))


def test_socket_no_title(cartridge_cmd, project_with_instances_no_cartridge):
    project = project_with_instances_no_cartridge.project
    instances = project_with_instances_no_cartridge.instances

    router = instances['router']
    console_sock_path = project.get_console_sock(router.name)

    cmd = [
        cartridge_cmd, 'connect', console_sock_path,
    ]

    assert_successful_piped_commands(project, cmd, exp_connect=console_sock_path)


def test_uri_instance_exited(cartridge_cmd, project_with_instances):
    project = project_with_instances.project
    instances = project_with_instances.instances

    router = instances['router']

    cmd = [
        cartridge_cmd, 'connect', router.advertise_uri,
        '--username', 'admin',
        '--password', DEFAULT_CLUSTER_COOKIE,
    ]

    assert_exited_piped_commands(project, cmd, exp_connect='%s.%s' % (project.name, router.name))


def test_socket_instance_exited(cartridge_cmd, project_with_instances):
    project = project_with_instances.project
    instances = project_with_instances.instances

    router = instances['router']
    console_sock_path = project.get_console_sock(router.name)

    cmd = [
        cartridge_cmd, 'connect', console_sock_path,
    ]

    assert_exited_piped_commands(project, cmd, exp_connect='%s.%s' % (project.name, router.name))


def test_socket_session_push(cartridge_cmd, project_with_instances):
    project = project_with_instances.project
    instances = project_with_instances.instances

    router = instances['router']
    console_sock_path = project.get_console_sock(router.name)

    cmd = [
        cartridge_cmd, 'connect', console_sock_path,
    ]

    assert_session_push_commands(project, cmd, exp_connect='%s.%s' % (project.name, router.name))


def test_version_update_at_roles_reload(cartridge_cmd, project_setuped_with_instances):
    project = project_setuped_with_instances.project
    instances = project_setuped_with_instances.instances

    router = instances['router']

    cmd = [
        cartridge_cmd, 'connect', router.advertise_uri,
        '--username', 'admin',
        '--password', DEFAULT_CLUSTER_COOKIE,
    ]

    command = Command("require('VERSION')", yaml_output=project.version)
    rc, output = run_commands_in_pipe(project, cmd, [command])
    assert rc == 0

    _, command_output = output.split('\n', maxsplit=1)
    assert command.exp_output in command_output

    # Update VERSION.lua file
    new_project_version = "100500"
    with open(os.path.join(project.path, 'VERSION.lua'), 'w') as f:
        f.write(f"return {new_project_version}")

    # Reload cartridge
    command = Command("require('cartridge').reload_roles()", yaml_output="true")
    rc, output = run_commands_in_pipe(project, cmd, [command])
    assert rc == 0

    _, command_output = output.split('\n', maxsplit=1)
    assert command.exp_output in command_output

    # Try with new VERSION.lua file
    command = Command("require('VERSION')", yaml_output=new_project_version)
    rc, output = run_commands_in_pipe(project, cmd, [command])
    assert rc == 0

    _, command_output = output.split('\n', maxsplit=1)
    assert command.exp_output in command_output
