import subprocess
from utils import run_command_and_get_output


def test_version_command(cartridge_cmd):
    for version_cmd in ["version", "-v", "--version"]:
        rc, output = run_command_and_get_output([cartridge_cmd, version_cmd])
        assert rc == 0
        assert 'Tarantool Cartridge CLI\n Version:\t2' in output
        assert 'Tarantool Cartridge\n Version:\t<unknown>' in output
        assert 'Failed to get the version of the Cartridge' in output


def test_version_command_with_project(project_with_cartridge, cartridge_cmd, tmpdir):
    project = project_with_cartridge

    cmd = [
        cartridge_cmd,
        "build",
        project.path
    ]

    process = subprocess.run(cmd, cwd=tmpdir)
    assert process.returncode == 0

    cmd = [
        cartridge_cmd, "version",
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
    cmd = [
        cartridge_cmd,
        "version", "--rocks",
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


def test_version_command_invalid_project(cartridge_cmd):
    cmd = [
        cartridge_cmd, "version",
        "--project-path=invalid_path"
    ]

    rc, output = run_command_and_get_output(cmd)
    assert rc == 0
    assert 'Failed to get the version of the Cartridge. Your project path is invalid' in output
