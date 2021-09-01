from utils import (get_log_lines, is_instance_expelled,
                   run_command_and_get_output)


def test_bad_instance_name(cartridge_cmd, project_with_vshard_replicasets):
    project = project_with_vshard_replicasets.project

    cmd = [
        cartridge_cmd, 'replicasets', 'expel',
        'unknown-instance',
    ]

    rc, output = run_command_and_get_output(cmd, cwd=project.path)
    assert rc == 1
    assert "Instance unknown-instance isn't found in cluster" in output


def test_expel_last_instance(cartridge_cmd, project_with_one_joined_instance):
    project = project_with_one_joined_instance.project
    instances = project_with_one_joined_instance.instances

    instance = instances['some-instance']

    cmd = [
        cartridge_cmd, 'replicasets', 'expel',
        instance.name,
    ]

    rc, output = run_command_and_get_output(cmd, cwd=project.path)
    assert rc == 1
    assert "Not found any other non-expelled instance joined to cluster" in output


def test_expel(cartridge_cmd, project_with_vshard_replicasets):
    project = project_with_vshard_replicasets.project
    instances = project_with_vshard_replicasets.instances

    router = instances['router']
    hot_replica = instances['hot-replica']
    admin_api_url = router.get_admin_api_url()

    # expel hot sotrage replica
    cmd = [
        cartridge_cmd, 'replicasets', 'expel', hot_replica.name,
    ]

    rc, output = run_command_and_get_output(cmd, cwd=project.path)
    assert rc == 0
    assert get_log_lines(output) == [
        '• Instance(s) %s have been successfully expelled' % hot_replica.name,
    ]

    assert is_instance_expelled(admin_api_url, hot_replica.name)


def test_expel_fails(cartridge_cmd, project_with_vshard_replicasets):
    project = project_with_vshard_replicasets.project
    instances = project_with_vshard_replicasets.instances

    # the replicaset leader can't be expelled if vshard is bootstrapped

    # bootstrap vshard
    cmd = [
        cartridge_cmd, 'replicasets', 'bootstrap-vshard',
    ]

    rc, _ = run_command_and_get_output(cmd, cwd=project.path)
    assert rc == 0

    hot_master = instances['hot-master']

    # expel cold sotrage master
    cmd = [
        cartridge_cmd, 'replicasets', 'expel', hot_master.name,
    ]

    rc, output = run_command_and_get_output(cmd, cwd=project.path)
    assert rc == 1
    assert "is the leader and can't be expelled" in output
