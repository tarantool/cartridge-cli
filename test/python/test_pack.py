#!/usr/bin/python3

import os
import pytest
import subprocess
import tarfile
import re
import shutil
import stat

from utils import tarantool_enterprise_is_used
from utils import Archive, find_archive
from utils import recursive_listdir
from utils import assert_distribution_dir_contents
from utils import assert_filemodes
from utils import assert_files_mode_and_owner_rpm
from utils import validate_version_file
from utils import check_package_files
from utils import assert_tarantool_dependency_deb
from utils import assert_tarantool_dependency_rpm
from utils import run_command_and_get_output


# ########
# Fixtures
# ########
@pytest.fixture(scope="function", params=['local', 'docker'])
def tgz_archive(cartridge_cmd, tmpdir, light_project, request):
    project = light_project

    cmd = [cartridge_cmd, "pack", "tgz", project.path]

    if request.param == 'docker':
        if project.deprecated_flow_is_used:
            pytest.skip()

        cmd.append('--use-docker')

    process = subprocess.run(cmd, cwd=tmpdir)
    assert process.returncode == 0, \
        "Error during creating of tgz archive with project"

    filepath = find_archive(tmpdir, project.name, 'tar.gz')
    assert filepath is not None, "TGZ archive isn't found in work directory"

    return Archive(filepath=filepath, project=project)


@pytest.fixture(scope="function", params=['local', 'docker'])
def rpm_archive(cartridge_cmd, tmpdir, light_project, request):
    project = light_project

    cmd = [cartridge_cmd, "pack", "rpm", project.path]

    if request.param == 'docker':
        if project.deprecated_flow_is_used:
            pytest.skip()

        cmd.append('--use-docker')

    process = subprocess.run(cmd, cwd=tmpdir)
    assert process.returncode == 0, \
        "Error during creating of rpm archive with project"

    filepath = find_archive(tmpdir, project.name, 'rpm')
    assert filepath is not None, "RPM archive isn't found in work directory"

    return Archive(filepath=filepath, project=project)


@pytest.fixture(scope="function", params=['local', 'docker'])
def deb_archive(cartridge_cmd, tmpdir, light_project, request):
    project = light_project

    cmd = [cartridge_cmd, "pack", "deb", project.path]

    if request.param == 'docker':
        if project.deprecated_flow_is_used:
            pytest.skip()

        cmd.append('--use-docker')

    process = subprocess.run(cmd, cwd=tmpdir)
    assert process.returncode == 0, \
        "Error during creating of deb archive with project"

    filepath = find_archive(tmpdir, project.name, 'deb')
    assert filepath is not None, "DEB archive isn't found in work directory"

    return Archive(filepath=filepath, project=project)


@pytest.fixture(scope="function")
def rpm_archive_with_custom_units(cartridge_cmd, tmpdir, light_project):
    project = light_project

    unit_template = '''
[Unit]
Description=Tarantool service: ${app_name}
SIMPLE_UNIT_TEMPLATE
[Service]
Type=simple
ExecStart=${bindir}/tarantool ${app_dir}/init.lua

Environment=TARANTOOL_WORK_DIR=${workdir}
Environment=TARANTOOL_CONSOLE_SOCK=/var/run/tarantool/${app_name}.control
Environment=TARANTOOL_PID_FILE=/var/run/tarantool/${app_name}.pid
Environment=TARANTOOL_INSTANCE_NAME=${app_name}

[Install]
WantedBy=multi-user.target
Alias=${app_name}
    '''

    instantiated_unit_template = '''
[Unit]
Description=Tarantool service: ${app_name} %i
INSTANTIATED_UNIT_TEMPLATE

[Service]
Type=simple
ExecStartPre=mkdir -p ${workdir}.%i
ExecStart=${bindir}/tarantool ${app_dir}/init.lua

Environment=TARANTOOL_WORK_DIR=${workdir}.%i
Environment=TARANTOOL_CONSOLE_SOCK=/var/run/tarantool/${app_name}.%i.control
Environment=TARANTOOL_PID_FILE=/var/run/tarantool/${app_name}.%i.pid
Environment=TARANTOOL_INSTANCE_NAME=${app_name}@%i

[Install]
WantedBy=multi-user.target
Alias=${app_name}
    '''
    unit_template_filepath = os.path.join(tmpdir, "unit_template.tmpl")
    with open(unit_template_filepath, 'w') as f:
        f.write(unit_template)

    inst_unit_template_filepath = os.path.join(tmpdir, "instantiated_unit_template.tmpl")
    with open(inst_unit_template_filepath, 'w') as f:
        f.write(instantiated_unit_template)

    process = subprocess.run([
            cartridge_cmd, "pack", "rpm",
            "--unit-template", "unit_template.tmpl",
            "--instantiated-unit-template", "instantiated_unit_template.tmpl",
            project.path
        ],
        cwd=tmpdir
    )
    assert process.returncode == 0, \
        "Error during creating of rpm archive with project"

    filepath = find_archive(tmpdir, project.name, 'rpm')
    assert filepath is not None, "RPM archive isn't found in work directory"

    return Archive(filepath=filepath, project=project)


