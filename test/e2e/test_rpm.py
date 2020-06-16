import pytest
import platform
import subprocess
import os
import shutil

from utils import Archive, find_archive
from utils import tarantool_repo_version, tarantool_enterprise_is_used
from utils import build_image
from utils import delete_image
from utils import check_systemd_service
from utils import ProjectContainer


# ########
# Fixtures
# ########
@pytest.fixture(scope="function")
def rpm_archive_with_cartridge(cartridge_cmd, tmpdir, original_project_with_cartridge, request):
    project = original_project_with_cartridge

    cmd = [
        cartridge_cmd,
        "pack", "rpm",
        project.path
    ]

    if platform.system() == 'Darwin':
        cmd.append("--use-docker")

    process = subprocess.run(cmd, cwd=tmpdir)
    assert process.returncode == 0, \
        "Error during creating of rpm archive with project"

    filepath = find_archive(tmpdir, project.name, 'rpm')
    assert filepath is not None, "RPM archive isn't found in work directory"

    return Archive(filepath=filepath, project=project)


@pytest.fixture(scope="function")
def container_with_installed_rpm(docker_client, rpm_archive_with_cartridge,
                                 request, tmpdir):
    project = rpm_archive_with_cartridge.project

    # build image with installed RPM
    build_path = os.path.join(tmpdir, 'build_image')
    os.makedirs(build_path)

    shutil.copy(rpm_archive_with_cartridge.filepath, build_path)

    dockerfile_layers = ["FROM centos:7"]
    if not tarantool_enterprise_is_used():
        dockerfile_layers.append('''RUN curl -s \
            https://packagecloud.io/install/repositories/tarantool/{}/script.rpm.sh | bash
        '''.format(tarantool_repo_version()))

    dockerfile_layers.append('''
        COPY {rpm_filename} /opt
        RUN yum install -y /opt/{rpm_filename}
    '''.format(rpm_filename=os.path.basename(rpm_archive_with_cartridge.filepath)))

    with open(os.path.join(build_path, 'Dockerfile'), 'w') as f:
        f.write('\n'.join(dockerfile_layers))

    image_name = '%s-test-rpm' % project.name
    build_image(build_path, image_name)

    request.addfinalizer(lambda: delete_image(docker_client, image_name))

    # create container
    http_port = '8183'

    container = docker_client.containers.create(
        image_name,
        command='/sbin/init',
        ports={http_port: http_port},
        name='%s-test-rpm' % project.name,
        detach=True,
        privileged=True,
        volumes=['/sys/fs/cgroup:/sys/fs/cgroup:ro'],
    )

    request.addfinalizer(lambda: container.remove(force=True))

    return ProjectContainer(project=project, container=container, http_port=http_port)


# #####
# Tests
# #####
def test_rpm(container_with_installed_rpm, tmpdir):
    container = container_with_installed_rpm.container
    project = container_with_installed_rpm.project
    http_port = container_with_installed_rpm.http_port

    container.start()
    check_systemd_service(container, project, http_port, tmpdir)
    container.stop()
