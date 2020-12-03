import os
import subprocess
import rpmfile
import re
import sys
import psutil
import glob
import json
import requests
import tenacity
import time
import yaml
import tarfile
import gzip

from docker import APIClient

__tarantool_version = None

DEFAULT_RUN_DIR = 'tmp/run'
DEFAULT_DATA_DIR = 'tmp/data'
DEFAULT_LOG_DIR = 'tmp/log'
DEFAULT_CFG = 'instances.yml'
DEFAULT_RPL_CFG = 'replicasets.yml'

DEFAULT_SCRIPT = 'init.lua'
DEFAULT_STATEBOARD_SCRIPT = 'stateboard.init.lua'

STATUS_NOT_STARTED = 'NOT STARTED'
STATUS_RUNNING = 'RUNNING'
STATUS_STOPPED = 'STOPPED'
STATUS_FAILED = 'FAILED'


def get_logs(output):
    rgx = re.compile(r'^\s+\S+\s+(?P<msg>\S+.*)$')
    logs = []

    for line in output.split('\n'):
        if line == '':
            continue

        m = rgx.match(line)
        assert m is not None
        logs.append(m.group("msg").strip())

    return logs


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
    def __init__(self, pid):
        self._env = {}
        self._process = None
        self.name = None
        self.cmd = None
        self.pid_not_exists = False

        if not psutil.pid_exists(pid):
            self.pid_not_exists = True
            return

        self._process = psutil.Process(pid)

        self.name = self._process.name()
        self.cmd = self._process.cmdline()
        self.cwd = self._process.cwd()

        env = self._process.environ()
        self._env = {
            'TARANTOOL_APP_NAME': env.get('TARANTOOL_APP_NAME'),
            'TARANTOOL_INSTANCE_NAME': env.get('TARANTOOL_INSTANCE_NAME'),
            'TARANTOOL_CFG': env.get('TARANTOOL_CFG'),
            'TARANTOOL_CONSOLE_SOCK': env.get('TARANTOOL_CONSOLE_SOCK'),
            'TARANTOOL_PID_FILE': env.get('TARANTOOL_PID_FILE'),
            'TARANTOOL_WORKDIR': env.get('TARANTOOL_WORKDIR'),
            'NOTIFY_SOCKET': env.get('NOTIFY_SOCKET')
        }

    def is_running(self):
        if self.pid_not_exists:
            return False

        return self._process.is_running() and self._process.status() != psutil.STATUS_ZOMBIE

    def getenv(self, name):
        return self._env.get(name)


def get_instance_id_by_pid_filepath(pid_filepath):
    filename = os.path.basename(pid_filepath)
    instance_id = filename.replace(".pid", "")
    return instance_id


