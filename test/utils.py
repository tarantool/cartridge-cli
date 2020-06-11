#!/usr/bin/python3

import os
import subprocess
import rpmfile
import re
import time
import sys
import psutil
import atexit
import glob
import json
import requests
import tenacity

__tarantool_version = None

# DEFAULT_RUN_DIR = 'tmp/run'
DEFAULT_RUN_DIR = 'tmp'
# DEFAULT_DATA_DIR = 'tmp/data'
DEFAULT_CFG = 'instances.yml'

DEFAULT_SCRIPT = 'init.lua'
DEFAULT_STATEBOARD_SCRIPT = 'stateboard.init.lua'

STATUS_NOT_STARTED = '\x1b[36mNot started\x1b[0m\x1b[0m'  # 'NOT STARTED'
STATUS_RUNNING = '\x1b[32mRunning\x1b[0m\x1b[0m'  # 'RUNNING'
STATUS_STOPPED = '\x1b[33mStopped\x1b[0m\x1b[0m'  # 'STOPPED'
# STATUS_FAILED = 'FAILED'


# #############
# Class Archive
# #############
class Archive:
    def __init__(self, filepath, project):
        self.filepath = filepath
        self.filename = os.path.basename(filepath)
        self.project = project


# ###########
# Class Image
# ###########
class Image:
    def __init__(self, name, project):
        self.name = name
        self.project = project


# #####################
# Class InstanceProcess
# #####################
class InstanceProcess():
    def __init__(self, process):
        self._process = process
        self._pid = process.pid
        self._ppid = process.ppid()

        self.name = process.name()
        self.cmd = process.cmdline()

        env = process.environ()

        assert 'TARANTOOL_APP_NAME' in env
        if 'TARANTOOL_INSTANCE_NAME' in env:
            self.id = get_instance_id(
                env['TARANTOOL_APP_NAME'],
                env['TARANTOOL_INSTANCE_NAME'],
            )
        else:
            self.id = env['TARANTOOL_APP_NAME']

        self._env = {
            'TARANTOOL_APP_NAME': env.get('TARANTOOL_APP_NAME'),
            'TARANTOOL_INSTANCE_NAME': env.get('TARANTOOL_INSTANCE_NAME'),
            'TARANTOOL_CFG': env.get('TARANTOOL_CFG'),
            'TARANTOOL_CONSOLE_SOCK': env.get('TARANTOOL_CONSOLE_SOCK'),
            'TARANTOOL_PID_FILE': env.get('TARANTOOL_PID_FILE'),
            'TARANTOOL_WORKDIR': env.get('TARANTOOL_WORKDIR'),
        }

    def is_running(self):
        return self._process.is_running() and self._process.status() != psutil.STATUS_ZOMBIE

    def getenv(self, name):
        return self._env.get(name)


