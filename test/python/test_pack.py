#!/usr/bin/python3

import os
import pytest
import subprocess
import tarfile
import shutil

from utils import basepath
from utils import tarantool_enterprise_is_used
from utils import find_archive
from utils import recursive_listdir
from utils import assert_distribution_dir_contents
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
def tgz_archive(module_tmpdir, light_project):
    cmd = [os.path.join(basepath, "cartridge"), "pack", "tgz", light_project.path]
    process = subprocess.run(cmd, cwd=module_tmpdir)
    assert process.returncode == 0, \
        "Error during creating of tgz archive with project"

    archive_name = find_archive(module_tmpdir, light_project.name, 'tar.gz')
    assert archive_name is not None, "TGZ archive isn't found in work directory"

    return {'name': archive_name}


@pytest.fixture(scope="module")
def rpm_archive(module_tmpdir, light_project):
    cmd = [os.path.join(basepath, "cartridge"), "pack", "rpm", light_project.path]
    process = subprocess.run(cmd, cwd=module_tmpdir)
    assert process.returncode == 0, \
        "Error during creating of rpm archive with project"

    archive_name = find_archive(module_tmpdir, light_project.name, 'rpm')
    assert archive_name is not None, "RPM archive isn't found in work directory"

    return {'name': archive_name}


@pytest.fixture(scope="module")
def deb_archive(module_tmpdir, light_project):
    cmd = [os.path.join(basepath, "cartridge"), "pack", "deb", light_project.path]
    process = subprocess.run(cmd, cwd=module_tmpdir)
    assert process.returncode == 0, \
        "Error during creating of deb archive with project"

    archive_name = find_archive(module_tmpdir, light_project.name, 'deb')
    assert archive_name is not None, "DEB archive isn't found in work directory"

    return {'name': archive_name}


@pytest.fixture(scope="module")
def rpm_archive_with_custom_units(module_tmpdir, light_project):
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
            light_project.path
        ],
        cwd=module_tmpdir
    )
    assert process.returncode == 0, \
        "Error during creating of rpm archive with project"

    archive_name = find_archive(module_tmpdir, light_project.name, 'rpm')
    assert archive_name is not None, "RPM archive isn't found in work directory"

    return {'name': archive_name}


# #####
# Tests
# #####
def test_tgz_pack(light_project, tgz_archive, tmpdir):
    with tarfile.open(name=tgz_archive['name']) as tgz_arch:
        # usr/share/tarantool is added to coorectly run assert_filemodes
        distribution_dir = os.path.join(tmpdir, 'usr/share/tarantool', light_project.name)
        os.makedirs(distribution_dir, exist_ok=True)

        tgz_arch.extractall(path=os.path.join(tmpdir, 'usr/share/tarantool'))
        assert_distribution_dir_contents(
            dir_contents=recursive_listdir(distribution_dir),
            project=light_project
        )

        validate_version_file(light_project, distribution_dir)
        assert_filemodes(light_project, tmpdir)


def test_rpm_pack(light_project, rpm_archive, tmpdir):
    ps = subprocess.Popen(
        ['rpm2cpio', rpm_archive['name']], stdout=subprocess.PIPE)
    subprocess.check_output(['cpio', '-idmv'], stdin=ps.stdout, cwd=tmpdir)
    ps.wait()
    assert ps.returncode == 0, "Error during extracting files from rpm archive"

    if not tarantool_enterprise_is_used():
        assert_tarantool_dependency_rpm(rpm_archive['name'])

    check_package_files(light_project, tmpdir)
    assert_files_mode_and_owner_rpm(light_project, rpm_archive['name'])


def test_deb_pack(light_project, deb_archive, tmpdir):
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
        check_package_files(light_project, data_dir)
        assert_filemodes(light_project, data_dir)

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
            assert 'chown -R root:root /usr/share/tarantool/{}'.format(light_project.name) in postinst_script
            assert 'chown root:root /etc/systemd/system/{}.service'.format(light_project.name) in postinst_script
            assert 'chown root:root /etc/systemd/system/{}@.service'.format(light_project.name) in postinst_script
            assert 'chown root:root /usr/lib/tmpfiles.d/{}.conf'.format(light_project.name) in postinst_script