# #########
# Class Cli
# #########
class Cli():
    def __init__(self, cartridge_cmd):
        self._cartridge_cmd = cartridge_cmd

        self._processes = []
        self._instance_pids = set()
        self._subprocess = None

    def start(self, project, instances=[], daemonized=False, stateboard=False, stateboard_only=False,
              cfg=None, script=None, run_dir=None, data_dir=None, log_dir=None, timeout=None,
              capture_output=False, exp_rc=0):
        cmd = [self._cartridge_cmd, 'start']
        if daemonized:
            cmd.append('-d')
        if stateboard:
            cmd.append('--stateboard')
        if stateboard_only:
            cmd.append('--stateboard-only')
        if timeout is not None:
            cmd.extend(['--timeout', timeout])
        if cfg is not None:
            cmd.extend(['--cfg', cfg])
        if script is not None:
            cmd.extend(['--script', script])
        if run_dir is not None:
            cmd.extend(['--run-dir', run_dir])
        if data_dir is not None:
            cmd.extend(['--data-dir', data_dir])
        if log_dir is not None:
            cmd.extend(['--log-dir', log_dir])

        cmd.extend(instances)

        if not capture_output:
            self._subprocess = subprocess.Popen(
                cmd, cwd=project.path,
                stdout=sys.stdout,
                stderr=sys.stderr,
            )
        else:
            self._subprocess = subprocess.Popen(
                cmd, cwd=project.path,
                stdout=subprocess.PIPE,
                stderr=subprocess.STDOUT,
            )

        self._process = psutil.Process(self._subprocess.pid)
        self._processes.append(self._process)

        run_dir = run_dir if run_dir is not None else DEFAULT_RUN_DIR

        if daemonized:
            rc = self.wait(project, run_dir=run_dir)
            assert rc == exp_rc
            if capture_output:
                output = self._subprocess.stdout.read().decode('utf-8')
                logs = get_logs(output)

                return logs

    def wait(self, project, capture_output=False, run_dir=DEFAULT_RUN_DIR):
        self._subprocess.wait(timeout=10)
        self.get_child_instances(project, run_dir=run_dir)

        return self._subprocess.returncode

    def stop(self, project, instances=[], run_dir=None, cfg=None, force=False,
             stateboard=False, stateboard_only=False):
        cmd = [self._cartridge_cmd, 'stop']
        if force:
            cmd.append('--force')
        if stateboard:
            cmd.append('--stateboard')
        if stateboard_only:
            cmd.append('--stateboard-only')
        if run_dir is not None:
            cmd.extend(['--run-dir', run_dir])
        if cfg is not None:
            cmd.extend(['--cfg', cfg])

        cmd.extend(instances)

        process = subprocess.run(
            cmd, cwd=project.path,
            stdout=sys.stdout,
            stderr=sys.stderr,
        )
        assert process.returncode == 0

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

        logs = get_logs(output)

        for msg in logs:
            m = re.match(r'^(\S+):\s+(.+)$', msg)
            assert m is not None

            instance_id = m.group(1)
            instance_status = m.group(2)

            assert instance_id not in status
            status[instance_id] = instance_status

        return status

    def get_logs(self, project, instances=[], n=None, log_dir=None, run_dir=None, cfg=None,
                 stateboard=False, stateboard_only=False):
        cmd = [self._cartridge_cmd, 'log']
        if n is not None:
            cmd.append('-n{}'.format(n))
        if stateboard:
            cmd.append('--stateboard')
        if stateboard_only:
            cmd.append('--stateboard-only')
        if log_dir is not None:
            cmd.extend(['--log-dir', log_dir])
        if run_dir is not None:
            cmd.extend(['--run-dir', run_dir])
        if cfg is not None:
            cmd.extend(['--cfg', cfg])

        cmd.extend(instances)

        rc, output = run_command_and_get_output(cmd, cwd=project.path)
        assert rc == 0

        logs = {}

        for line in output.split('\n'):
            m = re.match(r'^(\S+)\s+\|\s+(.+)$', line)
            if m is None:
                continue

            instance_id = m.group(1)
            instance_log_line = m.group(2)

            if instance_log_line == "entering the event loop":
                continue

            if instance_id not in logs:
                logs[instance_id] = []

            logs[instance_id].append(instance_log_line)

        return logs

    def clean(self, project, instances=[], log_dir=None, run_dir=None, cfg=None, data_dir=None,
              stateboard=False, stateboard_only=False, exp_rc=0):
        cmd = [self._cartridge_cmd, 'clean']
        if stateboard:
            cmd.append('--stateboard')
        if stateboard_only:
            cmd.append('--stateboard-only')
        if cfg is not None:
            cmd.extend(['--cfg', cfg])
        if run_dir is not None:
            cmd.extend(['--run-dir', run_dir])
        if data_dir is not None:
            cmd.extend(['--data-dir', data_dir])
        if log_dir is not None:
            cmd.extend(['--log-dir', log_dir])

        cmd.extend(instances)

        process = subprocess.Popen(
            cmd, cwd=project.path,
            stdout=subprocess.PIPE,
            stderr=subprocess.STDOUT,
        )

        process.wait(timeout=10)
        assert process.returncode == exp_rc

        output = process.stdout.read().decode('utf-8')
        logs = get_logs(output)

        return logs

    def get_child_instances(self, project, run_dir=DEFAULT_RUN_DIR):
        instances = dict()

        for pid_filepath in glob.glob(os.path.join(project.path, run_dir, "*.pid")):
            with open(pid_filepath) as pid_file:
                pid = int(pid_file.read().strip())
                self._instance_pids.add(pid)

                instance = InstanceProcess(pid)
                instance_id = get_instance_id_by_pid_filepath(pid_filepath)
                assert instance_id not in instances
                instances[instance_id] = instance

        return instances

    def is_running(self):
        return self._process.is_running() and self._process.status() != psutil.STATUS_ZOMBIE

    def terminate(self):
        for process in self._processes:
            if process.is_running():
                process.kill()

        # kill all instance processes
        for pid in self._instance_pids:
            if not psutil.pid_exists(pid):
                continue
            process = psutil.Process(pid)
            if process.is_running():
                process.kill()


