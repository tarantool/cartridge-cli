#!/usr/bin/python3

import os
import pytest
import subprocess
import configparser
import tarfile
import rpmfile
import docker
import re
import requests
import time

from utils import basepath
from utils import create_project
from utils import tarantool_version
from utils import tarantool_enterprise_is_used

project_name = "test_proj"

original_file_tree = set([
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


def check_systemd_dir(basedir):
    systemd_dir = (os.path.join(basedir, 'etc/systemd/system'))
    assert os.path.exists(systemd_dir)

    systemd_files = recursive_listdir(systemd_dir)

    assert len(systemd_files) == 2
    assert '{}.service'.format(project_name) in systemd_files
    assert '{}@.service'.format(project_name) in systemd_files


def wait_for_container_start(container, timeout=10):
    time_start = time.time()
    while True:
        now = time.time()
        if now > time_start + timeout:
            break

        container_logs = container.logs(since=int(time_start)).decode('utf-8')
        if 'entering the event loop' in container_logs:
            return True

        time.sleep(1)

    return False


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
    '!app/roles/custom.lua',
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
def docker_client():
    client = docker.from_env()
    return client


def find_image(docker_client, project_name):
    for image in docker_client.images.list():
        for t in image.tags:
            if t.startswith(project_name):
                return t


@pytest.fixture(scope="module")
def docker_image(module_tmpdir, project_path, prepare_ignore, request, docker_client):
    cmd = [os.path.join(basepath, "cartridge"), "pack", "docker", project_path]
    process = subprocess.run(cmd, cwd=module_tmpdir)
    assert process.returncode == 0, \
        "Error during creating of docker image"

    image_name = find_image(docker_client, project_name)
    assert image_name != None, "Docker image isn't found"

    def delete_image(image_name):
        if docker_client.images.list('myapp:0.1.0-0'):
            # remove all image containers
            containers = docker_client.containers.list(
                all=True,
                filters={'ancestor': image_name}
            )

            for c in containers:
                c.remove(force=True)

            # remove image itself
            docker_client.images.remove(image_name)

    request.addfinalizer(lambda: delete_image(image_name))
    return {'name': image_name}


@pytest.fixture(scope="module")
def deb_archive(module_tmpdir, project_path, prepare_ignore):
    cmd = [os.path.join(basepath, "cartridge"), "pack", "deb", project_path]
    process = subprocess.run(cmd, cwd=module_tmpdir)
    assert process.returncode == 0, \
        "Error during creating of deb archive with project"

    archive_name = find_archive(module_tmpdir, 'deb')
    assert archive_name != None, "DEB archive isn't found in work directory"

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


def validate_version_file(distribution_dir):
    original_keys = [
        project_name,
    ]
    if tarantool_enterprise_is_used():
        original_keys.append('TARANTOOL')
        original_keys.append('TARANTOOL_SDK')

    version_filepath = os.path.join(distribution_dir, 'VERSION')
    assert os.path.exists(version_filepath)

    with open(version_filepath, 'r') as version_file:
        version_content = version_file.read()
        for key in original_keys:
            assert '{}='.format(key) in version_content


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


def expected_filemode(filepath):
    filepath = os.path.join('/', filepath)

    if filepath == os.path.join('/usr/share/tarantool/', project_name, 'VERSION'):
        return '0644'

    if filepath.startswith('/usr/share/tarantool/'):
        return '0755'

    return '0644'


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

        filemode_raw = rpm.headers['filemodes'][i]
        filemode = oct((filemode_raw + 2**32) & 0o777).replace('0o', '0')

        assert filemode == expected_filemode(filepath)



def assert_file_modes(basedir):
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
        filemode = oct(file_stat.st_mode & 0o777).replace('0o', '0')

        assert filemode == expected_filemode(filename)


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
    project_tmpfiles_conf_file = os.path.join(basedir, 'usr/lib/tmpfiles.d', '%s.conf' % project_name )
    assert open(project_tmpfiles_conf_file).read().find('d /var/run/tarantool') != -1

    # check version file
    validate_version_file(distribution_dir)


def run_command_on_image(docker_client, image_name, command):
    command = '/bin/bash -c "{}"'.format(command.replace('"', '\\"'))
    output = docker_client.containers.run(
        image_name,
        command,
        remove=True
    )
    return output.decode("utf-8").strip()


def test_tgz_pack(project_path, tgz_archive, tmpdir):
    with tarfile.open(name=tgz_archive['name']) as tgz_arch:
        # usr/share/tarantool is added to coorectly run assert_file_modes
        distribution_dir = os.path.join(tmpdir, 'usr/share/tarantool', project_name)
        os.makedirs(distribution_dir, exist_ok=True)

        tgz_arch.extractall(path=os.path.join(tmpdir, 'usr/share/tarantool'))
        assert_dir_contents(recursive_listdir(distribution_dir))

        validate_version_file(distribution_dir)
        assert_file_modes(tmpdir)


def test_rpm_pack(project_path, rpm_archive, tmpdir):
    ps = subprocess.Popen(
        ['rpm2cpio', rpm_archive['name']], stdout=subprocess.PIPE)
    subprocess.check_output(['cpio', '-idmv'], stdin=ps.stdout, cwd=tmpdir)
    ps.wait()
    assert ps.returncode == 0, "Error during extracting files from rpm archive"

    if not tarantool_enterprise_is_used():
        assert_tarantool_dependency_rpm(rpm_archive['name'])

    check_package_files(tmpdir, project_path)
    assert_files_mode_and_owner_rpm(rpm_archive['name'])


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
        check_package_files(data_dir, project_path)
        assert_file_modes(data_dir)

    # check control.tar.xz
    with tarfile.open(name=os.path.join(tmpdir, 'control.tar.xz')) as control_arch:
        control_dir = os.path.join(tmpdir, 'control')
        control_arch.extractall(path=control_dir)

        for filename in ['control', 'preinst', 'postinst']:
            assert os.path.exists(os.path.join(control_dir, filename))

        if not tarantool_enterprise_is_used():
            assert_tarantool_dependency_deb(os.path.join(control_dir, 'control'))

        # check if postinst script set owners correctly
        with open(os.path.join(control_dir, 'postinst')) as postinst_script_file:
            postinst_script = postinst_script_file.read()
            assert 'chown -R root:root /usr/share/tarantool/{}'.format(project_name) in postinst_script
            assert 'chown root:root /etc/systemd/system/{}.service'.format(project_name) in postinst_script
            assert 'chown root:root /etc/systemd/system/{}@.service'.format(project_name) in postinst_script
            assert 'chown root:root /usr/lib/tmpfiles.d/{}.conf'.format(project_name) in postinst_script


def test_docker_pack(project_path, docker_image, tmpdir, docker_client):
    image_name = docker_image['name']
    container = docker_client.containers.create(image_name)

    container_distribution_dir = '/usr/share/tarantool/{}'.format(project_name)

    # check if distribution dir was created
    command = '[ -d "{}" ] && echo true || echo false'.format(container_distribution_dir)
    output = run_command_on_image(docker_client, image_name, command)
    assert output == 'true'

    # get distribution dir contents
    arhive_path = os.path.join(tmpdir, 'distribution_dir.tar')
    with open(arhive_path, 'wb') as f:
        bits, _ = container.get_archive(container_distribution_dir)
        for chunk in bits:
            f.write(chunk)

    with tarfile.open(arhive_path) as arch:
        arch.extractall(path=os.path.join(tmpdir, 'usr/share/tarantool'))
    os.remove(arhive_path)

    assert_dir_contents(
        recursive_listdir(os.path.join(tmpdir, 'usr/share/tarantool/', project_name)),
        skip_tarantool_binaries=True
    )

    assert_file_modes(tmpdir)
    container.remove()

    if tarantool_enterprise_is_used():
        # check tarantool and tarantoolctl binaries
        command = '[ -d "/usr/share/tarantool/tarantool-enterprise/" ] && echo true || echo false'
        output = run_command_on_image(docker_client, image_name, command)
        assert output == 'true'

        command = 'cd /usr/share/tarantool/tarantool-enterprise/ && find .'
        output = run_command_on_image(docker_client, image_name, command)

        files_list = output.split('\n')
        files_list.remove('.')

        dir_contents = [
            os.path.normpath(filename)
            for filename in files_list
        ]

        assert 'tarantool' in dir_contents
        assert 'tarantoolctl' in dir_contents
    else:
        # check if tarantool was installed
        command = 'yum list installed 2>/dev/null | grep tarantool'
        output = run_command_on_image(docker_client, image_name, command)

        packages_list = output.split('\n')
        assert any(['tarantool' in package for package in packages_list])

        # check tarantool version
        command = 'yum info tarantool'
        output = run_command_on_image(docker_client, image_name, command)

        m = re.search(r'Version\s+:\s+(\d+)\.(\d+).', output)
        assert m is not None
        installed_version = m.groups()

        m = re.search(r'(\d+)\.(\d+)\.\d+', tarantool_version())
        assert m is not None
        expected_version = m.groups()

        assert installed_version == expected_version


def test_docker_e2e(project_path, docker_image, tmpdir, docker_client):
    image_name = docker_image['name']
    environment = [
        'TARANTOOL_INSTANCE_NAME=instance-1',
        'TARANTOOL_ADVERTISE_URI=3302',
        'TARANTOOL_CLUSTER_COOKIE=secret',
        'TARANTOOL_HTTP_PORT=8082',
    ]

    container = docker_client.containers.run(
        image_name,
        environment=environment,
        ports={'8082': '8082'},
        name='{}-instance-1'.format(project_name),
        detach=True,
        # remove=True
    )

    assert container.status == 'created'
    assert wait_for_container_start(container)

    container_logs = container.logs().decode('utf-8')
    m = re.search(r'Auto-detected IP to be "(\d+\.\d+\.\d+\.\d+)', container_logs)
    assert m is not None
    ip = m.groups()[0]

    admin_api_url = 'http://127.0.0.1:8082/admin/api'

    # join instance
    query = '''
        mutation {{
        j1: join_server(
            uri:"{}:3302",
            roles: ["vshard-router", "app.roles.custom"]
            instance_uuid: "aaaaaaaa-aaaa-4000-b000-000000000001"
            replicaset_uuid: "aaaaaaaa-0000-4000-b000-000000000000"
        )
    }}
    '''.format(ip)

    r = requests.post(admin_api_url, json={'query': query})
    assert r.status_code == 200
    resp = r.json()
    assert 'data' in resp
    assert 'j1' in resp['data']
    assert resp['data']['j1'] is True

    # check status and alias
    query = '''
        query {
        instance: cluster {
            self {
                alias
            }
        }
        replicaset: replicasets(uuid: "aaaaaaaa-0000-4000-b000-000000000000") {
            status
        }
    }
    '''

    r = requests.post(admin_api_url, json={'query': query})
    assert r.status_code == 200
    resp = r.json()
    assert 'data' in resp
    assert 'replicaset' in resp['data'] and 'instance' in resp['data']
    assert resp['data']['replicaset'][0]['status'] == 'healthy'
    assert resp['data']['instance']['self']['alias'] == 'instance-1'

    # restart instance
    container.restart()
    wait_for_container_start(container)

    # check instance restarted
    r = requests.post(admin_api_url, json={'query': query})
    assert r.status_code == 200
    resp = r.json()
    assert 'data' in resp
    assert 'replicaset' in resp['data'] and 'instance' in resp['data']
    assert resp['data']['replicaset'][0]['status'] == 'healthy'
    assert resp['data']['instance']['self']['alias'] == 'instance-1'

    container.stop()


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
