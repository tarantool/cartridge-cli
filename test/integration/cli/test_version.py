import pytest
import subprocess
from utils import run_command_and_get_output
from project import remove_project_file


@pytest.mark.parametrize('version_cmd', ['version', '-v', '--version'])
def test_version_command(cartridge_cmd, version_cmd):
    rc, output = run_command_and_get_output([cartridge_cmd, version_cmd])
    assert rc == 0
    assert 'Tarantool Cartridge CLI\n Version:\t2' in output
    assert 'Failed to show Cartridge version: Project path . is not a project' in output


@pytest.mark.parametrize('version_cmd', ['version', '-v', '--version'])
def test_version_command_with_project(project_with_cartridge, version_cmd, cartridge_cmd, tmpdir):
    project = project_with_cartridge

    cmd = [
        cartridge_cmd,
        "build",
        project.path
    ]

    process = subprocess.run(cmd, cwd=tmpdir)
    assert process.returncode == 0

    cmd = [
        cartridge_cmd, version_cmd,
        f"--project-path={project.path}"
    ]

    rc, output = run_command_and_get_output(cmd)
    assert rc == 0

    assert 'Tarantool Cartridge CLI\n Version:\t2' in output
    assert 'Tarantool Cartridge\n Version:\t2' in output
    assert 'Rocks' not in output


@pytest.mark.parametrize('version_cmd', ['version', '-v', '--version'])
def test_version_command_with_rocks(project_with_cartridge, version_cmd, cartridge_cmd, tmpdir):
    project = project_with_cartridge

    cmd = [
        cartridge_cmd,
        "build",
        project.path
    ]

    process = subprocess.run(cmd, cwd=tmpdir)
    assert process.returncode == 0

    cmd = [
        cartridge_cmd,
        version_cmd, "--rocks",
        f"--project-path={project.path}"
    ]

    rc, output = run_command_and_get_output(cmd)
    assert rc == 0
    assert 'Tarantool Cartridge CLI\n Version:\t2' in output
    assert 'Tarantool Cartridge\n Version:\t2' in output
    assert 'Rocks' in output
    assert 'metrics ' in output
    assert 'checks ' in output
    assert project.name in output


@pytest.mark.parametrize('version_cmd', ['version', '-v', '--version'])
def test_version_command_invalid_project(project_without_dependencies, version_cmd, cartridge_cmd, tmpdir):
    project = project_without_dependencies
    remove_project_file(project, f'{project.name}-scm-1.rockspec')

    cmd = [
        cartridge_cmd,
        version_cmd, "--rocks",
        f"--project-path={tmpdir}"
    ]

    rc, output = run_command_and_get_output(cmd)
    assert rc == 1
    assert f'Failed to show Cartridge version: Project path {tmpdir} is not a project' in output


@pytest.mark.parametrize('version_cmd', ['version', '-v', '--version'])
def test_version_command_nonbuilded_project(project_without_dependencies, version_cmd, cartridge_cmd, tmpdir):
    project = project_without_dependencies

    cmd = [
        cartridge_cmd,
        version_cmd, "--rocks",
        f"--project-path={project.path}"
    ]

    rc, output = run_command_and_get_output(cmd)
    assert rc == 1
    assert 'Failed to show Cartridge version: ' \
        'Are dependencies in .rocks directory correct?' in output


@pytest.mark.parametrize('version_cmd', ['version', '-v', '--version'])
def test_version_command_invalid_path(cartridge_cmd, version_cmd):
    cmd = [
        cartridge_cmd, version_cmd,
        "--project-path=invalid_path"
    ]

    rc, output = run_command_and_get_output(cmd)
    assert rc == 1
    assert 'Failed to show Cartridge version: Specified project path doesn\'t exist' in output


@pytest.mark.parametrize('version_cmd', ['version', '-v', '--version'])
def test_duplicate_rocks(project_with_cartridge, cartridge_cmd, version_cmd, tmpdir):
    project = project_with_cartridge

    cmd = [
        cartridge_cmd,
        "build",
        project.path
    ]

    process = subprocess.run(cmd, cwd=tmpdir)
    assert process.returncode == 0

    # Cartridge already has graphql dependency
    cmd = [
        "tarantoolctl", "rocks", "install",
        "graphql", "0.1.0-1"
    ]

    process = subprocess.run(cmd, cwd=project.path)
    assert process.returncode == 0

    cmd = [
        cartridge_cmd,
        version_cmd, "--rocks",
        f"--project-path={project.path}"
    ]

    rc, output = run_command_and_get_output(cmd)
    assert rc == 0
    assert "graphql 0.1.0-1, 0.1.1-1" in output
    assert "Found multiple versions in rocks manifest" in output


@pytest.mark.parametrize('version_cmd', ['version', '-v', '--version'])
def test_duplicate_cartridge_no_rocks_flag(project_with_cartridge, cartridge_cmd, version_cmd, tmpdir):
    project = project_with_cartridge

    cmd = [
        cartridge_cmd, "build", project.path
    ]

    process = subprocess.run(cmd, cwd=tmpdir)
    assert process.returncode == 0

    # Install one more Cartridge
    cmd = [
        "tarantoolctl", "rocks", "install",
        "cartridge", "2.5.0"
    ]

    process = subprocess.run(cmd, cwd=project.path)
    assert process.returncode == 0

    cmd = [
        cartridge_cmd,
        version_cmd,
        f"--project-path={project.path}"
    ]

    rc, output = run_command_and_get_output(cmd)
    assert rc == 0
    assert "Version:\t2.5.0-1, 2.6.0-1" in output
    assert "Found multiple versions of Cartridge in rocks manifest" in output
