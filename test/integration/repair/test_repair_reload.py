import subprocess
import os
import requests

from utils import get_instance_id
from utils import check_instances_running
from utils import DEFAULT_CFG
from utils import DEFAULT_DATA_DIR
from utils import DEFAULT_RUN_DIR
from utils import write_conf
from utils import create_replicaset
from utils import run_command_and_get_output
from utils import wait_for_replicaset_is_healthy

from project import patch_cartridge_proc_titile


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

    advertise_uris = [cfg[id]['advertise_uri'] for id in cfg]

    # join instances to replicaset
    admin_api_url = 'http://localhost:{}/admin/api'.format(cfg[ID1]['http_port'])
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
        replicaset_uuid, new_leader_uuid,
    ]

    rc, output = run_command_and_get_output(cmd, cwd=tmpdir)
    assert rc == 0

    wait_for_replicaset_is_healthy(admin_api_url, replicaset_uuid)

    new_replicaset = get_replicaset(admin_api_url, replicaset_uuid)
    new_replicaset_leader_uuid = new_replicaset['master']['uuid']

    assert new_replicaset_leader_uuid == new_leader_uuid
