import subprocess
import os
import re

from utils import basepath
from utils import recursive_listdir
from utils import run_command_and_get_output


# #####
# Tests
# #####
def test_build(light_project, tmpdir):
    project = light_project

    project_files_before = recursive_listdir(project.path)

    cmd = [
        os.path.join(basepath, "cartridge"),
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


def test_using_both_flows(project_without_dependencies, tmpdir):
    # add deprecated flow files to the project
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
        os.path.join(basepath, "cartridge"),
        "build",
        project.path
    ]
    rc, outout = run_command_and_get_output(cmd, cwd=tmpdir)
    assert rc == 1
    assert re.search(r'You use deprecated .+ files and .+ files at the same time', outout)


def test_building_without_path_specifying(project_without_dependencies, tmpdir):
    project = project_without_dependencies

    # say `cartridge build` in project directory
    cmd = [
        os.path.join(basepath, "cartridge"),
        "build",
    ]
    process = subprocess.run(cmd, cwd=project.path)
    assert process.returncode == 0, 'Building project failed'

    # check that all expected rocks was installed
    files = recursive_listdir(project.path)
    assert '.rocks' in files
    assert all([rock in files for rock in project.rocks_content])
