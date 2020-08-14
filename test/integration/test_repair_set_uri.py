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


def test_uri_does_not_exist(cartridge_cmd, clusterwide_conf_non_existent_uri, tmpdir):
    data_dir = os.path.join(tmpdir, 'tmp', 'data')
    os.makedirs(data_dir)

    NEW_URI = 'new-uri:666'

    config = clusterwide_conf_non_existent_uri

    instances = ['instance-1', 'instance-2']
    write_instance_topology_conf(data_dir, APPNAME, config.conf, instances)

    cmd = [
        cartridge_cmd, 'repair', 'set-uri',
        '--name', APPNAME,
        '--data-dir', data_dir,
        config.instance_uri, NEW_URI,
    ]

    rc, output = run_command_and_get_output(cmd, cwd=tmpdir)
    assert rc == 1

    assert_for_all_instances(
        get_logs(output), APPNAME, instances, lambda line:
        "Instance with URI %s isn't found in the cluster" % config.instance_uri in line
    )


@pytest.mark.parametrize('conf_type', ['simple', 'srv-disabled'])
def test_set_uri(cartridge_cmd, conf_type, tmpdir,
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
    conf_paths = write_instance_topology_conf(data_dir, APPNAME, old_conf, instances)

    # create other app working directories
    other_instances = ['other-instance-1', 'other-instance-2']
    other_app_conf_paths = write_instance_topology_conf(data_dir, OTHER_APP_NAME, old_conf, other_instances)

    cmd = [
        cartridge_cmd, 'repair', 'set-uri',
        '--name', APPNAME,
        '--data-dir', data_dir,
        config.instance_uri, NEW_URI,
    ]

    rc, output = run_command_and_get_output(cmd, cwd=tmpdir)
    assert rc == 0

    # check logs
    logs = get_logs(output)
    assert len(logs) == len(instances) + 1
    assert logs[0] == "Update advertise URI %s -> %s" % (config.instance_uri, NEW_URI)
    assert all([line.strip().endswith('OK') for line in logs[1:]])
    assert_for_all_instances(
        logs[1:], APPNAME, instances, lambda line: line.strip().endswith('OK'),
    )

    # check app config changes
    new_conf = copy.deepcopy(old_conf)
    new_conf['servers'][config.instance_uuid]['uri'] = NEW_URI
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
    conf_paths = write_instance_topology_conf(data_dir, APPNAME, old_conf, instances)

    cmd = [
        cartridge_cmd, 'repair', 'set-uri',
        '--name', APPNAME,
        '--data-dir', data_dir,
        '--dry-run',
        config.instance_uri, NEW_URI,
    ]

    rc, output = run_command_and_get_output(cmd, cwd=tmpdir)
    assert rc == 0

    # check logs
    assert "Update advertise URI %s -> %s" % (config.instance_uri, NEW_URI) in output
    assert "Data directory is set to: %s" % data_dir in output

    assert all([
        'Topology config file: %s' % conf_path in output
        for conf_path in conf_paths
    ])

    exp_diff = '\n'.join([
        '-    uri: %s' % config.instance_uri,
        '+    uri: %s' % NEW_URI,
    ])
    assert exp_diff in output

    # check config wasn't changed
    assert_conf_not_changed(conf_paths, old_conf)
