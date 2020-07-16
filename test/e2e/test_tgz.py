import pytest
import os
import subprocess
import platform
import shutil

from utils import tarantool_enterprise_is_used
from utils import Archive, find_archive
from utils import ProjectContainer
from utils import tarantool_short_version
from utils import build_image
from utils import run_command_on_container
from utils import delete_image
from utils import write_instance_conf
from utils import check_global_running_instance
from utils import check_contains_file


# ########
# Fixtures
# ########
@pytest.fixture(scope="function")
def tgz_archive_with_cartridge(cartridge_cmd, tmpdir, project_with_cartridge, request):
    project = project_with_cartridge

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
def container_with_unpacked_tgz(docker_client, tmpdir, tgz_archive_with_cartridge,
                                cartridge_cmd_for_linux, request):
    project = tgz_archive_with_cartridge.project

    # build image with installed Tarantool
    build_path = os.path.join(tmpdir, 'build_image')
    os.makedirs(build_path)

    # copy tgz archive with an application
    shutil.copy(tgz_archive_with_cartridge.filepath, build_path)

    # copy cartridge cli for linux
    shutil.copyfile(cartridge_cmd_for_linux, os.path.join(build_path, "cartridge"))

    tgz_filename = os.path.basename(tgz_archive_with_cartridge.filepath)

    dockerfile_layers = ["FROM centos:7"]
    if not tarantool_enterprise_is_used():
        dockerfile_layers.append('''RUN curl -L \
            https://tarantool.io/installer.sh | VER={} bash
        '''.format(tarantool_short_version()))

    dockerfile_layers.extend([
        "COPY cartridge /opt",
        "RUN chmod +x /opt/cartridge",
        "ENV PATH=/opt:${PATH}",
        "COPY %s /tmp" % tgz_filename,
        "RUN mkdir -p /usr/share/tarantool && tar -zxf /tmp/%s -C /usr/share/tarantool" % tgz_filename,
    ])

    with open(os.path.join(build_path, 'Dockerfile'), 'w') as f:
        f.write('\n'.join(dockerfile_layers))

    image_name = '%s-test-tgz' % project.name
    build_image(build_path, image_name)

    request.addfinalizer(lambda: delete_image(docker_client, image_name))

    # create container
    http_port = '8183'

    container = docker_client.containers.create(
        image_name,
        command='/sbin/init',
        ports={http_port: http_port},
        name='%s-test-tgz' % project.name,
        detach=True,
    )

    request.addfinalizer(lambda: container.remove(force=True))

    return ProjectContainer(project=project, container=container, http_port=http_port)


# #####
# Tests
# #####
def test_tgz(container_with_unpacked_tgz, tmpdir):
    project = container_with_unpacked_tgz.project

    container = container_with_unpacked_tgz.container
    container.start()

    assert container.status == 'created'

    instance_name = 'instance-1'
    advertise_uri = 'localhost:3303'
    http_port = container_with_unpacked_tgz.http_port

    write_instance_conf(container, tmpdir, project, instance_name, http_port, advertise_uri)

    run_command_on_container(container, "cartridge start -g -d --name %s %s" % (project.name, instance_name))
    output = run_command_on_container(container, "cartridge status -g --name %s %s" % (project.name, instance_name))
    assert "RUNNING" in output

    check_global_running_instance(container, project, instance_name, http_port, advertise_uri)
    check_contains_file(container, '/var/log/tarantool/%s.%s.log' % (project.name, instance_name))

    container.stop()
