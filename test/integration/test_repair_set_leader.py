import os
import copy
import pytest

from utils import run_command_and_get_output
from utils import get_logs
from utils import assert_for_all_instances

from clusterwide_conf import write_instances_topology_conf
from clusterwide_conf import assert_conf_changed
from clusterwide_conf import assert_conf_not_changed

APPNAME = 'myapp'
OTHER_APP_NAME = 'other-app'


@pytest.mark.parametrize('conf_type', [
    'non-existent-srv', 'non-existent-rpl', 'srv-disabled', 'srv-expelled', 'srv-from-other-rpl',
])
def test_bad_args(cartridge_cmd, conf_type, tmpdir,
                  clusterwide_conf_non_existent_instance,
                  clusterwide_conf_non_existent_rpl,
                  clusterwide_conf_srv_disabled,
                  clusterwide_conf_srv_expelled,
                  clusterwide_conf_srv_from_other_rpl):
    data_dir = os.path.join(tmpdir, 'tmp', 'data')
    os.makedirs(data_dir)

    configs = {
        'non-existent-srv': clusterwide_conf_non_existent_instance,
        'non-existent-rpl': clusterwide_conf_non_existent_rpl,
        'srv-disabled': clusterwide_conf_srv_disabled,
        'srv-expelled': clusterwide_conf_srv_expelled,
        'srv-from-other-rpl': clusterwide_conf_srv_from_other_rpl,
    }

    config = configs[conf_type]

    instances = ['instance-1', 'instance-2']
    write_instances_topology_conf(data_dir, APPNAME, config.conf, instances)

    cmd = [
        cartridge_cmd, 'repair', 'set-leader',
        '--name', APPNAME,
        '--data-dir', data_dir,
        config.replicaset_uuid, config.instance_uuid,
    ]

    rc, output = run_command_and_get_output(cmd, cwd=tmpdir)
    assert rc == 1

    exp_errors = {
        'non-existent-srv': "Instance %s isn't found in cluster" % config.instance_uuid,
        'non-existent-rpl': "Replicaset %s isn't found in the cluster" % config.replicaset_uuid,
        'srv-disabled': "Instance %s is disabled" % config.instance_uuid,
        'srv-expelled': "Instance %s is expelled" % config.instance_uuid,
        'srv-from-other-rpl': "Instance %s doesn't belong to replicaset %s"
        % (config.instance_uuid, config.replicaset_uuid),
    }

    exp_error = exp_errors[conf_type]
    assert_for_all_instances(
        get_logs(output), APPNAME, instances, lambda line: exp_error in line
    )


@pytest.mark.parametrize('conf_type', ['simple', 'not-in-leaders'])
def test_set_leader(cartridge_cmd, conf_type, tmpdir,
                    clusterwide_conf_simple,
                    clusterwide_conf_srv_not_in_leaders):
    data_dir = os.path.join(tmpdir, 'tmp', 'data')
    os.makedirs(data_dir)

    configs = {
        'simple': clusterwide_conf_simple,
        'not-in-leaders': clusterwide_conf_srv_not_in_leaders,
    }

    config = configs[conf_type]
    old_conf = copy.deepcopy(config.conf)

    # create app working directories
    instances = ['instance-1', 'instance-2']
    conf_paths = write_instances_topology_conf(data_dir, APPNAME, old_conf, instances)

    # create other app working directories
    other_instances = ['other-instance-1', 'other-instance-2']
    other_app_conf_paths = write_instances_topology_conf(data_dir, OTHER_APP_NAME, old_conf, other_instances)

    cmd = [
        cartridge_cmd, 'repair', 'set-leader',
        '--name', APPNAME,
        '--data-dir', data_dir,
        config.replicaset_uuid, config.instance_uuid,
    ]

    rc, output = run_command_and_get_output(cmd, cwd=tmpdir)
    assert rc == 0

    # check logs
    logs = get_logs(output)
    assert len(logs) == len(instances) + 1
    assert logs[0] == "Set %s master to %s" % (config.replicaset_uuid, config.instance_uuid)
    assert all([line.strip().endswith('OK') for line in logs[1:]])
    assert_for_all_instances(
        logs[1:], APPNAME, instances, lambda line: line.strip().endswith('OK'),
    )

    # check app config changes
    new_conf = copy.deepcopy(old_conf)
    new_leaders = new_conf['replicasets'][config.replicaset_uuid]['master']
    if config.instance_uuid in new_leaders:
        new_leaders.remove(config.instance_uuid)

    new_leaders.insert(0, config.instance_uuid)
    assert_conf_changed(conf_paths, other_app_conf_paths, old_conf, new_conf)


@pytest.mark.parametrize('conf_type', ['simple', 'not-in-leaders'])
def test_set_leader_dry_run(cartridge_cmd, conf_type, tmpdir,
                            clusterwide_conf_simple,
                            clusterwide_conf_srv_not_in_leaders):

    data_dir = os.path.join(tmpdir, 'tmp', 'data')
    os.makedirs(data_dir)

    configs = {
        'simple': clusterwide_conf_simple,
        'not-in-leaders': clusterwide_conf_srv_not_in_leaders,
    }

    config = configs[conf_type]
    old_conf = copy.deepcopy(config.conf)

    exp_rpl_diffs = {
        'simple':  '\n'.join([
            '   %s:' % config.replicaset_uuid,
            '     alias: unnamed',
            '     master:',
            '+    - %s' % config.instance_uuid,
            '     - srv-1',
            '     - srv-2',
            '-    - %s' % config.instance_uuid,
        ]),
        'not-in-leaders':  '\n'.join([
            '   %s:' % config.replicaset_uuid,
            '     alias: unnamed',
            '     master:',
            '+    - %s' % config.instance_uuid,
            '     - srv-1',
            '     - srv-2',
        ]),
    }

    # create app working directories
    instances = ['instance-1', 'instance-2']
    conf_paths = write_instances_topology_conf(data_dir, APPNAME, old_conf, instances)

    cmd = [
        cartridge_cmd, 'repair', 'set-leader',
        '--name', APPNAME,
        '--data-dir', data_dir,
        '--dry-run',
        config.replicaset_uuid, config.instance_uuid,
    ]

    rc, output = run_command_and_get_output(cmd, cwd=tmpdir)
    assert rc == 0

    # check logs
    assert "Set %s master to %s" % (config.replicaset_uuid, config.instance_uuid) in output
    assert "Data directory is set to: %s" % data_dir in output

    assert all([
        'Topology config file: %s' % conf_path in output
        for conf_path in conf_paths
    ])

    exp_rpl_diff = exp_rpl_diffs[conf_type]
    assert exp_rpl_diff in output

    # check config wasn't changed
    assert_conf_not_changed(conf_paths, old_conf)