# ########
# Helpers
# ########
def extract_rpm(rpm_archive_path, extract_dir):
    ps = subprocess.Popen(
        ['rpm2cpio', rpm_archive_path],
        stdout=subprocess.PIPE
    )
    subprocess.check_output(['cpio', '-idmv'], stdin=ps.stdout, cwd=extract_dir)
    ps.wait()
    assert ps.returncode == 0, "Error during extracting files from rpm archive"


def extract_deb(deb_archive_path, extract_dir):
    process = subprocess.run([
            'ar', 'x', deb_archive_path
        ],
        cwd=extract_dir
    )
    assert process.returncode == 0, 'Error during unpacking of deb archive'


# #####
# Tests
# #####
def test_tgz(tgz_archive, tmpdir):
    project = tgz_archive.project

    # archive files should be extracted to the empty directory
    # to correctly check archive contents
    extract_dir = os.path.join(tmpdir, 'extract')
    os.makedirs(extract_dir)

    with tarfile.open(name=tgz_archive.filepath) as tgz_arch:
        # usr/share/tarantool is added to correctly run assert_filemodes
        distribution_dir = os.path.join(extract_dir, 'usr/share/tarantool', project.name)
        os.makedirs(distribution_dir, exist_ok=True)

        tgz_arch.extractall(path=os.path.join(extract_dir, 'usr/share/tarantool'))
        assert_distribution_dir_contents(
            dir_contents=recursive_listdir(distribution_dir),
            project=project
        )

        validate_version_file(project, distribution_dir)
        assert_filemodes(project, extract_dir)


def test_rpm(rpm_archive, tmpdir):
    project = rpm_archive.project

    # archive files should be extracted to the empty directory
    # to correctly check archive contents
    extract_dir = os.path.join(tmpdir, 'extract')
    os.makedirs(extract_dir)

    extract_rpm(rpm_archive.filepath, extract_dir)

    if not tarantool_enterprise_is_used():
        assert_tarantool_dependency_rpm(rpm_archive.filepath)

    check_package_files(project, extract_dir)
    assert_files_mode_and_owner_rpm(project, rpm_archive.filepath)

    # check rpm signature
    cmd = [
        'rpm', '--checksig', '-v', rpm_archive.filepath
    ]
    process = subprocess.run(cmd)
    assert process.returncode == 0, "RPM signature isn't correct"


