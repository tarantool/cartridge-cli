import os
import pytest
import subprocess
import tarfile
import platform

from utils import mark_only_opensource
from utils import find_archive
from utils import assert_tarantool_dependency_deb
from utils import assert_tarantool_dependency_rpm
from utils import extract_deb
from utils import run_command_and_get_output
from utils import get_tarantool_version


def get_other_tarantool_version():
    tarantool_version = get_tarantool_version()
    if tarantool_version.startswith('1'):
        return '2.3.3-0-g5be85a37f'

    return '1.10.8-0-g2f18757b7'


def assert_archive_dependency(archive_path, pack_format, tarantool_version, tmpdir):
    if pack_format == 'rpm':
        assert_tarantool_dependency_rpm(archive_path, tarantool_version)
    elif pack_format == 'deb':
        extract_dir = os.path.join(tmpdir, 'extract')
        os.makedirs(extract_dir)

        extract_deb(archive_path, extract_dir)

        with tarfile.open(name=os.path.join(extract_dir, 'control.tar.gz')) as control_arch:
            control_dir = os.path.join(extract_dir, 'control')
            control_arch.extractall(path=control_dir)

        control_file_path = os.path.join(control_dir, 'control')
        assert_tarantool_dependency_deb(control_file_path, tarantool_version)


@pytest.mark.parametrize('pack_format', ['deb', 'rpm', 'tgz'])
def test_specified_both_flags(cartridge_cmd, tmpdir, light_project, pack_format):
    project = light_project

    cmd = [
        cartridge_cmd, "pack", pack_format,
        "--tarantool-version", "6.6.6-6-gdeadbee",
        "--sdk-version", "7.6.6-6-gdeadbee",
        project.path
    ]

    if pack_format in ['rpm', 'deb'] and platform.system() == 'Darwin':
        cmd.append('--use-docker')

    rc, output = run_command_and_get_output(cmd, cwd=tmpdir)
    assert rc == 1
    assert "You can specify only one of --tarantool-version and --sdk-version" in output


@pytest.mark.parametrize('pack_format', ['deb', 'rpm', 'tgz'])
def test_specified_both_sections_in_file(cartridge_cmd, tmpdir, light_project, pack_format):
    project = light_project

    with open(os.path.join(project.path, "tarantool.txt"), "w") as f:
        f.write('TARANTOOL=6.6.6-6-gdeadbee\nTARANTOOL_SDK=7.6.6-6-gdeadbee\n')

    cmd = [
        cartridge_cmd, "pack", pack_format,
        project.path
    ]

    if pack_format in ['rpm', 'deb'] and platform.system() == 'Darwin':
        cmd.append('--use-docker')

    rc, output = run_command_and_get_output(cmd, cwd=tmpdir)
    assert rc == 1
    assert "You can specify only one of TARANTOOL and TARANTOOL_SDK in tarantool.txt file" in output


@mark_only_opensource
@pytest.mark.parametrize('pack_format', ['deb', 'rpm'])
def test_tarantool_version_specified(cartridge_cmd, tmpdir, light_project, pack_format):
    project = light_project

    tarantool_version = get_other_tarantool_version()

    cmd = [
        cartridge_cmd, "pack", pack_format,
        "--tarantool-version", tarantool_version,
        project.path
    ]

    if platform.system() == 'Darwin':
        cmd.append('--use-docker')

    process = subprocess.run(cmd, cwd=tmpdir)
    assert process.returncode == 0

    filepath = find_archive(tmpdir, project.name, pack_format)
    assert filepath is not None

    assert_archive_dependency(filepath, pack_format, tarantool_version, tmpdir)


@mark_only_opensource
@pytest.mark.parametrize('pack_format', ['deb', 'rpm'])
def test_tarantool_version_from_file(cartridge_cmd, tmpdir, light_project, pack_format):
    project = light_project

    tarantool_version = get_other_tarantool_version()

    with open(os.path.join(project.path, "tarantool.txt"), "w") as f:
        f.write('TARANTOOL=%s\n' % tarantool_version)

    cmd = [
        cartridge_cmd, "pack", pack_format,
        project.path
    ]

    if platform.system() == 'Darwin':
        cmd.append('--use-docker')

    process = subprocess.run(cmd, cwd=tmpdir)
    assert process.returncode == 0

    filepath = find_archive(tmpdir, project.name, pack_format)
    assert filepath is not None

    assert_archive_dependency(filepath, pack_format, tarantool_version, tmpdir)
