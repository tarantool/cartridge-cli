import os
import copy
import pytest

from utils import write_instance_topology_conf
from utils import run_command_and_get_output
from utils import get_logs
from utils import assert_for_all_instances


from clusterwide_conf import assert_conf_changed
from clusterwide_conf import assert_conf_not_changed


APPNAME = 'myapp'
OTHER_APP_NAME = 'other-app'


def test_remove_uuid_does_not_exist(cartridge_cmd, clusterwide_conf_non_existent_instance, tmpdir):
    data_dir = os.path.join(tmpdir, 'tmp', 'data')
    os.makedirs(data_dir)

    clusterwide_conf = clusterwide_conf_non_existent_instance

    instances = ['instance-1', 'instance-2']
    write_instance_topology_conf(data_dir, APPNAME, clusterwide_conf.conf, instances)

    cmd = [
        cartridge_cmd, 'repair', 'remove-instance',
        '--name', APPNAME,
        '--data-dir', data_dir,
        clusterwide_conf.instance_uuid,
    ]

    rc, output = run_command_and_get_output(cmd, cwd=tmpdir)
    assert rc == 1

    assert_for_all_instances(
        get_logs(output), APPNAME, instances, lambda line:
        "Instance %s isn't found in cluster" % clusterwide_conf.instance_uuid in line
    )


@pytest.mark.parametrize('conf_type', ['simple', 'disabled', 'expelled', 'not-in-leaders', 'non-existent-rpl'])
def test_remove(cartridge_cmd, conf_type, tmpdir,
                clusterwide_conf_simple,
                clusterwide_conf_srv_disabled,
                clusterwide_conf_srv_expelled,
                clusterwide_conf_srv_not_in_leaders,
                clusterwide_conf_non_existent_rpl):
    data_dir = os.path.join(tmpdir, 'tmp', 'data')
    os.makedirs(data_dir)

    configs = {
        'simple': clusterwide_conf_simple,
        'disabled': clusterwide_conf_srv_disabled,
        'expelled': clusterwide_conf_srv_expelled,
        'not-in-leaders': clusterwide_conf_srv_not_in_leaders,
        'non-existent-rpl': clusterwide_conf_non_existent_rpl,
    }

    config = configs[conf_type]
    old_conf = copy.deepcopy(config.conf)
    instance_uuid = config.instance_uuid

    # create app working directories
    instances = ['instance-1', 'instance-2']
    conf_paths = write_instance_topology_conf(data_dir, APPNAME, old_conf, instances)

    # create other app working directories
    other_instances = ['other-instance-1', 'other-instance-2']
    other_app_conf_paths = write_instance_topology_conf(data_dir, OTHER_APP_NAME, old_conf, other_instances)

    cmd = [
        cartridge_cmd, 'repair', 'remove-instance',
        '--name', APPNAME,
        '--data-dir', data_dir,
        instance_uuid,
    ]

    rc, output = run_command_and_get_output(cmd, cwd=tmpdir)
    assert rc == 0

    # check logs
    logs = get_logs(output)
    assert len(logs) == len(instances) + 1
    assert logs[0] == "Remove instance with UUID %s" % config.instance_uuid
    assert_for_all_instances(
        logs[1:], APPNAME, instances, lambda line:
        line.strip().endswith('OK')
    )

    # check config changes
    new_conf = copy.deepcopy(old_conf)

    if conf_type != 'expelled':
        replicaset_uuid = new_conf['servers'][instance_uuid]['replicaset_uuid']
        if replicaset_uuid in new_conf['replicasets']:
            new_leaders = new_conf['replicasets'][replicaset_uuid]['master']
            if instance_uuid in new_leaders:
                new_leaders.remove(instance_uuid)

    del new_conf['servers'][instance_uuid]
    assert_conf_changed(conf_paths, other_app_conf_paths, old_conf, new_conf)


@pytest.mark.parametrize('conf_type', ['simple', 'disabled', 'expelled', 'not-in-leaders', 'non-existent-rpl'])
def test_remove_dry_run(cartridge_cmd, conf_type, tmpdir,
                        clusterwide_conf_simple,
                        clusterwide_conf_srv_disabled,
                        clusterwide_conf_srv_expelled,
                        clusterwide_conf_srv_not_in_leaders,
                        clusterwide_conf_non_existent_rpl):
    data_dir = os.path.join(tmpdir, 'tmp', 'data')
    os.makedirs(data_dir)

    configs = {
        'simple': clusterwide_conf_simple,
        'disabled': clusterwide_conf_srv_disabled,
        'expelled': clusterwide_conf_srv_expelled,
        'not-in-leaders': clusterwide_conf_srv_not_in_leaders,
        'non-existent-rpl': clusterwide_conf_non_existent_rpl,
    }

    config = configs[conf_type]
    old_conf = copy.deepcopy(config.conf)
    instance_uuid = config.instance_uuid

    instance_conf = old_conf['servers'][instance_uuid]

    instances = ['instance-1', 'instance-2']
    conf_paths = write_instance_topology_conf(data_dir, APPNAME, old_conf, instances)

    cmd = [
        cartridge_cmd, 'repair', 'remove-instance',
        '--name', APPNAME,
        '--data-dir', data_dir,
        '--dry-run',
        instance_uuid,
    ]

    rc, output = run_command_and_get_output(cmd, cwd=tmpdir)
    assert rc == 0

    # check logs
    assert "Remove instance with UUID %s" % instance_uuid in output
    assert "Data directory is set to: %s" % data_dir in output

    assert all([
        'Topology config file: %s' % conf_path in output
        for conf_path in conf_paths
    ])

    if conf_type == 'expelled':
        exp_srv_diff = '-  %s: expelled' % instance_uuid
    else:
        exp_srv_diff = '\n'.join([
            '-  %s:' % instance_uuid,
            '-    disabled: %s' % ('true' if instance_conf['disabled'] else 'false'),
            '-    replicaset_uuid: %s' % instance_conf['replicaset_uuid'],
            '-    uri: %s' % instance_conf['uri'],
        ])
    assert exp_srv_diff in output

    if conf_type == 'simple':
        replicaset_conf = old_conf['replicasets'][config.replicaset_uuid]
        exp_rpl_diff = '\n'.join([
            '   %s:' % config.replicaset_uuid,
            '     alias: %s' % replicaset_conf['alias'],
            '     master:',
            '     - srv-1',
            '     - srv-2',
            '-    - %s' % instance_uuid,
        ])
        assert exp_rpl_diff in output

    # check config wasn't changed
    assert_conf_not_changed(conf_paths, old_conf)
