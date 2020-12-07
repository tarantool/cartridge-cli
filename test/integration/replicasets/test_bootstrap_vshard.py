from utils import run_command_and_get_output
from utils import get_log_lines
from utils import is_vshard_bootstrapped


def test_bootstrap(cartridge_cmd, project_with_vshard_replicasets):
    project = project_with_vshard_replicasets.project
    instances = project_with_vshard_replicasets.instances

    # bootstrap vshard
    cmd = [
        cartridge_cmd, 'replicasets', 'bootstrap-vshard',
    ]

    rc, output = run_command_and_get_output(cmd, cwd=project.path)
    assert rc == 0

    assert get_log_lines(output) == [
        "• Vshard is bootstrapped successfully"
    ]

    router = instances['router']
    admin_api_url = router.get_admin_api_url()
    assert is_vshard_bootstrapped(admin_api_url)

    # bootstrap again
    cmd = [
        cartridge_cmd, 'replicasets', 'bootstrap-vshard',
    ]

    rc, output = run_command_and_get_output(cmd, cwd=project.path)
    assert rc == 1

    assert "already bootstrapped" in output


def test_no_vshard_roles_avaliable(cartridge_cmd, project_with_replicaset_no_roles):
    project = project_with_replicaset_no_roles.project

    # bootstrap vshard
    cmd = [
        cartridge_cmd, 'replicasets', 'bootstrap-vshard',
    ]

    rc, output = run_command_and_get_output(cmd, cwd=project.path)
    assert rc == 1

    assert 'No remotes with role "vshard-router" available' in output