# #########
# Class Cli
# #########
class Cli():
    def __init__(self, cartridge_cmd):
        self._cartridge_cmd = cartridge_cmd
        self._children = []
        self._instances = dict()

    def start(self, project, instances=[], daemonized=False, stateboard=False, stateboard_only=False,
              cfg=None, script=None, run_dir=None):

        cmd = [self._cartridge_cmd, 'start']
        if daemonized:
            cmd.append('-d')
        if stateboard:
            cmd.append('--stateboard')
        if stateboard_only:
            cmd.append('--stateboard-only')
        if cfg is not None:
            cmd.extend(['--cfg', cfg])
        if script is not None:
            cmd.extend(['--script', script])
        if run_dir is not None:
            cmd.extend(['--run-dir', run_dir])
        # if data_dir is not None:
        #     cmd.extend(['--data-dir', data_dir])

        cmd.extend(instances)

        self._subprocess = subprocess.Popen(
            cmd, cwd=project.path,
            stdout=sys.stdout,
            stderr=sys.stderr,
        )

        self._pid = self._subprocess.pid
        self._process = psutil.Process(self._pid)

        time.sleep(5)  # let cli to start instances

        if not self._process.is_running():
            assert self._subprocess.returncode == 0

        self._collect_instances(project, run_dir)

    def stop(self, project, instances=[], run_dir=None, cfg=None, stateboard=False, stateboard_only=False):
        cmd = [self._cartridge_cmd, 'stop']
        if stateboard:
            cmd.append('--stateboard')
        if stateboard_only:
            cmd.append('--stateboard-only')
        if run_dir is not None:
            cmd.extend(['--run-dir', run_dir])
        if cfg is not None:
            cmd.extend(['--cfg', cfg])

        cmd.extend(instances)

        self._subprocess = subprocess.Popen(
            cmd, cwd=project.path,
            stdout=sys.stdout,
            stderr=sys.stderr,
        )
        self._pid = self._subprocess.pid
        self._process = psutil.Process(self._pid)

        time.sleep(0.5)  # let cli to terminate instances

    def get_status(self, project, instances=[], run_dir=None, cfg=None,
                   stateboard=False, stateboard_only=False):
        cmd = [self._cartridge_cmd, 'status']
        if stateboard:
            cmd.append('--stateboard')
        if stateboard_only:
            cmd.append('--stateboard-only')
        if run_dir is not None:
            cmd.extend(['--run-dir', run_dir])
        if cfg is not None:
            cmd.extend(['--cfg', cfg])

        cmd.extend(instances)

        rc, output = run_command_and_get_output(cmd, cwd=project.path)
        assert rc == 0

        status = {}

        for line in output.split('\n'):
            if line == '':
                continue

            m = re.match(r'^\x1b\[36m(\S+):\s+(.+)$', line)
            assert m is not None

            instance_id = m.group(1)
            instance_status = m.group(2)

            # msg = logfmt.parse_line(line)['msg']
            # m = re.match(r'^(\S+):\s+(.+)$', msg)
            # assert m is not None

            # instance_id = m.group(1)
            # instance_status = m.group(2)

            assert instance_id not in status
            status[instance_id] = instance_status

        return status

    def _collect_instances(self, project, run_dir):
        if run_dir is None:
            run_dir = DEFAULT_RUN_DIR

        for pid_filepath in glob.glob(os.path.join(project.path, run_dir, "*.pid")):
            with open(pid_filepath) as pid_file:
                pid = int(pid_file.read().strip())
                self._children.append(psutil.Process(pid))

        for child in self._children:
            instance = InstanceProcess(child)
            assert instance.id not in self._instances
            self._instances[instance.id] = instance

        atexit.register(self.terminate)

    def get_child_instances(self):
        return self._instances

    def is_running(self):
        return self._process.is_running() and self._process.status() != psutil.STATUS_ZOMBIE

    def terminate(self):
        self._subprocess.terminate()
        for child in self._children:
            if child.is_running():
                child.terminate()


# #######
# Helpers
# #######
def tarantool_version():
    global __tarantool_version
    if __tarantool_version is None:
        __tarantool_version = subprocess.check_output(['tarantool', '-V']).decode('ascii').split('\n')[0]

    return __tarantool_version


def tarantool_repo_version():
    m = re.search(r'(\d+).(\d+)', tarantool_version())
    assert m is not None
    major, minor = m.groups()
    return '{}_{}'.format(major, minor)


def tarantool_enterprise_is_used():
    return tarantool_version().startswith('Tarantool Enterprise')


def create_project(cartridge_cmd, module_tmpdir, project_name, template):
    cmd = [
        cartridge_cmd, "create",
        "--name", project_name,
        "--template", template
    ]
    process = subprocess.run(cmd, cwd=module_tmpdir)
    assert process.returncode == 0, \
        "Error during creating the project"
    return os.path.join(module_tmpdir, project_name)


def find_archive(path, project_name, arch_ext):
    with os.scandir(path) as it:
        for entry in it:
            if entry.name.startswith(project_name) and entry.name.endswith('.' + arch_ext) and entry.is_file():
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