def test_deb(deb_archive, tmpdir):
    project = deb_archive.project

    # archive files should be extracted to the empty directory
    # to correctly check archive contents
    extract_dir = os.path.join(tmpdir, 'extract')
    os.makedirs(extract_dir)

    extract_deb(deb_archive.filepath, extract_dir)

    for filename in ['debian-binary', 'control.tar.xz', 'data.tar.xz']:
        assert os.path.exists(os.path.join(extract_dir, filename))

    # check debian-binary
    with open(os.path.join(extract_dir, 'debian-binary')) as debian_binary_file:
        assert debian_binary_file.read() == '2.0\n'

    # check data.tar.xz
    with tarfile.open(name=os.path.join(extract_dir, 'data.tar.xz')) as data_arch:
        data_dir = os.path.join(extract_dir, 'data')
        data_arch.extractall(path=data_dir)
        check_package_files(project, data_dir)
        assert_filemodes(project, data_dir)

    # check control.tar.xz
    with tarfile.open(name=os.path.join(extract_dir, 'control.tar.xz')) as control_arch:
        control_dir = os.path.join(extract_dir, 'control')
        control_arch.extractall(path=control_dir)

        for filename in ['control', 'preinst', 'postinst']:
            assert os.path.exists(os.path.join(control_dir, filename))

        if not tarantool_enterprise_is_used():
            assert_tarantool_dependency_deb(os.path.join(control_dir, 'control'))

        # check if postinst script set owners correctly
        with open(os.path.join(control_dir, 'postinst')) as postinst_script_file:
            postinst_script = postinst_script_file.read()
            assert 'chown -R root:root /usr/share/tarantool/{}'.format(project.name) in postinst_script
            assert 'chown root:root /etc/systemd/system/{}.service'.format(project.name) in postinst_script
            assert 'chown root:root /etc/systemd/system/{}@.service'.format(project.name) in postinst_script
            assert 'chown root:root /usr/lib/tmpfiles.d/{}.conf'.format(project.name) in postinst_script


def test_systemd_units(rpm_archive_with_custom_units, tmpdir):
    project = rpm_archive_with_custom_units.project

    # archive files should be extracted to the empty directory
    # to correctly check archive contents
    extract_dir = os.path.join(tmpdir, 'extract')
    os.makedirs(extract_dir)

    extract_rpm(rpm_archive_with_custom_units.filepath, extract_dir)

    project_unit_file = os.path.join(extract_dir, 'etc/systemd/system', "%s.service" % project.name)
    with open(project_unit_file) as f:
        assert f.read().find('SIMPLE_UNIT_TEMPLATE') != -1

    project_inst_file = os.path.join(extract_dir, 'etc/systemd/system', "%s@.service" % project.name)
    with open(project_inst_file) as f:
        assert f.read().find('INSTANTIATED_UNIT_TEMPLATE') != -1

    project_tmpfiles_conf_file = os.path.join(extract_dir, 'usr/lib/tmpfiles.d', '%s.conf' % project.name)
    with open(project_tmpfiles_conf_file) as f:
        assert f.read().find('d /var/run/tarantool') != -1


def test_packing_without_git(cartridge_cmd, project_without_dependencies, tmpdir):
    project = project_without_dependencies

    # remove .git directory
    shutil.rmtree(os.path.join(project.path, '.git'))

    # try to build rpm w/o --version
    cmd = [
        cartridge_cmd,
        "pack", "rpm",
        project.path,
    ]
    rc, output = run_command_and_get_output(cmd, cwd=tmpdir)
    assert rc == 1
    assert 'Failed to detect version' in output
    assert 'Please pass it explicitly' in output

    # pass version explicitly
    cmd = [
        cartridge_cmd,
        "pack", "rpm",
        "--version", "0.1.0",
        project.path,
    ]
    process = subprocess.run(cmd, cwd=tmpdir)
    assert process.returncode == 0
    assert '{}-0.1.0-0.rpm'.format(project.name) in os.listdir(tmpdir)


def test_packing_with_git_file(cartridge_cmd, project_without_dependencies, tmpdir):
    project = project_without_dependencies

    # remove .git directory
    shutil.rmtree(os.path.join(project.path, '.git'))

    # create file with name .git
    git_filepath = os.path.join(project.path, '.git')
    with open(git_filepath, 'w') as f:
        f.write("I am .git file")

    cmd = [
        cartridge_cmd,
        "pack", "rpm",
        "--version", "0.1.0",  # we have to pass version explicitly
        project.path,
    ]
    process = subprocess.run(cmd, cwd=tmpdir)
    assert process.returncode == 0


