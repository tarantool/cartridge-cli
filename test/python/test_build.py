import subprocess
import os


from utils import recursive_listdir


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


def test_building_without_path_specifying(cartridge_cmd, project_without_dependencies, tmpdir):
    project = project_without_dependencies

    # say `cartridge build` in project directory
    cmd = [
        cartridge_cmd,
        "build",
    ]
    process = subprocess.run(cmd, cwd=project.path)
    assert process.returncode == 0, 'Building project failed'

    # check that all expected rocks was installed
    files = recursive_listdir(project.path)
    assert '.rocks' in files
    assert all([rock in files for rock in project.rocks_content])


def test_files_with_bad_symbols(cartridge_cmd, project_without_dependencies, tmpdir):
    project = project_without_dependencies

    BAD_FILENAME = 'I \'am\' "the" $worst (file) [ever]'

    with open(os.path.join(project.path, BAD_FILENAME), 'w') as f:
        f.write('Hi!')

    cmd = [
        cartridge_cmd,
        "build",
    ]
    process = subprocess.run(cmd, cwd=project.path)
    assert process.returncode == 0, 'Building project failed'
