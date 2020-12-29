from utils import run_command_and_get_output
from utils import get_replicasets, get_known_roles
from utils import get_log_lines

from integration.replicasets.utils import get_list_from_log_lines
from integration.replicasets.utils import get_replicaset_by_alias


def assert_add_roles_log(output, replicaset_alias, roles_to_add, res_roles_list):
    roles_to_add_str = ', '.join(roles_to_add)

    log_lines = get_log_lines(output)

    assert log_lines[:2] == [
        '• Add role(s) %s to replica set %s' % (roles_to_add_str, replicaset_alias),
        '• Replica set %s now has these roles enabled:' % replicaset_alias,
    ]

    roles_list = get_list_from_log_lines(log_lines[2:])
    assert set(roles_list) == set(res_roles_list)


def assert_remove_roles_log(output, replicaset_alias, roles_to_remove, res_roles_list):
    roles_to_remove_str = ', '.join(roles_to_remove)

    log_lines = get_log_lines(output)

    assert log_lines[:2] == [
        '• Remove role(s) %s from replica set %s' % (roles_to_remove_str, replicaset_alias),
        '• Replica set %s now has these roles enabled:' % replicaset_alias,
    ]

    roles_list = get_list_from_log_lines(log_lines[2:])
    assert set(roles_list) == set(res_roles_list)


def test_bad_replicaset_name(cartridge_cmd, project_with_replicaset_no_roles):
    project = project_with_replicaset_no_roles.project

    # add-roles
    cmd = [
        cartridge_cmd, 'replicasets', 'add-roles',
        '--replicaset', 'unknown-replicaset',
        'vshard-router',
    ]

    rc, output = run_command_and_get_output(cmd, cwd=project.path)
    assert rc == 1
    assert "Replica set unknown-replicaset isn't found in current topology" in output

    # remove-roles
    cmd = [
        cartridge_cmd, 'replicasets', 'remove-roles',
        '--replicaset', 'unknown-replicaset',
        'vshard-router',
    ]

    rc, output = run_command_and_get_output(cmd, cwd=project.path)
    assert rc == 1
    assert "Replica set unknown-replicaset isn't found in current topology" in output


def test_list_roles(cartridge_cmd, project_with_instances):
    project = project_with_instances.project
    instances = project_with_instances.instances

    router = instances['router']
    admin_api_url = router.get_admin_api_url()

    # get list of roles
    cmd = [
        cartridge_cmd, 'replicasets', 'list-roles',
    ]

    rc, output = run_command_and_get_output(cmd, cwd=project.path)
    assert rc == 0

    log_lines = get_log_lines(output)
    assert log_lines[:1] == [
        "• Available roles:",
    ]

    known_roles = get_known_roles(admin_api_url)
    exp_roles_list = []
    for known_role in known_roles:
        if not known_role['dependencies']:
            exp_roles_list.append(known_role['name'])
        else:
            exp_roles_list.append("%s (+ %s)" % (
                known_role['name'],
                ', '.join(known_role['dependencies'])
            ))

    roles_list = get_list_from_log_lines(log_lines[1:])
    assert set(roles_list) == set(exp_roles_list)


