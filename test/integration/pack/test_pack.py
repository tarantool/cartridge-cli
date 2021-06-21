import os
import subprocess
import tarfile
import re
import shutil
import stat
import platform
import hashlib
import yaml
import pytest

from utils import tarantool_enterprise_is_used
from utils import Archive, find_archive
from utils import recursive_listdir
from utils import assert_distribution_dir_contents
from utils import assert_filemodes
from utils import assert_files_mode_and_owner_rpm
from utils import validate_version_file
from utils import check_package_files
from utils import assert_tarantool_dependency_deb, assert_dependencies_deb
from utils import assert_tarantool_dependency_rpm, assert_dependencies_rpm
from utils import assert_pre_and_post_install_scripts_rpm, assert_pre_and_post_install_scripts_deb
from utils import run_command_and_get_output
from utils import get_rockspec_path
from utils import tarantool_version
from utils import extract_app_files, extract_rpm, extract_deb
from utils import check_fd_limits_in_unit_files, check_param_in_unit_files
from utils import clear_project_rocks_cache, get_rocks_cache_path

from project import set_and_return_whoami_on_build, replace_project_file, remove_project_file


# ########
# Fixtures
# ########
@pytest.fixture(scope="function", params=['local', 'docker'])
def tgz_archive(cartridge_cmd, tmpdir, light_project, request):
    project = light_project

    cmd = [cartridge_cmd, "pack", "tgz", project.path]

    if request.param == 'docker':
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
        cmd.append('--use-docker')

    rc, output = run_command_and_get_output(cmd, cwd=tmpdir)

    if request.param == 'local' and platform.system() == 'Darwin':
        assert rc == 1
        assert "It's not possible to pack application into RPM or DEB on non-linux OS" in output

        pytest.skip("Packing RPM and DEB locally should fail for Darwin")

    assert rc == 0

    filepath = find_archive(tmpdir, project.name, 'rpm')
    assert filepath is not None, "RPM archive isn't found in work directory"

    return Archive(filepath=filepath, project=project)


@pytest.fixture(scope="function", params=['local', 'docker'])
def deb_archive(cartridge_cmd, tmpdir, light_project, request):
    project = light_project

    cmd = [cartridge_cmd, "pack", "deb", project.path]

    if request.param == 'docker':
        cmd.append('--use-docker')

    rc, output = run_command_and_get_output(cmd, cwd=tmpdir)

    if request.param == 'local' and platform.system() == 'Darwin':
        assert rc == 1
        assert "It's not possible to pack application into RPM or DEB on non-linux OS" in output

        pytest.skip("Packing RPM and DEB locally should fail for Darwin")

    assert rc == 0

    filepath = find_archive(tmpdir, project.name, 'deb')
    assert filepath is not None, "DEB archive isn't found in work directory"

    return Archive(filepath=filepath, project=project)


@pytest.fixture(scope="session")
def tarantool_versions():
    min_deb_version = re.findall(r'\d+\.\d+\.\d+-\d+-\S+', tarantool_version())[0]
    max_deb_version = str(int(re.findall(r'\d+', tarantool_version())[0]) + 1)
    min_rpm_version = re.findall(r'\d+\.\d+\.\d+', tarantool_version())[0]
    max_rpm_version = max_deb_version  # Their format is the same

    return {"min": {"deb": min_deb_version, "rpm": min_rpm_version},
            "max": {"deb": max_deb_version, "rpm": max_rpm_version}}


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

    for filename in ['debian-binary', 'control.tar.gz', 'data.tar.gz']:
        assert os.path.exists(os.path.join(extract_dir, filename))

    # check debian-binary
    with open(os.path.join(extract_dir, 'debian-binary')) as debian_binary_file:
        assert debian_binary_file.read() == '2.0\n'

    # check data.tar.gz
    with tarfile.open(name=os.path.join(extract_dir, 'data.tar.gz')) as data_arch:
        data_dir = os.path.join(extract_dir, 'data')
        data_arch.extractall(path=data_dir)
        check_package_files(project, data_dir)
        assert_filemodes(project, data_dir)

    # check control.tar.gz
    with tarfile.open(name=os.path.join(extract_dir, 'control.tar.gz')) as control_arch:
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


@pytest.mark.parametrize('unit', ['unit', 'instantiated-unit', 'stateboard-unit'])
@pytest.mark.parametrize('pack_format', ['deb', 'rpm'])
def test_custom_unit_files(cartridge_cmd, project_without_dependencies, tmpdir, unit, pack_format):
    project = project_without_dependencies

    files_by_units = {
        'unit': "%s.service" % project.name,
        'instantiated-unit': "%s@.service" % project.name,
        'stateboard-unit': "%s-stateboard.service" % project.name,
    }

    CUSTOM_UNIT_TEMPLATE = "CUSTOM UNIT"

    unit_template_filepath = os.path.join(tmpdir, "systemd-unit-template")
    with open(unit_template_filepath, 'w') as f:
        f.write(CUSTOM_UNIT_TEMPLATE)

    # pass non-existent path
    cmd = [
        cartridge_cmd, "pack", pack_format,
        "--%s-template" % unit, "non-existent-path",
        project.path
    ]

    if platform.system() == 'Darwin':
        cmd.append('--use-docker')

    rc, output = run_command_and_get_output(cmd, cwd=tmpdir)
    assert rc == 1
    assert re.search(r'Failed to read specified .*unit template', output) is not None

    # pass correct path
    cmd = [
        cartridge_cmd, "pack", pack_format,
        "--%s-template" % unit, unit_template_filepath,
        project.path
    ]

    if platform.system() == 'Darwin':
        cmd.append('--use-docker')

    process = subprocess.run(cmd, cwd=tmpdir)
    assert process.returncode == 0

    # extract files from archive
    archive_path = find_archive(tmpdir, project.name, pack_format)
    extract_dir = os.path.join(tmpdir, 'extract')
    extract_app_files(archive_path, pack_format, extract_dir)

    filename = files_by_units[unit]
    filepath = os.path.join(extract_dir, 'etc/systemd/system', filename)
    with open(filepath) as f:
        assert f.read() == CUSTOM_UNIT_TEMPLATE


@pytest.mark.parametrize('pack_format', ['tgz'])
def test_packing_without_git(cartridge_cmd, project_without_dependencies, tmpdir, pack_format):
    project = project_without_dependencies

    # remove .git directory
    shutil.rmtree(os.path.join(project.path, '.git'))

    # try to build rpm w/o --version
    cmd = [
        cartridge_cmd,
        "pack", pack_format,
        project.path,
    ]
    rc, output = run_command_and_get_output(cmd, cwd=tmpdir)
    assert rc == 1
    assert 'Project is not a git project' in output
    assert 'Please pass version explicitly' in output

    # pass version explicitly
    cmd = [
        cartridge_cmd,
        "pack", pack_format,
        "--version", "0.1.0",
        project.path,
    ]
    process = subprocess.run(cmd, cwd=tmpdir)
    assert process.returncode == 0


