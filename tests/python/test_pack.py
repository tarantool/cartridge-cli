#!/usr/bin/python3

import os
import pytest
import subprocess
import configparser
import tarfile

from utils import tarantool_enterprise_is_used

project_name = "test_proj"

original_file_tree = set([
    '.editorconfig',
    '.gitignore',
    'deps.sh',
    'init.lua',
    'test',
    'test/helper',
    'test/integration',
    'test/unit',
    'tmp',
    'tmp/.keep',
    project_name + '-scm-1.rockspec',
    'tarantool',
    'tarantoolctl',
    'VERSION',
    'ignored',  # special folder for test work tarantoolapp ignore
    'ignored/asterisk'
])

original_rocks_content = set([
    '.rocks',
    '.rocks/share/tarantool/rocks/manifest',
    '.rocks/share/tarantool/rocks/' + project_name,
    '.rocks/share/tarantool/rocks/' + project_name + '/scm-1',
    '.rocks/share/tarantool/rocks/' + project_name + '/scm-1/rock_manifest',
    '.rocks/share/tarantool/rocks/' + project_name + '/scm-1/' + project_name + '-scm-1.rockspec',

    '.rocks/bin/luatest',
    '.rocks/share/tarantool/checks.lua',
    '.rocks/share/tarantool/luarocks/test/luatest.lua',
    '.rocks/share/tarantool/luatest',
    '.rocks/share/tarantool/rocks/checks',
    '.rocks/share/tarantool/rocks/luatest',
])

def assert_dir_contents(files_list):
    without_rocks = {x for x in files_list if not x.startswith('.rocks')}

    file_tree = original_file_tree
    if not tarantool_enterprise_is_used():
        file_tree = {x for x in file_tree if x not in ['tarantool', 'tarantoolctl']}

    assert file_tree == without_rocks
    assert all(x in files_list for x in original_rocks_content)

ignored_data = [
    {
        'dir': '',
        'file': 'ignored.txt'
    },
    {
        'dir': '',
        'file': 'asterisk'
    },
    {
        'dir': '',
        'file': 'ignored.lua'
    },
    {
        'dir': '',
        'file': 'ignored_by.format'
    },
    {
        'dir': 'ignored',
        'file': 'sample.txt'
    },
    {
        'dir': 'ignored/folder',
        'file': 'sample.txt'
    },
    {
        'dir': 'ignored/asterisk',
        'file': 'star.txt'
    },
    {
        'dir': 'ignored/asterisk',
        'file': 'simple'
    },
    {
        'dir': 'ignored/sample',
        'file': 'test'
    },
    {
        'dir': 'ignored',
        'file': '#test'
    }
]


patterns = [
    # patterns that match the patterns from whitelist
    '.rocks/share/tarantool/rocks/**',
    '*.lua',
    'deps.sh',
    # whitelist
    '!*.sh',
    '!.rocks/**',
    '!init.lua',
    '!asterisk/',
    # for ignore
    'ignored.txt',
    '*.format',
    'ignored/*.txt',
    'ignored/folder/',
    '**/*.txt',
    'simple',
    'sample',
    'asterisk',
    # comment example
    '# /scm-1',
    # escaping \#
    '\\#test'
]


tarantoolapp_ignore_text = '\n'.join(patterns)

@pytest.fixture(scope="session")
def test_project_path(session_tmpdir):
    cmd = ["tarantoolapp", "create",
        "--name", project_name,
        "--template", "plain"]
    process = subprocess.run(cmd, cwd=session_tmpdir)
    assert process.returncode == 0, \
        "Error during creating the project"
    return os.path.join(session_tmpdir, project_name)

@pytest.fixture(scope="session")
def prepare_ignore(test_project_path):
    """function creates files and directories
    to check the work .tarantoolapp.ignore"""

    def create_file(path, text=None):
        with open(path, 'w') as f:
            if text:
                f.write(text)

    for item in ignored_data:
        directory = os.path.join(test_project_path, item['dir'])
        if not os.path.exists(directory):
            os.makedirs(directory)
        create_file(os.path.join(directory, item['file']))

    create_file(
        os.path.join(test_project_path, ".tarantoolapp.ignore"),
        tarantoolapp_ignore_text)


@pytest.fixture(scope="session")
def tgz_archive(session_tmpdir, test_project_path, prepare_ignore):
    cmd = ["tarantoolapp", "pack", "tgz", test_project_path]
    process = subprocess.run(cmd, cwd=session_tmpdir)
    assert process.returncode == 0, \
        "Error during creating of tgz archive with project"

    archive_name = find_archive(session_tmpdir, 'tar.gz')
    assert archive_name != None, "TGZ archive isn't founded in work directory"

    return {'name': archive_name}