@pytest.mark.parametrize('pack_format', ['tgz'])
def test_result_filename_version(cartridge_cmd, project_without_dependencies, tmpdir, pack_format):
    project = project_without_dependencies

    versions = ['0.1.0', '0.1.0-42', '0.1.0-gdeadbeaf', '0.1.0-42-gdeadbeaf']
    version_to_normalized = {
        '0.1.0':              '0.1.0-0',
        '0.1.0-42':           '0.1.0-42',
        '0.1.0-gdeadbeaf':    '0.1.0-gdeadbeaf',
        '0.1.0-42-gdeadbeaf': '0.1.0-42-gdeadbeaf'
    }

    ext = pack_format if pack_format != 'tgz' else 'tar.gz'

    for version in versions:
        normalized_version = version_to_normalized[version]

        expected_filename = '{name}-{version}.{ext}'.format(
            name=project.name,
            version=normalized_version,
            ext=ext
        )

        cmd = [
            cartridge_cmd,
            "pack", pack_format,
            "--version", version,
            "--suffix", "",
            project.path,
        ]
        process = subprocess.run(cmd, cwd=tmpdir)
        assert process.returncode == 0
        assert expected_filename in os.listdir(tmpdir)


@pytest.mark.parametrize('pack_format', ['tgz'])
def test_result_filename_suffix(cartridge_cmd, project_without_dependencies, tmpdir, pack_format):
    project = project_without_dependencies

    version = '0.1.0-42'
    ext = pack_format if pack_format != 'tgz' else 'tar.gz'

    suffixes_to_filenames = {
        '':        '{name}-{version}.{ext}',
        '  ':      '{name}-{version}.{ext}',
        'dev':     '{name}-{version}-dev.{ext}',
        '  prod ': '{name}-{version}-dev.{ext}',
    }

    for suffix, filename in suffixes_to_filenames.items():
        expected_filename = filename.format(
            name=project.name,
            version=version,
            ext=ext
        )

        cmd = [
            cartridge_cmd,
            "pack", pack_format,
            "--version", version,
            "--suffix", suffix,
            project.path,
        ]
        process = subprocess.run(cmd, cwd=tmpdir)
        assert process.returncode == 0
        assert expected_filename in os.listdir(tmpdir)


def test_packing_with_wrong_filemodes(cartridge_cmd, project_without_dependencies, tmpdir):
    project = project_without_dependencies

    # add file with invalid (700) mode
    filename = 'wrong-mode-file.lua'
    filepath = os.path.join(project.path, filename)
    with open(filepath, 'w') as f:
        f.write("return 'My filemode is wrong'")
    os.chmod(filepath, 0o700)

    # run `cartridge pack`
    cmd = [cartridge_cmd, "pack", "rpm", project.path]
    rc, output = run_command_and_get_output(cmd, cwd=tmpdir)
    assert rc == 1
    assert '{} has invalid mode'.format(filename) in output


def test_builddir(cartridge_cmd, project_without_dependencies, tmpdir):
    project = project_without_dependencies

    cmd = [
        cartridge_cmd,
        "pack", "rpm",
        project.path,
    ]

    env = os.environ.copy()

    # pass application path as a builddir
    env['CARTRIDGE_BUILDDIR'] = project.path
    rc, output = run_command_and_get_output(cmd, cwd=tmpdir, env=env)
    assert rc == 1
    assert "Build directory can't be project subdirectory" in output

    # pass application subdirectory as a builddir
    env['CARTRIDGE_BUILDDIR'] = os.path.join(project.path, 'sub', 'sub', 'directory')
    rc, output = run_command_and_get_output(cmd, cwd=tmpdir, env=env)
    assert rc == 1
    assert "Build directory can't be project subdirectory" in output

    # pass correct directory as a builddir
    builddir = os.path.join(tmpdir, 'build')
    env['CARTRIDGE_BUILDDIR'] = builddir
    rc, output = run_command_and_get_output(cmd, cwd=tmpdir, env=env)
    assert rc == 0
    assert re.search(r'[Bb]uild directory .*{}'.format(builddir), output) is not None


