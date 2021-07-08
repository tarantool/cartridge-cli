import pytest
import subprocess
import os
import shutil

from utils import Archive, find_archive
from utils import tarantool_short_version, tarantool_enterprise_is_used
from utils import build_image
from utils import delete_image
from utils import check_systemd_service
from utils import ProjectContainer, run_command_on_container
from utils import check_contains_file

from project import patch_init_to_net_msg_max


# ########
# Fixtures
# ########
@pytest.fixture(scope="function")
def deb_archive_with_cartridge(cartridge_cmd, tmpdir, project_with_cartridge):
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

    deps_filepath = os.path.join(tmpdir, "deps.txt")
    with open(deps_filepath, "w") as f:
        f.write("unzip>1,<=7\n" +
                "stress\n" +
                "neofetch < 25")

    net_msg_max = 1024

    systemd_unit_params = os.path.join(tmpdir, "systemd-unit-params.yml")
    with open(systemd_unit_params, "w") as f:
        f.write(f"""
                instance-env:
                    net_msg_max: {net_msg_max}
                """)

    patch_init_to_net_msg_max(project, net_msg_max)

    cmd = [
        cartridge_cmd,
        "pack", "deb",
        "--deps-file", deps_filepath,
        "--preinst", pre_install_filepath,
        "--postinst", post_install_filepath,
        "--unit-params-file", systemd_unit_params,
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
    else:
        dockerfile_layers.append("RUN apt-get update")

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

    run_command_on_container(container, "unzip")
    run_command_on_container(container, "stress")
    run_command_on_container(container, "neofetch")

    assert check_contains_file(container, '$HOME/hello-bin-sh.txt')
    assert check_contains_file(container, '$HOME/hello-absolute.txt')
    assert check_contains_file(container, '$HOME/bye-bin-sh.txt')
    assert check_contains_file(container, '$HOME/bye-absolute.txt')

    container.stop()
