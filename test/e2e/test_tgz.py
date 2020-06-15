import pytest
import os
import subprocess
import gzip
import platform

from utils import tarantool_enterprise_is_used
from utils import Archive, find_archive
from utils import InstanceContainer, examine_application_instance_container
from utils import tarantool_repo_version
from utils import delete_image


# ########
# Fixtures
# ########
@pytest.fixture(scope="function")
def tgz_archive_with_cartridge(cartridge_cmd, tmpdir, original_project_with_cartridge, request):
    project = original_project_with_cartridge

    cmd = [
        cartridge_cmd,
        "pack", "tgz",
        project.path
    ]

    if platform.system() == 'Darwin':
        cmd.append("--use-docker")

    process = subprocess.run(cmd, cwd=tmpdir)
    assert process.returncode == 0, \
        "Error during creating of tgz archive with project"

    filepath = find_archive(tmpdir, project.name, 'tar.gz')
    assert filepath is not None, "TGZ archive isn't found in work directory"

    return Archive(filepath=filepath, project=project)


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
def instance_container_with_unpacked_tgz(docker_client, image_name_for_tests,
                                         tmpdir, tgz_archive_with_cartridge, request):
    project = tgz_archive_with_cartridge.project

    instance_name = 'instance-1'
    http_port = '8183'
    advertise_port = '3302'

    environment = [
        'TARANTOOL_APP_NAME=%s' % project.name,
        'TARANTOOL_INSTANCE_NAME=%s' % instance_name,
        'TARANTOOL_ADVERTISE_URI=%s' % advertise_port,
        'TARANTOOL_HTTP_PORT=%s' % http_port,
    ]

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


# #####
# Tests
# #####
def test_tgz(instance_container_with_unpacked_tgz):
    container = instance_container_with_unpacked_tgz.container
    container.start()

    assert container.status == 'created'
    examine_application_instance_container(instance_container_with_unpacked_tgz)

    container.stop()