def test_packing_without_path_specifying(cartridge_cmd, project_without_dependencies, tmpdir):
    project = project_without_dependencies

    # say `cartridge pack rpm` in project directory
    cmd = [
        cartridge_cmd,
        "pack", "rpm",
    ]
    process = subprocess.run(cmd, cwd=project.path)
    assert process.returncode == 0, 'Packing application failed'

    filepath = find_archive(project.path, project.name, 'rpm')
    assert filepath is not None,  'Package not found'


@pytest.mark.parametrize('pack_format', ['tgz'])
def test_using_both_flows(cartridge_cmd, project_without_dependencies, pack_format, tmpdir):
    project = project_without_dependencies

    deprecated_files = [
        '.cartridge.ignore',
        '.cartridge.pre',
    ]

    for filename in deprecated_files:
        filepath = os.path.join(project.path, filename)
        with open(filepath, 'w') as f:
            f.write('# I am deprecated file')

    cmd = [
        cartridge_cmd,
        "pack", pack_format,
        project.path
    ]

    rc, output = run_command_and_get_output(cmd, cwd=tmpdir)
    assert rc == 1
    assert re.search(r'You use deprecated .+ files and .+ files at the same time', output)


@pytest.mark.parametrize('pack_format', ['tgz'])
def test_build_in_docker_sdk_path_ee(cartridge_cmd, project_without_dependencies, pack_format, tmpdir):
    if not tarantool_enterprise_is_used():
        pytest.skip()

    project = project_without_dependencies

    # remove TARANTOOL_SDK_PATH from env
    env = os.environ.copy()
    del env['TARANTOOL_SDK_PATH']

    def get_pack_cmd(sdk_path):
        return [
            cartridge_cmd,
            "pack", pack_format,
            "--use-docker",
            "--sdk-path", sdk_path,
            project.path
        ]

    def create_binary(path, name, executable=False):
        binary_filepath = os.path.join(path, name)
        with open(binary_filepath, 'w') as f:
            f.write('I am {} binary'.format(name))

        if executable:
            st = os.stat(binary_filepath)
            os.chmod(binary_filepath, st.st_mode | stat.S_IEXEC)

    # pass non-exitent path
    cmd = get_pack_cmd(sdk_path='non-existent-path')
    rc, output = run_command_and_get_output(cmd, cwd=tmpdir, env=env)
    assert rc == 1
    assert 'Specified SDK path does not exists' in output

    # pass a file
    sdk_filepath = os.path.join(tmpdir, 'sdk-file')
    with open(sdk_filepath, 'w') as f:
        f.write('I am SDK file')

    cmd = get_pack_cmd(sdk_path=sdk_filepath)
    rc, output = run_command_and_get_output(cmd, cwd=tmpdir, env=env)
    assert rc == 1
    assert 'Specified SDK path is not a directory' in output

    # create empty SDK directory
    empty_sdk_path = os.path.join(tmpdir, 'SDK-empty')
    os.mkdir(empty_sdk_path)

    cmd = get_pack_cmd(sdk_path=empty_sdk_path)
    rc, output = run_command_and_get_output(cmd, cwd=tmpdir, env=env)
    assert rc == 1
    assert re.search(r'Specified SDK directory \S+ does not contain \S+ binary', output) is not None

    # check that both binaries should exists
    for binary in ['tarantool', 'tarantoolctl']:
        sdk_path = os.path.join(tmpdir, 'SDK-with-only-{}-binary'.format(binary))
        os.mkdir(sdk_path)

        create_binary(sdk_path, binary, executable=True)

        cmd = get_pack_cmd(sdk_path=sdk_path)
        rc, output = run_command_and_get_output(cmd, cwd=tmpdir, env=env)
        assert rc == 1
        assert re.search(r'Specified SDK directory \S+ does not contain \S+ binary', output) is not None

    # check that both binaries should be executable
    sdk_path = os.path.join(tmpdir, 'SDK-with-one-binary-non-exec')
    os.mkdir(sdk_path)
    create_binary(sdk_path, 'tarantool', executable=True)
    create_binary(sdk_path, 'tarantoolctl', executable=False)

    cmd = get_pack_cmd(sdk_path=sdk_path)
    rc, output = run_command_and_get_output(cmd, cwd=tmpdir, env=env)
    assert rc == 1
    assert 'Specified SDK directory contains tarantoolctl binary that is not executable' in output