@pytest.fixture(scope="session")
def rpm_archive(session_tmpdir, test_project_path, prepare_ignore):
    cmd = ["tarantoolapp", "pack", "rpm", test_project_path]
    process = subprocess.run(cmd, cwd=session_tmpdir)
    assert process.returncode == 0, \
        "Error during creating of rpm archive with project"

    archive_name = find_archive(session_tmpdir, 'rpm')
    assert archive_name != None, "RPM archive isn't founded in work directory"

    return {'name': archive_name}

@pytest.fixture(scope="session")
def rpm_archive_with_custom_units(session_tmpdir, test_project_path, prepare_ignore):
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
    open(os.path.join(session_tmpdir, "unit_template.tmpl"), 'w').write(unit_template)
    open(os.path.join(session_tmpdir, "instantiated_unit_template.tmpl"), 'w').write(instantiated_unit_template)

    process = subprocess.run([
            "tarantoolapp", "pack", "rpm", "--unit_template", "unit_template.tmpl",
            "--instantiated_unit_template", "instantiated_unit_template.tmpl", test_project_path
        ],
        cwd=session_tmpdir
    )
    assert process.returncode == 0, \
        "Error during creating of rpm archive with project"

    archive_name = find_archive(session_tmpdir, 'rpm')
    assert archive_name != None, "RPM archive isn't founded in work directory"

    return {'name': archive_name}

def find_archive(path, arch_ext):
    with os.scandir(path) as it:
        for entry in it:
            if entry.name.endswith('.' + arch_ext) and entry.is_file():
                return os.path.join(path, entry.name)


def filter_archive_files(arch_files):
    def get_tail(path): return path.split('/', maxsplit=1)[1]

    def has_tail(path): return len(path.split('/', maxsplit=1)) > 1

    return set(map(get_tail, filter(has_tail, arch_files)))


def validate_version_file(file_path):
    original_keys = [
        project_name,
    ]
    if tarantool_enterprise_is_used():
        original_keys.append('TARANTOOL')
        original_keys.append('TARANTOOL_SDK')

    default_section_name = 'tnt-version'
    version_props = '[{}]\n'.format(default_section_name)

    with open(file_path) as version_file:
        version_props = version_props + version_file.read()

    parser = configparser.ConfigParser()
    parser.read_string(version_props)
    assert set(parser[default_section_name].keys()) == set(
        map(lambda x: x.lower(), original_keys))


def test_tgz_pack(test_project_path, tgz_archive, tmpdir):

    with tarfile.open(name=tgz_archive['name']) as tgz_arch:
        assert_dir_contents(
            filter_archive_files(map(lambda x: x.name, tgz_arch.getmembers()))
        )

        tgz_arch.extract(os.path.join(project_name, 'VERSION'), path=tmpdir)
        validate_version_file(os.path.join(tmpdir, project_name, 'VERSION'))


def recursive_listdir(root_dir):
    files = set()
    for root, directories, filenames in os.walk(root_dir):
        rel_dir = os.path.relpath(root, root_dir)
        if rel_dir == '.':
            rel_dir = ''

        for directory in directories:
            files.add(os.path.join(rel_dir, directory))

        for filename in filenames:
            files.add(os.path.join(rel_dir, filename))

    return files


def test_rpm_pack(test_project_path, rpm_archive, tmpdir):
    ps = subprocess.Popen(
        ['rpm2cpio', rpm_archive['name']], stdout=subprocess.PIPE)
    subprocess.check_output(['cpio', '-idmv'], stdin=ps.stdout, cwd=tmpdir)
    ps.wait()
    assert ps.returncode == 0, "Error during extracting files from rpm archive"

    project_dir = os.path.join(tmpdir, 'usr/share/tarantool', project_name)
    assert_dir_contents(recursive_listdir(project_dir))

    target_version_file = os.path.join(test_project_path, 'VERSION')
    with open(os.path.join(project_dir, 'VERSION'), 'r') as version_file:
        with open(target_version_file, 'w') as xvf:
            xvf.write(version_file.read())

    validate_version_file(target_version_file)

def test_unit_templates(test_project_path, rpm_archive_with_custom_units, tmpdir):
    ps = subprocess.Popen(
        ['rpm2cpio', rpm_archive_with_custom_units['name']], stdout=subprocess.PIPE)
    subprocess.check_output(['cpio', '-idmv'], stdin=ps.stdout, cwd=tmpdir)
    ps.wait()
    assert ps.returncode == 0, "Error during extracting files from rpm archive"

    project_unit_file = os.path.join(tmpdir, 'etc/systemd/system', "%s.service" % project_name )
    assert open(project_unit_file).read().find('SIMPLE_UNIT_TEMPLATE') != -1

    project_inst_file = os.path.join(tmpdir, 'etc/systemd/system', "%s@.service" % project_name )
    assert open(project_inst_file).read().find('INSTANTIATED_UNIT_TEMPLATE') != -1