def test_add_remove_roles(cartridge_cmd, project_with_replicaset_no_roles):
    project = project_with_replicaset_no_roles.project
    instances = project_with_replicaset_no_roles.instances
    replicasets = project_with_replicaset_no_roles.replicasets

    VSHARD_ROUTER_ROLE = 'vshard-router'
    APP_CUSTOM_ROLE = 'app.roles.custom'
    FAILOVER_COORDINATOR_ROLE = 'failover-coordinator'

    rpl = replicasets['some-rpl']
    instance = instances['some-instance']
    admin_api_url = instance.get_admin_api_url()

    # add vshard-router and app.roles.custom roles to replicaset
    cmd = [
        cartridge_cmd, 'replicasets', 'add-roles',
        '--replicaset', rpl.name,
        VSHARD_ROUTER_ROLE, APP_CUSTOM_ROLE,
    ]

    rc, output = run_command_and_get_output(cmd, cwd=project.path)
    assert rc == 0

    exp_res_roles_list = [VSHARD_ROUTER_ROLE, APP_CUSTOM_ROLE]
    assert_add_roles_log(
        output, rpl.name,
        roles_to_add=[VSHARD_ROUTER_ROLE, APP_CUSTOM_ROLE],
        res_roles_list=exp_res_roles_list,
    )

    replicasets = get_replicasets(admin_api_url)
    router_replicaset = get_replicaset_by_alias(replicasets, rpl.name)
    assert router_replicaset is not None
    assert set(router_replicaset['roles']) == set(exp_res_roles_list)

    # add failover-coordinator and app.roles.custom (again) roles to replicaset
    cmd = [
        cartridge_cmd, 'replicasets', 'add-roles',
        '--replicaset', rpl.name,
        FAILOVER_COORDINATOR_ROLE, APP_CUSTOM_ROLE,
    ]

    rc, output = run_command_and_get_output(cmd, cwd=project.path)
    assert rc == 0

    exp_res_roles_list = [
        VSHARD_ROUTER_ROLE, FAILOVER_COORDINATOR_ROLE, APP_CUSTOM_ROLE,
    ]
    assert_add_roles_log(
        output, rpl.name,
        roles_to_add=[FAILOVER_COORDINATOR_ROLE, APP_CUSTOM_ROLE],
        res_roles_list=exp_res_roles_list,
    )

    replicasets = get_replicasets(admin_api_url)
    router_replicaset = get_replicaset_by_alias(replicasets, rpl.name)
    assert router_replicaset is not None
    assert set(router_replicaset['roles']) == set(exp_res_roles_list)

    # remove metrics and failover-coordinator roles from replicaset
    cmd = [
        cartridge_cmd, 'replicasets', 'remove-roles',
        '--replicaset', rpl.name,
        FAILOVER_COORDINATOR_ROLE,
    ]

    rc, output = run_command_and_get_output(cmd, cwd=project.path)
    assert rc == 0

    exp_res_roles_list = [
        VSHARD_ROUTER_ROLE, APP_CUSTOM_ROLE,
    ]

    assert_remove_roles_log(
        output, rpl.name,
        roles_to_remove=[FAILOVER_COORDINATOR_ROLE],
        res_roles_list=exp_res_roles_list,
    )

    replicasets = get_replicasets(admin_api_url)
    router_replicaset = get_replicaset_by_alias(replicasets, rpl.name)
    assert router_replicaset is not None
    assert set(router_replicaset['roles']) == set(exp_res_roles_list)


def test_add_roles_vshard_group(cartridge_cmd, project_with_replicaset_no_roles):
    project = project_with_replicaset_no_roles.project
    instances = project_with_replicaset_no_roles.instances
    replicasets = project_with_replicaset_no_roles.replicasets

    VSHARD_STORAGE_ROLE = 'vshard-storage'
    HOT_GROUP_NAME = 'hot'

    rpl = replicasets['some-rpl']
    instance = instances['some-instance']
    admin_api_url = instance.get_admin_api_url()

    # add vshard-storage role
    cmd = [
        cartridge_cmd, 'replicasets', 'add-roles',
        '--replicaset', rpl.name,
        '--vshard-group', HOT_GROUP_NAME,
        VSHARD_STORAGE_ROLE,
    ]

    rc, output = run_command_and_get_output(cmd, cwd=project.path)
    assert rc == 0

    exp_res_roles = [VSHARD_STORAGE_ROLE]
    exp_res_roles_list = ['%s (%s)' % (VSHARD_STORAGE_ROLE, HOT_GROUP_NAME)]

    assert_add_roles_log(
        output, rpl.name,
        roles_to_add=[VSHARD_STORAGE_ROLE],
        res_roles_list=exp_res_roles_list,
    )

    replicasets = get_replicasets(admin_api_url)
    router_replicaset = get_replicaset_by_alias(replicasets, rpl.name)
    assert router_replicaset is not None
    assert router_replicaset['vshard_group'] == HOT_GROUP_NAME
    assert set(router_replicaset['roles']) == set(exp_res_roles)
