import os
import re
import shutil
import subprocess
import tarfile
import time

import pytest
from project import INIT_PRINT_ENV_FILEPATH, replace_project_file
from utils import (Image, assert_distribution_dir_contents, assert_filemodes,
                   delete_image, find_image, mark_only_opensource,
                   recursive_listdir, run_command_and_get_output,
                   tarantool_enterprise_is_used, tarantool_version,
                   wait_for_container_start)


# #######
# Helpers
# #######
def run_command_on_image(docker_client, image_name, command):
    command = '/bin/bash -c "{}"'.format(command.replace('"', '\\"'))
    output = docker_client.containers.run(
        image_name,
        command,
        remove=True
    )
    return output.decode("utf-8").strip()


def add_runtime_requirements_file(project):
    # add a file with runtime requirements
    runtime_requirements_filename = 'runtime-requirements.txt'
    runtime_requirements_filepath = os.path.join(project.path, runtime_requirements_filename)
    with open(runtime_requirements_filepath, 'w') as f:
        f.write('''
            # runtime requirements
        ''')

    # update distribution files
    project.distribution_files.add(runtime_requirements_filename)

    # copy this file to image in runtime nase dockerfile
    runtime_dockerfile_path = os.path.join(project.path, 'Dockerfile.cartridge')
    image_runtime_requirements_filepath = os.path.join('/tmp', runtime_requirements_filename)
    with open(runtime_dockerfile_path, 'w') as f:
        f.write('''
            FROM centos:8
            COPY {} {}
        '''.format(runtime_requirements_filename, image_runtime_requirements_filepath))

    project.image_runtime_requirements_filepath = image_runtime_requirements_filepath


# ########
# Fixtures
# ########
@pytest.fixture(scope="session")
def docker_image(cartridge_cmd, session_tmpdir, session_light_project, request, docker_client):
    project = session_light_project
    add_runtime_requirements_file(project)

    cmd = [cartridge_cmd, "pack", "docker", project.path]
    process = subprocess.run(cmd, cwd=session_tmpdir)
    assert process.returncode == 0, \
        "Error during creating of docker image"

    image_name = find_image(docker_client, project.name)
    assert image_name is not None, "Docker image isn't found"

    request.addfinalizer(lambda: delete_image(docker_client, image_name))

    image = Image(image_name, project)
    return image


@pytest.fixture(scope="function")
def docker_image_print_environment(cartridge_cmd, tmpdir, project_without_dependencies, request, docker_client):
    project = project_without_dependencies
    replace_project_file(project, 'init.lua', INIT_PRINT_ENV_FILEPATH)

    cmd = [
        cartridge_cmd,
        "pack", "docker",
        "--tag", project.name,
        project.path,
    ]

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
def test_pack(docker_image, tmpdir, docker_client):
    project = docker_image.project
    image_name = docker_image.name

    container = docker_client.containers.create(docker_image.name)
    container_distribution_dir = '/usr/share/tarantool/{}'.format(project.name)

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

    distribution_dir_contents = recursive_listdir(os.path.join(tmpdir, 'usr/share/tarantool/', project.name))

    # The runtime image is built using Dockerfile.<random-string> in the
    #   distribution directory
    # This dockerfile name should be added to project distribution files set
    #   to correctly check distribution directory contents
    for f in distribution_dir_contents:
        if f.startswith('Dockerfile') and f not in ['Dockerfile.build.cartridge', 'Dockerfile.cartridge']:
            project.distribution_files.add(f)
            break

    assert_distribution_dir_contents(
        dir_contents=recursive_listdir(os.path.join(tmpdir, 'usr/share/tarantool/', project.name)),
        project=project,
    )

    assert_filemodes(project, tmpdir)
    container.remove()

    if project.image_runtime_requirements_filepath is not None:
        command = 'ls {}'.format(project.image_runtime_requirements_filepath)
        output = run_command_on_image(docker_client, image_name, command)
        assert output == project.image_runtime_requirements_filepath

    if not tarantool_enterprise_is_used():
        # check if tarantool was installed
        command = 'yum list installed 2>/dev/null | grep tarantool'
        output = run_command_on_image(docker_client, image_name, command)

        packages_list = output.split('\n')
        assert any(['tarantool' in package for package in packages_list])

        # check tarantool version
        command = 'tarantool --version'
        output = run_command_on_image(docker_client, image_name, command)

        m = re.search(r'Tarantool\s+(\d+.\d+)', output)
        assert m is not None
        installed_version = m.group(1)

        m = re.search(r'(\d+.\d+)', tarantool_version())
        assert m is not None
        expected_version = m.group(1)

        assert installed_version == expected_version


