import os
import shutil
import subprocess

import pytest
from utils import (Archive, ProjectContainer, build_image,
                   check_contains_regular_file, check_systemd_service,
                   delete_image, find_archive, get_tarantool_installer_cmd,
                   run_command_on_container, tarantool_enterprise_is_used)


# ########
# Fixtures
# ########
@pytest.fixture(scope="function")
def rpm_archive_with_cartridge(cartridge_cmd, tmpdir, project_with_cartridge):
    project = project_with_cartridge

    pre_install_filepath = os.path.join(tmpdir, "pre.sh")
    with open(pre_install_filepath, "w") as f:
        f.write("""
                /bin/sh -c 'touch $HOME/hello-bin-sh.txt'
                /bin/touch $HOME/hello-absolute.txt
                """)

    post_install_filepath = os.path.join(tmpdir, "post.sh")
    with open(post_install_filepath, "w") as f:
        f.write("""
                /bin/sh -c 'touch $HOME/bye-bin-sh.txt'
                /bin/touch $HOME/bye-absolute.txt
                """)

    cmd = [
        cartridge_cmd,
        "pack", "rpm",
        "--deps", "unzip>1,unzip<=7",
        "--deps", "wget",
        "--deps", "make>0.1.0",
        "--preinst", pre_install_filepath,
        "--postinst", post_install_filepath,
        project.path,
        "--use-docker",
    ]

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
        installer_cmd = get_tarantool_installer_cmd("yum")
        dockerfile_layers.append(f"RUN {installer_cmd}")
    else:
        dockerfile_layers.append("RUN yum update -y")

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

    run_command_on_container(container, "unzip")
    run_command_on_container(container, "wget --version")
    run_command_on_container(container, "make --version")

    assert check_contains_regular_file(container, '$HOME/hello-bin-sh.txt')
    assert check_contains_regular_file(container, '$HOME/hello-absolute.txt')
    assert check_contains_regular_file(container, '$HOME/bye-bin-sh.txt')
    assert check_contains_regular_file(container, '$HOME/bye-absolute.txt')

    container.stop()
