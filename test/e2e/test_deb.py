import pytest
import subprocess
import os
import shutil

from utils import Archive, find_archive
from utils import tarantool_short_version, tarantool_enterprise_is_used
from utils import build_image
from utils import delete_image
from utils import check_systemd_service
from utils import ProjectContainer


# ########
# Fixtures
# ########
@pytest.fixture(scope="function")
def deb_archive_with_cartridge(cartridge_cmd, tmpdir, project_with_cartridge, request):
    project = project_with_cartridge

    cmd = [
        cartridge_cmd,
        "pack", "deb",
        project.path,
        "--use-docker",
    ]

    process = subprocess.run(cmd, cwd=tmpdir)
    assert process.returncode == 0, \
        "Error during creating of deb archive with project"

    filepath = find_archive(tmpdir, project.name, 'deb')
    assert filepath is not None, "DEB archive isn't found in work directory"

    return Archive(filepath=filepath, project=project)


@pytest.fixture(scope="function")
def container_with_installed_deb(docker_client, deb_archive_with_cartridge,
                                 request, tmpdir):
    project = deb_archive_with_cartridge.project

    # build image with installed DEB
    build_path = os.path.join(tmpdir, 'build_image')
    os.makedirs(build_path)

    shutil.copy(deb_archive_with_cartridge.filepath, build_path)

    dockerfile_layers = ["FROM jrei/systemd-ubuntu"]
    if not tarantool_enterprise_is_used():
        dockerfile_layers.append('''RUN apt-get update && apt-get install -y curl \
            && DEBIAN_FRONTEND="noninteractive" apt-get -y install tzdata \
            && curl -L https://tarantool.io/installer.sh | VER={} bash
        '''.format(tarantool_short_version()))

    dockerfile_layers.append('''
        COPY {deb_filename} /opt
        RUN apt-get install -y /opt/{deb_filename}
    '''.format(deb_filename=os.path.basename(deb_archive_with_cartridge.filepath)))

    with open(os.path.join(build_path, 'Dockerfile'), 'w') as f:
        f.write('\n'.join(dockerfile_layers))

    image_name = '%s-test-deb' % project.name
    build_image(build_path, image_name)

    request.addfinalizer(lambda: delete_image(docker_client, image_name))

    # create container
    http_port = '8183'

    container = docker_client.containers.create(
        image_name,
        command='/lib/systemd/systemd',
        ports={http_port: http_port},
        name='%s-test-deb' % project.name,
        detach=True,
        privileged=True,
        volumes=['/sys/fs/cgroup:/sys/fs/cgroup:ro'],
    )

    request.addfinalizer(lambda: container.remove(force=True))

    return ProjectContainer(project=project, container=container, http_port=http_port)


# #####
# Tests
# #####
def test_deb(container_with_installed_deb, tmpdir):
    container = container_with_installed_deb.container
    project = container_with_installed_deb.project
    http_port = container_with_installed_deb.http_port

    container.start()
    check_systemd_service(container, project, http_port, tmpdir)
    container.stop()