# #######################
# Class InstanceContainer
# #######################
class InstanceContainer:
    def __init__(self, container, instance_name, http_port, advertise_port):
        self.container = container
        self.instance_name = instance_name
        self.http_port = http_port
        self.advertise_port = advertise_port


class ProjectContainer:
    def __init__(self, container, project, http_port):
        self.container = container
        self.project = project
        self.http_port = http_port


# #######
# Helpers
# #######
def tarantool_version():
    global __tarantool_version
    if __tarantool_version is None:
        __tarantool_version = subprocess.check_output(['tarantool', '-V']).decode('ascii').split('\n')[0]

    return __tarantool_version


def tarantool_short_version():
    m = re.search(r'(\d+).(\d+)', tarantool_version())
    assert m is not None
    major, minor = m.groups()

    short_version = '{}.{}'.format(major, minor)
    return short_version


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

        min_version = re.findall(r'\d+\.\d+\.\d+-\d+-\S+', tarantool_version())[0]
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
                           daemonized=False,
                           cfg=DEFAULT_CFG,
                           script=DEFAULT_SCRIPT,
                           run_dir=DEFAULT_RUN_DIR,
                           data_dir=DEFAULT_DATA_DIR,
                           log_dir=DEFAULT_LOG_DIR,
                           skip_env_checks=False):
    assert instance_id in child_instances
    instance = child_instances[instance_id]

    assert instance.is_running()
    assert instance.cwd == app_path

    if skip_env_checks:
        return

    assert instance.cmd == ["tarantool", os.path.join(app_path, script)]

    instance_name = instance_id.split('.', 1)[1]

    assert instance.getenv('TARANTOOL_APP_NAME') == app_name
    assert instance.getenv('TARANTOOL_INSTANCE_NAME') == instance_name
    assert instance.getenv('TARANTOOL_CFG') == os.path.join(app_path, cfg)
    assert instance.getenv('TARANTOOL_PID_FILE') == os.path.join(app_path, run_dir, '%s.pid' % instance_id)
    assert instance.getenv('TARANTOOL_CONSOLE_SOCK') == os.path.join(app_path, run_dir, '%s.control' % instance_id)
    assert instance.getenv('TARANTOOL_WORKDIR') == os.path.join(app_path, data_dir, instance_id)

    if daemonized:
        assert os.path.exists(os.path.join(app_path, log_dir, '%s.log' % instance_id))

        notify_socket_path = os.path.join(app_path, run_dir, '%s.notify' % instance_id)
        assert(os.path.exists(notify_socket_path))
        assert instance.getenv('NOTIFY_SOCKET') == notify_socket_path