@pytest.mark.parametrize('pack_format', ['tgz', 'docker'])
def test_project_without_build_dockerfile(cartridge_cmd, project_without_dependencies, tmpdir, pack_format):
    project = project_without_dependencies

    os.remove(os.path.join(project.path, 'Dockerfile.build.cartridge'))

    cmd = [
        cartridge_cmd,
        "pack", pack_format,
        "--use-docker",
        project.path,
    ]

    process = subprocess.run(cmd, cwd=tmpdir)
    assert process.returncode == 0


def test_project_without_runtime_dockerfile(cartridge_cmd, project_without_dependencies, tmpdir):
    project = project_without_dependencies

    os.remove(os.path.join(project.path, 'Dockerfile.cartridge'))

    cmd = [
        cartridge_cmd,
        "pack", "docker",
        project.path,
    ]

    process = subprocess.run(cmd, cwd=tmpdir)
    assert process.returncode == 0


@pytest.mark.parametrize('pack_format', ['tgz'])
def test_invalid_base_build_dockerfile(cartridge_cmd, project_without_dependencies, pack_format, tmpdir):
    invalid_dockerfile_path = os.path.join(tmpdir, 'Dockerfile')
    with open(invalid_dockerfile_path, 'w') as f:
        f.write('''
            # Invalid dockerfile
            FROM ubuntu:xenial
        ''')

    cmd = [
        cartridge_cmd,
        "pack", pack_format,
        "--use-docker",
        "--build-from", invalid_dockerfile_path,
        project_without_dependencies.path,
    ]

    rc, output = run_command_and_get_output(cmd, cwd=tmpdir)
    assert rc == 1
    assert 'Base Dockerfile validation failed' in output
    assert 'base image must be centos:8' in output


def test_invalid_base_runtime_dockerfile(cartridge_cmd, project_without_dependencies, tmpdir):
    invalid_dockerfile_path = os.path.join(tmpdir, 'Dockerfile')
    with open(invalid_dockerfile_path, 'w') as f:
        f.write('''
            # Invalid dockerfile
            FROM ubuntu:xenial
        ''')

    cmd = [
        cartridge_cmd,
        "pack", "docker",
        "--from", invalid_dockerfile_path,
        project_without_dependencies.path,
    ]

    rc, output = run_command_and_get_output(cmd, cwd=tmpdir)
    assert rc == 1
    assert 'Base Dockerfile validation failed' in output
    assert 'base image must be centos:8' in output


@pytest.mark.parametrize('pack_format', ['tgz'])
def test_base_build_dockerfile_with_env_vars(cartridge_cmd, project_without_dependencies, pack_format, tmpdir):
    # The main idea of this test is to check that using `${name}` constructions
    #   in the base Dockerfile doesn't break the `pack` command running.
    # So, it's not about testing that the ENV option works, it's about
    #   testing that `pack docker` command wouldn't fail if the base Dockerfile
    #   contains `${name}` constructions.
    # The problem is the `expand` function.
    # Base Dockerfile with `${name}` shouldn't be passed to this function,
    #   otherwise it will raise an error or substitute smth wrong.
    dockerfile_with_env_path = os.path.join(tmpdir, 'Dockerfile')
    with open(dockerfile_with_env_path, 'w') as f:
        f.write('''
            FROM centos:8
            # comment this string to use cached image
            # ENV TEST_VARIABLE=${TEST_VARIABLE}
        ''')

    cmd = [
        cartridge_cmd,
        "pack", pack_format,
        "--use-docker",
        "--build-from", dockerfile_with_env_path,
        project_without_dependencies.path,
    ]

    rc, output = run_command_and_get_output(cmd, cwd=tmpdir)
    assert rc == 0
    assert 'Detected base Dockerfile {}'.format(dockerfile_with_env_path) in output


