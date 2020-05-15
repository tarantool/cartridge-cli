import subprocess
import yaml
import os
import requests
import pytest

from utils import Cli
from utils import DEFAULT_CFG
from utils import get_stateboard_name, get_instance_id
from utils import check_instances_running
from utils import patch_cartridge_proc_titile
from utils import create_replicaset, wait_for_replicaset_is_healthy, get_replicaset_roles
from utils import bootstrap_vshard


# #######
# Helpers
# #######
def get_instances_from_conf(project):
    res = dict()

    with open(os.path.join(project.path, DEFAULT_CFG)) as f:
        conf = yaml.safe_load(f)

    for instance_id, instance_conf in conf.items():
        if instance_id == get_stateboard_name(project.name):
            continue

        id_parts = instance_id.split(".")
        assert len(id_parts) == 2
        assert id_parts[0] == project.name

        assert 'http_port' in instance_conf
        assert 'workdir' in instance_conf
        assert 'advertise_uri' in instance_conf

        assert instance_id not in res
        res[instance_id] = instance_conf

    return res


# #####
# Tests
# #####
def test_project(cartridge_cmd, project_getting_started):
    project = project_getting_started

    # build app
    process = subprocess.run([cartridge_cmd, 'build'], cwd=project.path)
    assert process.returncode == 0

    # install test deps
    process = subprocess.run(["./deps.sh"], cwd=project.path)
    assert process.returncode == 0

    # run luacheck
    process = subprocess.run([".rocks/bin/luacheck", "."], cwd=project.path)
    assert process.returncode == 0

    # run luatest
    process = subprocess.run([".rocks/bin/luatest", "-v"], cwd=project.path)
    assert process.returncode == 0


@pytest.mark.skip()
def test_api(cartridge_cmd, project_getting_started):
    project = project_getting_started
    cli = Cli(cartridge_cmd)

    APP_INSTANCES = [get_instance_id(project.name, 'router')]
    S1_INSTANCES = [get_instance_id(project.name, 's1-master'), get_instance_id(project.name, 's1-replica')]
    S2_INSTANCES = [get_instance_id(project.name, 's2-master'), get_instance_id(project.name, 's2-replica')]

    # build app
    process = subprocess.run([cartridge_cmd, 'build'], cwd=project.path)
    assert process.returncode == 0

    patch_cartridge_proc_titile(project)

    # check config and get instances
    instances_conf = get_instances_from_conf(project)
    instances = list(instances_conf.keys())

    assert all([
        instance in instances
        for instance in APP_INSTANCES + S1_INSTANCES + S2_INSTANCES
    ])

    router_http_port = instances_conf[APP_INSTANCES[0]]['http_port']
    admin_api_url = 'http://localhost:{}/admin/api'.format(router_http_port)

    # start application in interactive mode (to easily check logs on debug)
    cli.start(project)
    check_instances_running(cli, project, instances)

    # create app replicaset
    uris = [instances_conf[instance]['advertise_uri'] for instance in APP_INSTANCES]
    roles = ['api']
    app_replicaset_uuid = create_replicaset(admin_api_url, uris, roles)
    wait_for_replicaset_is_healthy(admin_api_url, app_replicaset_uuid)

    replicaset_roles = get_replicaset_roles(admin_api_url, app_replicaset_uuid)
    # api role should contain vshard-router dependency
    assert set(replicaset_roles) == set(['api', 'vshard-router'])

    # create s1 replicaset
    uris = [instances_conf[instance]['advertise_uri'] for instance in S1_INSTANCES]
    roles = ['storage']
    s1_replicaset_uuid = create_replicaset(admin_api_url, uris, roles)
    wait_for_replicaset_is_healthy(admin_api_url, s1_replicaset_uuid)

    replicaset_roles = get_replicaset_roles(admin_api_url, s1_replicaset_uuid)
    # storage role should contain vshard-storage dependency
    assert set(replicaset_roles) == set(['storage', 'vshard-storage'])

    # create s2 replicaset
    uris = [instances_conf[instance]['advertise_uri'] for instance in S2_INSTANCES]
    roles = ['storage']
    s2_replicaset_uuid = create_replicaset(admin_api_url, uris, roles)
    wait_for_replicaset_is_healthy(admin_api_url, s2_replicaset_uuid)

    replicaset_roles = get_replicaset_roles(admin_api_url, s2_replicaset_uuid)
    # storage role should contain vshard-storage dependency
    assert set(replicaset_roles) == set(['storage', 'vshard-storage'])

    # bootstrap vshard
    bootstrap_vshard(admin_api_url)

    # test HTTP API
    CUSTOMER_ID = 10
    CUSTOMER_NAME = 'Elizabeth'
    customer = {
        'customer_id': CUSTOMER_ID,
        'name': CUSTOMER_NAME
    }

    # create new customer
    url = 'http://localhost:{}/storage/customers/create'.format(router_http_port)
    r = requests.post(url, json=customer)
    assert r.status_code == requests.status_codes.codes.CREATED
    resp = r.json()
    assert 'info' in resp
    assert resp['info'] == 'Successfully created'

    # # create the same customer again
    # r = requests.post(url, json=customer)
    # # XXX: r.status_code is 500 now

    # get customer
    url = 'http://localhost:{}/storage/customers/{}'.format(router_http_port, CUSTOMER_ID)
    r = requests.get(url, json=customer)
    assert r.status_code == requests.status_codes.codes.OK
    resp = r.json()
    assert resp == {
        'customer_id': CUSTOMER_ID,
        'name': CUSTOMER_NAME,
        'accounts': [],
    }

    # get customer by wrong id
    url = 'http://localhost:{}/storage/customers/{}'.format(router_http_port, CUSTOMER_ID+1)
    r = requests.get(url, json=customer)
    assert r.status_code == requests.status_codes.codes.NOT_FOUND
    resp = r.json()
    assert 'info' in resp
    assert resp['info'] == 'Customer not found'

    # update customer balance
    # XXX: now I have no idea how to perform this call
    # it requires account_id field, but now I don't even
    # understand how to add an account to customer
