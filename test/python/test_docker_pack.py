#!/usr/bin/python3

import pytest
import os
import docker
import re
import requests
import time
import subprocess
import tarfile

from utils import basepath
from utils import tarantool_version
from utils import tarantool_enterprise_is_used
from utils import recursive_listdir
from utils import assert_dir_contents
from utils import assert_filemodes


# #######
# Helpers
# #######
def find_image(docker_client, project_name):
    for image in docker_client.images.list():
        for t in image.tags:
            if t.startswith(project_name):
                return t


def run_command_on_image(docker_client, image_name, command):
    command = '/bin/bash -c "{}"'.format(command.replace('"', '\\"'))
    output = docker_client.containers.run(
        image_name,
        command,
        remove=True
    )
    return output.decode("utf-8").strip()


def wait_for_container_start(container, timeout=10):
    time_start = time.time()
    while True:
        now = time.time()
        if now > time_start + timeout:
            break

        container_logs = container.logs(since=int(time_start)).decode('utf-8')
        if 'entering the event loop' in container_logs:
            return True

        time.sleep(1)

    return False


# ########
# Fixtures
# ########
@pytest.fixture(scope="module")
def docker_client():
    client = docker.from_env()
    return client


@pytest.fixture(scope="module")
def docker_image(module_tmpdir, project, request, docker_client):
    if project['name'] == 'deprecated-project':
        pytest.skip("Docker packing isn't allowed for deprecated project")

    cmd = [os.path.join(basepath, "cartridge"), "pack", "docker", project['path']]
    process = subprocess.run(cmd, cwd=module_tmpdir)
    assert process.returncode == 0, \
        "Error during creating of docker image"

    image_name = find_image(docker_client, project['name'])
    assert image_name is not None, "Docker image isn't found"

    def delete_image(image_name):
        if docker_client.images.list(image_name):
            # remove all image containers
            containers = docker_client.containers.list(
                all=True,
                filters={'ancestor': image_name}
            )

            for c in containers:
                c.remove(force=True)

            # remove image itself
            docker_client.images.remove(image_name)

    request.addfinalizer(lambda: delete_image(image_name))
    return {'name': image_name}


# #####
# Tests
# #####
def test_invalid_base_dockerfile(project, module_tmpdir, tmpdir):
    if project['name'] == 'deprecated-project':
        pytest.skip("Docker packing isn't allowed for deprecated project")

    invalid_dockerfile_path = os.path.join(tmpdir, 'Dockerfile')
    with open(invalid_dockerfile_path, 'w') as f:
        f.write('''
            # Invalid dockerfile
            FROM ubuntu:xenial
        ''')

    cmd = [
        os.path.join(basepath, "cartridge"),
        "pack", "docker",
        "--from", invalid_dockerfile_path,
        project['path'],
    ]
    process = subprocess.run(cmd, cwd=module_tmpdir)
    assert process.returncode == 1


def test_docker_pack(project, docker_image, tmpdir, docker_client):
    image_name = docker_image['name']
    container = docker_client.containers.create(image_name)

    container_distribution_dir = '/usr/share/tarantool/{}'.format(project['name'])

    # check if distribution dir was created
    command = '[ -d "{}" ] && echo true || echo false'.format(container_distribution_dir)
    output = run_command_on_image(docker_client, image_name, command)
    assert output == 'true'

    # get distribution dir contents
    arhive_path = os.path.join(tmpdir, 'distribution_dir.tar')
    with open(arhive_path, 'wb') as f:
        bits, _ = container.get_archive(container_distribution_dir)
        for chunk in bits:
            f.write(chunk)

    with tarfile.open(arhive_path) as arch:
        arch.extractall(path=os.path.join(tmpdir, 'usr/share/tarantool'))
    os.remove(arhive_path)

    assert_dir_contents(
        files_list=recursive_listdir(os.path.join(tmpdir, 'usr/share/tarantool/', project['name'])),
        exp_files_list=project['distribution_files_list'],
        exp_rocks_content=project['rocks_content'],
        skip_tarantool_binaries=True
    )

    assert_filemodes(project, tmpdir)
    container.remove()

    if tarantool_enterprise_is_used():
        # check tarantool and tarantoolctl binaries
        command = '[ -d "/usr/share/tarantool/tarantool-enterprise/" ] && echo true || echo false'
        output = run_command_on_image(docker_client, image_name, command)
        assert output == 'true'

        command = 'cd /usr/share/tarantool/tarantool-enterprise/ && find .'
        output = run_command_on_image(docker_client, image_name, command)

        files_list = output.split('\n')
        files_list.remove('.')

        dir_contents = [
            os.path.normpath(filename)
            for filename in files_list
        ]

        assert 'tarantool' in dir_contents
        assert 'tarantoolctl' in dir_contents
    else:
        # check if tarantool was installed
        command = 'yum list installed 2>/dev/null | grep tarantool'
        output = run_command_on_image(docker_client, image_name, command)

        packages_list = output.split('\n')
        assert any(['tarantool' in package for package in packages_list])

        # check tarantool version
        command = 'yum info tarantool'
        output = run_command_on_image(docker_client, image_name, command)

        m = re.search(r'Version\s+:\s+(\d+)\.(\d+).', output)
        assert m is not None
        installed_version = m.groups()

        m = re.search(r'(\d+)\.(\d+)\.\d+', tarantool_version())
        assert m is not None
        expected_version = m.groups()

        assert installed_version == expected_version


