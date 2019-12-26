#!/usr/bin/python3

import os
import pytest
import subprocess
import tarfile
import shutil

from utils import project_name
from utils import basepath
from utils import tarantool_enterprise_is_used
from utils import find_archive
from utils import recursive_listdir
from utils import assert_dir_contents
from utils import assert_filemodes
from utils import assert_files_mode_and_owner_rpm
from utils import validate_version_file
from utils import check_package_files
from utils import assert_tarantool_dependency_deb
from utils import assert_tarantool_dependency_rpm


# ########
# Fixtures
# ########
@pytest.fixture(scope="module")
def tgz_archive(module_tmpdir, project_path, prepare_ignore):
    cmd = [os.path.join(basepath, "cartridge"), "pack", "tgz", project_path]
    process = subprocess.run(cmd, cwd=module_tmpdir)
    assert process.returncode == 0, \
        "Error during creating of tgz archive with project"

    archive_name = find_archive(module_tmpdir, 'tar.gz')
    assert archive_name is not None, "TGZ archive isn't founded in work directory"

    return {'name': archive_name}


@pytest.fixture(scope="module")
def rpm_archive(module_tmpdir, project_path, prepare_ignore):
    cmd = [os.path.join(basepath, "cartridge"), "pack", "rpm", project_path]
    process = subprocess.run(cmd, cwd=module_tmpdir)
    assert process.returncode == 0, \
        "Error during creating of rpm archive with project"

    archive_name = find_archive(module_tmpdir, 'rpm')
    assert archive_name is not None, "RPM archive isn't founded in work directory"

    return {'name': archive_name}


@pytest.fixture(scope="module")
def deb_archive(module_tmpdir, project_path, prepare_ignore):
    cmd = [os.path.join(basepath, "cartridge"), "pack", "deb", project_path]
    process = subprocess.run(cmd, cwd=module_tmpdir)
    assert process.returncode == 0, \
        "Error during creating of deb archive with project"

    archive_name = find_archive(module_tmpdir, 'deb')
    assert archive_name is not None, "DEB archive isn't found in work directory"

    return {'name': archive_name}


@pytest.fixture(scope="module")
def rpm_archive_with_custom_units(module_tmpdir, project_path, prepare_ignore):
    unit_template = '''
[Unit]
Description=Tarantool service: ${name}
SIMPLE_UNIT_TEMPLATE
[Service]
Type=simple
ExecStart=${dir}/tarantool ${dir}/init.lua

Environment=TARANTOOL_WORK_DIR=${workdir}
Environment=TARANTOOL_CONSOLE_SOCK=/var/run/tarantool/${name}.control
Environment=TARANTOOL_PID_FILE=/var/run/tarantool/${name}.pid
Environment=TARANTOOL_INSTANCE_NAME=${name}

[Install]
WantedBy=multi-user.target
Alias=${name}
    '''

    instantiated_unit_template = '''
[Unit]
Description=Tarantool service: ${name} %i
INSTANTIATED_UNIT_TEMPLATE

[Service]
Type=simple
ExecStartPre=mkdir -p ${workdir}.%i
ExecStart=${dir}/tarantool ${dir}/init.lua

Environment=TARANTOOL_WORK_DIR=${workdir}.%i
Environment=TARANTOOL_CONSOLE_SOCK=/var/run/tarantool/${name}.%i.control
Environment=TARANTOOL_PID_FILE=/var/run/tarantool/${name}.%i.pid
Environment=TARANTOOL_INSTANCE_NAME=${name}@%i

[Install]
WantedBy=multi-user.target
Alias=${name}
    '''
    unit_template_filepath = os.path.join(module_tmpdir, "unit_template.tmpl")
    with open(unit_template_filepath, 'w') as f:
        f.write(unit_template)

    inst_unit_template_filepath = os.path.join(module_tmpdir, "instantiated_unit_template.tmpl")
    with open(inst_unit_template_filepath, 'w') as f:
        f.write(instantiated_unit_template)

    process = subprocess.run([
            os.path.join(basepath, "cartridge"), "pack", "rpm",
            "--unit_template", "unit_template.tmpl",
            "--instantiated_unit_template", "instantiated_unit_template.tmpl",
            project_path
        ],
        cwd=module_tmpdir
    )
    assert process.returncode == 0, \
        "Error during creating of rpm archive with project"

    archive_name = find_archive(module_tmpdir, 'rpm')
    assert archive_name is not None, "RPM archive isn't founded in work directory"

    return {'name': archive_name}


def test_tgz_pack(project_path, tgz_archive, tmpdir):
    with tarfile.open(name=tgz_archive['name']) as tgz_arch:
        # usr/share/tarantool is added to coorectly run assert_file_modes
        distribution_dir = os.path.join(tmpdir, 'usr/share/tarantool', project_name)
        os.makedirs(distribution_dir, exist_ok=True)

        tgz_arch.extractall(path=os.path.join(tmpdir, 'usr/share/tarantool'))
        assert_dir_contents(recursive_listdir(distribution_dir))

        validate_version_file(distribution_dir)
        assert_filemodes(tmpdir)


def test_rpm_pack(project_path, rpm_archive, tmpdir):
    ps = subprocess.Popen(
        ['rpm2cpio', rpm_archive['name']], stdout=subprocess.PIPE)
    subprocess.check_output(['cpio', '-idmv'], stdin=ps.stdout, cwd=tmpdir)
    ps.wait()
    assert ps.returncode == 0, "Error during extracting files from rpm archive"

    if not tarantool_enterprise_is_used():
        assert_tarantool_dependency_rpm(rpm_archive['name'])

    check_package_files(tmpdir, project_path)
    assert_files_mode_and_owner_rpm(rpm_archive['name'])