def assert_distribution_dir_contents(dir_contents, project, exclude_files=set()):
    without_rocks = {x for x in dir_contents if not x.startswith('.rocks')}

    assert without_rocks == project.distribution_files.difference(exclude_files)
    assert all(x in dir_contents for x in project.rocks_content)


def assert_filemode(project, filepath, filemode):
    filepath = os.path.join('/', filepath)

    if filepath == os.path.join('/usr/share/tarantool/', project.name, 'VERSION'):
        assert filemode & 0o777 == 0o644
    elif filepath.startswith('/etc/systemd/system/'):
        assert filemode & 0o777 == 0o644
    elif filepath.startswith('/usr/lib/tmpfiles.d/'):
        assert filemode & 0o777 == 0o644
    elif filepath.startswith('/usr/share/tarantool/'):
        # a+r for files, a+rx for directories
        required_bits = 0o555 if os.path.isdir(filepath) else 0o444
        assert filemode & required_bits == required_bits


def assert_filemodes(project, basedir):
    known_dirs = {
        'etc', 'etc/systemd', 'etc/systemd/system',
        'usr', 'usr/share', 'usr/share/tarantool',
        'usr/lib', 'usr/lib/tmpfiles.d'
    }
    filenames = recursive_listdir(basedir) - known_dirs

    for filename in filenames:
        # we don't check fileowner here because it's set in postinst script

        # check filemode
        if filename.startswith(os.path.join('usr/share/tarantool/', project.name, '.rocks')):
            continue

        # get filestat
        file_stat = os.stat(os.path.join(basedir, filename))
        filemode = file_stat.st_mode
        assert_filemode(project, filename, filemode)


def validate_version_file(project, distribution_dir):
    version_filepath = os.path.join(distribution_dir, 'VERSION')
    assert os.path.exists(version_filepath)

    version_file_content = {}
    with open(version_filepath, 'r') as version_file:
        file_lines = version_file.read().strip().split('\n')
        for line in file_lines:
            m = re.match(r'^([^=]+)=([^=]+)$', line)
            assert m is not None

            key, version = m.groups()
            version_file_content[key] = version

    for key in project.version_file_keys:
        assert key in version_file_content


def assert_files_mode_and_owner_rpm(project, filename):
    DIRNAMES_TAG = 1118
    DIRINDEXES_TAG = 1116
    PAYLOADDIGEST_TAG = 5092
    PAYLOADDIGESTALGO_TAG = 5093

    expected_tags = [
        'basenames', DIRNAMES_TAG, DIRINDEXES_TAG, 'filemodes',
        'fileusername', 'filegroupname',
        PAYLOADDIGEST_TAG, PAYLOADDIGESTALGO_TAG,
    ]

    with rpmfile.open(filename) as rpm:
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
            if filepath.startswith(os.path.join('/usr/share/tarantool/', project.name, '.rocks')):
                continue

            filemode = rpm.headers['filemodes'][i]
            assert_filemode(project, filepath, filemode)


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
    with rpmfile.open(filename) as rpm:
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


def check_systemd_dir(project, basedir):
    systemd_dir = (os.path.join(basedir, 'etc/systemd/system'))
    assert os.path.exists(systemd_dir)

    systemd_files = recursive_listdir(systemd_dir)

    assert len(systemd_files) == 3
    assert '{}.service'.format(project.name) in systemd_files
    assert '{}@.service'.format(project.name) in systemd_files
    assert '{}-stateboard.service'.format(project.name) in systemd_files


def check_package_files(project, basedir):
    # check if only theese files are delivered
    for filename in recursive_listdir(basedir):
        assert any([
            filename.startswith(prefix) or prefix.startswith(filename)
            for prefix in [
                os.path.join('usr/share/tarantool', project.name),
                'etc/systemd/system',
                'usr/lib/tmpfiles.d'
            ]
        ])

    # check distribution dir content
    distribution_dir = os.path.join(basedir, 'usr/share/tarantool', project.name)
    assert os.path.exists(distribution_dir)
    assert_distribution_dir_contents(
        dir_contents=recursive_listdir(distribution_dir),
        project=project,
    )

    # check systemd dir content
    check_systemd_dir(project, basedir)

    # check tmpfiles conf
    project_tmpfiles_conf_file = os.path.join(basedir, 'usr/lib/tmpfiles.d', '%s.conf' % project.name)
    with open(project_tmpfiles_conf_file) as f:
        assert f.read().find('d /var/run/tarantool') != -1

    # check version file
    validate_version_file(project, distribution_dir)


