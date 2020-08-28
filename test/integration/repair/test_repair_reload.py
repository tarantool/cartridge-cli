import subprocess
import os
import requests
import pytest

from utils import get_instance_id
from utils import check_instances_running
from utils import check_instances_stopped
from utils import DEFAULT_CFG
from utils import DEFAULT_DATA_DIR
from utils import DEFAULT_RUN_DIR
from utils import write_conf
from utils import create_replicaset
from utils import run_command_and_get_output
from utils import wait_for_replicaset_is_healthy

from project import patch_cartridge_proc_titile
from project import patch_cartridge_version


def get_replicaset(admin_api_url, replicaset_uuid):
    query = '''
        query {{
        replicasets: replicasets(uuid: "{uuid}") {{
            uuid
            status
            servers {{
                uuid
                uri
                priority
            }}
            master {{
                uuid
            }}
        }}
    }}
    '''.format(uuid=replicaset_uuid)

    r = requests.post(admin_api_url, json={'query': query})
    assert r.status_code == 200
    resp = r.json()

    return resp['data']['replicasets'][0]


@pytest.mark.parametrize('cartridge_version', ['1.2.0', None])
def test_repair_reload_old_cartridge(cartridge_cmd, start_stop_cli, project_with_cartridge, cartridge_version, tmpdir):
    project = project_with_cartridge
    cli = start_stop_cli

    patch_cartridge_version(project, cartridge_version)

    cmd = [
        cartridge_cmd,
        "build",
        project.path
    ]
    process = subprocess.run(cmd, cwd=tmpdir)
    assert process.returncode == 0, "Error during building the project"

    # patch cartridge.cfg to don't change process title
    patch_cartridge_proc_titile(project)

    # start instances
    INSTANCE1 = 'instance-1'
    INSTANCE2 = 'instance-2'

    ID1 = get_instance_id(project.name, INSTANCE1)
    ID2 = get_instance_id(project.name, INSTANCE2)

    cfg = {
        ID1: {
            'advertise_uri': 'localhost:3301',
            'http_port': 8081,
        },
        ID2: {
            'advertise_uri': 'localhost:3302',
            'http_port': 8082,
        },
    }

    write_conf(os.path.join(project.path, DEFAULT_CFG), cfg)

    # start instance-1 and instance-2
    cli.start(project, daemonized=True)
    check_instances_running(cli, project, [INSTANCE1, INSTANCE2], daemonized=True)

    data_dir = os.path.join(project.path, DEFAULT_DATA_DIR)
    run_dir = os.path.join(project.path, DEFAULT_RUN_DIR)

    cmd = [
        cartridge_cmd, 'repair', 'set-leader',
        '--name', project.name,
        '--data-dir', data_dir,
        '--run-dir', run_dir,
        '--reload',
        '--verbose',
        'some-rpl', 'some-instance',
    ]

    rc, output = run_command_and_get_output(cmd, cwd=tmpdir)
    assert rc == 1

    if cartridge_version is None:
        version_err = "Cartridge version is less than 2.0.0."
    else:
        version_err = "Cartridge version (%s) is less than 2.0.0." % cartridge_version

    exp_err = "Configurations reload isn't possible: %s Please, specify --no-reload flag" % version_err
    assert exp_err in output