def test_base_runtime_dockerfile_with_env_vars(cartridge_cmd, project_without_dependencies, tmpdir):
    # The main idea of this test is to check that using `${name}` constructions
    #   in the base Dockerfile doesn't break the `pack` command running.
    # So, it's not about testing that the ENV option works, it's about
    #   testing that `pack docker` command wouldn't fail if the base Dockerfile
    #   contains `${name}` constructions.
    # The problem is the `expand` function.
    # Base Dockerfile with `${name}` shouldn't be passed to this function,
    #   otherwise it will raise an error or substitute smth wrong.
    dockerfile_with_env_path = os.path.join(tmpdir, 'Dockerfile')
    with open(dockerfile_with_env_path, 'w') as f:
        f.write('''
            FROM centos:8
            # comment this string to use cached image
            # ENV TEST_VARIABLE=${TEST_VARIABLE}
        ''')

    cmd = [
        cartridge_cmd,
        "pack", "docker",
        "--from", dockerfile_with_env_path,
        project_without_dependencies.path,
    ]
    rc, output = run_command_and_get_output(cmd, cwd=tmpdir)
    assert rc == 0
    assert 'Detected base Dockerfile {}'.format(dockerfile_with_env_path) in output


@pytest.mark.parametrize('pack_format', ['tgz'])
def test_builddir_is_removed(cartridge_cmd, project_without_dependencies, pack_format, tmpdir):
    project = project_without_dependencies

    cmd = [
        cartridge_cmd,
        "pack", pack_format,
        project.path,
    ]

    env = os.environ.copy()

    # pass correct directory as a builddir
    builddir = os.path.join(tmpdir, 'build')
    env['CARTRIDGE_BUILDDIR'] = builddir
    process = subprocess.run(cmd, cwd=tmpdir)
    assert process.returncode == 0

    assert not os.path.exists(builddir)


@pytest.mark.parametrize('pack_format', ['rpm', 'deb'])
def test_project_without_stateboard(cartridge_cmd, project_without_dependencies, pack_format, tmpdir):
    project = project_without_dependencies

    STATEBOARD_ENTRYPOINT_NAME = 'stateboard.init.lua'

    # remove stateboard entrypoint from project
    os.remove(os.path.join(project.path, STATEBOARD_ENTRYPOINT_NAME))
    project.distribution_files.remove(STATEBOARD_ENTRYPOINT_NAME)

    cmd = [
        cartridge_cmd,
        "pack", pack_format,
        project.path,
    ]

    # call cartridge pack
    process = subprocess.run(cmd, cwd=tmpdir)
    assert process.returncode == 0

    # packing should succeed with warning
    rc, output = run_command_and_get_output(cmd, cwd=tmpdir)
    assert rc == 0
    assert "App directory doesn't contain stateboard entrypoint script" in output

    # extract files from archive
    archive_path = find_archive(tmpdir, project.name, pack_format)
    extract_dir = os.path.join(tmpdir, 'extract')
    os.makedirs(extract_dir)

    if pack_format == 'rpm':
        extract_rpm(archive_path, extract_dir)
    elif pack_format == 'deb':
        extract_deb(archive_path, extract_dir)
        with tarfile.open(name=os.path.join(extract_dir, 'data.tar.xz')) as data_arch:
            data_arch.extractall(path=extract_dir)

    # check that stateboard unit file wasn't delivered
    systemd_dir = (os.path.join(extract_dir, 'etc/systemd/system'))
    assert os.path.exists(systemd_dir)

    systemd_files = recursive_listdir(systemd_dir)

    assert len(systemd_files) == 2
    assert '{}-stateboard.service'.format(project.name) not in systemd_files
