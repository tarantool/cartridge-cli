import os
import pytest

from utils import run_command_and_get_output
from utils import get_logs
from utils import assert_ok_for_all_instances
from utils import assert_ok_for_instances_group
from clusterwide_conf import assert_conf_not_changed

from clusterwide_conf import get_srv_conf
from clusterwide_conf import get_rpl_conf
from clusterwide_conf import get_topology_conf
from clusterwide_conf import get_conf_with_new_uri
from clusterwide_conf import get_conf_with_removed_instance
from clusterwide_conf import get_conf_with_new_leader
from clusterwide_conf import ClusterwideConfig
from clusterwide_conf import write_instances_topology_conf
from clusterwide_conf import assert_conf_changed


APPNAME = 'myapp'


simple_args = {
    'set-uri': ['srv-1', 'localhost:3310'],
    'remove-instance': ['srv-2'],
    'set-leader': ['rpl-1', 'srv-1']
}


##########
# FIXTURES
@pytest.fixture(scope="function")
def clusterwide_conf_simple_v1():
    INSTANCE_UUID = 'srv-1'
    REPLICASET_UUID = 'rpl-1'

    conf = get_topology_conf(
        instances=[
            get_srv_conf(INSTANCE_UUID, uri='localhost:3301', rpl_uuid=REPLICASET_UUID),
            get_srv_conf('srv-2', rpl_uuid='rpl-1'),
            get_srv_conf('srv-3', rpl_uuid='rpl-2'),
        ],
        replicasets=[
            get_rpl_conf(REPLICASET_UUID, leaders=[INSTANCE_UUID]),
            get_rpl_conf('rpl-2', leaders=['srv-3']),
        ]
    )

    return ClusterwideConfig(conf, instance_uuid=INSTANCE_UUID,
                             replicaset_uuid=REPLICASET_UUID)


@pytest.fixture(scope="function")
def clusterwide_conf_simple_v2():
    INSTANCE_UUID = 'srv-1'
    REPLICASET_UUID = 'rpl-1'

    conf = get_topology_conf(
        instances=[
            get_srv_conf(INSTANCE_UUID, uri='localhost:3301', rpl_uuid=REPLICASET_UUID),
            get_srv_conf('srv-2', rpl_uuid='rpl-1'),
            get_srv_conf('srv-3', rpl_uuid='rpl-2'),
            get_srv_conf('srv-4', rpl_uuid='rpl-2'),  # <= one more instance in rpl-2
        ],
        replicasets=[
            get_rpl_conf(REPLICASET_UUID, leaders=[INSTANCE_UUID]),
            get_rpl_conf('rpl-2', leaders=['srv-3']),
        ]
    )

    return ClusterwideConfig(conf, instance_uuid=INSTANCE_UUID,
                             replicaset_uuid=REPLICASET_UUID)


#######
# TESTS
@pytest.mark.parametrize('repair_cmd', ['set-uri', 'remove-instance', 'set-leader', 'list-topology'])
def test_no_force(cartridge_cmd, repair_cmd, tmpdir, clusterwide_conf_simple_v1, clusterwide_conf_simple_v2):
    data_dir = os.path.join(tmpdir, 'tmp', 'data')
    os.makedirs(data_dir)

    config1 = clusterwide_conf_simple_v1
    config2 = clusterwide_conf_simple_v2

    # create app configs
    conf1_instances = ['instance-1', 'instance-2']
    write_instances_topology_conf(data_dir, APPNAME, config1.conf, conf1_instances)

    conf2_instances = ['instance-3', 'instance-4']
    write_instances_topology_conf(data_dir, APPNAME, config2.conf, conf2_instances)

    args = simple_args.get(repair_cmd, [])
    cmd = [
        cartridge_cmd, 'repair', repair_cmd,
        '--name', APPNAME,
        '--data-dir', data_dir,
    ]
    cmd.extend(args)

    rc, output = run_command_and_get_output(cmd, cwd=tmpdir)
    assert rc == 1

    assert "Clusterwide config is diverged between instances" in output
    assert '''--- instance-3, instance-4
+++ instance-1, instance-2
@@ -27,10 +27,6 @@
     uri: srv-2-uri
   srv-3:
     disabled: false
     replicaset_uuid: rpl-2
     uri: srv-3-uri
-  srv-4:
-    disabled: false
-    replicaset_uuid: rpl-2
-    uri: srv-4-uri
''' in output


