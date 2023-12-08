import copy
import os

import pytest
from clusterwide_conf import (assert_conf_changed, assert_conf_not_changed,
                              write_instances_topology_conf)
from utils import (assert_for_instances_group, assert_ok_for_all_instances,
                   get_logs, run_command_and_get_output)

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
    assert_for_instances_group(
        get_logs(output), instances, lambda line: exp_error in line
    )


@pytest.mark.parametrize('conf_type', ['simple', 'not-in-leaders', 'leader-is-string', 'one-file-config'])
def test_set_leader(cartridge_cmd, conf_type, tmpdir,
                    clusterwide_conf_simple,
                    clusterwide_conf_srv_not_in_leaders,
                    clusterwide_conf_other_leader_is_string,
                    clusterwide_conf_one_file):
    data_dir = os.path.join(tmpdir, 'tmp', 'data')
    os.makedirs(data_dir)

    configs = {
        'simple': clusterwide_conf_simple,
        'not-in-leaders': clusterwide_conf_srv_not_in_leaders,
        'leader-is-string': clusterwide_conf_other_leader_is_string,
        'one-file-config': clusterwide_conf_one_file,
    }

    config = configs[conf_type]
    old_conf = copy.deepcopy(config.conf)

    # create app configs
    instances = ['instance-1', 'instance-2']
    conf_paths = write_instances_topology_conf(data_dir, APPNAME, old_conf, instances, config.one_file)

    # create other app configs
    other_instances = ['other-instance-1', 'other-instance-2']
    other_app_conf_paths = write_instances_topology_conf(
        data_dir, OTHER_APP_NAME, old_conf, other_instances, config.one_file,
    )

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
    assert logs[0] == "Set %s leader to %s" % (config.replicaset_uuid, config.instance_uuid)

    instances_logs = logs[-len(instances):]
    assert_ok_for_all_instances(instances_logs, instances)

    # check app config changes
    new_conf = copy.deepcopy(old_conf)

    # apply expected changes to topology conf
    new_topology_conf = new_conf
    if config.one_file:
        new_topology_conf = new_conf['topology']

    new_leaders = new_topology_conf['replicasets'][config.replicaset_uuid]['master']
    if isinstance(new_leaders, list):
        if config.instance_uuid in new_leaders:
            new_leaders.remove(config.instance_uuid)

        new_leaders.insert(0, config.instance_uuid)
    else:
        new_topology_conf['replicasets'][config.replicaset_uuid]['master'] = config.instance_uuid

    assert_conf_changed(conf_paths, other_app_conf_paths, old_conf, new_conf)


@pytest.mark.parametrize('conf_type', ['simple', 'not-in-leaders', 'leader-is-string'])
def test_set_leader_dry_run(cartridge_cmd, conf_type, tmpdir,
                            clusterwide_conf_simple,
                            clusterwide_conf_srv_not_in_leaders,
                            clusterwide_conf_other_leader_is_string):
    data_dir = os.path.join(tmpdir, 'tmp', 'data')
    os.makedirs(data_dir)

    configs = {
        'simple': clusterwide_conf_simple,
        'not-in-leaders': clusterwide_conf_srv_not_in_leaders,
        'leader-is-string': clusterwide_conf_other_leader_is_string,
    }

    config = configs[conf_type]
    old_conf = copy.deepcopy(config.conf)

    exp_rpl_diffs = {
        'simple': '\n'.join([
            '   %s:' % config.replicaset_uuid,
            '     alias: unnamed',
            '     master:',
            '+    - %s' % config.instance_uuid,
            '     - srv-1',
            '     - srv-2',
            '-    - %s' % config.instance_uuid,
        ]),
        'not-in-leaders': '\n'.join([
            '   %s:' % config.replicaset_uuid,
            '     alias: unnamed',
            '     master:',
            '+    - %s' % config.instance_uuid,
            '     - srv-1',
            '     - srv-2',
        ]),
        'leader-is-string': '\n'.join([
            '   %s:' % config.replicaset_uuid,
            '     alias: unnamed',
            '-    master: srv-1',
            '+    master: %s' % config.instance_uuid,
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
    assert "Set %s leader to %s" % (config.replicaset_uuid, config.instance_uuid) in output

    exp_rpl_diff = exp_rpl_diffs[conf_type]
    assert exp_rpl_diff in output

    # check config wasn't changed
    assert_conf_not_changed(conf_paths, old_conf)