def test_custom_base_runtime_dockerfile(
    cartridge_cmd, project_without_dependencies, module_tmpdir, custom_base_image, tmpdir
):
    custom_base_dockerfile_path = os.path.join(tmpdir, 'Dockerfile')
    with open(custom_base_dockerfile_path, 'w') as f:
        f.write(f"""
            # Non standard base image
            FROM {custom_base_image["image_name"]}
        """)

    cmd = [
        cartridge_cmd,
        "pack", "docker",
        "--from", custom_base_dockerfile_path,
        project_without_dependencies.path,
    ]

    rc, output = run_command_and_get_output(cmd, cwd=module_tmpdir)
    assert rc == 0
    assert 'Image based on centos:8 is expected to be used' in output


def test_project_witout_runtime_dockerfile(cartridge_cmd, project_without_dependencies, tmpdir):
    project = project_without_dependencies

    os.remove(os.path.join(project.path, 'Dockerfile.cartridge'))

    cmd = [
        cartridge_cmd,
        "pack", "docker",
        project.path,
    ]

    process = subprocess.run(cmd, cwd=tmpdir)
    assert process.returncode == 0


def test_image_tag_without_git(cartridge_cmd, project_without_dependencies, tmpdir):
    project = project_without_dependencies

    # remove .git directory
    shutil.rmtree(os.path.join(project.path, '.git'))

    # pass image tags
    tag1 = 'my-cute-tag:xxx'
    tag2 = 'your-cute-tag:yyy'

    expected_image_tags = '{}, {}'.format(tag1, tag2)

    cmd = [
        cartridge_cmd,
        "pack", 'docker',
        "--tag", tag1,
        "--tag", tag2,
        project.path,
    ]
    rc, output = run_command_and_get_output(cmd, cwd=tmpdir)
    assert rc == 0
    assert 'Created result image with tags {}'.format(expected_image_tags) in output


def test_customized_environment_variables(docker_image_print_environment, docker_client, request):
    project = docker_image_print_environment.project

    instance_name = 'instance-1'
    http_port = '8182'
    advertise_port = '3302'

    # Custom TARANTOOL_WORKDIR, TARANTOOL_PID_FILE and TARANTOOL_CONSOLE_SOCK
    workdir = "test_workdir"
    pidfile = "test_pidfile"
    console_sock = "test_console_sock"

    environment = [
        f"TARANTOOL_APP_NAME={project.name}",
        f"TARANTOOL_INSTANCE_NAME={instance_name}",
        f"TARANTOOL_ADVERTISE_URI={advertise_port}",
        f"TARANTOOL_HTTP_PORT={http_port}",
        f"TARANTOOL_WORKDIR={workdir}",
        f"TARANTOOL_PID_FILE={pidfile}",
        f"TARANTOOL_CONSOLE_SOCK={console_sock}",
    ]

    container = docker_client.containers.run(
        project.name,
        environment=environment,
        ports={http_port: http_port},
        name='{}-{}'.format(project.name, instance_name),
        detach=True,
    )

    request.addfinalizer(lambda: container.remove(force=True))

    assert container.status == 'created'
    container_message = f"{console_sock}\n{workdir}\n{pidfile}\n"

    wait_for_container_start(container, time.time(), message=container_message)


def test_customized_data_and_run_dir(docker_image_print_environment, docker_client, request):
    project = docker_image_print_environment.project

    instance_name = 'instance-1'
    http_port = '8182'
    advertise_port = '3302'

    # Custom CARTRIDGE_RUN_DIR and CARTRIDGE_DATA_DIR
    run_dir = "/var/lib/tarantool/custom_run"
    data_dir = "/var/lib/tarantool/custom_data"

    environment = [
        f"TARANTOOL_APP_NAME={project.name}",
        f"TARANTOOL_INSTANCE_NAME={instance_name}",
        f"TARANTOOL_ADVERTISE_URI={advertise_port}",
        f"TARANTOOL_HTTP_PORT={http_port}",
        f"CARTRIDGE_DATA_DIR={data_dir}",
        f"CARTRIDGE_RUN_DIR={run_dir}"
    ]

    container = docker_client.containers.run(
        project.name,
        environment=environment,
        ports={http_port: http_port},
        name='new-{}-{}'.format(project.name, instance_name),
        detach=True,
    )

    request.addfinalizer(lambda: container.remove(force=True))
    assert container.status == 'created'

    console_sock_path = f"{run_dir}/{project.name}.{instance_name}.control"
    pidfile_path = f"{run_dir}/{project.name}.{instance_name}.pid"
    workdir_path = f"{data_dir}/{project.name}.{instance_name}"
    container_message = f"{console_sock_path}\n{workdir_path}\n{pidfile_path}\n"

    wait_for_container_start(container, time.time(), message=container_message)