def test_deb_pack(project_path, deb_archive, tmpdir):
    # unpack ar
    process = subprocess.run([
            'ar', 'x', deb_archive['name']
        ],
        cwd=tmpdir
    )
    assert process.returncode == 0, 'Error during unpacking of deb archive'

    for filename in ['debian-binary', 'control.tar.xz', 'data.tar.xz']:
        assert os.path.exists(os.path.join(tmpdir, filename))

    # check debian-binary
    with open(os.path.join(tmpdir, 'debian-binary')) as debian_binary_file:
        assert debian_binary_file.read() == '2.0\n'

    # check data.tar.xz
    with tarfile.open(name=os.path.join(tmpdir, 'data.tar.xz')) as data_arch:
        data_dir = os.path.join(tmpdir, 'data')
        data_arch.extractall(path=data_dir)
        check_package_files(data_dir, project_path)
        assert_filemodes(data_dir)

    # check control.tar.xz
    with tarfile.open(name=os.path.join(tmpdir, 'control.tar.xz')) as control_arch:
        control_dir = os.path.join(tmpdir, 'control')
        control_arch.extractall(path=control_dir)

        for filename in ['control', 'preinst', 'postinst']:
            assert os.path.exists(os.path.join(control_dir, filename))

        if not tarantool_enterprise_is_used():
            assert_tarantool_dependency_deb(os.path.join(control_dir, 'control'))

        # check if postinst script set owners correctly
        with open(os.path.join(control_dir, 'postinst')) as postinst_script_file:
            postinst_script = postinst_script_file.read()
            assert 'chown -R root:root /usr/share/tarantool/{}'.format(project_name) in postinst_script
            assert 'chown root:root /etc/systemd/system/{}.service'.format(project_name) in postinst_script
            assert 'chown root:root /etc/systemd/system/{}@.service'.format(project_name) in postinst_script
            assert 'chown root:root /usr/lib/tmpfiles.d/{}.conf'.format(project_name) in postinst_script


def test_systemd_units(project_path, rpm_archive_with_custom_units, tmpdir):
    ps = subprocess.Popen(
        ['rpm2cpio', rpm_archive_with_custom_units['name']], stdout=subprocess.PIPE)
    subprocess.check_output(['cpio', '-idmv'], stdin=ps.stdout, cwd=tmpdir)
    ps.wait()
    assert ps.returncode == 0, "Error during extracting files from rpm archive"

    project_unit_file = os.path.join(tmpdir, 'etc/systemd/system', "%s.service" % project_name)
    with open(project_unit_file) as f:
        assert f.read().find('SIMPLE_UNIT_TEMPLATE') != -1

    project_inst_file = os.path.join(tmpdir, 'etc/systemd/system', "%s@.service" % project_name)
    with open(project_inst_file) as f:
        assert f.read().find('INSTANTIATED_UNIT_TEMPLATE') != -1

    project_tmpfiles_conf_file = os.path.join(tmpdir, 'usr/lib/tmpfiles.d', '%s.conf' % project_name)
    with open(project_tmpfiles_conf_file) as f:
        assert f.read().find('d /var/run/tarantool') != -1


def test_packing_without_git(tmpdir):
    # create project and remove .git
    cmd = [
        os.path.join(basepath, "cartridge"), "create",
        "--name", project_name
    ]
    process = subprocess.run(cmd, cwd=tmpdir)
    assert process.returncode == 0, \
        "Error during creating the project"
    project_path = os.path.join(tmpdir, project_name)
    shutil.rmtree(os.path.join(project_path, '.git'))

    # try to build rpm w/o --version
    cmd = [
        os.path.join(basepath, "cartridge"),
        "pack", "rpm",
        project_path,
    ]
    process = subprocess.run(cmd, cwd=tmpdir)
    assert process.returncode == 1

    # remove dependenciies to speed up test
    rockspec_path = os.path.join(project_path, '{}-scm-1.rockspec'.format(project_name))
    with open(rockspec_path, 'w') as f:
        f.write('''
            package = '{}'
            version = 'scm-1'
            source  = {{ url = '/dev/null' }}
            dependencies = {{ 'tarantool' }}
            build = {{ type = 'none' }}
        '''.format(project_name))

    # pass version explicitly
    cmd = [
        os.path.join(basepath, "cartridge"),
        "pack", "rpm",
        "--version", "0.1.0",
        project_path,
    ]
    process = subprocess.run(cmd, cwd=tmpdir)


def test_packing_with_wrong_filemodes(tmpdir):
    project_name = 'test-project'

    # create project
    cmd = [
        os.path.join(basepath, "cartridge"), "create",
        "--name", project_name
    ]
    process = subprocess.run(cmd, cwd=tmpdir)
    assert process.returncode == 0, "Error during creating the project"
    project_path = os.path.join(tmpdir, project_name)

    # add file with invalid (700) mode
    filepath = os.path.join(project_path, 'wrong-mode-file.lua')
    with open(filepath, 'w') as f:
        f.write("return 'Hi'")
    os.chmod(filepath, 0o700)

    # run `cartridge pack`
    cmd = [os.path.join(basepath, "cartridge"), "pack", "rpm", project_path]
    process = subprocess.run(cmd, cwd=tmpdir)
    assert process.returncode == 1, "Packing project with invalid filemode must fail"
