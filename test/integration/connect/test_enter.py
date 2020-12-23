from integration.connect.utils import assert_successful_piped_commands
from integration.connect.utils import assert_exited_piped_commands
from integration.connect.utils import assert_session_push_commands
from integration.connect.utils import assert_error


def test_bad_instance_name(cartridge_cmd, project_with_instances):
    project = project_with_instances.project

    cmd = [
        cartridge_cmd, 'enter', 'unknown-instance',
    ]

    assert_error(project, cmd, "Instance unknown-instance is not running")


def test_enter_piped(cartridge_cmd, project_with_instances):
    project = project_with_instances.project
    instances = project_with_instances.instances

    router = instances['router']

    cmd = [
        cartridge_cmd, 'enter', router.name,
    ]

    assert_successful_piped_commands(project, cmd, exp_connect='%s.%s' % (project.name, router.name))


def test_instance_exited(cartridge_cmd, project_with_instances):
    project = project_with_instances.project
    instances = project_with_instances.instances

    router = instances['router']

    cmd = [
        cartridge_cmd, 'enter', router.name,
    ]

    assert_exited_piped_commands(project, cmd, exp_connect='%s.%s' % (project.name, router.name))


def test_session_push(cartridge_cmd, project_with_instances):
    project = project_with_instances.project
    instances = project_with_instances.instances

    router = instances['router']

    cmd = [
        cartridge_cmd, 'enter', router.name,
    ]

    assert_session_push_commands(project, cmd, exp_connect='%s.%s' % (project.name, router.name))
