from utils import get_replicasets
from utils import run_command_and_get_output
from utils import get_log_lines

from integration.replicasets.utils import get_replicaset_by_alias


def assert_join_instances_logs(output, replicaset_alias, instances):
    format_strings = (
        ', '.join([i.name for i in instances]),
        replicaset_alias,
    )
    assert get_log_lines(output) == [
        '• Join instance(s) %s to replica set %s' % format_strings,
        '• Instance(s) %s have been successfully joined to replica set %s' % format_strings,
    ]


def assert_joined_replicaset(admin_api_url, replicaset_alias, instances, exp_replicasets_num=1):
    replicasets = get_replicasets(admin_api_url)
    assert len(replicasets) == exp_replicasets_num

    replicaset = get_replicaset_by_alias(replicasets, replicaset_alias)
    assert replicaset is not None

    assert replicaset == {
        'alias': replicaset_alias,
        'roles': [],
        'vshard_group': None,
        'all_rw': False,
        'weight': None,
        'servers': [
            {'uri': i.advertise_uri, 'alias': i.name} for i in instances
        ]
    }


def test_join_first_not_started_instance(cartridge_cmd, project_with_replicaset_no_roles):
    project = project_with_replicaset_no_roles.project
    replicasets = project_with_replicaset_no_roles.replicasets

    rpl = replicasets['some-rpl']

    # join instance w/ bad name
    cmd = [
        cartridge_cmd, 'replicasets', 'join',
        '--replicaset', rpl.name,
        'unknown-instance',
    ]

    # since this instance is first joined, CLI ties to connect to it
    rc, output = run_command_and_get_output(cmd, cwd=project.path)
    assert rc == 1
    assert "Configuration for instance unknown-instance hasn't found in instances.yml" in output


def test_join_unknown_instance(cartridge_cmd, project_with_instances):
    project = project_with_instances.project

    # join instance w/ bad name
    cmd = [
        cartridge_cmd, 'replicasets', 'join',
        '--replicaset', 'some-replicaset',
        'unknown-instance',
    ]

    # since this instance is first joined, CLI ties to connect to it
    rc, output = run_command_and_get_output(cmd, cwd=project.path)
    assert rc == 1
    assert "Failed to connect to Tarantool instance:" in output


def test_join(cartridge_cmd, project_with_instances):
    project = project_with_instances.project
    instances = project_with_instances.instances

    # create router replicaset
    ROUTER_RPL_ALIAS = 'router'
    router = instances['router']

    admin_api_url = router.get_admin_api_url()

    cmd = [
        cartridge_cmd, 'replicasets', 'join',
        '--replicaset', ROUTER_RPL_ALIAS,
        router.name,
    ]

    rc, output = run_command_and_get_output(cmd, cwd=project.path)
    assert rc == 0
    assert_join_instances_logs(output, ROUTER_RPL_ALIAS, [router])

    assert_joined_replicaset(admin_api_url, ROUTER_RPL_ALIAS, [router])

    # create s-1 replicaset
    S1_RPL_ALIAS = 's-1'
    s1_master = instances['s1-master']
    s1_replica = instances['s1-replica']
    s1_replica2 = instances['s1-replica-2']

    # join s1-master, s1-replica
    cmd = [
        cartridge_cmd, 'replicasets', 'join',
        '--replicaset', S1_RPL_ALIAS,
        s1_master.name, s1_replica.name,
    ]

    rc, output = run_command_and_get_output(cmd, cwd=project.path)
    assert rc == 0
    assert_join_instances_logs(output, S1_RPL_ALIAS, [s1_master, s1_replica])

    assert_joined_replicaset(admin_api_url, S1_RPL_ALIAS, [s1_master, s1_replica], exp_replicasets_num=2)

    # join s1-replica-2
    cmd = [
        cartridge_cmd, 'replicasets', 'join',
        '--replicaset', S1_RPL_ALIAS,
        s1_replica2.name,
    ]

    rc, output = run_command_and_get_output(cmd, cwd=project.path)
    assert rc == 0
    assert_join_instances_logs(output, S1_RPL_ALIAS, [s1_replica2])

    assert_joined_replicaset(admin_api_url, S1_RPL_ALIAS, [s1_master, s1_replica, s1_replica2], exp_replicasets_num=2)
