import pytest
import os
import subprocess
import gzip

from utils import tarantool_enterprise_is_used
from utils import Archive, find_archive
from utils import InstanceContainer, examine_application_instance_container
from utils import tarantool_short_version
from utils import build_image
from utils import delete_image


# ########
# Fixtures
# ########
@pytest.fixture(scope="function")
def tgz_archive_with_cartridge(cartridge_cmd, tmpdir, project_with_cartridge, request):
    project = project_with_cartridge

    cmd = [
        cartridge_cmd,
        "pack", "tgz",
        project.path,
        "--use-docker",
    ]

    process = subprocess.run(cmd, cwd=tmpdir)
    assert process.returncode == 0, \
        "Error during creating of tgz archive with project"

    filepath = find_archive(tmpdir, project.name, 'tar.gz')
    assert filepath is not None, "TGZ archive isn't found in work directory"

    return Archive(filepath=filepath, project=project)


@pytest.fixture(scope="function")
def instance_container_with_unpacked_tgz(docker_client, tmpdir, tgz_archive_with_cartridge, request):
    project = tgz_archive_with_cartridge.project

    # build image with installed Tarantool
    build_path = os.path.join(tmpdir, 'build_image')
    os.makedirs(build_path)

    dockerfile_layers = ["FROM centos:7"]
    if not tarantool_enterprise_is_used():
        dockerfile_layers.append('''RUN curl -L \
            https://tarantool.io/installer.sh | VER={} bash
        '''.format(tarantool_short_version()))

    with open(os.path.join(build_path, 'Dockerfile'), 'w') as f:
        f.write('\n'.join(dockerfile_layers))

    image_name = '%s-test-rpm' % project.name
    build_image(build_path, image_name)

    request.addfinalizer(lambda: delete_image(docker_client, image_name))

    # create container
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
        image_name,
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
