#!/usr/bin/python3

import os
import subprocess
import rpmfile
import re

__tarantool_version = None

basepath = os.path.realpath(
    os.path.join(os.path.dirname(__file__), '..', '..')
)

project_name = 'test_project'


# #######
# Helpers
# #######
def tarantool_version():
    global __tarantool_version
    if __tarantool_version is None:
        __tarantool_version = subprocess.check_output(['tarantool', '-V']).decode('ascii').split('\n')[0]

    return __tarantool_version


def tarantool_enterprise_is_used():
    return tarantool_version().startswith('Tarantool Enterprise')


def create_project(module_tmpdir, project_name, template):
    cmd = [
        os.path.join(basepath, "cartridge"), "create",
        "--name", project_name,
        "--template", template
    ]
    process = subprocess.run(cmd, cwd=module_tmpdir)
    assert process.returncode == 0, \
        "Error during creating the project"
    return os.path.join(module_tmpdir, project_name)


def find_archive(path, arch_ext):
    with os.scandir(path) as it:
        for entry in it:
            if entry.name.endswith('.' + arch_ext) and entry.is_file():
                return os.path.join(path, entry.name)


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


original_file_tree = set([
    'Dockerfile.cartridge',
    '.cartridge.yml',
    '.editorconfig',
    '.gitignore',
    '.luacheckrc',
    'deps.sh',
    'init.lua',
    'instances.yml',
    'app',
    'app/roles',
    'app/roles/custom.lua',
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


def assert_dir_contents(files_list, skip_tarantool_binaries=False):
    without_rocks = {x for x in files_list if not x.startswith('.rocks')}

    file_tree = original_file_tree
    if skip_tarantool_binaries or not tarantool_enterprise_is_used():
        file_tree = {x for x in file_tree if x not in ['tarantool', 'tarantoolctl']}

    assert file_tree == without_rocks
    assert all(x in files_list for x in original_rocks_content)


def assert_filemode(filepath, filemode):
    filepath = os.path.join('/', filepath)

    if filepath == os.path.join('/usr/share/tarantool/', project_name, 'VERSION'):
        assert filemode & 0o777 == 0o644
    elif filepath.startswith('/etc/systemd/system/'):
        assert filemode & 0o777 == 0o644
    elif filepath.startswith('/usr/lib/tmpfiles.d/'):
        assert filemode & 0o777 == 0o644
    elif filepath.startswith('/usr/share/tarantool/'):
        # a+r for files, a+rx for directories
        required_bits = 0o555 if os.path.isdir(filepath) else 0o444
        assert filemode & required_bits == required_bits


def assert_filemodes(basedir):
    known_dirs = set([
        'etc', 'etc/systemd', 'etc/systemd/system',
        'usr', 'usr/share', 'usr/share/tarantool',
        'usr/lib', 'usr/lib/tmpfiles.d'
    ])
    filenames = recursive_listdir(basedir) - known_dirs

    for filename in filenames:
        # we don't check fileowner here because it's set in postinst script

        # check filemode
        if filename.startswith(os.path.join('usr/share/tarantool/', project_name, '.rocks')):
            continue

        # get filestat
        file_stat = os.stat(os.path.join(basedir, filename))
        filemode = file_stat.st_mode
        assert_filemode(filename, filemode)


def validate_version_file(distribution_dir):
    original_keys = [
        'TARANTOOL',
        project_name,
        # default app dependencies
        'luatest',
        'cartridge',
        # known cartridge dependencies
        'membership',
        'checks',
        'vshard',
        'http',
        'frontend-core',
    ]

    if tarantool_enterprise_is_used():
        original_keys.append('TARANTOOL_SDK')

    version_filepath = os.path.join(distribution_dir, 'VERSION')
    assert os.path.exists(version_filepath)

    version_file_content = {}
    with open(version_filepath, 'r') as version_file:
        file_lines = version_file.read().strip().split('\n')
        for l in file_lines:
            m = re.match(r'^([^=]+)=([^=]+)$', l)
            assert m is not None

            key, version = m.groups()
            version_file_content[key] = version

    for key in original_keys:
        assert key in version_file_content


def assert_files_mode_and_owner_rpm(filename):
    DIRNAMES_TAG = 1118
    DIRINDEXES_TAG = 1116

    expected_tags = [
        'basenames', DIRNAMES_TAG, DIRINDEXES_TAG, 'filemodes',
        'fileusername', 'filegroupname'
    ]

    rpm = rpmfile.open(filename)
    for key in expected_tags:
        assert key in rpm.headers

    for i, basename in enumerate(rpm.headers['basenames']):
        # get filepath
        basename = basename.decode("utf-8")
        dirindex = rpm.headers[DIRINDEXES_TAG][i]
        dirname = rpm.headers[DIRNAMES_TAG][dirindex].decode("utf-8")

        filepath = os.path.join(dirname, basename)

        # check fileowner
        assert rpm.headers['fileusername'][i].decode("utf-8") == 'root'
        assert rpm.headers['filegroupname'][i].decode("utf-8") == 'root'

        # check filemodes
        if filepath.startswith(os.path.join('/usr/share/tarantool/', project_name, '.rocks')):
            continue

        filemode = rpm.headers['filemodes'][i]
        assert_filemode(filepath, filemode)


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


def check_systemd_dir(basedir):
    systemd_dir = (os.path.join(basedir, 'etc/systemd/system'))
    assert os.path.exists(systemd_dir)

    systemd_files = recursive_listdir(systemd_dir)

    assert len(systemd_files) == 2
    assert '{}.service'.format(project_name) in systemd_files
    assert '{}@.service'.format(project_name) in systemd_files


def check_package_files(basedir, project_path):
    # check if only theese files are delivered
    for filename in recursive_listdir(basedir):
        assert any([
            filename.startswith(prefix) or prefix.startswith(filename)
            for prefix in [
                os.path.join('usr/share/tarantool', project_name),
                'etc/systemd/system',
                'usr/lib/tmpfiles.d'
            ]
        ])

    # check distribution dir content
    distribution_dir = os.path.join(basedir, 'usr/share/tarantool', project_name)
    assert os.path.exists(distribution_dir)
    assert_dir_contents(recursive_listdir(distribution_dir))

    # check systemd dir content
    check_systemd_dir(basedir)

    # check tmpfiles conf
    project_tmpfiles_conf_file = os.path.join(basedir, 'usr/lib/tmpfiles.d', '%s.conf' % project_name)
    assert open(project_tmpfiles_conf_file).read().find('d /var/run/tarantool') != -1

    # check version file
    validate_version_file(distribution_dir)
