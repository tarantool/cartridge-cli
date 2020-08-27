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
        cmd.append('--use-docker')

    process = subprocess.run(cmd, cwd=tmpdir)
    assert process.returncode == 0, \
        "Error during creating of deb archive with project"

    filepath = find_archive(tmpdir, project.name, 'deb')
    assert filepath is not None, "DEB archive isn't found in work directory"

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


def extract_app_files(archive_path, pack_format, extract_dir):
    os.makedirs(extract_dir)

    if pack_format == 'rpm':
        extract_rpm(archive_path, extract_dir)
    elif pack_format == 'deb':
        extract_deb(archive_path, extract_dir)
        with tarfile.open(name=os.path.join(extract_dir, 'data.tar.gz')) as data_arch:
            data_arch.extractall(path=extract_dir)


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
    rc, output = run_command_and_get_output(cmd, cwd=tmpdir)
    assert rc == 1
    assert re.search(r'Failed to read specified .*unit template', output) is not None

    # pass correct path
    process = subprocess.run([
            cartridge_cmd, "pack", pack_format,
            "--%s-template" % unit, unit_template_filepath,
            project.path
        ],
        cwd=tmpdir
    )
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


@pytest.mark.parametrize('pack_format', ['tgz'])
def test_packing_without_path_specifying(cartridge_cmd, project_without_dependencies, pack_format, tmpdir):
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
def test_invalid_base_build_dockerfile(cartridge_cmd, project_without_dependencies, pack_format, tmpdir):
    bad_dockerfiles = [
        "FROM ubuntu:xenial\n",
        "I am FROM centos:8",
    ]

    invalid_dockerfile_path = os.path.join(tmpdir, 'Dockerfile')
    for bad_dockerfile in bad_dockerfiles:
        with open(invalid_dockerfile_path, 'w') as f:
            f.write(bad_dockerfile)

        cmd = [
            cartridge_cmd,
            "pack", pack_format,
            "--use-docker",
            "--build-from", invalid_dockerfile_path,
            project_without_dependencies.path,
        ]

        rc, output = run_command_and_get_output(cmd, cwd=tmpdir)
        assert rc == 1
        assert 'Invalid base build Dockerfile' in output
        assert 'base image must be centos:8' in output


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
