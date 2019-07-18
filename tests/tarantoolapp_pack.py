#!/usr/bin/python3

import pytest
import os
import subprocess
import configparser
import tarfile

abspath = os.path.realpath(
    os.path.dirname(__file__)
)

project_name = "test_proj"
target_version_dir = os.path.join(abspath, project_name)
target_version_file = os.path.join(target_version_dir, 'VERSION')


original_file_tree = set([
    '.editorconfig',
    '.gitignore',
    'deps.sh',
    'init.lua',

    '.rocks',
    '.rocks/share',
    '.rocks/share/tarantool',
    '.rocks/share/tarantool/rocks',
    '.rocks/share/tarantool/rocks/manifest',
    '.rocks/share/tarantool/rocks/' + project_name,
    '.rocks/share/tarantool/rocks/' + project_name + '/scm-1',
    '.rocks/share/tarantool/rocks/' + project_name + '/scm-1/rock_manifest',
    '.rocks/share/tarantool/rocks/' + project_name +
    '/scm-1/' + project_name + '-scm-1.rockspec',

    project_name + '-scm-1.rockspec',
    'tarantool',
    'tarantoolctl',
    'VERSION',
    'ignored',  # special folder for test work tarantoolapp ignore
    'ignored/asterisk'
])

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
def prepare_ignore():
    """function creates files and directories
    to check the work .tarantoolapp.ignore"""

    def create_file(path, text=None):
        with open(path, 'w') as f:
            if text:
                f.write(text)

    project_path = os.path.join(abspath, project_name)

    for item in ignored_data:
        directory = os.path.join(project_path, item['dir'])
        if not os.path.exists(directory):
            os.makedirs(directory)
        create_file(os.path.join(directory, item['file']))

    create_file(
        os.path.join(project_path, ".tarantoolapp.ignore"),
        tarantoolapp_ignore_text)


@pytest.fixture(scope="module")
def tgz_archive(prepare_ignore):

    process = subprocess.run(
        ["tarantool", "./tarantoolapp", "pack", "tgz", project_name])
    assert process.returncode == 0, \
        "Error during creating of tgz archive with project"

    archive_name = find_archive(abspath, 'tar.gz')
    assert archive_name != None, "TGZ archive isn't founded in work directory"

    return {'name': archive_name}


@pytest.fixture(scope="module")
def rpm_archive(prepare_ignore):
    process = subprocess.run(
        ["tarantool", "./tarantoolapp", "pack", "rpm", project_name])
    assert process.returncode == 0, \
        "Error during creating of rpm archive with project"

    archive_name = find_archive(abspath, 'rpm')
    assert archive_name != None, "RPM archive isn't founded in work directory"

    return {'name': archive_name}

@pytest.fixture(scope="module")
def rpm_archive_with_custom_units(prepare_ignore):
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
    open("unit_template.tmpl", 'w').write(unit_template)
    open("instantiated_unit_template.tmpl", 'w').write(instantiated_unit_template)

    process = subprocess.run([
            "tarantool", "./tarantoolapp", "pack", "rpm", "--unit_template", "unit_template.tmpl",
            "--instantiated_unit_template", "instantiated_unit_template.tmpl", project_name
        ])
    assert process.returncode == 0, \
        "Error during creating of rpm archive with project"

    archive_name = find_archive(abspath, 'rpm')
    assert archive_name != None, "RPM archive isn't founded in work directory"

    return {'name': archive_name}

def find_archive(path, arch_ext):
    with os.scandir(path) as it:
        for entry in it:
            if entry.name.endswith('.' + arch_ext) and entry.is_file():
                return entry.name


@pytest.fixture(scope="function")
def remove_target_dir():
    yield
    try:
        os.rmdir(target_version_dir)
    except:
        pass


def filter_archive_files(arch_files):
    def get_tail(path): return path.split('/', maxsplit=1)[1]

    def has_tail(path): return len(path.split('/', maxsplit=1)) > 1

    return set(map(get_tail, filter(has_tail, arch_files)))


def validate_version_file(file_path):
    default_section_name = 'tnt-version'
    original_keys = [
        'TARANTOOL',
        'TARANTOOL_SDK',
        project_name,
    ]

    version_props = '[{}]\n'.format(default_section_name)

    with open(file_path) as version_file:
        version_props = version_props + version_file.read()

    parser = configparser.ConfigParser()
    parser.read_string(version_props)
    assert set(parser[default_section_name].keys()) == set(
        map(lambda x: x.lower(), original_keys))


def test_tgz_pack(tgz_archive, remove_target_dir):

    with tarfile.open(name=tgz_archive['name']) as tgz_arch:
        assert original_file_tree == filter_archive_files(map(lambda x: x.name, tgz_arch.getmembers(
        ))), "File tree structure from tgz archive isn't equal to original"

        # tgz_arch.extract(os.path.join(project_name, 'VERSION'), path=abspath)
        # validate_version_file(target_version_file)


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


def test_rpm_pack(rpm_archive, remove_target_dir):

    ps = subprocess.Popen(
        ['rpm2cpio', rpm_archive['name']], stdout=subprocess.PIPE)
    output = subprocess.check_output(['cpio', '-idmv'], stdin=ps.stdout)
    ps.wait()
    assert ps.returncode == 0, "Error during extracting files from rpm archive"

    project_dir = os.path.join('./usr/share/tarantool', project_name)
    assert original_file_tree == recursive_listdir(
        project_dir), "File tree structure from rpm archive isn't equal to original"

    # with open(os.path.join(project_dir, 'VERSION'), 'r') as version_file:
        # with open(target_version_file, 'w') as xvf:
            # xvf.write(version_file.read())

    # validate_version_file(target_version_file)

def test_unit_templates(rpm_archive_with_custom_units, remove_target_dir):
    ps = subprocess.Popen(
        ['rpm2cpio', rpm_archive_with_custom_units['name']], stdout=subprocess.PIPE)
    output = subprocess.check_output(['cpio', '-idmv'], stdin=ps.stdout)
    ps.wait()
    assert ps.returncode == 0, "Error during extracting files from rpm archive"

    project_unit_file = os.path.join('./etc/systemd/system', "%s.service" % project_name )
    assert open(project_unit_file).read().find('SIMPLE_UNIT_TEMPLATE') != -1

    project_inst_file = os.path.join('./etc/systemd/system', "%s@.service" % project_name )
    assert open(project_inst_file).read().find('INSTANTIATED_UNIT_TEMPLATE') != -1