def check_started_stateboard(child_instances, app_path, app_name,
                             daemonized=False,
                             cfg=DEFAULT_CFG,
                             script=DEFAULT_STATEBOARD_SCRIPT,
                             run_dir=DEFAULT_RUN_DIR,
                             data_dir=DEFAULT_DATA_DIR,
                             log_dir=DEFAULT_LOG_DIR,
                             skip_env_checks=False):
    stateboard_name = get_stateboard_name(app_name)

    assert stateboard_name in child_instances
    instance = child_instances[stateboard_name]

    assert instance.is_running()

    assert instance.cmd[0].endswith("tarantool")

    if skip_env_checks:
        return

    assert instance.getenv('TARANTOOL_APP_NAME') == stateboard_name
    assert instance.getenv('TARANTOOL_CFG') == os.path.join(app_path, cfg)
    assert instance.getenv('TARANTOOL_PID_FILE') == os.path.join(app_path, run_dir, '%s.pid' % stateboard_name)
    assert instance.getenv('TARANTOOL_CONSOLE_SOCK') == os.path.join(app_path, run_dir, '%s.control' % stateboard_name)
    assert instance.getenv('TARANTOOL_WORKDIR') == os.path.join(app_path, data_dir, stateboard_name)

    if daemonized:
        assert os.path.exists(os.path.join(app_path, log_dir, '%s.log' % stateboard_name))

        notify_socket_path = os.path.join(app_path, run_dir, '%s.notify' % stateboard_name)
        assert(os.path.exists(notify_socket_path))
        assert instance.getenv('NOTIFY_SOCKET') == notify_socket_path


@tenacity.retry(stop=tenacity.stop_after_delay(15), wait=tenacity.wait_fixed(1))
def wait_instances(cli, project, instance_ids=[], run_dir=DEFAULT_RUN_DIR, stateboard=False, stateboard_only=False):
    exp_instances = instance_ids.copy()
    if stateboard or stateboard_only:
        exp_instances.append(get_stateboard_name(project.name))

    child_instances = cli.get_child_instances(project, run_dir)

    assert all([
        instance in child_instances
        for instance in exp_instances
    ])

    return child_instances


def check_instances_running(cli, project, instance_names=[],
                            stateboard=False, stateboard_only=False,
                            daemonized=False,
                            cfg=DEFAULT_CFG,
                            script=DEFAULT_SCRIPT,
                            run_dir=DEFAULT_RUN_DIR,
                            data_dir=DEFAULT_DATA_DIR,
                            log_dir=DEFAULT_LOG_DIR,
                            skip_env_checks=False):

    instance_ids = [get_instance_id(project.name, instance_name) for instance_name in instance_names]
    child_instances = wait_instances(cli, project, instance_ids, run_dir, stateboard, stateboard_only)

    # check that there is no extra instances running
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
        check_started_stateboard(child_instances, project.path, project.name, daemonized=daemonized,
                                 cfg=cfg, run_dir=run_dir, data_dir=data_dir, log_dir=log_dir,
                                 skip_env_checks=skip_env_checks)
    if not stateboard_only:
        for instance_id in instance_ids:
            check_running_instance(child_instances, project.path, project.name, instance_id, daemonized=daemonized,
                                   script=script, cfg=cfg, run_dir=run_dir, data_dir=data_dir, log_dir=log_dir,
                                   skip_env_checks=skip_env_checks)

    if not daemonized:
        assert cli.is_running()
    else:
        assert not cli.is_running()


@tenacity.retry(stop=tenacity.stop_after_delay(5), wait=tenacity.wait_fixed(1))
def check_instances_stopped(cli, project, instance_names=[], run_dir=DEFAULT_RUN_DIR,
                            stateboard=False, stateboard_only=False):
    instance_ids = [get_instance_id(project.name, instance_name) for instance_name in instance_names]
    child_instances = cli.get_child_instances(project, run_dir)

    if not stateboard_only:
        for instance_id in instance_ids:
            if instance_id in child_instances:
                instance = child_instances[instance_id]
                assert not instance.is_running()

    if stateboard:
        stateboard_name = get_stateboard_name(project.name)

        if stateboard_name in child_instances:
            instance = child_instances[stateboard_name]
            assert not instance.is_running()

    assert not cli.is_running()


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


@tenacity.retry(stop=tenacity.stop_after_delay(10), wait=tenacity.wait_fixed(1))
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


def get_replicasets(admin_api_url):
    query = '''
        query {
        replicasets: replicasets {
            alias
            roles
            vshard_group
            weight
            all_rw
            servers {
                alias
                uri
            }
        }
    }
    '''

    r = requests.post(admin_api_url, json={'query': query})
    assert r.status_code == 200
    resp = r.json()

    return resp['data']['replicasets']


