from integration.connect.utils import (assert_error,
                                       assert_exited_piped_commands,
                                       assert_session_push_commands,
                                       assert_successful_piped_commands)
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

    assert_successful_piped_commands(project, cmd, exp_connect='%s.%s' % (project.name, router.name),
                                     remote_control=True)


def test_socket_piped(cartridge_cmd, project_with_instances):
    project = project_with_instances.project
    instances = project_with_instances.instances

    router = instances['router']
    console_sock_path = project.get_console_sock(router.name)

    cmd = [
        cartridge_cmd, 'connect', console_sock_path,
    ]

    assert_successful_piped_commands(project, cmd, exp_connect='%s.%s' % (project.name, router.name),
                                     remote_control=False)


def test_socket_no_title(cartridge_cmd, project_with_instances_no_cartridge):
    project = project_with_instances_no_cartridge.project
    instances = project_with_instances_no_cartridge.instances

    router = instances['router']
    console_sock_path = project.get_console_sock(router.name)

    cmd = [
        cartridge_cmd, 'connect', console_sock_path,
    ]

    assert_successful_piped_commands(project, cmd, exp_connect=console_sock_path, remote_control=False)


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
