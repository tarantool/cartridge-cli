from utils import run_command_and_get_output
from utils import get_log_lines
from utils import get_vshard_group_names

from integration.replicasets.utils import get_list_from_log_lines


def test_list_groups(cartridge_cmd, project_with_instances):
    project = project_with_instances.project
    instances = project_with_instances.instances

    router = instances['router']
    admin_api_url = router.get_admin_api_url()
    vshard_group_names = get_vshard_group_names(admin_api_url)

    # bootstrap vshard
    cmd = [
        cartridge_cmd, 'replicasets', 'list-vshard-groups',
    ]

    rc, output = run_command_and_get_output(cmd, cwd=project.path)
    assert rc == 0

    log_lines = get_log_lines(output)

    assert log_lines[:1] == [
        '• Available vshard groups:',
    ]

    groups_list = get_list_from_log_lines(log_lines[1:])
    assert set(groups_list) == set(vshard_group_names)