def get_known_roles(admin_api_url):
    query = '''
        query {
        cluster {
            known_roles {
                name
                dependencies
            }
        }
    }
    '''

    r = requests.post(admin_api_url, json={'query': query})
    assert r.status_code == 200
    resp = r.json()

    known_roles = resp['data']['cluster']['known_roles']
    return known_roles


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


def is_vshard_bootstrapped(admin_api_url):
    query = '''
        query {
            cluster {
                vshard_groups {
                    bootstrapped
                }
            }
        }
    '''

    r = requests.post(admin_api_url, json={'query': query})
    assert r.status_code == 200
    resp = r.json()
    assert 'data' in resp

    vshard_groups = resp['data']['cluster']['vshard_groups']
    return all([g['bootstrapped'] for g in vshard_groups])


def get_vshard_group_names(admin_api_url):
    query = '''
        query {
            cluster {
                vshard_groups {
                    name
                }
            }
        }
    '''

    r = requests.post(admin_api_url, json={'query': query})
    assert r.status_code == 200
    resp = r.json()
    assert 'data' in resp

    vshard_groups = resp['data']['cluster']['vshard_groups']
    return [g['name'] for g in vshard_groups]


def is_instance_expelled(admin_api_url, instance_name):
    query = '''
        query {
            servers {
                alias
            }
        }
    '''

    r = requests.post(admin_api_url, json={'query': query})
    assert r.status_code == 200
    resp = r.json()
    assert 'data' in resp

    instance_names = [s['alias'] for s in resp['data']['servers']]
    return instance_name not in instance_names


@tenacity.retry(stop=tenacity.stop_after_delay(15), wait=tenacity.wait_fixed(1))
def wait_for_container_start(container, time_start):
    container_logs = container.logs(since=int(time_start)).decode('utf-8')
    assert 'entering the event loop' in container_logs


def examine_application_instance_container(instance_container):
    container = instance_container.container
    wait_for_container_start(container, time.time())

    container_logs = container.logs().decode('utf-8')
    m = re.search(r'Auto-detected IP to be "(\d+\.\d+\.\d+\.\d+)', container_logs)
    assert m is not None
    ip = m.groups()[0]

    admin_api_url = 'http://localhost:{}/admin/api'.format(instance_container.http_port)
    advertise_uri = '{}:{}'.format(ip, instance_container.advertise_port)
    roles = ["vshard-router", "app.roles.custom"]

    replicaset_uuid = create_replicaset(admin_api_url, [advertise_uri], roles)
    wait_for_replicaset_is_healthy(admin_api_url, replicaset_uuid)

    # restart instance
    container.restart()
    wait_for_container_start(container, time.time())

    # check instance restarted
    wait_for_replicaset_is_healthy(admin_api_url, replicaset_uuid)


def write_conf(path, conf):
    with open(path, 'w') as f:
        yaml.dump(conf, f, default_flow_style=False)


def run_command_on_container(container, command):
    command = '/bin/bash -c "{}"'.format(command.replace('"', '\\"'))
    rc, output = container.exec_run(command)
    assert rc == 0, output
    return output.decode("utf-8").strip()


def check_contains_dir(container, dirpath):
    command = '[ -d "{}" ] && echo true || echo false'.format(dirpath)
    return run_command_on_container(container, command)


def check_contains_file(container, filepath):
    command = '[ -f "{}" ] && echo true || echo false'.format(filepath)
    return run_command_on_container(container, command)


@tenacity.retry(stop=tenacity.stop_after_delay(10), wait=tenacity.wait_fixed(1))
def wait_for_systemd_service(container, service_name):
    show_logs_command = "journalctl --unit=%s -n 100 --no-pager" % service_name
    instance_logs = run_command_on_container(container, show_logs_command)
    assert 'entering the event loop' in instance_logs

    output = run_command_on_container(container, "systemctl status %s" % service_name)
    assert 'active (running)' in output