def test_repair_reload_set_leader(cartridge_cmd, start_stop_cli, project_with_cartridge, tmpdir):
    project = project_with_cartridge
    cli = start_stop_cli

    cmd = [
        cartridge_cmd,
        "build",
        project.path
    ]
    process = subprocess.run(cmd, cwd=tmpdir)
    assert process.returncode == 0, "Error during building the project"

    # patch cartridge.cfg to don't change process title
    patch_cartridge_proc_titile(project)

    # start instances
    INSTANCE1 = 'instance-1'
    INSTANCE2 = 'instance-2'

    ID1 = get_instance_id(project.name, INSTANCE1)
    ID2 = get_instance_id(project.name, INSTANCE2)

    ADMIN_HTTP_PORT = 8081

    cfg = {
        ID1: {
            'advertise_uri': 'localhost:3301',
            'http_port': ADMIN_HTTP_PORT,
        },
        ID2: {
            'advertise_uri': 'localhost:3302',
            'http_port': 8082,
        },
    }

    write_conf(os.path.join(project.path, DEFAULT_CFG), cfg)

    # start instance-1 and instance-2
    cli.start(project, daemonized=True)
    check_instances_running(cli, project, [INSTANCE1, INSTANCE2], daemonized=True)

    advertise_uris = [cfg[id]['advertise_uri'] for id in cfg]

    # join instances to replicaset
    admin_api_url = 'http://localhost:%s/admin/api' % ADMIN_HTTP_PORT
    replicaset_uuid = create_replicaset(admin_api_url, advertise_uris, ['vshard-storage'])

    replicaset = get_replicaset(admin_api_url, replicaset_uuid)
    cluster_instances = replicaset['servers']

    # change leader
    instances_by_priority = sorted(cluster_instances, key=lambda i: i['priority'])
    new_leader_uuid = instances_by_priority[-1]['uuid']

    data_dir = os.path.join(project.path, DEFAULT_DATA_DIR)
    run_dir = os.path.join(project.path, DEFAULT_RUN_DIR)

    cmd = [
        cartridge_cmd, 'repair', 'set-leader',
        '--name', project.name,
        '--data-dir', data_dir,
        '--run-dir', run_dir,
        '--reload',
        '--verbose',
        replicaset_uuid, new_leader_uuid,
    ]

    rc, output = run_command_and_get_output(cmd, cwd=tmpdir)
    assert rc == 0

    wait_for_replicaset_is_healthy(admin_api_url, replicaset_uuid)

    new_replicaset = get_replicaset(admin_api_url, replicaset_uuid)
    new_replicaset_leader_uuid = new_replicaset['master']['uuid']

    assert new_replicaset_leader_uuid == new_leader_uuid


def test_repair_reload_remove_instance(cartridge_cmd, start_stop_cli, project_with_cartridge, tmpdir):
    project = project_with_cartridge
    cli = start_stop_cli

    cmd = [
        cartridge_cmd,
        "build",
        project.path
    ]
    process = subprocess.run(cmd, cwd=tmpdir)
    assert process.returncode == 0, "Error during building the project"

    # patch cartridge.cfg to don't change process title
    patch_cartridge_proc_titile(project)

    # start instances
    INSTANCE1 = 'instance-1'
    INSTANCE2 = 'instance-2'

    ID1 = get_instance_id(project.name, INSTANCE1)
    ID2 = get_instance_id(project.name, INSTANCE2)

    ADVERTISE_URI_TO_REMOVE = 'localhost:3302'
    ADMIN_HTTP_PORT = 8081

    cfg = {
        # this instance shouldn't be removed since we use it's http port
        ID1: {
            'advertise_uri': 'localhost:3301',
            'http_port': ADMIN_HTTP_PORT,
        },
        ID2: {
            'advertise_uri': ADVERTISE_URI_TO_REMOVE,
            'http_port': 8082,
        },
    }

    write_conf(os.path.join(project.path, DEFAULT_CFG), cfg)

    # start instance-1 and instance-2
    cli.start(project, daemonized=True)
    check_instances_running(cli, project, [INSTANCE1, INSTANCE2], daemonized=True)

    advertise_uris = [cfg[id]['advertise_uri'] for id in cfg]

    admin_api_url = 'http://localhost:%s/admin/api' % ADMIN_HTTP_PORT

    # join instances to replicaset
    replicaset_uuid = create_replicaset(admin_api_url, advertise_uris, ['vshard-storage'])

    replicaset = get_replicaset(admin_api_url, replicaset_uuid)
    cluster_instances = replicaset['servers']

    # change leader
    instance_to_remove_uuid = None
    for instance in cluster_instances:
        if instance['uri'] == ADVERTISE_URI_TO_REMOVE:
            instance_to_remove_uuid = instance['uuid']
            break

    assert instance_to_remove_uuid is not None

    data_dir = os.path.join(project.path, DEFAULT_DATA_DIR)
    run_dir = os.path.join(project.path, DEFAULT_RUN_DIR)

    cmd = [
        cartridge_cmd, 'repair', 'remove-instance',
        '--name', project.name,
        '--data-dir', data_dir,
        '--run-dir', run_dir,
        '--reload',
        '--verbose',
        instance_to_remove_uuid,
    ]

    rc, output = run_command_and_get_output(cmd, cwd=tmpdir)
    assert rc == 0

    wait_for_replicaset_is_healthy(admin_api_url, replicaset_uuid)

    new_replicaset = get_replicaset(admin_api_url, replicaset_uuid)
    new_replicaset_instances_uuids = [
        instance['uuid'] for instance in new_replicaset['servers']
    ]

    assert instance_to_remove_uuid not in new_replicaset_instances_uuids


