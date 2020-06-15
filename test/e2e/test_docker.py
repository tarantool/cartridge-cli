import pytest
import subprocess

from utils import Image, find_image, delete_image
from utils import InstanceContainer, examine_application_instance_container


# ########
# Fixtures
# ########
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


def test_docker(docker_image_with_cartridge, tmpdir, docker_client, request):
    image_name = docker_image_with_cartridge.name
    project = docker_image_with_cartridge.project

    instance_name = 'instance-1'
    http_port = '8182'
    advertise_port = '3302'

    environment = [
        'TARANTOOL_APP_NAME=%s' % project.name,
        'TARANTOOL_INSTANCE_NAME=%s' % instance_name,
        'TARANTOOL_ADVERTISE_URI=%s' % advertise_port,
        'TARANTOOL_HTTP_PORT=%s' % http_port,
    ]

    container = docker_client.containers.run(
        image_name,
        environment=environment,
        ports={http_port: http_port},
        name='{}-{}'.format(project.name, instance_name),
        detach=True,
    )

    request.addfinalizer(lambda: container.remove(force=True))

    assert container.status == 'created'
    examine_application_instance_container(InstanceContainer(
        container=container,
        instance_name=instance_name,
        http_port=http_port,
        advertise_port=advertise_port
    ))

    container.stop()
