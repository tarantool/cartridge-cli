from utils import run_command_and_get_output
from utils import write_conf


def test_default_application(cartridge_cmd, default_project_with_instances):
    project = default_project_with_instances.project

    # setup replicasets
    cmd = [
        cartridge_cmd, 'replicasets', 'setup',
        '--bootstrap-vshard',
    ]

    rc, output = run_command_and_get_output(cmd, cwd=project.path)
    assert rc == 0

    # list replicasets
    cmd = [
        cartridge_cmd, 'replicasets', 'list',
    ]

    rc, output = run_command_and_get_output(cmd, cwd=project.path)
    assert rc == 0

    assert output.strip() == """• Current replica sets:
• router
  Role: failover-coordinator | vshard-router | metrics | app.roles.custom
    ★ router localhost:3301
• s-1                             default | 1
  Role: vshard-storage | metrics
    ★ s1-master localhost:3302
    • s1-replica localhost:3303
• s-2                             default | 1
  Role: vshard-storage | metrics
    ★ s2-master localhost:3304
    • s2-replica localhost:3305"""


def test_no_joined_instances(cartridge_cmd, project_with_instances):
    project = project_with_instances.project

    # list replicasets
    cmd = [
        cartridge_cmd, 'replicasets', 'list',
    ]

    rc, output = run_command_and_get_output(cmd, cwd=project.path)
    assert rc == 1
    assert "Failed to find some instance joined to cluster" in output


def test_list(cartridge_cmd, project_with_instances):
    project = project_with_instances.project
    instances = project_with_instances.instances

    router = instances['router']
    s1_master = instances['s1-master']
    s1_replica = instances['s1-replica']
    s1_replica2 = instances['s1-replica-2']

    # setup replicasets
    rpl_cfg_path = project.get_replicasets_cfg_path()
    rpl_cfg = {
        'router': {
            'roles': ['vshard-router', 'app.roles.custom', 'failover-coordinator'],
            'instances': [router.name],
        },
        's-1': {
            'roles': ['vshard-storage'],
            'instances': [s1_master.name, s1_replica.name, s1_replica2.name],
            'weight': 1.234,
            'vshard_group': 'hot',
            'all_rw': True,
        },
    }

    write_conf(rpl_cfg_path, rpl_cfg)

    cmd = [
        cartridge_cmd, 'replicasets', 'setup',
    ]

    rc, output = run_command_and_get_output(cmd, cwd=project.path)
    assert rc == 0

    # get current topology
    cmd = [
        cartridge_cmd, 'replicasets', 'list',
    ]

    rc, output = run_command_and_get_output(cmd, cwd=project.path)
    assert rc == 0

    assert output.strip() == """• Current replica sets:
• router
  Role: failover-coordinator | vshard-router | app.roles.custom
    ★ router localhost:3301
• s-1                             hot | 1.234 | ALL RW
  Role: vshard-storage
    ★ s1-master localhost:3302
    • s1-replica localhost:3303
    • s1-replica-2 localhost:3304"""