def test_systemd_units(light_project, rpm_archive_with_custom_units, tmpdir):
    ps = subprocess.Popen(
        ['rpm2cpio', rpm_archive_with_custom_units['name']], stdout=subprocess.PIPE)
    subprocess.check_output(['cpio', '-idmv'], stdin=ps.stdout, cwd=tmpdir)
    ps.wait()
    assert ps.returncode == 0, "Error during extracting files from rpm archive"

    project_unit_file = os.path.join(tmpdir, 'etc/systemd/system', "%s.service" % light_project.name)
    with open(project_unit_file) as f:
        assert f.read().find('SIMPLE_UNIT_TEMPLATE') != -1

    project_inst_file = os.path.join(tmpdir, 'etc/systemd/system', "%s@.service" % light_project.name)
    with open(project_inst_file) as f:
        assert f.read().find('INSTANTIATED_UNIT_TEMPLATE') != -1

    project_tmpfiles_conf_file = os.path.join(tmpdir, 'usr/lib/tmpfiles.d', '%s.conf' % light_project.name)
    with open(project_tmpfiles_conf_file) as f:
        assert f.read().find('d /var/run/tarantool') != -1


def test_packing_without_git(project_without_dependencies, tmpdir):
    project_path = project_without_dependencies.path
    shutil.rmtree(os.path.join(project_path, '.git'))

    # try to build rpm w/o --version
    cmd = [
        os.path.join(basepath, "cartridge"),
        "pack", "rpm",
        project_path,
    ]
    process = subprocess.run(cmd, cwd=tmpdir)
    assert process.returncode == 1

    # pass version explicitly
    cmd = [
        os.path.join(basepath, "cartridge"),
        "pack", "rpm",
        "--version", "0.1.0",
        project_path,
    ]
    process = subprocess.run(cmd, cwd=tmpdir)
    assert process.returncode == 0
    project_name = project_without_dependencies.name
    assert '{}-0.1.0-0.rpm'.format(project_name) in os.listdir(tmpdir)


@pytest.mark.parametrize('version,pack_format,expected_postfix',
                         [
                             ('0.1.0', 'rpm', '0.1.0-0.rpm'),
                             ('0.1.0', 'deb', '0.1.0-0.deb'),
                             ('0.1.0-42', 'rpm', '0.1.0-42.rpm'),
                             ('0.1.0-42', 'deb', '0.1.0-42.deb'),
                             ('0.1.0-42-g8bce594e', 'rpm', '0.1.0-42-g8bce594e.rpm'),
                             ('0.1.0-42-g8bce594e', 'deb', '0.1.0-42-g8bce594e.deb'),
                             ('0.1.0-g8bce594e', 'rpm', '0.1.0-g8bce594e.rpm'),
                             ('0.1.0-g8bce594e', 'deb', '0.1.0-g8bce594e.deb'),
                         ])
def test_packing_with_version(project_without_dependencies, tmpdir, version, pack_format, expected_postfix):
    # pass version explicitly
    cmd = [
        os.path.join(basepath, "cartridge"),
        "pack", pack_format,
        "--version", version,
        project_without_dependencies.path,
    ]
    process = subprocess.run(cmd, cwd=tmpdir)
    assert process.returncode == 0
    expected_file = '{name}-{postfix}'.format(name=project_without_dependencies.name, postfix=expected_postfix)
    assert expected_file in os.listdir(tmpdir)


def test_packing_with_wrong_filemodes(project_without_dependencies, tmpdir):
    project_path = project_without_dependencies.path

    # add file with invalid (700) mode
    filepath = os.path.join(project_path, 'wrong-mode-file.lua')
    with open(filepath, 'w') as f:
        f.write("return 'Hi'")
    os.chmod(filepath, 0o700)

    # run `cartridge pack`
    cmd = [os.path.join(basepath, "cartridge"), "pack", "rpm", project_path]
    process = subprocess.run(cmd, cwd=tmpdir)
    assert process.returncode == 1, "Packing project with invalid filemode must fail"


def test_rpm_checksig(rpm_archive):
    cmd = [
        'rpm', '--checksig', '-v', rpm_archive['name']
    ]
    process = subprocess.run(cmd)
    assert process.returncode == 0, "RPM signature isn't correct"
