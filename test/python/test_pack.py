#!/usr/bin/python3

import os
import pytest
import subprocess
import configparser
import tarfile
import rpmfile
import re

from utils import basepath
from utils import create_project
from utils import tarantool_version
from utils import tarantool_enterprise_is_used

project_name = "test_proj"

original_file_tree = set([
    '.editorconfig',
    '.gitignore',
    '.luacheckrc',
    'deps.sh',
    'init.lua',
    'app',
    'app/roles',
    'test',
    'test/helper',
    'test/integration',
    'test/unit',
    'tmp',
    'tmp/.keep',
    project_name + '-scm-1.rockspec',
    'VERSION',
    'ignored',  # special folder for test work cartridge ignore
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

if tarantool_enterprise_is_used():
    original_file_tree |= set([
        'cartridge',
        'tarantool',
        'tarantoolctl',
    ])
    original_rocks_content |= set([
        '.rocks/share/tarantool/rocks/cartridge-cli/1.0.0-1/bin/cartridge',
    ])

def assert_dir_contents(files_list):
    without_rocks = {x for x in files_list if not x.startswith('.rocks')}
    assert original_file_tree == without_rocks
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


cartridge_ignore_text = '\n'.join(patterns)

@pytest.fixture(scope="module")
def project_path(module_tmpdir):
    return create_project(module_tmpdir, project_name, 'cartridge')

@pytest.fixture(scope="module")
def prepare_ignore(project_path):
    """function creates files and directories
    to check the work .cartridge.ignore"""

    def create_file(path, text=None):
        with open(path, 'w') as f:
            if text:
                f.write(text)

    for item in ignored_data:
        directory = os.path.join(project_path, item['dir'])
        if not os.path.exists(directory):
            os.makedirs(directory)
        create_file(os.path.join(directory, item['file']))

    create_file(
        os.path.join(project_path, ".cartridge.ignore"),
        cartridge_ignore_text)


@pytest.fixture(scope="module")
def tgz_archive(module_tmpdir, project_path, prepare_ignore):
    cmd = [os.path.join(basepath, "cartridge"), "pack", "tgz", project_path]
    process = subprocess.run(cmd, cwd=module_tmpdir)
    assert process.returncode == 0, \
        "Error during creating of tgz archive with project"

    archive_name = find_archive(module_tmpdir, 'tar.gz')
    assert archive_name != None, "TGZ archive isn't founded in work directory"

    return {'name': archive_name}


@pytest.fixture(scope="module")
def rpm_archive(module_tmpdir, project_path, prepare_ignore):
    cmd = [os.path.join(basepath, "cartridge"), "pack", "rpm", project_path]
    process = subprocess.run(cmd, cwd=module_tmpdir)
    assert process.returncode == 0, \
        "Error during creating of rpm archive with project"

    archive_name = find_archive(module_tmpdir, 'rpm')
    assert archive_name != None, "RPM archive isn't founded in work directory"

    return {'name': archive_name}


@pytest.fixture(scope="module")
def deb_archive(module_tmpdir, project_path, prepare_ignore):
    cmd = [os.path.join(basepath, "cartridge"), "pack", "deb", project_path]
    process = subprocess.run(cmd, cwd=module_tmpdir)
    assert process.returncode == 0, \
        "Error during creating of deb archive with project"

    archive_name = find_archive(module_tmpdir, 'deb')
    assert archive_name != None, "DEB archive isn't founded in work directory"

    return {'name': archive_name}


@pytest.fixture(scope="module")
def rpm_archive_with_custom_units(module_tmpdir, project_path, prepare_ignore):
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
    open(os.path.join(module_tmpdir, "unit_template.tmpl"), 'w').write(unit_template)
    open(os.path.join(module_tmpdir, "instantiated_unit_template.tmpl"), 'w').write(instantiated_unit_template)

    process = subprocess.run([
            os.path.join(basepath, "cartridge"), "pack", "rpm",
            "--unit_template", "unit_template.tmpl",
            "--instantiated_unit_template", "instantiated_unit_template.tmpl",
            project_path
        ],
        cwd=module_tmpdir
    )
    assert process.returncode == 0, \
        "Error during creating of rpm archive with project"

    archive_name = find_archive(module_tmpdir, 'rpm')
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


def test_tgz_pack(project_path, tgz_archive, tmpdir):

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


def assert_tarantool_dependency_rpm(filename):
    rpm = rpmfile.open(filename)
    dependency_keys = ['requirename', 'requireversion', 'requireflags']
    for key in dependency_keys:
        assert key in rpm.headers

    assert len(rpm.headers['requirename']) == 2
    assert len(rpm.headers['requireversion']) == 2
    assert len(rpm.headers['requireversion']) == 2

    min_version = re.findall(r'\d+\.\d+\.\d+', tarantool_version())[0]
    max_version = str(int(re.findall(r'\d+', tarantool_version())[0]) + 1)

    assert rpm.headers['requirename'][0].decode('ascii') == 'tarantool'
    assert rpm.headers['requireversion'][0].decode('ascii') == min_version
    assert rpm.headers['requireflags'][0] == 0x08 | 0x04  # >=

    assert rpm.headers['requirename'][1].decode('ascii') == 'tarantool'
    assert rpm.headers['requireversion'][1].decode('ascii') == max_version
    assert rpm.headers['requireflags'][1] == 0x02  # <


def assert_tarantool_dependency_deb(filename):
    with open(filename) as control:
        control_info = control.read()

        depends_str = re.search('Depends: (.*)', control_info)
        assert depends_str is not None

        min_version = re.findall(r'\d+\.\d+\.\d+', tarantool_version())[0]
        max_version = str(int(re.findall(r'\d+', tarantool_version())[0]) + 1)

        deps = depends_str.group(1)
        assert 'tarantool (>= {})'.format(min_version) in deps
        assert 'tarantool (<< {})'.format(max_version) in deps


def test_rpm_pack(project_path, rpm_archive, tmpdir):
    ps = subprocess.Popen(
        ['rpm2cpio', rpm_archive['name']], stdout=subprocess.PIPE)
    subprocess.check_output(['cpio', '-idmv'], stdin=ps.stdout, cwd=tmpdir)
    ps.wait()
    assert ps.returncode == 0, "Error during extracting files from rpm archive"

    project_dir = os.path.join(tmpdir, 'usr/share/tarantool', project_name)
    assert_dir_contents(recursive_listdir(project_dir))

    target_version_file = os.path.join(project_path, 'VERSION')
    with open(os.path.join(project_dir, 'VERSION'), 'r') as version_file:
        with open(target_version_file, 'w') as xvf:
            xvf.write(version_file.read())

    if not tarantool_enterprise_is_used():
        assert_tarantool_dependency_rpm(rpm_archive['name'])

    validate_version_file(target_version_file)


def test_deb_pack(project_path, deb_archive, tmpdir):
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
        project_dir = os.path.join(data_dir, 'usr/share/tarantool', project_name)
        assert_dir_contents(recursive_listdir(project_dir))

    target_version_file = os.path.join(project_path, 'VERSION')
    with open(os.path.join(project_dir, 'VERSION'), 'r') as version_file:
        with open(target_version_file, 'w') as xvf:
            xvf.write(version_file.read())

    validate_version_file(target_version_file)

    # check control.tar.xz
    with tarfile.open(name=os.path.join(tmpdir, 'control.tar.xz')) as control_arch:
        control_dir = os.path.join(tmpdir, 'control')
        control_arch.extractall(path=control_dir)

        for filename in ['control', 'preinst', 'postinst']:
            assert os.path.exists(os.path.join(control_dir, filename))

        if not tarantool_enterprise_is_used():
            assert_tarantool_dependency_deb(os.path.join(control_dir, 'control'))


def test_systemd_units(project_path, rpm_archive_with_custom_units, tmpdir):
    ps = subprocess.Popen(
        ['rpm2cpio', rpm_archive_with_custom_units['name']], stdout=subprocess.PIPE)
    subprocess.check_output(['cpio', '-idmv'], stdin=ps.stdout, cwd=tmpdir)
    ps.wait()
    assert ps.returncode == 0, "Error during extracting files from rpm archive"

    project_unit_file = os.path.join(tmpdir, 'etc/systemd/system', "%s.service" % project_name )
    assert open(project_unit_file).read().find('SIMPLE_UNIT_TEMPLATE') != -1

    project_inst_file = os.path.join(tmpdir, 'etc/systemd/system', "%s@.service" % project_name )
    assert open(project_inst_file).read().find('INSTANTIATED_UNIT_TEMPLATE') != -1

    project_tmpfiles_conf_file = os.path.join(tmpdir, 'usr/lib/tmpfiles.d', '%s.conf' % project_name )
    assert open(project_tmpfiles_conf_file).read().find('d /var/run/tarantool') != -1
