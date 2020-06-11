import pytest
import os
import tarfile
import subprocess
import re
import time
import gzip
import tenacity

from utils import tarantool_enterprise_is_used
from utils import Image, find_image, delete_image
from utils import Archive, find_archive
from utils import tarantool_repo_version
from utils import create_replicaset, wait_for_replicaset_is_healthy


# #######################
# Class InstanceContainer
# #######################
class InstanceContainer:
    def __init__(self, container, instance_name, http_port, advertise_port):
        self.container = container
        self.instance_name = instance_name
        self.http_port = http_port
        self.advertise_port = advertise_port


# ########
# Helpers
# ########
@tenacity.retry(stop=tenacity.stop_after_delay(10))
def wait_for_container_start(container, time_start):
    container_logs = container.logs(since=int(time_start)).decode('utf-8')
    assert 'entering the event loop' in container_logs


def examine_application_instance_container(container, instance_name, http_port, advertise_port):
    wait_for_container_start(container, time.time())

    container_logs = container.logs().decode('utf-8')
    m = re.search(r'Auto-detected IP to be "(\d+\.\d+\.\d+\.\d+)', container_logs)
    assert m is not None
    ip = m.groups()[0]

    admin_api_url = 'http://localhost:{}/admin/api'.format(http_port)
    advertise_uri = '{}:{}'.format(ip, advertise_port)
    roles = ["vshard-router", "app.roles.custom"]

    replicaset_uuid = create_replicaset(admin_api_url, [advertise_uri], roles)
    wait_for_replicaset_is_healthy(admin_api_url, replicaset_uuid)

    # restart instance
    container.restart()
    wait_for_container_start(container, time.time())

    # check instance restarted
    wait_for_replicaset_is_healthy(admin_api_url, replicaset_uuid)


# ########
# Fixtures
# ########
@pytest.fixture(scope="function")
def image_name_for_tests(docker_client, tmpdir, request):
    if tarantool_enterprise_is_used():
        docker_client.images.pull('centos', '8')
        return 'centos:8'

    build_path = os.path.join(tmpdir, 'build_image')
    os.makedirs(build_path)

    test_image_dockerfile_path = os.path.join(build_path, 'Dockerfile')
    with open(test_image_dockerfile_path, 'w') as f:
        f.write('''
            FROM centos:8
            RUN curl -s \
                https://packagecloud.io/install/repositories/tarantool/{}/script.rpm.sh | bash \
                && yum -y install tarantool tarantool-devel
        '''.format(tarantool_repo_version()))

    IMAGE_NAME = 'test-image'
    docker_client.images.build(
        path=build_path,
        forcerm=True,
        tag=IMAGE_NAME,
    )

    request.addfinalizer(lambda: delete_image(docker_client, IMAGE_NAME))

    return IMAGE_NAME


@pytest.fixture(scope="function")
def tgz_archive_with_cartridge(cartridge_cmd, tmpdir, original_project_with_cartridge, request):
    project = original_project_with_cartridge

    cmd = [
        cartridge_cmd,
        "pack", "tgz",
        "--use-docker",
        project.path
    ]

    process = subprocess.run(cmd, cwd=tmpdir)
    assert process.returncode == 0, \
        "Error during creating of tgz archive with project"

    filepath = find_archive(tmpdir, project.name, 'tar.gz')
    assert filepath is not None, "TGZ archive isn't found in work directory"

    return Archive(filepath=filepath, project=project)


@pytest.fixture(scope="function")
def instance_container_with_unpacked_tgz(docker_client, image_name_for_tests,
                                         tmpdir, tgz_archive_with_cartridge, request):
    project = tgz_archive_with_cartridge.project

    instance_name = 'instance-1'
    http_port = '8183'
    advertise_port = '3302'

    environment = [
        'TARANTOOL_INSTANCE_NAME={}'.format(instance_name),
        'TARANTOOL_ADVERTISE_URI={}'.format(advertise_port),
        'TARANTOOL_CLUSTER_COOKIE=secret',
        'TARANTOOL_HTTP_PORT={}'.format(http_port),
    ]

    distribution_dir = os.path.join(tmpdir, 'distribution_dir')

    with tarfile.open(name=tgz_archive_with_cartridge.filepath) as tgz_arch:
        tgz_arch.extractall(path=distribution_dir)

    container_proj_path = os.path.join('/opt', project.name)
    init_script_path = os.path.join(container_proj_path, 'init.lua')
    tarantool_executable = \
        os.path.join(container_proj_path, 'tarantool') \
        if tarantool_enterprise_is_used() \
        else 'tarantool'

    cmd = [tarantool_executable, init_script_path]

    container = docker_client.containers.create(
        image_name_for_tests,
        cmd,
        environment=environment,
        ports={http_port: http_port},
        name='{}-{}'.format(project.name, instance_name),
        detach=True,
    )

    with gzip.open(tgz_archive_with_cartridge.filepath, 'rb') as f:
        container.put_archive('/opt', f.read())

    request.addfinalizer(lambda: container.remove(force=True))

    return InstanceContainer(
        container=container,
        instance_name=instance_name,
        http_port=http_port,
        advertise_port=advertise_port
    )


@pytest.fixture(scope="function")
def docker_image_with_cartridge(cartridge_cmd, tmpdir, original_project_with_cartridge, request, docker_client):
    project = original_project_with_cartridge

    cmd = [cartridge_cmd, "pack", "docker", project.path]
    process = subprocess.run(cmd, cwd=tmpdir)
    assert process.returncode == 0, \
        "Error during creating of docker image"

    image_name = find_image(docker_client, project.name)
    assert image_name is not None, "Docker image isn't found"

    request.addfinalizer(lambda: delete_image(docker_client, image_name))

    image = Image(image_name, project)
    return image


# #####
# Tests
# #####
def test_tgz(instance_container_with_unpacked_tgz, tmpdir, docker_client, request):
    container = instance_container_with_unpacked_tgz.container
    container.start()

    assert container.status == 'created'
    examine_application_instance_container(
        container,
        instance_container_with_unpacked_tgz.instance_name,
        instance_container_with_unpacked_tgz.http_port,
        instance_container_with_unpacked_tgz.advertise_port
    )

    container.stop()


def test_docker(docker_image_with_cartridge, tmpdir, docker_client):
    image_name = docker_image_with_cartridge.name
    project = docker_image_with_cartridge.project

    instance_name = 'instance-1'
    http_port = '8182'
    advertise_port = '3302'

    environment = [
        'TARANTOOL_INSTANCE_NAME={}'.format(instance_name),
        'TARANTOOL_ADVERTISE_URI={}'.format(advertise_port),
        'TARANTOOL_CLUSTER_COOKIE=secret',
        'TARANTOOL_HTTP_PORT={}'.format(http_port),
    ]

    container = docker_client.containers.run(
        image_name,
        environment=environment,
        ports={http_port: http_port},
        name='{}-{}'.format(project.name, instance_name),
        detach=True,
    )

    assert container.status == 'created'
    examine_application_instance_container(container, instance_name, http_port, advertise_port)

    container.stop()