def run_command_and_get_output(cmd, cwd=None, env=None):
    process = subprocess.Popen(
        cmd,
        env=env,
        cwd=cwd,
        stderr=subprocess.STDOUT,
        stdout=subprocess.PIPE
    )

    stdout, _ = process.communicate()
    stdout = stdout.decode('utf-8')

    # This print is here to make running tests with -s flag more verbose
    print(stdout)

    return process.returncode, stdout


def find_image(docker_client, project_name):
    for image in docker_client.images.list():
        for t in image.tags:
            if t.startswith(project_name):
                return t


def delete_image(docker_client, image_name):
    if docker_client.images.list(image_name):
        # remove all image containers
        containers = docker_client.containers.list(
            all=True,
            filters={'ancestor': image_name}
        )

        for c in containers:
            c.remove(force=True)

        # remove image itself
        docker_client.images.remove(image_name)


def get_instance_id(app_name, instance_name):
    return '{}.{}'.format(app_name, instance_name)


def get_stateboard_name(app_name):
    return '{}-stateboard'.format(app_name)


def check_running_instance(child_instances, app_path, app_name, instance_id,
                           cfg=DEFAULT_CFG,
                           script=DEFAULT_SCRIPT,
                           run_dir=DEFAULT_RUN_DIR):
    assert instance_id in child_instances
    instance = child_instances[instance_id]

    assert instance.is_running()

    # assert instance.cmd == ["tarantool", os.path.join(app_path, script)]
    assert len(instance.cmd) == 2
    assert instance.cmd[0].endswith("tarantool")
    assert instance.cmd[1] == os.path.join(app_path, script)

    instance_name = instance_id.split('.', 1)[1]

    assert instance.getenv('TARANTOOL_APP_NAME') == app_name
    assert instance.getenv('TARANTOOL_INSTANCE_NAME') == instance_name
    assert instance.getenv('TARANTOOL_CFG') == os.path.join(app_path, cfg)
    assert instance.getenv('TARANTOOL_PID_FILE') == os.path.join(app_path, run_dir, '%s.pid' % instance_id)
    assert instance.getenv('TARANTOOL_CONSOLE_SOCK') == os.path.join(app_path, run_dir, '%s.sock' % instance_id)
    # assert instance.getenv('TARANTOOL_WORKDIR') == os.path.join(app_path, data_dir, instance_id)


def check_started_stateboard(child_instances, app_path, app_name,
                             cfg=DEFAULT_CFG,
                             script=DEFAULT_STATEBOARD_SCRIPT,
                             run_dir=DEFAULT_RUN_DIR):
    stateboard_name = get_stateboard_name(app_name)

    assert stateboard_name in child_instances
    instance = child_instances[stateboard_name]

    assert instance.is_running()

    # assert instance.cmd == ["tarantool",  os.path.join(app_path, script)]
    assert len(instance.cmd) == 2
    assert instance.cmd[0].endswith("tarantool")
    assert instance.cmd[1] == os.path.join(app_path, script)

    assert instance.getenv('TARANTOOL_APP_NAME') == stateboard_name
    assert instance.getenv('TARANTOOL_CFG') == os.path.join(app_path, cfg)
    assert instance.getenv('TARANTOOL_PID_FILE') == os.path.join(app_path, run_dir, '%s.pid' % stateboard_name)
    assert instance.getenv('TARANTOOL_CONSOLE_SOCK') == os.path.join(app_path, run_dir, '%s.sock' % stateboard_name)
    # assert instance.getenv('TARANTOOL_WORKDIR') == os.path.join(app_path, data_dir, stateboard_name)


