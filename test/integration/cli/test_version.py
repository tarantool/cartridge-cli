import subprocess
from utils import run_command_and_get_output
from project import remove_project_file


def test_version_command(cartridge_cmd):
    for version_cmd in ["version", "-v", "--version"]:
        rc, output = run_command_and_get_output([cartridge_cmd, version_cmd])
        assert rc == 0
        assert 'Tarantool Cartridge CLI\n Version:\t2' in output
        assert 'Failed to show Cartridge version: Project path . is not a project' in output


def test_version_command_with_project(project_with_cartridge, cartridge_cmd, tmpdir):
    project = project_with_cartridge

    cmd = [
        cartridge_cmd,
        "build",
        project.path
    ]

    process = subprocess.run(cmd, cwd=tmpdir)
    assert process.returncode == 0

    for version_cmd in ["version", "-v", "--version"]:
        cmd = [
            cartridge_cmd, version_cmd,
            f"--project-path={project.path}"
        ]

        rc, output = run_command_and_get_output(cmd)
        assert rc == 0

        assert 'Tarantool Cartridge CLI\n Version:\t2' in output
        assert 'Tarantool Cartridge\n Version:\t2' in output


def test_version_command_with_rocks(project_with_cartridge, cartridge_cmd, tmpdir):
    project = project_with_cartridge

    cmd = [
        cartridge_cmd,
        "build",
        project.path
    ]

    process = subprocess.run(cmd, cwd=tmpdir)
    assert process.returncode == 0

    for version_cmd in ["version", "-v", "--version"]:
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


def test_version_command_invalid_project(project_without_dependencies, cartridge_cmd, tmpdir):
    project = project_without_dependencies
    remove_project_file(project, f'{project.name}-scm-1.rockspec')

    for version_cmd in ["version", "-v", "--version"]:
        cmd = [
            cartridge_cmd,
            version_cmd, "--rocks",
            f"--project-path={tmpdir}"
        ]

        rc, output = run_command_and_get_output(cmd)
        assert rc == 1
        assert f'Failed to show Cartridge and other rocks versions: Project path {tmpdir} is not a project' in output


def test_version_command_nonbuilded_project(project_without_dependencies, cartridge_cmd, tmpdir):
    project = project_without_dependencies

    for version_cmd in ["version", "-v", "--version"]:
        cmd = [
            cartridge_cmd,
            version_cmd, "--rocks",
            f"--project-path={project.path}"
        ]

        rc, output = run_command_and_get_output(cmd)
        assert rc == 1
        assert 'Failed to show Cartridge and other rocks versions: ' \
            'Are dependencies in .rocks directory correct?' in output


def test_version_command_invalid_path(cartridge_cmd):
    for version_cmd in ["version", "-v", "--version"]:
        cmd = [
            cartridge_cmd, version_cmd,
            "--project-path=invalid_path"
        ]

        rc, output = run_command_and_get_output(cmd)
        assert rc == 1
        assert 'Failed to show Cartridge version: Specified project path doesn\'t exist' in output