@pytest.mark.parametrize('repair_cmd', ['set-uri', 'remove-instance', 'set-leader'])
def test_force_patch(cartridge_cmd, repair_cmd, tmpdir, clusterwide_conf_simple_v1, clusterwide_conf_simple_v2):
    data_dir = os.path.join(tmpdir, 'tmp', 'data')
    os.makedirs(data_dir)

    config1 = clusterwide_conf_simple_v1
    config2 = clusterwide_conf_simple_v2

    # create app configs
    conf1_instances = ['instance-1', 'instance-2']
    conf1_paths = write_instances_topology_conf(data_dir, APPNAME, config1.conf, conf1_instances)

    conf2_instances = ['instance-3', 'instance-4']
    conf2_paths = write_instances_topology_conf(data_dir, APPNAME, config2.conf, conf2_instances)

    instances = conf1_instances + conf2_instances

    args = simple_args.get(repair_cmd, [])
    cmd = [
        cartridge_cmd, 'repair', repair_cmd,
        '--name', APPNAME,
        '--data-dir', data_dir,
        '--force',
    ]
    cmd.extend(args)

    rc, output = run_command_and_get_output(cmd, cwd=tmpdir)
    assert rc == 0

    if repair_cmd == 'set-uri':
        first_log_line = "Set %s advertise URI to %s" % (args[0], args[1])
    elif repair_cmd == 'remove-instance':
        first_log_line = "Remove instance with UUID %s" % args[0]
    elif repair_cmd == 'set-leader':
        first_log_line = "Set %s leader to %s" % (args[0], args[1])

    logs = get_logs(output)
    assert logs[0] == first_log_line
    assert "Clusterwide config is diverged between instances" in logs[1]
    assert logs[2] == "Process application cluster-wide configurations..."

    process_conf_logs = logs[3:5]
    assert_ok_for_instances_group(process_conf_logs, conf1_instances)
    assert_ok_for_instances_group(process_conf_logs, conf2_instances)

    assert logs[5] == "Write application cluster-wide configurations..."

    write_conf_logs = logs[6:]
    assert_ok_for_all_instances(write_conf_logs, instances)

    # check config changes independently
    if repair_cmd == 'set-uri':
        new_conf1 = get_conf_with_new_uri(config1.conf, config1.instance_uuid, args[1])
        new_conf2 = get_conf_with_new_uri(config2.conf, config2.instance_uuid, args[1])
    elif repair_cmd == 'remove-instance':
        new_conf1 = get_conf_with_removed_instance(config1.conf, args[0])
        new_conf2 = get_conf_with_removed_instance(config2.conf, args[0])
    elif repair_cmd == 'set-leader':
        new_conf1 = get_conf_with_new_leader(config1.conf, args[0], args[1])
        new_conf2 = get_conf_with_new_leader(config2.conf, args[0], args[1])

    assert_conf_changed(conf1_paths, None, config1.conf, new_conf1)
    assert_conf_changed(conf2_paths, None, config2.conf, new_conf2)


def test_force_list_topology(cartridge_cmd, tmpdir, clusterwide_conf_simple_v1, clusterwide_conf_simple_v2):
    data_dir = os.path.join(tmpdir, 'tmp', 'data')
    os.makedirs(data_dir)

    config1 = clusterwide_conf_simple_v1
    config2 = clusterwide_conf_simple_v2

    # create app configs
    conf1_instances = ['instance-1', 'instance-2']
    conf1_paths = write_instances_topology_conf(data_dir, APPNAME, config1.conf, conf1_instances)

    conf2_instances = ['instance-3', 'instance-4']
    conf2_paths = write_instances_topology_conf(data_dir, APPNAME, config2.conf, conf2_instances)

    # instances = conf1_instances + conf2_instances

    cmd = [
        cartridge_cmd, 'repair', 'list-topology',
        '--name', APPNAME,
        '--data-dir', data_dir,
        '--force',
    ]

    rc, output = run_command_and_get_output(cmd, cwd=tmpdir)
    assert rc == 0

    assert_conf_not_changed(conf1_paths, config1.conf)
    assert_conf_not_changed(conf2_paths, config2.conf)

    lines = output.split('\n')
    logs = get_logs('\n'.join(lines[:3]))

    assert logs[0] == "Get current topology"
    assert "Clusterwide config is diverged between instances" in logs[1]
    assert logs[2] == "Process application cluster-wide configurations..."

    assert "Write application cluster-wide configurations..." not in output

    exp_summary_conf1 = '''   • instance-1, instance-2... OK
Instances
  * srv-1
    URI: localhost:3301
    replicaset: rpl-1
  * srv-2
    URI: srv-2-uri
    replicaset: rpl-1
  * srv-3
    URI: srv-3-uri
    replicaset: rpl-2
Replicasets
  * rpl-1
    roles:
     * vshard-storage
    instances:
     * srv-1
     * srv-2
  * rpl-2
    roles:
     * vshard-storage
    instances:
     * srv-3
'''

    exp_summary_conf2 = '''
   • instance-3, instance-4... OK
Instances
  * srv-1
    URI: localhost:3301
    replicaset: rpl-1
  * srv-2
    URI: srv-2-uri
    replicaset: rpl-1
  * srv-3
    URI: srv-3-uri
    replicaset: rpl-2
  * srv-4
    URI: srv-4-uri
    replicaset: rpl-2
Replicasets
  * rpl-1
    roles:
     * vshard-storage
    instances:
     * srv-1
     * srv-2
  * rpl-2
    roles:
     * vshard-storage
    instances:
     * srv-3
     * srv-4'''

    assert exp_summary_conf1 in output
    assert exp_summary_conf2 in output
