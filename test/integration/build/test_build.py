import subprocess
import os
import pytest


from utils import recursive_listdir
from utils import run_command_and_get_output


# #####
# Tests
# #####
def test_build(cartridge_cmd, light_project, tmpdir):
    project = light_project

    project_files_before = recursive_listdir(project.path)

    cmd = [
        cartridge_cmd,
        "build",
        project.path
    ]
    process = subprocess.run(cmd, cwd=tmpdir)
    assert process.returncode == 0, "Error during building the project"

    # check that all expected rocks was installed
    files = recursive_listdir(project.path)
    assert '.rocks' in files
    assert all([rock in files for rock in project.rocks_content])

    project_files_after = recursive_listdir(project.path)

    # check that nothing was deleted
    assert all([f in project_files_after for f in project_files_before])


def test_building_without_path_specifying(cartridge_cmd, project_without_dependencies):
    project = project_without_dependencies

    # say `cartridge build` in project directory
    cmd = [
        cartridge_cmd,
        "build",
    ]
    process = subprocess.run(cmd, cwd=project.path)
    assert process.returncode == 0

    # check that all expected rocks was installed
    files = recursive_listdir(project.path)
    assert '.rocks' in files
    assert all([rock in files for rock in project.rocks_content])


def test_files_with_bad_symbols(cartridge_cmd, project_without_dependencies):
    project = project_without_dependencies

    BAD_FILENAME = 'I \'am\' "the" $worst (file) [ever]'

    with open(os.path.join(project.path, BAD_FILENAME), 'w') as f:
        f.write('Hi!')

    cmd = [
        cartridge_cmd,
        "build",
    ]
    process = subprocess.run(cmd, cwd=project.path)
    assert process.returncode == 0


def test_app_without_rockspec(cartridge_cmd, project_without_dependencies):
    project = project_without_dependencies

    os.remove(project.rockspec_path)
    cmd = [
        cartridge_cmd,
        "build",
    ]

    rc, output = run_command_and_get_output(cmd, cwd=project.path)
    assert rc == 1, 'Building project should fail'
    assert 'Application directory should contain rockspec' in output


@pytest.mark.parametrize('hook', ['cartridge.pre-build'])
def test_app_with_non_executable_hook(cartridge_cmd, project_without_dependencies, hook):
    project = project_without_dependencies

    hook_path = os.path.join(project.path, hook)
    os.chmod(hook_path, 0o0644)

    cmd = [
        cartridge_cmd,
        "build",
    ]

    rc, output = run_command_and_get_output(cmd, cwd=project.path)
    assert rc == 1, 'Building project should fail'
    assert 'Hook `{}` should be executable'.format(hook) in output


def test_verbosity(cartridge_cmd, project_without_dependencies):
    project = project_without_dependencies

    prebuild_output = "pre-build hook output"
    rocks_make_output = "{} scm-1 is now installed".format(project.name)

    with open(os.path.join(project.path, 'cartridge.pre-build'), 'w') as f:
        prebuild_script_lines = [
            "#!/bin/sh",
            "echo \"{}\"".format(prebuild_output)
        ]
        f.write('\n'.join(prebuild_script_lines))

    build_logs = [
        'Build application in',
        'Running `cartridge.pre-build`',
        'Running `tarantoolctl rocks make`',
        'Application was successfully built',
    ]

    # w/o flags
    cmd = [
        cartridge_cmd,
        "build",
    ]

    rc, output = run_command_and_get_output(cmd, cwd=project.path)
    assert rc == 0
    assert all([log in output for log in build_logs])
    assert prebuild_output not in output
    assert rocks_make_output not in output

    # with --verbose
    cmd = [
        cartridge_cmd,
        "build",
        "--verbose",
    ]

    rc, output = run_command_and_get_output(cmd, cwd=project.path)
    assert rc == 0
    assert all([log in output for log in build_logs])
    assert prebuild_output in output
    assert rocks_make_output in output

    # with --quiet
    cmd = [
        cartridge_cmd,
        "build",
        "--quiet",
    ]

    rc, output = run_command_and_get_output(cmd, cwd=project.path)
    assert rc == 0
    assert output == ''

    # hook error with --quiet
    cmd = [
        cartridge_cmd,
        "build",
        "--quiet",
    ]

    with open(os.path.join(project.path, 'cartridge.pre-build'), 'w') as f:
        prebuild_script_lines = [
            "#!/bin/sh",
            "echo \"{}\"".format(prebuild_output),
            "exit 1"
        ]
        f.write('\n'.join(prebuild_script_lines))

    rc, output = run_command_and_get_output(cmd, cwd=project.path)
    assert rc == 1, 'Building project should fail'
    assert all([log not in output for log in build_logs])
    assert 'Failed to run pre-build hook' in output
    assert prebuild_output in output