def test_repair_reload_set_uri(cartridge_cmd, start_stop_cli, project_with_cartridge, tmpdir):
    project = project_with_cartridge
    cli = start_stop_cli

    cmd = [
        cartridge_cmd,
        "build",
        project.path
    ]
    process = subprocess.run(cmd, cwd=tmpdir)
    assert process.returncode == 0, "Error during building the project"

    # patch cartridge.cfg to don't change process title
    patch_cartridge_proc_titile(project)

    # start instances
    INSTANCE1 = 'instance-1'
    INSTANCE2 = 'instance-2'

    ID1 = get_instance_id(project.name, INSTANCE1)
    ID2 = get_instance_id(project.name, INSTANCE2)

    ADMIN_HTTP_PORT = 8081
    ADVERTISE_URI_TO_CHANGE = 'localhost:3302'
    NEW_ADVERTISE_URI = 'localhost:3322'
    INSTANCE_TO_SET_URI = INSTANCE2

    cfg = {
        ID1: {
            'advertise_uri': 'localhost:3301',
            'http_port': ADMIN_HTTP_PORT,
            'replication_connect_quorum': 0,
            'custom_proc_title': '',
        },
        ID2: {
            'advertise_uri': ADVERTISE_URI_TO_CHANGE,
            'http_port': 8082,
            'replication_connect_quorum': 0,
            'custom_proc_title': '',
        },
    }

    write_conf(os.path.join(project.path, DEFAULT_CFG), cfg)

    # start instance-1 and instance-2
    cli.start(project, daemonized=True)
    check_instances_running(cli, project, [INSTANCE1, INSTANCE2], daemonized=True)

    advertise_uris = [cfg[id]['advertise_uri'] for id in cfg]

    admin_api_url = 'http://localhost:%s/admin/api' % ADMIN_HTTP_PORT

    # join instances to replicaset
    replicaset_uuid = create_replicaset(admin_api_url, advertise_uris, ['vshard-storage'])

    replicaset = get_replicaset(admin_api_url, replicaset_uuid)
    instance_to_set_uri_uuid = None
    for instance in replicaset['servers']:
        if instance['uri'] == ADVERTISE_URI_TO_CHANGE:
            instance_to_set_uri_uuid = instance['uuid']
            break

    assert instance_to_set_uri_uuid is not None

    # first, change URI in config and restart instance
    INSTANCE_ID = get_instance_id(project.name, INSTANCE_TO_SET_URI)

    cfg[INSTANCE_ID]['advertise_uri'] = NEW_ADVERTISE_URI
    write_conf(os.path.join(project.path, DEFAULT_CFG), cfg)

    cli.stop(project, [INSTANCE_TO_SET_URI])
    check_instances_stopped(cli, project, [INSTANCE_TO_SET_URI])

    cli.start(project, [INSTANCE_TO_SET_URI], daemonized=True)
    check_instances_running(cli, project, [INSTANCE1, INSTANCE2],  daemonized=True, skip_env_checks=True)

    # then, update cluster-wide configs
    data_dir = os.path.join(project.path, DEFAULT_DATA_DIR)
    run_dir = os.path.join(project.path, DEFAULT_RUN_DIR)

    cmd = [
        cartridge_cmd, 'repair', 'set-advertise-uri',
        '--name', project.name,
        '--data-dir', data_dir,
        '--run-dir', run_dir,
        '--reload',
        '--verbose',
        instance_to_set_uri_uuid, NEW_ADVERTISE_URI,
    ]

    rc, output = run_command_and_get_output(cmd, cwd=tmpdir)
    assert rc == 0

    wait_for_replicaset_is_healthy(admin_api_url, replicaset_uuid)

    new_replicaset = get_replicaset(admin_api_url, replicaset_uuid)
    new_replicaset_instance_advertise_uri = None
    for instance in new_replicaset['servers']:
        if instance['uuid'] == instance_to_set_uri_uuid:
            new_replicaset_instance_advertise_uri = instance['uri']
            break

    assert new_replicaset_instance_advertise_uri == NEW_ADVERTISE_URI