@pytest.mark.parametrize('pack_format', ['tgz'])
def test_packing_with_git_file(cartridge_cmd, project_without_dependencies, tmpdir, pack_format):
    project = project_without_dependencies

    # remove .git directory
    shutil.rmtree(os.path.join(project.path, '.git'))

    # create file with name .git
    git_filepath = os.path.join(project.path, '.git')
    with open(git_filepath, 'w') as f:
        f.write("I am .git file")

    cmd = [
        cartridge_cmd,
        "pack", pack_format,
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


@pytest.mark.parametrize('pack_format', ['tgz'])
def test_invalid_version(cartridge_cmd, project_without_dependencies, tmpdir, pack_format):
    project = project_without_dependencies

    bad_versions = [
        'bad', '0-1-0', '0.1.0.42', '0.1.0.gdeadbeaf', '0.1.0.42.gdeadbeaf',
        'xx1', '1xx',
        'xx1.2', '1.2xx',
        'xx1.2.3', '1.2.3xx',
        'xx1.2.3-4', '1.2.3-4xxx',
        'xx1.2.3-4-gdeadbeaf',
    ]

    for bad_version in bad_versions:
        cmd = [
            cartridge_cmd,
            "pack", pack_format,
            "--version", bad_version,
            project.path,
        ]
        rc, output = run_command_and_get_output(cmd, cwd=tmpdir)
        assert rc == 1
        rgx = r"Version should be semantic \(major\.minor\.patch\[\-count\]\[\-commit\]\)"
        assert re.search(rgx, output) is not None


@pytest.mark.parametrize('pack_format', ['tgz'])
def test_packing_with_wrong_filemodes(cartridge_cmd, project_without_dependencies, tmpdir, pack_format):
    project = project_without_dependencies

    # add file with invalid (700) mode
    filename = 'wrong-mode-file.lua'
    filepath = os.path.join(project.path, filename)
    with open(filepath, 'w') as f:
        f.write("return 'My filemode is wrong'")
    os.chmod(filepath, 0o700)

    # run `cartridge pack`
    cmd = [cartridge_cmd, "pack", pack_format, project.path]
    rc, output = run_command_and_get_output(cmd, cwd=tmpdir)
    assert rc == 1
    assert '{} has invalid mode'.format(filename) in output


@pytest.mark.parametrize('pack_format', ['tgz'])
def test_tempdir(cartridge_cmd, project_without_dependencies, tmpdir, pack_format):
    project = project_without_dependencies

    cmd = [
        cartridge_cmd,
        "pack", pack_format,
        project.path,
    ]

    env = os.environ.copy()

    # pass application path as a cartridge_tempdir
    env['CARTRIDGE_TEMPDIR'] = project.path
    rc, output = run_command_and_get_output(cmd, cwd=tmpdir, env=env)
    assert rc == 0

    # pass application subdirectory as a cartridge_tempdir
    env['CARTRIDGE_TEMPDIR'] = os.path.join(project.path, 'sub', 'sub', 'directory')
    rc, output = run_command_and_get_output(cmd, cwd=tmpdir, env=env)
    assert rc == 0

    # pass correct directory as a cartridge_tempdir
    cartridge_tempdir = os.path.join(tmpdir, 'build')
    env['CARTRIDGE_TEMPDIR'] = cartridge_tempdir
    rc, output = run_command_and_get_output(cmd, cwd=tmpdir, env=env)
    assert rc == 0
    assert re.search(r'Temporary directory is set to {}'.format(cartridge_tempdir), output) is not None


@pytest.mark.parametrize('pack_format', ['tGz'])
def test_pack_type_mixed_case(cartridge_cmd, project_without_dependencies, tmpdir, pack_format):
    project = project_without_dependencies

    cmd = [
        cartridge_cmd,
        "pack", pack_format,
        project.path,
    ]

    process = subprocess.run(cmd, cwd=tmpdir)
    assert process.returncode == 0


@pytest.mark.parametrize('pack_format', ['tgz'])
def test_packing_without_path_specifying(cartridge_cmd, project_without_dependencies, pack_format):
    project = project_without_dependencies

    # say `cartridge pack <type>` in project directory
    cmd = [
        cartridge_cmd,
        "pack", pack_format,
    ]
    process = subprocess.run(cmd, cwd=project.path)
    assert process.returncode == 0, 'Packing application failed'


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
    assert 'Unable to use specified SDK' in output

    # pass a file
    sdk_filepath = os.path.join(tmpdir, 'sdk-file')
    with open(sdk_filepath, 'w') as f:
        f.write('I am SDK file')

    cmd = get_pack_cmd(sdk_path=sdk_filepath)
    rc, output = run_command_and_get_output(cmd, cwd=tmpdir, env=env)
    assert rc == 1
    assert 'Unable to use specified SDK: Is not a directory' in output

    # create empty SDK directory
    empty_sdk_path = os.path.join(tmpdir, 'SDK-empty')
    os.mkdir(empty_sdk_path)

    cmd = get_pack_cmd(sdk_path=empty_sdk_path)
    rc, output = run_command_and_get_output(cmd, cwd=tmpdir, env=env)
    assert rc == 1
    assert re.search(r'Unable to use specified SDK: \S+ binary is missed', output) is not None

    # check that both binaries should exists
    for binary in ['tarantool', 'tarantoolctl']:
        sdk_path = os.path.join(tmpdir, 'SDK-with-only-{}-binary'.format(binary))
        os.mkdir(sdk_path)

        create_binary(sdk_path, binary, executable=True)

        cmd = get_pack_cmd(sdk_path=sdk_path)
        rc, output = run_command_and_get_output(cmd, cwd=tmpdir, env=env)
        assert rc == 1
        assert re.search(r'Unable to use specified SDK: \S+ binary is missed', output) is not None

    # check that both binaries should be executable
    sdk_path = os.path.join(tmpdir, 'SDK-with-one-binary-non-exec')
    os.mkdir(sdk_path)
    create_binary(sdk_path, 'tarantool', executable=True)
    create_binary(sdk_path, 'tarantoolctl', executable=False)

    cmd = get_pack_cmd(sdk_path=sdk_path)
    rc, output = run_command_and_get_output(cmd, cwd=tmpdir, env=env)
    assert rc == 1
    assert 'Unable to use specified SDK: tarantoolctl binary is not executable' in output


# @pytest.mark.parametrize('pack_format', ['tgz', 'docker'])
@pytest.mark.parametrize('pack_format', ['tgz'])
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


@pytest.mark.parametrize('pack_format', ['tgz'])
def test_custom_base_image_build_dockerfile(
    cartridge_cmd, project_without_dependencies, pack_format, custom_base_image, tmpdir
):
    custom_base_image_dockerfile = f"FROM {custom_base_image['image_name']}"

    custom_dockerfile_path = os.path.join(tmpdir, 'Dockerfile')
    with open(custom_dockerfile_path, 'w') as f:
        f.write(custom_base_image_dockerfile)

    cmd = [
        cartridge_cmd,
        "pack", pack_format,
        "--use-docker",
        "--build-from", custom_dockerfile_path,
        project_without_dependencies.path,
    ]

    rc, output = run_command_and_get_output(cmd, cwd=tmpdir)
    assert rc == 0
    assert 'Image based on centos:8 is expected to be used' in output


@pytest.mark.parametrize('pack_format', ['tgz'])
def test_pack_tempdir_is_removed(cartridge_cmd, project_without_dependencies, pack_format, tmpdir):
    project = project_without_dependencies
    PACK_TEMPDIR_RGX = re.compile(r'Temporary directory is set to ([\w\-\_\.\/\~]+)')

    # pass correct directory as a tempdir
    cartridge_tempdir = os.path.join(tmpdir, 'build')
    env = os.environ.copy()
    env['CARTRIDGE_TEMPDIR'] = cartridge_tempdir

    # w/o --debug flag
    cmd = [
        cartridge_cmd,
        "pack", pack_format,
        project.path,
    ]

    rc, output = run_command_and_get_output(cmd, cwd=tmpdir, env=env)
    assert rc == 0

    m = re.search(PACK_TEMPDIR_RGX, output)
    assert m is not None
    pack_tempdir = m.group(1)

    assert not os.path.exists(pack_tempdir)

    # w/ --debug flag
    cmd = [
        cartridge_cmd,
        "pack", pack_format,
        project.path,
        "--debug",
    ]

    rc, output = run_command_and_get_output(cmd, cwd=tmpdir, env=env)
    assert rc == 0

    m = re.search(PACK_TEMPDIR_RGX, output)
    assert m is not None
    pack_tempdir = m.group(1)

    assert os.path.exists(pack_tempdir)


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

    if platform.system() == 'Darwin':
        cmd.append('--use-docker')

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
    extract_app_files(archive_path, pack_format, extract_dir)

    # check that stateboard unit file wasn't delivered
    systemd_dir = (os.path.join(extract_dir, 'etc/systemd/system'))
    assert os.path.exists(systemd_dir)

    systemd_files = recursive_listdir(systemd_dir)

    assert len(systemd_files) == 2
    assert '{}-stateboard.service'.format(project.name) not in systemd_files


@pytest.mark.parametrize('build', ['docker', 'local'])
@pytest.mark.parametrize('pack_format', ['tgz'])
def test_pack_with_spec_specified(cartridge_cmd, project_without_dependencies, pack_format, build, tmpdir):
    project = project_without_dependencies

    version = 'scm-2'
    rockspec_path = get_rockspec_path(project.path, project.name, version)
    who_am_i = set_and_return_whoami_on_build(rockspec_path, project.name, version)

    cmd = [
        cartridge_cmd,
        "pack", pack_format,
        project.path,
        "--spec",
        rockspec_path,
        "--verbose",
    ]

    if build == "docker":
        cmd.append("--use-docker")

    rc, output = run_command_and_get_output(cmd, cwd=tmpdir)
    assert rc == 0
    # tarantoolctl performs build with the oldest version of rockspec files
    assert who_am_i in output


@pytest.mark.parametrize('pack_format', ['tgz'])
def test_packing_with_rockspec_from_other_dir(cartridge_cmd, project_without_dependencies, pack_format, tmpdir):
    project = project_without_dependencies

    dir_path = os.path.join(project.path, 'some_dir')
    os.mkdir(dir_path)

    version = 'scm-2'
    rockspec_path = get_rockspec_path(dir_path, project.name, version)

    cmd = [
        cartridge_cmd,
        "pack", pack_format,
        project.path,
        "--spec",
        rockspec_path,
        "--verbose",
    ]

    rc, output = run_command_and_get_output(cmd, cwd=tmpdir)
    assert rc == 1, 'Building project should fail'

    rocks_make_output = "Rockspec %s should be in project root" % rockspec_path
    assert rocks_make_output in output


@pytest.mark.parametrize('pack_format', ['tgz'])
def test_pack_with_rockspec_bad_name(cartridge_cmd, project_without_dependencies, pack_format):
    project = project_without_dependencies

    bad_rockspec_name = "bad_rockspec-scm-1.rockspec"
    bad_rockspec_path = os.path.join(project.path, bad_rockspec_name)
    rocks_make_output = "Rockspec %s doesn't exist" % bad_rockspec_path

    # with --spec
    cmd = [
        cartridge_cmd,
        "pack", pack_format,
        project.path,
        "--spec",
        bad_rockspec_path,
    ]

    rc, output = run_command_and_get_output(cmd, cwd=project.path)
    assert rc == 1, 'Building project should fail'
    assert rocks_make_output in output


@pytest.mark.parametrize('pack_format', ['rpm', 'deb', 'tgz', 'docker'])
def test_project_without_init(cartridge_cmd, project_without_dependencies, pack_format, tmpdir):
    project = project_without_dependencies

    ENTRYPOINT_NAME = 'init.lua'

    # remove entrypoint from project
    os.remove(os.path.join(project.path, ENTRYPOINT_NAME))

    cmd = [
        cartridge_cmd,
        "pack", pack_format,
        project.path,
    ]

    if pack_format in ['rpm', 'deb'] and platform.system() == 'Darwin':
        cmd.append('--use-docker')

    rc, output = run_command_and_get_output(cmd, cwd=tmpdir)
    if pack_format == 'tgz':
        assert rc == 0
    else:
        assert rc == 1
        assert "Application doesn't contain entrypoint script" in output


@pytest.mark.parametrize('pack_format', ['rpm', 'deb', 'tgz', 'docker'])
def test_files_with_bad_symbols(cartridge_cmd, project_without_dependencies, pack_format, tmpdir):
    project = project_without_dependencies

    BAD_FILENAME = 'I \'am\' "the" $worst (file) [ever]'

    with open(os.path.join(project.path, BAD_FILENAME), 'w') as f:
        f.write('Hi!')

    cmd = [
        cartridge_cmd,
        "pack", pack_format,
        project.path,
    ]

    if pack_format in ['rpm', 'deb'] and platform.system() == 'Darwin':
        cmd.append('--use-docker')

    # call cartridge pack
    process = subprocess.run(cmd, cwd=tmpdir)
    assert process.returncode == 0


@pytest.mark.parametrize('pack_format', ['rpm', 'deb', 'tgz'])
def test_tempdir_with_bad_symbols(cartridge_cmd, project_without_dependencies, pack_format, tmpdir):
    project = project_without_dependencies

    BAD_DIRNAME = 'I \'am\' "the" $worst (directory) [ever]'
    cartridge_tempdir = os.path.join(tmpdir, BAD_DIRNAME)
    os.makedirs(cartridge_tempdir)

    cmd = [
        cartridge_cmd,
        "pack", pack_format,
        project.path,
    ]

    if pack_format in ['rpm', 'deb'] and platform.system() == 'Darwin':
        cmd.append('--use-docker')

    env = os.environ.copy()
    env['CARTRIDGE_TEMPDIR'] = cartridge_tempdir

    # call cartridge pack
    process = subprocess.run(cmd, cwd=tmpdir)
    assert process.returncode == 0


@pytest.mark.parametrize('hook', ['cartridge.pre-build', 'cartridge.post-build'])
@pytest.mark.parametrize('pack_format', ['tgz'])
def test_app_with_non_executable_hook(cartridge_cmd, project_without_dependencies, hook, pack_format, tmpdir):
    project = project_without_dependencies

    hook_path = os.path.join(project.path, hook)
    os.chmod(hook_path, 0o0644)

    cmd = [
        cartridge_cmd,
        "pack", pack_format,
        project.path,
    ]

    rc, output = run_command_and_get_output(cmd, cwd=tmpdir)
    assert rc == 1, 'Packing project should fail'
    assert 'Hook `{}` should be executable'.format(hook) in output


@pytest.mark.parametrize('build', ['docker', 'local'])
@pytest.mark.parametrize('hook', ['cartridge.pre-build', 'cartridge.post-build'])
@pytest.mark.parametrize('pack_format', ['tgz'])
def test_app_with_non_existing_hook(cartridge_cmd, project_without_dependencies, hook, pack_format, build, tmpdir):
    project = project_without_dependencies

    hook_path = os.path.join(project.path, hook)
    os.remove(hook_path)

    cmd = [
        cartridge_cmd,
        "pack", pack_format,
        project.path,
    ]

    if build == "docker":
        cmd.append("--use-docker")

    process = subprocess.run(cmd, cwd=tmpdir)
    assert process.returncode == 0


@pytest.mark.parametrize('pack_format', ['tgz'])
def test_verbosity(cartridge_cmd, project_without_dependencies, pack_format):
    project = project_without_dependencies

    prebuild_output = "pre-build hook output"
    postbuild_output = "post-build hook output"
    rocks_make_output = "{} scm-1 is now installed".format(project.name)

    with open(os.path.join(project.path, 'cartridge.pre-build'), 'w') as f:
        prebuild_script_lines = [
            "#!/bin/sh",
            "echo \"{}\"".format(prebuild_output)
        ]
        f.write('\n'.join(prebuild_script_lines))

    with open(os.path.join(project.path, 'cartridge.post-build'), 'w') as f:
        postbuild_script_lines = [
            "#!/bin/sh",
            "echo \"{}\"".format(postbuild_output)
        ]
        f.write('\n'.join(postbuild_script_lines))

    build_logs = [
        'Packing empty-project into %s' % pack_format,
        'Build application in',
        'Running `cartridge.pre-build`',
        'Running `tarantoolctl rocks make`',
        'Running `cartridge.post-build`',
        'Application was successfully built',
    ]

    # w/o flags
    cmd = [
        cartridge_cmd,
        "pack", pack_format,
    ]

    rc, output = run_command_and_get_output(cmd, cwd=project.path)
    assert rc == 0
    assert all([log in output for log in build_logs])
    assert prebuild_output not in output
    assert rocks_make_output not in output

    # with --verbose
    cmd = [
        cartridge_cmd,
        "pack", pack_format,
        "--verbose",
    ]

    project_dir = hashlib.sha1(project.path.encode('utf-8')).hexdigest()[:10]
    clear_project_rocks_cache(project_dir)
    rc, output = run_command_and_get_output(cmd, cwd=project.path)
    assert rc == 0
    assert all([log in output for log in build_logs])
    assert prebuild_output in output
    assert rocks_make_output in output

    # with --quiet
    cmd = [
        cartridge_cmd,
        "pack", pack_format,
        "--quiet",
    ]

    rc, output = run_command_and_get_output(cmd, cwd=project.path)
    assert rc == 0, 'Building project failed'
    assert output == ''

    # hook error with --quiet
    cmd = [
        cartridge_cmd,
        "pack", pack_format,
        "--quiet",
    ]

    with open(os.path.join(project.path, 'cartridge.pre-build'), 'w') as f:
        prebuild_script_lines = [
            "#!/bin/sh",
            "echo \"{}\"".format(prebuild_output),
            "exit 1"
        ]
        f.write('\n'.join(prebuild_script_lines))

    clear_project_rocks_cache(project_dir)
    rc, output = run_command_and_get_output(cmd, cwd=project.path)
    assert rc == 1, 'Building project should fail'
    assert all([log not in output for log in build_logs])
    assert 'Failed to run pre-build hook' in output
    assert prebuild_output in output


@pytest.mark.parametrize('pack_format', ['deb', 'rpm'])
def test_dependencies(cartridge_cmd, project_without_dependencies, pack_format, tarantool_versions, tmpdir):
    project = project_without_dependencies

    cmd = [
        cartridge_cmd,
        "pack", pack_format,
        "--deps", "dependency01>=1.2,dependency01< 3",
        "--deps", "dependency02 == 2.5 ",
        "--deps", "\tdependency03    ",
        project.path,
    ]

    if platform.system() == 'Darwin':
        cmd.append('--use-docker')

    rc, output = run_command_and_get_output(cmd, cwd=tmpdir)
    assert rc == 0

    if pack_format == 'deb':
        deps = (
            "dependency01 (>= 1.2)",
            "dependency01 (<< 3)",
            "dependency02 (= 2.5)",
            "dependency03"
        )

        assert_dependencies_deb(find_archive(tmpdir, project.name, 'deb'), deps, tarantool_versions, tmpdir)
    else:
        deps = (
            ("dependency01", 0x08 | 0x04, "1.2"),  # >=
            ("dependency01", 0x02, "3"),  # <
            ("dependency02", 0x08, "2.5"),  # =
            ("dependency03", 0, "")
        )

        assert_dependencies_rpm(find_archive(tmpdir, project.name, 'rpm'), deps, tarantool_versions)


@pytest.mark.parametrize('pack_format', ['deb', 'rpm'])
def test_standard_dependencies_file(
    cartridge_cmd, project_without_dependencies, pack_format, tarantool_versions, tmpdir
):
    project = project_without_dependencies

    deps_filepath = os.path.join(tmpdir, "deps.txt")
    with open(deps_filepath, "w") as f:
        f.write("cool_dependency_01 >= 1.2, < 3 \n" +
                "\t\n" +
                " // comment\n" +
                "\n" +
                "  cool_dependency_02 == 2.5\n" +
                "\tcool_dependency_03  ")

    replace_project_file(project, 'package-deps.txt', deps_filepath)

    cmd = [
        cartridge_cmd,
        "pack", pack_format,
        project.path,
    ]

    if platform.system() == 'Darwin':
        cmd.append('--use-docker')

    rc, output = run_command_and_get_output(cmd, cwd=tmpdir)
    assert rc == 0

    if pack_format == 'deb':
        deps = (
            "cool_dependency_01 (>= 1.2)",
            "cool_dependency_01 (<< 3)",
            "cool_dependency_02 (= 2.5)",
            "cool_dependency_03"
        )

        assert_dependencies_deb(find_archive(tmpdir, project.name, 'deb'), deps, tarantool_versions, tmpdir)
    else:
        deps = (
            ("cool_dependency_01", 0x08 | 0x04, "1.2"),  # >=
            ("cool_dependency_01", 0x02, "3"),  # <
            ("cool_dependency_02", 0x08, "2.5"),  # =
            ("cool_dependency_03", 0, "")
        )

        assert_dependencies_rpm(find_archive(tmpdir, project.name, 'rpm'), deps, tarantool_versions)


@pytest.mark.parametrize('pack_format', ['deb', 'rpm'])
def test_custom_dependencies_file(cartridge_cmd, project_without_dependencies, pack_format, tarantool_versions, tmpdir):
    project = project_without_dependencies

    deps_filepath = os.path.join(tmpdir, "deps.txt")
    with open(deps_filepath, "w") as f:
        f.write("dependency01 >= 1.2, < 3 \n" +
                "  dependency02 == 2.5\n" +
                "\tdependency03  ")

    cmd = [
        cartridge_cmd,
        "pack", pack_format,
        "--deps-file", deps_filepath,
        project.path,
    ]

    if platform.system() == 'Darwin':
        cmd.append('--use-docker')

    rc, output = run_command_and_get_output(cmd, cwd=tmpdir)
    assert rc == 0

    if pack_format == 'deb':
        deps = (
            "dependency01 (>= 1.2)",
            "dependency01 (<< 3)",
            "dependency02 (= 2.5)",
            "dependency03"
        )

        assert_dependencies_deb(find_archive(tmpdir, project.name, 'deb'), deps, tarantool_versions, tmpdir)
    else:
        deps = (
            ("dependency01", 0x08 | 0x04, "1.2"),  # >=
            ("dependency01", 0x02, "3"),  # <
            ("dependency02", 0x08, "2.5"),  # =
            ("dependency03", 0, "")
        )

        assert_dependencies_rpm(find_archive(tmpdir, project.name, 'rpm'), deps, tarantool_versions)


@pytest.mark.parametrize('pack_format', ['docker', 'tgz'])
def test_dependencies_not_rpm_deb(cartridge_cmd, project_without_dependencies, pack_format, tmpdir):
    project = project_without_dependencies

    cmd = [
        cartridge_cmd,
        "pack", pack_format,
        "--deps-file", "dependencies.txt",
        project.path,
    ]

    warning_message = "You specified the --deps-file flag, but you are not packaging RPM or DEB. Flag will be ignored"
    rc, output = run_command_and_get_output(cmd, cwd=tmpdir)
    assert rc == 0
    assert warning_message in output

    cmd = [
        cartridge_cmd,
        "pack", pack_format,
        "--deps", "dependencies01",
        project.path
    ]

    warning_message = "You specified the --deps flag, but you are not packaging RPM or DEB. Flag will be ignored"
    rc, output = run_command_and_get_output(cmd, cwd=tmpdir)
    assert rc == 0
    assert warning_message in output


@pytest.mark.parametrize('pack_format', ['deb', 'rpm'])
def test_no_default_deps_file(cartridge_cmd, project_without_dependencies, pack_format, tmpdir):
    project = project_without_dependencies
    os.remove(os.path.join(project.path, 'package-deps.txt'))

    cmd = [
        cartridge_cmd,
        "pack", pack_format,
        project.path,
    ]

    if platform.system() == 'Darwin':
        cmd.append('--use-docker')

    rc, output = run_command_and_get_output(cmd, cwd=tmpdir)
    assert rc == 0


@pytest.mark.parametrize('pack_format', ['deb', 'rpm'])
def test_dependencies_same_flags(cartridge_cmd, project_without_dependencies, pack_format, tmpdir):
    project = project_without_dependencies

    deps_filepath = os.path.join(tmpdir, "deps.txt")
    with open(deps_filepath, "w") as f:
        f.write("")

    cmd = [
        cartridge_cmd,
        "pack", pack_format,
        "--deps-file", deps_filepath,
        "--deps", "dependency01",
        project.path,
    ]

    if platform.system() == 'Darwin':
        cmd.append('--use-docker')

    rc, output = run_command_and_get_output(cmd, cwd=tmpdir)
    assert rc == 1
    assert "You can't specify --deps and --deps-file flags at the same time" in output


@pytest.mark.parametrize('pack_format', ['deb', 'rpm'])
def test_dependencies_file_not_exist(cartridge_cmd, project_without_dependencies, pack_format, tmpdir):
    project = project_without_dependencies

    cmd = [
        cartridge_cmd,
        "pack", pack_format,
        "--deps-file", "not_exist_file.txt",
        project.path,
    ]

    if platform.system() == 'Darwin':
        cmd.append('--use-docker')

    rc, output = run_command_and_get_output(cmd, cwd=tmpdir)
    assert rc == 1
    assert "Invalid path to file with dependencies" in output


@pytest.mark.parametrize('pack_format', ['deb', 'rpm'])
def test_broken_dependencies(cartridge_cmd, project_without_dependencies, pack_format, tmpdir):
    project = project_without_dependencies

    broken_filepath = os.path.join(tmpdir, "broken.txt")
    with open(broken_filepath, "w") as f:
        f.write("dep01 >= 14, <= 25, > 14\n")

    cmd = [
        cartridge_cmd,
        "pack", pack_format,
        "--deps-file", broken_filepath,
        project.path,
    ]

    if platform.system() == 'Darwin':
        cmd.append('--use-docker')

    error_message = "Failed to parse dependencies file: Error during parse dependencies file"
    rc, output = run_command_and_get_output(cmd, cwd=tmpdir)
    assert rc == 1
    assert error_message in output


@pytest.mark.parametrize('pack_format', ['deb', 'rpm'])
def test_pre_and_post_install_scripts(cartridge_cmd, project_without_dependencies, pack_format, tmpdir):
    project = project_without_dependencies

    pre_install_script = os.path.join(tmpdir, "pre.sh")
    with open(pre_install_script, "w") as f:
        f.write("""
                /bin/sh -c 'touch $HOME/hello-bin-sh.txt'
                /bin/touch $HOME/hello-absolute.txt
                """)

    post_install_script = os.path.join(tmpdir, "post.sh")
    with open(post_install_script, "w") as f:
        f.write("""
                /bin/sh -c 'touch $HOME/bye-bin-sh.txt'
                /bin/touch $HOME/bye-absolute.txt
                """)

    cmd = [
        cartridge_cmd,
        "pack", pack_format,
        "--preinst", pre_install_script,
        "--postinst", post_install_script,
        project.path,
    ]

    if platform.system() == 'Darwin':
        cmd.append('--use-docker')

    warning_message = "You specified flag for pre/post install script, " \
                      "but you are not packaging RPM or DEB. Flag will be ignored"

    rc, output = run_command_and_get_output(cmd, cwd=tmpdir)
    warning_message not in output
    assert rc == 0

    if pack_format == 'rpm':
        assert_pre_and_post_install_scripts_rpm(find_archive(tmpdir, project.name, 'rpm'),
                                                pre_install_script, post_install_script)
    else:
        assert_pre_and_post_install_scripts_deb(find_archive(tmpdir, project.name, 'deb'),
                                                pre_install_script, post_install_script, tmpdir)


@pytest.mark.parametrize('pack_format', ['deb', 'rpm'])
def test_pre_and_post_install_scripts_default_files(cartridge_cmd, project_without_dependencies, pack_format, tmpdir):
    project = project_without_dependencies

    pre_install_script = os.path.join(tmpdir, "pre.sh")
    with open(pre_install_script, "w") as f:
        f.write("""
                /bin/sh -c 'touch $HOME/hello-bin-sh.txt'
                /bin/touch $HOME/hello-absolute.txt
                """)

    replace_project_file(project, 'preinst.sh', pre_install_script)

    post_install_script = os.path.join(tmpdir, "postinst.sh")
    with open(post_install_script, "w") as f:
        f.write("""
                /bin/sh -c 'touch $HOME/bye-bin-sh.txt'
                /bin/touch $HOME/bye-absolute.txt
                """)

    replace_project_file(project, 'postinst.sh', post_install_script)

    cmd = [
        cartridge_cmd,
        "pack", pack_format,
        project.path,
    ]

    if platform.system() == 'Darwin':
        cmd.append('--use-docker')

    warning_message = "You specified flag for pre/post install script, " \
                      "but you are not packaging RPM or DEB. Flag will be ignored"

    rc, output = run_command_and_get_output(cmd, cwd=tmpdir)
    warning_message not in output
    assert rc == 0

    if pack_format == 'rpm':
        assert_pre_and_post_install_scripts_rpm(find_archive(tmpdir, project.name, 'rpm'),
                                                pre_install_script, post_install_script)
    else:
        assert_pre_and_post_install_scripts_deb(find_archive(tmpdir, project.name, 'deb'),
                                                pre_install_script, post_install_script, tmpdir)


@pytest.mark.parametrize('pack_format', ['deb', 'rpm'])
def test_pre_and_post_install_scripts_file_not_exist(cartridge_cmd, project_without_dependencies, pack_format, tmpdir):
    project = project_without_dependencies

    cmd = [
        cartridge_cmd,
        "pack", pack_format,
        "--preinst", "not_exist_file.txt",
        project.path,
    ]

    if platform.system() == 'Darwin':
        cmd.append('--use-docker')

    rc, output = run_command_and_get_output(cmd, cwd=tmpdir)
    assert rc == 1
    assert "Specified pre-install script not_exist_file.txt doesn't exists" in output

    cmd = [
        cartridge_cmd,
        "pack", pack_format,
        "--postinst", "not_exist_file.txt",
        project.path,
    ]

    if platform.system() == 'Darwin':
        cmd.append('--use-docker')

    rc, output = run_command_and_get_output(cmd, cwd=tmpdir)
    assert rc == 1
    assert "Specified post-install script not_exist_file.txt doesn't exists" in output


@pytest.mark.parametrize('pack_format', ['docker', 'tgz'])
def test_pre_and_post_install_scripts_not_rpm_deb(cartridge_cmd, project_without_dependencies, pack_format, tmpdir):
    project = project_without_dependencies

    pre_install_script = os.path.join(tmpdir, "pre.sh")
    with open(pre_install_script, "w") as f:
        f.write("""
                /bin/sh -c 'touch $HOME/hello-bin-sh.txt'
                /bin/touch $HOME/hello-absolute.txt
                """)

    post_install_script = os.path.join(tmpdir, "post.sh")
    with open(post_install_script, "w") as f:
        f.write("""
                /bin/sh -c 'touch $HOME/bye-bin-sh.txt'
                /bin/touch $HOME/bye-absolute.txt
                """)

    cmd = [
        cartridge_cmd,
        "pack", pack_format,
        "--preinst", pre_install_script,
        project.path,
    ]

    warning_message = "You specified flag for pre/post install script, " \
                      "but you are not packaging RPM or DEB. Flag will be ignored"
    rc, output = run_command_and_get_output(cmd, cwd=tmpdir)
    assert rc == 0
    assert warning_message in output

    cmd = [
        cartridge_cmd,
        "pack", pack_format,
        "--postinst", post_install_script,
        project.path,
    ]

    rc, output = run_command_and_get_output(cmd, cwd=tmpdir)
    assert rc == 0
    assert warning_message in output


@pytest.mark.parametrize('pack_format', ['deb', 'rpm'])
def test_version_file(cartridge_cmd, project_without_dependencies, tmpdir, pack_format):
    project = project_without_dependencies
    version_filename = 'VERSION.lua'
    app_version = '1.2.3'

    cmd = [
        cartridge_cmd,
        "pack", pack_format,
        "--version", app_version,
        project.path,
    ]

    if platform.system() == 'Darwin':
        cmd.append('--use-docker')

    rc, output = run_command_and_get_output(cmd, cwd=tmpdir)
    assert rc == 0
    assert f'Generate {version_filename} file' in output

    archive_path = find_archive(tmpdir, project.name, pack_format)
    extract_dir = os.path.join(tmpdir, 'extract')
    extract_app_files(archive_path, pack_format, extract_dir)

    version_lua_filepath = os.path.join(extract_dir, 'usr', 'share', 'tarantool', project.name, 'VERSION.lua')
    assert os.path.exists(version_lua_filepath)

    with open(version_lua_filepath, 'r') as f:
        assert f.read() == f"return '{app_version}-0'"


@pytest.mark.parametrize('pack_format', ['deb', 'rpm'])
def test_overwritten_version_file(cartridge_cmd, project_without_dependencies, tmpdir, pack_format):
    project = project_without_dependencies
    version_filename = 'VERSION.lua'
    app_version = '1.2.3'

    cmd = [
        cartridge_cmd,
        "pack", pack_format,
        "--version", app_version,
        project.path,
    ]

    with open(os.path.join(project.path, version_filename), 'w') as f:
        f.write('dummy text')

    if platform.system() == 'Darwin':
        cmd.append('--use-docker')

    rc, output = run_command_and_get_output(cmd, cwd=tmpdir)
    assert rc == 0
    assert f'Generate {version_filename} file' in output
    assert f'File {version_filename} will be overwritten' in output

    archive_path = find_archive(tmpdir, project.name, pack_format)
    extract_dir = os.path.join(tmpdir, 'extract')
    extract_app_files(archive_path, pack_format, extract_dir)

    version_lua_filepath = os.path.join(extract_dir, 'usr', 'share', 'tarantool', project.name, 'VERSION.lua')
    assert os.path.exists(version_lua_filepath)

    with open(version_lua_filepath, 'r') as f:
        assert f.read() == f"return '{app_version}-0'"


@pytest.mark.parametrize('pack_format', ['rpm', 'deb'])
def test_fd_limit_default_file(cartridge_cmd, project_without_dependencies, pack_format, tmpdir):
    project = project_without_dependencies

    fd_limit = 1024
    stateboard_fd_limit = 2048

    systemd_unit_params = os.path.join(tmpdir, "systemd-unit-params.yml")
    with open(systemd_unit_params, "w") as f:
        f.write(f"""
                fd-limit: {fd_limit}
                stateboard-fd-limit: {stateboard_fd_limit}
                """)

    replace_project_file(project, 'systemd-unit-params.yml', systemd_unit_params)

    cmd = [
        cartridge_cmd,
        "pack", pack_format,
        project.path,
    ]

    if platform.system() == 'Darwin':
        cmd.append('--use-docker')

    process = subprocess.run(cmd, cwd=tmpdir)
    assert process.returncode == 0

    check_fd_limits_in_unit_files(fd_limit, stateboard_fd_limit, project.name, pack_format, tmpdir)


@pytest.mark.parametrize('pack_format', ['rpm', 'deb'])
def test_fd_limit_specified_with_flag(cartridge_cmd, project_without_dependencies, pack_format, tmpdir):
    project = project_without_dependencies

    fd_limit = 1024
    stateboard_fd_limit = 2048

    systemd_unit_params = os.path.join(tmpdir, "not-default-systemd-unit-params.yml")
    with open(systemd_unit_params, "w") as f:
        f.write(f"""
                fd-limit: {fd_limit}
                stateboard-fd-limit: {stateboard_fd_limit}
                """)

    cmd = [
        cartridge_cmd,
        "pack", pack_format,
        "--unit-params-file", systemd_unit_params,
        project.path,
    ]

    if platform.system() == 'Darwin':
        cmd.append('--use-docker')

    process = subprocess.run(cmd, cwd=tmpdir)
    assert process.returncode == 0

    check_fd_limits_in_unit_files(fd_limit, stateboard_fd_limit, project.name, pack_format, tmpdir)


@pytest.mark.parametrize('pack_format', ['rpm', 'deb'])
def test_fd_limit_default_values(cartridge_cmd, project_without_dependencies, pack_format, tmpdir):
    project = project_without_dependencies

    default_fd_limit = 65535
    default_stateboard_fd_limit = 65535

    cmd = [
        cartridge_cmd,
        "pack", pack_format,
        project.path,
    ]

    if platform.system() == 'Darwin':
        cmd.append('--use-docker')

    process = subprocess.run(cmd, cwd=tmpdir)
    assert process.returncode == 0

    check_fd_limits_in_unit_files(default_fd_limit, default_stateboard_fd_limit, project.name, pack_format, tmpdir)


@pytest.mark.parametrize('pack_format', ['rpm', 'deb'])
def test_fd_limit_without_default_file(cartridge_cmd, project_without_dependencies, pack_format, tmpdir):
    project = project_without_dependencies

    default_fd_limit = 65535
    default_stateboard_fd_limit = 65535

    remove_project_file(project, 'systemd-unit-params.yml')

    cmd = [
        cartridge_cmd,
        "pack", pack_format,
        project.path,
    ]

    if platform.system() == 'Darwin':
        cmd.append('--use-docker')

    process = subprocess.run(cmd, cwd=tmpdir)
    assert process.returncode == 0

    check_fd_limits_in_unit_files(default_fd_limit, default_stateboard_fd_limit, project.name, pack_format, tmpdir)


@pytest.mark.parametrize('pack_format', ['rpm', 'deb'])
def test_fd_limit_file_not_exist(cartridge_cmd, project_without_dependencies, pack_format, tmpdir):
    project = project_without_dependencies

    not_exist_file_path = "not_exist_file_path.yml"

    cmd = [
        cartridge_cmd,
        "pack", pack_format,
        "--unit-params-file", not_exist_file_path,
        project.path,
    ]

    if platform.system() == 'Darwin':
        cmd.append('--use-docker')

    rc, output = run_command_and_get_output(cmd, cwd=tmpdir)
    assert rc == 1
    assert f"Specified file with system unit params {not_exist_file_path} doesn't exists" in output


@pytest.mark.parametrize('pack_format', ['rpm', 'deb'])
def test_fd_limit_invalid_values(cartridge_cmd, project_without_dependencies, pack_format, tmpdir):
    project = project_without_dependencies

    fd_limit = -1

    systemd_unit_params = os.path.join(tmpdir, "systemd-unit-params.yml")
    with open(systemd_unit_params, "w") as f:
        f.write(f"""
                fd-limit: {fd_limit}
                """)

    replace_project_file(project, 'systemd-unit-params.yml', systemd_unit_params)

    cmd = [
        cartridge_cmd,
        "pack", pack_format,
        project.path,
    ]

    if platform.system() == 'Darwin':
        cmd.append('--use-docker')

    rc, output = run_command_and_get_output(cmd, cwd=tmpdir)
    assert rc == 1
    assert "Incorrect value for fd-limit: minimal value is 1024" in output

    stateboard_fd_limit = -2

    systemd_unit_params = os.path.join(tmpdir, "systemd-unit-params.yml")
    with open(systemd_unit_params, "w") as f:
        f.write(f"""
                stateboard-fd-limit: {stateboard_fd_limit}
                """)

    replace_project_file(project, 'systemd-unit-params.yml', systemd_unit_params)

    cmd = [
        cartridge_cmd,
        "pack", pack_format,
        project.path,
    ]

    if platform.system() == 'Darwin':
        cmd.append('--use-docker')

    rc, output = run_command_and_get_output(cmd, cwd=tmpdir)
    assert rc == 1
    assert "Incorrect value for stateboard-fd-limit: minimal value is 1024" in output


@pytest.mark.parametrize('pack_format', ['deb', 'rpm'])
def test_paths_caching(cartridge_cmd, light_project, tmpdir, pack_format):
    project = light_project
    project_dir = hashlib.sha1(project.path.encode('utf-8')).hexdigest()[:10]
    clear_project_rocks_cache(project_dir)

    cmd = [
        cartridge_cmd,
        "pack", pack_format,
        project.path,
    ]

    if platform.system() == 'Darwin':
        cmd.append('--use-docker')

    rc, output = run_command_and_get_output(cmd, cwd=tmpdir)
    assert rc == 0
    assert "Using cached path .rocks" not in output

    # Checking rocks are caching
    rc, output = run_command_and_get_output(cmd, cwd=tmpdir)
    assert rc == 0
    assert "Using cached path .rocks" in output

    project_path_cache = os.path.join(get_rocks_cache_path(), project_dir, ".rocks")
    cache_dir_items = os.listdir(project_path_cache)
    assert len(cache_dir_items) == 1

    # Changing rockspec file -> changing hash
    with open(os.path.join(project.path, f"{project.name}-scm-1.rockspec"), "a") as f:
        f.write("\n\n")

    rc, output = run_command_and_get_output(cmd, cwd=tmpdir)
    assert rc == 0
    assert "Using cached path .rocks" not in output

    # Сheck that only one rocks cache is saved for one project
    new_cache_dir_items = os.listdir(project_path_cache)
    assert len(new_cache_dir_items) == 1
    assert cache_dir_items != new_cache_dir_items


@pytest.mark.parametrize('pack_format', ['tgz'])
def test_paths_cache_evicting(cartridge_cmd, light_project, tmpdir, pack_format):
    project = light_project
    project_dir = hashlib.sha1(project.path.encode('utf-8')).hexdigest()[:10]

    cache_path = get_rocks_cache_path()
    if os.path.exists(cache_path):
        shutil.rmtree(cache_path)

    max_cache_size = 5

    for i in range(max_cache_size):
        os.makedirs(os.path.join(cache_path, str(i)))

    cache_dir_items = os.listdir(cache_path)
    assert len(cache_dir_items) == 5

    cmd = [
        cartridge_cmd,
        "pack", pack_format,
        project.path,
    ]

    if platform.system() == 'Darwin':
        cmd.append('--use-docker')

    rc, output = run_command_and_get_output(cmd, cwd=tmpdir)
    assert rc == 0
    assert "Using cached path .rocks" not in output

    new_cache_dir_items = os.listdir(cache_path)
    assert len(new_cache_dir_items) == 5
    assert project_dir in new_cache_dir_items
    assert "0" not in new_cache_dir_items


@pytest.mark.parametrize('pack_format', ['tgz'])
def test_paths_noncache_flag(cartridge_cmd, light_project, tmpdir, pack_format):
    project = light_project

    if os.path.exists(get_rocks_cache_path()):
        shutil.rmtree(get_rocks_cache_path())

    cmd = [
        cartridge_cmd,
        "pack", pack_format,
        project.path, "--no-cache"
    ]

    if platform.system() == 'Darwin':
        cmd.append('--use-docker')

    rc, output = run_command_and_get_output(cmd, cwd=tmpdir)
    assert rc == 0
    assert "Using cached path .rocks" not in output
    assert len(os.listdir(get_rocks_cache_path())) == 0

    rc, output = run_command_and_get_output(cmd, cwd=tmpdir)
    assert rc == 0
    assert "Using cached path .rocks" not in output
    assert len(os.listdir(get_rocks_cache_path())) == 0


@pytest.mark.parametrize('pack_format', ['tgz'])
def test_no_cache_yml_file(cartridge_cmd, light_project, tmpdir, pack_format):
    project = light_project
    remove_project_file(project, "pack-cache-config.yml")

    if os.path.exists(get_rocks_cache_path()):
        shutil.rmtree(get_rocks_cache_path())

    cmd = [
        cartridge_cmd,
        "pack", pack_format,
        project.path
    ]

    if platform.system() == 'Darwin':
        cmd.append('--use-docker')

    rc, output = run_command_and_get_output(cmd, cwd=tmpdir)
    assert rc == 0
    assert "Failed to process pack-cache-config.yml file which contain cache paths" not in output
    assert len(os.listdir(get_rocks_cache_path())) == 0


@pytest.mark.parametrize('pack_format', ['tgz'])
def test_always_cache_path(cartridge_cmd, light_project, tmpdir, pack_format):
    project = light_project
    project_dir = hashlib.sha1(project.path.encode('utf-8')).hexdigest()[:10]

    new_cache_yml_path = os.path.join(tmpdir, "new-pack-cache.yml")
    with open(new_cache_yml_path, "w") as f:
        yaml.dump([{"path": ".rocks", "always-cache": True}], f)

    replace_project_file(project, "pack-cache-config.yml", new_cache_yml_path)
    if os.path.exists(get_rocks_cache_path()):
        shutil.rmtree(get_rocks_cache_path())

    cmd = [
        cartridge_cmd,
        "pack", pack_format,
        project.path
    ]

    if platform.system() == 'Darwin':
        cmd.append('--use-docker')

    rc, output = run_command_and_get_output(cmd, cwd=tmpdir)
    assert rc == 0
    assert "Using cached path .rocks" not in output

    project_path_cache = os.path.join(get_rocks_cache_path(), project_dir, ".rocks")
    cache_dir_items = os.listdir(project_path_cache)
    assert len(cache_dir_items) == 1

    rc, output = run_command_and_get_output(cmd, cwd=tmpdir)
    assert rc == 0
    assert "Using cached path .rocks" in output

    project_path_cache = os.path.join(get_rocks_cache_path(), project_dir, ".rocks")
    new_cache_dir_items = os.listdir(project_path_cache)
    assert len(new_cache_dir_items) == 1
    assert cache_dir_items == new_cache_dir_items


@pytest.mark.parametrize('pack_format', ['tgz'])
def test_multiple_paths_in_pack_file(cartridge_cmd, light_project, tmpdir, pack_format):
    project = light_project
    project_dir = hashlib.sha1(project.path.encode('utf-8')).hexdigest()[:10]
    project_path_cache = os.path.join(get_rocks_cache_path(), project_dir)

    first_cached_dir_path = os.path.join(project.path, "first-dir")
    second_cached_dir_path = os.path.join(project.path, "second-dir", "nested")
    os.mkdir(first_cached_dir_path)
    os.makedirs(second_cached_dir_path)

    with open(os.path.join(first_cached_dir_path, "fst.txt"), "w") as f:
        f.write("Dummy text 123")

    with open(os.path.join(second_cached_dir_path, "snd.txt"), "w") as f:
        f.write("321 txet ymmuD")

    new_cache_yml_path = os.path.join(tmpdir, "new-pack-cache.yml")
    with open(new_cache_yml_path, "w") as f:
        rockspec_path = get_rockspec_path(project.path, project.name, "scm-1")
        yaml.dump([
            {"path": ".rocks", "always-cache": True},
            {"path": "first-dir", "key": "dummy-key"},
            {"path": "second-dir/nested", "key-path": os.path.basename(rockspec_path)},
        ], f)

    replace_project_file(project, "pack-cache-config.yml", new_cache_yml_path)

    if os.path.exists(get_rocks_cache_path()):
        shutil.rmtree(get_rocks_cache_path())

    cmd = [
        cartridge_cmd,
        "pack", pack_format,
        project.path
    ]

    if platform.system() == 'Darwin':
        cmd.append('--use-docker')

    rc, output = run_command_and_get_output(cmd, cwd=tmpdir)
    assert rc == 0
    assert "Using cached path .rocks" not in output
    assert "Using cached path first-dir" not in output
    assert "Using cached path second-dir/nested" not in output

    cache_dir_items = os.listdir(project_path_cache)
    assert len(cache_dir_items) == 3

    nested_dir_items = os.listdir(os.path.join(project_path_cache, "second-dir/nested"))
    assert len(nested_dir_items) == 1

    rc, output = run_command_and_get_output(cmd, cwd=tmpdir)
    assert rc == 0
    assert "Using cached path .rocks" in output
    assert "Using cached path first-dir" in output
    assert "Using cached path second-dir/nested" in output

    new_cache_dir_items = os.listdir(project_path_cache)
    assert len(new_cache_dir_items) == 3
    assert cache_dir_items == new_cache_dir_items


@pytest.mark.parametrize('pack_format', ['tgz'])
def test_invalid_yml_params(cartridge_cmd, light_project, tmpdir, pack_format):
    project = light_project

    # Combine always-cache and key params
    new_cache_yml_path = os.path.join(tmpdir, "new-pack-cache.yml")
    with open(new_cache_yml_path, "w") as f:
        yaml.dump([{"path": ".rocks", "always-cache": True, "key": "just-key"}], f)

    replace_project_file(project, "pack-cache-config.yml", new_cache_yml_path)
    if os.path.exists(get_rocks_cache_path()):
        shutil.rmtree(get_rocks_cache_path())

    cmd = [
        cartridge_cmd,
        "pack", pack_format,
        project.path
    ]

    if platform.system() == 'Darwin':
        cmd.append('--use-docker')

    rc, output = run_command_and_get_output(cmd, cwd=tmpdir)
    assert rc == 1
    assert "Please, specify one and only one of `always-true`, `key` and `key-path` for path .rocks" in output
    assert len(os.listdir(get_rocks_cache_path())) == 0

    # Combine key-path and key params
    with open(new_cache_yml_path, "w") as f:
        yaml.dump([{
            "path": ".rocks",
            "key-path": os.path.basename(get_rockspec_path(project.path, project.name, "scm-1")),
            "key": "just-key",
        }], f)

    replace_project_file(project, "pack-cache-config.yml", new_cache_yml_path)
    rc, output = run_command_and_get_output(cmd, cwd=tmpdir)
    assert rc == 1
    assert "Please, specify one and only one of `always-true`, `key` and `key-path` for path .rocks" in output
    assert len(os.listdir(get_rocks_cache_path())) == 0

    # Invalid key-path file (non exists)
    with open(new_cache_yml_path, "w") as f:
        yaml.dump([{"path": ".rocks", "key-path": "invalid_path"}], f)

    replace_project_file(project, "pack-cache-config.yml", new_cache_yml_path)
    rc, output = run_command_and_get_output(cmd, cwd=tmpdir)
    assert rc == 1
    assert "Failed to get specified cache key file for path .rocks" in output
    assert len(os.listdir(get_rocks_cache_path())) == 0

    # always-cache: false without any keys
    with open(new_cache_yml_path, "w") as f:
        yaml.dump([{"path": ".rocks", "always-cache": False}], f)

    replace_project_file(project, "pack-cache-config.yml", new_cache_yml_path)
    rc, output = run_command_and_get_output(cmd, cwd=tmpdir)
    assert rc == 1
    assert "Please, specify one and only one of `always-true`, `key` and `key-path` for path .rocks" in output
    assert len(os.listdir(get_rocks_cache_path())) == 0


@pytest.mark.parametrize('pack_format', ['tgz'])
def test_multiple_specified_path(cartridge_cmd, project_without_dependencies, tmpdir, pack_format):
    project = project_without_dependencies

    new_cache_yml_path = os.path.join(tmpdir, "new-pack-cache.yml")
    with open(new_cache_yml_path, "w") as f:
        yaml.dump([{"path": ".rocks", "always-cache": True}, {"path": ".rocks", "key": "simple-key"}], f)

    replace_project_file(project, "pack-cache-config.yml", new_cache_yml_path)
    if os.path.exists(get_rocks_cache_path()):
        shutil.rmtree(get_rocks_cache_path())

    cmd = [
        cartridge_cmd,
        "pack", pack_format,
        project.path
    ]

    if platform.system() == 'Darwin':
        cmd.append('--use-docker')

    rc, output = run_command_and_get_output(cmd, cwd=tmpdir)
    assert rc == 1
    assert "Cache path .rocks specified multiple times" in output
    assert len(os.listdir(get_rocks_cache_path())) == 0


@pytest.mark.parametrize('pack_format', ['tgz'])
def test_path_is_not_directory(cartridge_cmd, light_project, tmpdir, pack_format):
    project = light_project
    project_dir = hashlib.sha1(project.path.encode('utf-8')).hexdigest()[:10]
    project_path_cache = os.path.join(get_rocks_cache_path(), project_dir)

    shutil.make_archive(
        os.path.join(project.path, 'nested/dir/test/zip_arch'),
        'zip', os.path.join(project.path, "app")
    )

    new_cache_yml_path = os.path.join(tmpdir, "new-pack-cache.yml")
    with open(new_cache_yml_path, "w") as f:
        rockspec_path = get_rockspec_path(project.path, project.name, "scm-1")
        yaml.dump([
            {"path": "instances.yml", "always-cache": True},
            {"path": "nested/dir/test/zip_arch.zip", "key-path": os.path.basename(rockspec_path)},
            {"path": ".rocks", "key": "simple-key"},
        ], f)

    new_instances_yml = os.path.join(tmpdir, "new-instances.yml")
    with open(new_instances_yml, "w") as f:
        f.write("Dummy text")

    replace_project_file(project, "pack-cache-config.yml", new_cache_yml_path)
    replace_project_file(project, "instances.yml", new_instances_yml)

    if os.path.exists(get_rocks_cache_path()):
        shutil.rmtree(get_rocks_cache_path())

    cmd = [
        cartridge_cmd,
        "pack", pack_format,
        project.path
    ]

    if platform.system() == 'Darwin':
        cmd.append('--use-docker')

    rc, output = run_command_and_get_output(cmd, cwd=tmpdir)
    assert rc == 0
    assert "Using cached path .rocks" not in output
    assert "Using cached path instances.yml" not in output
    assert "Using cached path nested/dir/test/zip_arch.zip" not in output

    cache_items = os.listdir(project_path_cache)
    assert len(cache_items) == 3
    assert ".rocks" in cache_items
    assert "instances.yml" in cache_items
    assert "nested" in cache_items

    nested_dir_items = os.listdir(os.path.join(project_path_cache, "nested/dir/test"))
    assert len(nested_dir_items) == 1

    instances_cached_path = os.path.join(project_path_cache, "instances.yml", "always", "instances.yml")
    assert os.path.exists(instances_cached_path)
    with open(instances_cached_path) as f:
        assert f.read() == "Dummy text"

    rc, output = run_command_and_get_output(cmd, cwd=tmpdir)
    assert rc == 0
    assert "Using cached path .rocks" in output
    assert "Using cached path instances.yml" in output
    assert "Using cached path nested/dir/test/zip_arch.zip" in output


@pytest.mark.parametrize('pack_format', ['rpm', 'deb'])
def test_net_msg_max_specified(cartridge_cmd, project_without_dependencies, pack_format, tmpdir):
    project = project_without_dependencies

    instance_net_msg_max = 1024
    stateboard_net_msg_max = 2048

    systemd_unit_params = os.path.join(tmpdir, "systemd-unit-params.yml")
    with open(systemd_unit_params, "w") as f:
        f.write(f"""
                 instance-env:
                     net-msg-max: {instance_net_msg_max}
                 stateboard-env:
                     net-msg-max: {stateboard_net_msg_max}
                 """)

    cmd = [
        cartridge_cmd,
        "pack", pack_format,
        "--unit-params-file", systemd_unit_params,
        project.path,
    ]

    if platform.system() == 'Darwin':
        cmd.append('--use-docker')

    rc, output = run_command_and_get_output(cmd, cwd=tmpdir)
    assert rc == 0

    check_param_in_unit_files(instance_net_msg_max, stateboard_net_msg_max,
                              "Environment=TARANTOOL_NET_MSG_MAX",
                              project.name, pack_format, tmpdir)


@pytest.mark.parametrize('pack_format', ['rpm', 'deb'])
def test_net_msg_max_invalid_type(cartridge_cmd, project_without_dependencies, pack_format, tmpdir):
    project = project_without_dependencies

    invalid_net_msg_max = "string_value"

    systemd_unit_params = os.path.join(tmpdir, "systemd-unit-params.yml")
    with open(systemd_unit_params, "w") as f:
        f.write(f"""
                 instance-env:
                     net-msg-max: {invalid_net_msg_max}
                 """)

    cmd = [
        cartridge_cmd,
        "pack", pack_format,
        "--unit-params-file", systemd_unit_params,
        project.path,
    ]

    if platform.system() == 'Darwin':
        cmd.append('--use-docker')

    error_message = "net-msg-max parameter type should be integer"
    rc, output = run_command_and_get_output(cmd, cwd=tmpdir)
    assert rc == 1
    assert error_message in output


@pytest.mark.parametrize('pack_format', ['rpm', 'deb'])
def test_net_msg_max_invalid_value(cartridge_cmd, project_without_dependencies, pack_format, tmpdir):
    project = project_without_dependencies

    invalid_net_msg_max = -1

    systemd_unit_params = os.path.join(tmpdir, "systemd-unit-params.yml")
    with open(systemd_unit_params, "w") as f:
        f.write(f"""
                 instance-env:
                     net-msg-max: {invalid_net_msg_max}
                 """)

    cmd = [
        cartridge_cmd,
        "pack", pack_format,
        "--unit-params-file", systemd_unit_params,
        project.path,
    ]

    if platform.system() == 'Darwin':
        cmd.append('--use-docker')

    error_message = "Incorrect value for net-msg-max: minimal value is 2"
    rc, output = run_command_and_get_output(cmd, cwd=tmpdir)
    assert rc == 1
    assert error_message in output