def check_systemd_service(container, project, http_port, tmpdir):
    instance_name = 'instance-1'
    advertise_uri = 'localhost:3303'

    instance_id = '{}.{}'.format(project.name, instance_name)

    conf_path = os.path.join(tmpdir, 'conf.yml')
    write_conf(conf_path, {
        instance_id: {
            'http_port': http_port,
            'advertise_uri': advertise_uri,
        }
    })

    archived_conf_path = os.path.join(tmpdir, 'conf.tar.gz')
    with tarfile.open(archived_conf_path, 'w:gz') as tar:
        tar.add(conf_path, arcname=os.path.basename(conf_path))

    CFG_DIR = '/etc/tarantool/conf.d'
    with gzip.open(archived_conf_path, 'r') as f:
        container.put_archive(CFG_DIR, f.read())

    service_name = '{}@{}'.format(project.name, instance_name)

    run_command_on_container(container, "systemctl start %s" % service_name)
    run_command_on_container(container, "systemctl enable %s" % service_name)

    wait_for_systemd_service(container, service_name)

    check_contains_dir(container, '/var/lib/tarantool/%s' % instance_id)
    check_contains_file(container, '/var/run/tarantool/%s.control' % instance_id)
    check_contains_file(container, '/var/run/tarantool/%s.pid' % instance_id)

    admin_api_url = 'http://localhost:%s/admin/api' % http_port
    roles = ["vshard-router", "app.roles.custom"]

    replicaset_uuid = create_replicaset(admin_api_url, [advertise_uri], roles)
    wait_for_replicaset_is_healthy(admin_api_url, replicaset_uuid)

    container.restart()
    wait_for_replicaset_is_healthy(admin_api_url, replicaset_uuid)


def build_image(path, tag):
    cli = APIClient(base_url=os.environ.get('DOCKER_HOST'))
    response = cli.build(path=path, forcerm=True, tag=tag)
    for r in response:
        for part in r.decode('utf-8').split('\r\n'):
            if part == '':
                continue
            part = json.loads(part)
            if 'stream' in part:
                print(part['stream'].replace('\n', ''))
            else:
                print(part)


# func(line) should be true for each line in logs
# each line should startswith `<appname>.<instance-name>...` for any instance
def assert_for_all_instances(logs, instances, func):
    assert all([
        any([
            line.strip().startswith("%s" % i) and func(line)
            for line in logs
        ])
        for i in instances
    ])


def assert_ok_for_all_instances(logs, instances):
    assert_for_all_instances(logs, instances, lambda line: line.strip().endswith('OK'))


# func(line) should be true for each line in logs
# any line should startswith `<instance1-name>, <instance2-name>:`
# where all group instances are mentioned
def assert_for_instances_group(logs, instances, func):
    instances.sort()
    group_title = ', '.join([i for i in instances])
    assert any([
        line.strip().startswith("%s" % group_title) and func(line)
        for line in logs
    ])


def assert_ok_for_instances_group(logs, group):
    assert_for_instances_group(logs, group, lambda line: line.strip().endswith('OK'))


def start_instances(cartridge_cmd, cli, project, cfg=None):
    if cfg is None:
        INSTANCE1 = 'instance-1'
        INSTANCE2 = 'instance-2'

        ID1 = get_instance_id(project.name, INSTANCE1)
        ID2 = get_instance_id(project.name, INSTANCE2)

        cfg = {
            ID1: {
                'advertise_uri': 'localhost:3301',
                'http_port': 8081,
            },
            ID2: {
                'advertise_uri': 'localhost:3302',
                'http_port': 8082,
            },
        }

    instance_names = [instance_id.split('.')[1] for instance_id in cfg]

    write_conf(os.path.join(project.path, DEFAULT_CFG), cfg)

    # start instance-1 and instance-2
    cli.start(project, daemonized=True)
    check_instances_running(cli, project, instance_names, daemonized=True)


def get_log_lines(output):
    stripped_lines = [
        line.strip() for line in output.split('\n')
    ]

    return [
        line for line in stripped_lines if line != ''
    ]
