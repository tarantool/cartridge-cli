import os
import pytest

from utils import run_command_and_get_output
from utils import get_logs
from utils import assert_ok_for_all_instances
from utils import assert_for_instances_group

from clusterwide_conf import write_instances_topology_conf
from clusterwide_conf import assert_conf_changed
from clusterwide_conf import assert_conf_not_changed
from clusterwide_conf import get_conf_with_new_uri


APPNAME = 'myapp'
OTHER_APP_NAME = 'other-app'


@pytest.mark.parametrize('conf_type', ['non-existent-srv', 'srv-expelled'])
def test_bad_args(cartridge_cmd, conf_type, tmpdir,
                  clusterwide_conf_non_existent_instance,
                  clusterwide_conf_srv_expelled):
    data_dir = os.path.join(tmpdir, 'tmp', 'data')
    os.makedirs(data_dir)

    configs = {
        'non-existent-srv': clusterwide_conf_non_existent_instance,
        'srv-expelled': clusterwide_conf_srv_expelled,
    }

    config = configs[conf_type]

    instances = ['instance-1', 'instance-2']
    write_instances_topology_conf(data_dir, APPNAME, config.conf, instances)

    cmd = [
        cartridge_cmd, 'repair', 'set-advertise-uri',
        '--name', APPNAME,
        '--data-dir', data_dir,
        config.instance_uuid, 'new-uri:666'
    ]

    rc, output = run_command_and_get_output(cmd, cwd=tmpdir)
    assert rc == 1

    exp_errors = {
        'non-existent-srv': "Instance %s isn't found in cluster" % config.instance_uuid,
        'srv-expelled': "Instance %s is expelled" % config.instance_uuid,
    }

    exp_error = exp_errors[conf_type]
    assert_for_instances_group(get_logs(output), instances, lambda line: exp_error in line)


@pytest.mark.parametrize('conf_type', ['simple', 'srv-disabled', 'one-file-config'])
def test_set_uri(cartridge_cmd, conf_type, tmpdir,
                 clusterwide_conf_simple,
                 clusterwide_conf_srv_disabled,
                 clusterwide_conf_one_file):
    data_dir = os.path.join(tmpdir, 'tmp', 'data')
    os.makedirs(data_dir)

    NEW_URI = 'new-uri:666'

    configs = {
        'simple': clusterwide_conf_simple,
        'srv-disabled': clusterwide_conf_srv_disabled,
        'one-file-config': clusterwide_conf_one_file,
    }

    config = configs[conf_type]
    old_conf = config.conf

    # create app configs
    instances = ['instance-1', 'instance-2']
    conf_paths = write_instances_topology_conf(data_dir, APPNAME, old_conf, instances, config.one_file)

    # create other app configs
    other_instances = ['other-instance-1', 'other-instance-2']
    other_app_conf_paths = write_instances_topology_conf(
        data_dir, OTHER_APP_NAME, old_conf, other_instances, config.one_file,
    )

    cmd = [
        cartridge_cmd, 'repair', 'set-advertise-uri',
        '--name', APPNAME,
        '--data-dir', data_dir,
        config.instance_uuid, NEW_URI,
    ]

    rc, output = run_command_and_get_output(cmd, cwd=tmpdir)
    assert rc == 0

    # check logs
    logs = get_logs(output)
    assert logs[0] == "Set %s advertise URI to %s" % (config.instance_uuid, NEW_URI)

    instances_logs = logs[-len(instances):]
    assert_ok_for_all_instances(instances_logs, instances)

    # check app config changes
    new_conf = get_conf_with_new_uri(old_conf, config.instance_uuid, NEW_URI)
    assert_conf_changed(conf_paths, other_app_conf_paths, old_conf, new_conf)


@pytest.mark.parametrize('conf_type', ['simple', 'srv-disabled'])
def test_set_uri_dry_run(cartridge_cmd, conf_type, tmpdir,
                         clusterwide_conf_simple,
                         clusterwide_conf_srv_disabled):
    data_dir = os.path.join(tmpdir, 'tmp', 'data')
    os.makedirs(data_dir)

    NEW_URI = 'new-uri:666'

    configs = {
        'simple': clusterwide_conf_simple,
        'srv-disabled': clusterwide_conf_srv_disabled,
    }

    config = configs[conf_type]
    old_conf = config.conf

    # create app working directories
    instances = ['instance-1', 'instance-2']
    conf_paths = write_instances_topology_conf(data_dir, APPNAME, old_conf, instances)

    cmd = [
        cartridge_cmd, 'repair', 'set-advertise-uri',
        '--name', APPNAME,
        '--data-dir', data_dir,
        '--dry-run',
        config.instance_uuid, NEW_URI,
    ]

    rc, output = run_command_and_get_output(cmd, cwd=tmpdir)
    assert rc == 0

    # check logs
    assert "Set %s advertise URI to %s" % (config.instance_uuid, NEW_URI) in output

    exp_diff = '\n'.join([
        '-    uri: %s' % config.instance_uri,
        '+    uri: %s' % NEW_URI,
    ])
    assert exp_diff in output

    # check config wasn't changed
    assert_conf_not_changed(conf_paths, old_conf)