def test_base_dockerfile_with_env_vars(project, module_tmpdir, tmpdir):
    # The main idea of this test is to check that using `${name}` constructions
    #   in the base Dockerfile doesn't break the `pack docker` command running.
    # So, it's not about testing that the ENV option works, it's about
    #   testing that `pack docker` command wouldn't fail if the base Dockerfile
    #   contains `${name}` constructions.
    # The problem is the `expand` function.
    # Base Dockerfile with `${name}` shouldn't be passed to this function,
    #   otherwise it will raise an error or substitute smth wrong.
    if project['name'] == 'deprecated-project':
        pytest.skip('Deprecated project structure test')

    dockerfile_with_env = os.path.join(tmpdir, 'Dockerfile')
    with open(dockerfile_with_env, 'w') as f:
        f.write('''
            FROM centos:8
            # comment this string to use cached image
            # ENV TEST_VARIABLE=${TEST_VARIABLE}
        ''')

    cmd = [
        os.path.join(basepath, "cartridge"),
        "pack", "docker",
        "--from", dockerfile_with_env,
        project['path'],
    ]
    process = subprocess.run(cmd, cwd=module_tmpdir)
    assert process.returncode == 0


def test_docker_e2e(project, docker_image, tmpdir, docker_client):
    image_name = docker_image['name']
    environment = [
        'TARANTOOL_INSTANCE_NAME=instance-1',
        'TARANTOOL_ADVERTISE_URI=3302',
        'TARANTOOL_CLUSTER_COOKIE=secret',
        'TARANTOOL_HTTP_PORT=8082',
    ]

    container = docker_client.containers.run(
        image_name,
        environment=environment,
        ports={'8082': '8082'},
        name='{}-instance-1'.format(project['name']),
        detach=True,
        remove=True
    )

    assert container.status == 'created'
    assert wait_for_container_start(container)

    container_logs = container.logs().decode('utf-8')
    m = re.search(r'Auto-detected IP to be "(\d+\.\d+\.\d+\.\d+)', container_logs)
    assert m is not None
    ip = m.groups()[0]

    admin_api_url = 'http://localhost:8082/admin/api'

    # join instance
    query = '''
        mutation {{
        j1: join_server(
            uri:"{}:3302",
            roles: ["vshard-router", "app.roles.custom"]
            instance_uuid: "aaaaaaaa-aaaa-4000-b000-000000000001"
            replicaset_uuid: "aaaaaaaa-0000-4000-b000-000000000000"
        )
    }}
    '''.format(ip)

    r = requests.post(admin_api_url, json={'query': query})
    assert r.status_code == 200
    resp = r.json()
    assert 'data' in resp
    assert 'j1' in resp['data']
    assert resp['data']['j1'] is True

    # check status and alias
    query = '''
        query {
        instance: cluster {
            self {
                alias
            }
        }
        replicaset: replicasets(uuid: "aaaaaaaa-0000-4000-b000-000000000000") {
            status
        }
    }
    '''

    r = requests.post(admin_api_url, json={'query': query})
    assert r.status_code == 200
    resp = r.json()
    assert 'data' in resp
    assert 'replicaset' in resp['data'] and 'instance' in resp['data']
    assert resp['data']['replicaset'][0]['status'] == 'healthy'
    assert resp['data']['instance']['self']['alias'] == 'instance-1'

    # restart instance
    container.restart()
    wait_for_container_start(container)

    # check instance restarted
    r = requests.post(admin_api_url, json={'query': query})
    assert r.status_code == 200
    resp = r.json()
    assert 'data' in resp
    assert 'replicaset' in resp['data'] and 'instance' in resp['data']
    assert resp['data']['replicaset'][0]['status'] == 'healthy'
    assert resp['data']['instance']['self']['alias'] == 'instance-1'

    container.stop()