def test_tarantool_uid_and_gid(docker_image, docker_client):
    image_name = docker_image.name
    docker_client.containers.create(docker_image.name)

    command = 'whoami'
    output = run_command_on_image(docker_client, image_name, command)
    assert output == 'tarantool'

    command = 'id -u tarantool'
    output = run_command_on_image(docker_client, image_name, command)
    assert output == '1200'

    command = 'id -g tarantool'
    output = run_command_on_image(docker_client, image_name, command)
    assert output == '1200'


@mark_only_opensource
def test_image_specific_tarantool_versions(cartridge_cmd, project_without_dependencies, tmpdir, request, docker_client):
    project = project_without_dependencies

    tarantool_versions = [
        {"input": "2", "expect": "Tarantool 2"},
        {"input": "2.8", "expect": "Tarantool 2.8.4-0-"},
        {"input": "2.8.1", "expect": "Tarantool 2.8.1-0-"},
        {"input": "2.10.0-beta1", "expect": "Tarantool 2.10.0-beta1"},
        {"input": "2.10.0~beta1", "expect": "Tarantool 2.10.0-beta1"}
        ]

    for test_version in tarantool_versions:
        cmd = [cartridge_cmd, "pack", "--tarantool-version", test_version["input"], "docker", project.path]

        process = subprocess.run(cmd, cwd=tmpdir)
        assert process.returncode == 0, \
            "Error during creating of docker image"

        image_name = find_image(docker_client, project.name)
        assert image_name is not None, "Docker image isn't found"

        request.addfinalizer(lambda: delete_image(docker_client, image_name))

        command = 'tarantool --version'
        output = run_command_on_image(docker_client, image_name, command)
        assert test_version["expect"] in output


@mark_only_opensource
def test_image_specific_tarantool_version_from_file(cartridge_cmd, project_without_dependencies, tmpdir,
                                                    request, docker_client):
    project = project_without_dependencies

    tarantool_version = "2.8.2"
    with open(os.path.join(project.path, "tarantool.txt"), "w") as tarantool_version_file:
        tarantool_version_file.write(f"TARANTOOL={tarantool_version}")

    cmd = [cartridge_cmd, "pack", "docker", project.path]

    process = subprocess.run(cmd, cwd=tmpdir)
    assert process.returncode == 0, \
        "Error during creating of docker image"

    image_name = find_image(docker_client, project.name)
    assert image_name is not None, "Docker image isn't found"

    request.addfinalizer(lambda: delete_image(docker_client, image_name))

    command = 'tarantool --version'
    output = run_command_on_image(docker_client, image_name, command)
    assert tarantool_version in output


@mark_only_opensource
def test_tarantool_version_cli_option_validation(cartridge_cmd, project_without_dependencies, tmpdir):
    project = project_without_dependencies

    # sdk-local with tarantool-version
    cmd = [cartridge_cmd, "pack", "--tarantool-version", "2.7.3", "--sdk-local", "docker", project.path]

    rc, output = run_command_and_get_output(cmd, cwd=tmpdir)

    assert rc == 1
    assert "You can specify only one of --tarantool-version,--sdk-path or --sdk-local" in output

    # sdk-path with tarantool-version
    cmd = [cartridge_cmd, "pack", "--tarantool-version", "2.7.3", "--sdk-path",
           tmpdir, "docker", project.path]

    rc, output = run_command_and_get_output(cmd, cwd=tmpdir)

    assert rc == 1
    assert "You can specify only one of --tarantool-version,--sdk-path or --sdk-local" in output

    # --tarantool-version with rpm pack type
    cmd = [cartridge_cmd, "pack", "--tarantool-version", "2.7.3", "rpm", project.path]

    rc, output = run_command_and_get_output(cmd, cwd=tmpdir)

    assert rc == 1
    assert "--tarantool-version option can be used only with docker type" in output

    # --tarantool-version with deb pack type
    cmd = [cartridge_cmd, "pack", "--tarantool-version", "2.7.3", "deb", project.path]

    rc, output = run_command_and_get_output(cmd, cwd=tmpdir)

    assert rc == 1
    assert "--tarantool-version option can be used only with docker type" in output

    # entrypoint tarantool version
    cmd = [cartridge_cmd, "pack", "--tarantool-version", "2.7.3-entrypoint", "docker", project.path]

    rc, output = run_command_and_get_output(cmd, cwd=tmpdir)

    assert rc == 1
    assert "Entrypoint build cannot be used for --tarantool-version" in output