def check_instances_running(cli, project, instance_ids=[],
                            stateboard=False, stateboard_only=False,
                            daemonized=False,
                            cfg=DEFAULT_CFG,
                            script=DEFAULT_SCRIPT,
                            run_dir=DEFAULT_RUN_DIR):
    child_instances = cli.get_child_instances()

    running_instances_count = len([
        instance
        for instance in child_instances.values()
        if instance.is_running()
    ])

    if stateboard_only:
        assert running_instances_count == 1
    elif stateboard:
        assert running_instances_count == len(instance_ids) + 1
    else:
        assert running_instances_count == len(instance_ids)

    if stateboard:
        check_started_stateboard(child_instances, project.path, project.name,
                                 cfg=cfg, run_dir=run_dir)
    if not stateboard_only:
        for instance_id in instance_ids:
            check_running_instance(child_instances, project.path, project.name, instance_id,
                                   script=script, cfg=cfg, run_dir=run_dir)

    if not daemonized:
        assert cli.is_running()
    else:
        assert not cli.is_running()


def check_instances_stopped(cli, project, instance_ids=[], run_dir=DEFAULT_RUN_DIR,
                            stateboard=False, stateboard_only=False):
    child_instances = cli.get_child_instances()

    if not stateboard_only:
        for instance_id in instance_ids:
            assert instance_id in child_instances
            instance = child_instances[instance_id]

            assert not instance.is_running()

    if stateboard:
        stateboard_name = get_stateboard_name(project.name)

        assert stateboard_name in child_instances
        instance = child_instances[stateboard_name]

        assert not instance.is_running()

    assert not cli.is_running()


def patch_cartridge_proc_titile(project):
    filepath = os.path.join(project.path, '.rocks/share/tarantool/cartridge.lua')
    with open(filepath) as f:
        data = f.read()

    patched_data = data.replace(
        'title.update(box_opts.custom_proc_title)',
        '-- title.update(box_opts.custom_proc_title)'
    )

    with open(filepath, 'w') as f:
        f.write(patched_data)


def create_replicaset(admin_api_url, advertise_uris, roles):
    query = '''
        mutation {{
        j1: cluster{{ edit_topology(
            replicasets: [{{
                join_servers: [{servers}],
                roles: {roles},
            }}]
        ) {{ replicasets {{ uuid }} }}
        }}
    }}
    '''.format(
        servers=", ".join([
            '{{ uri: "{uri}" }}'.format(uri=uri)
            for uri in advertise_uris
        ]),
        roles=json.dumps(roles),
    )

    r = requests.post(admin_api_url, json={'query': query})
    assert r.status_code == 200
    resp = r.json()
    assert 'data' in resp

    replicaset_uuid = resp['data']['j1']['edit_topology']['replicasets'][0]['uuid']

    return replicaset_uuid


@tenacity.retry(stop=tenacity.stop_after_delay(10))
def wait_for_replicaset_is_healthy(admin_api_url, replicaset_uuid):
    query = '''
        query {{
        replicaset: replicasets(uuid: "{uuid}") {{
            status
        }}
    }}
    '''.format(uuid=replicaset_uuid)

    r = requests.post(admin_api_url, json={'query': query})
    assert r.status_code == 200
    resp = r.json()

    status = resp['data']['replicaset'][0]['status']
    assert status == 'healthy'


def get_replicaset_roles(admin_api_url, replicaset_uuid):
    query = '''
        query {{
        replicaset: replicasets(uuid: "{uuid}") {{
            roles
        }}
    }}
    '''.format(uuid=replicaset_uuid)

    r = requests.post(admin_api_url, json={'query': query})
    assert r.status_code == 200
    resp = r.json()

    return resp['data']['replicaset'][0]['roles']


def bootstrap_vshard(admin_api_url):
    query = '''
        mutation {
            bootstrap: bootstrap_vshard
        }
    '''

    r = requests.post(admin_api_url, json={'query': query})
    assert r.status_code == 200
    resp = r.json()
    assert 'data' in resp

    assert resp['data']['bootstrap'] is True
