import subprocess
import os
import yaml
import time
import sys
import psutil
import atexit
import shutil
import glob
import logfmt
import re

from utils import run_command_and_get_output


DEFAULT_RUN_DIR = 'tmp/run'
DEFAULT_DATA_DIR = 'tmp/data'
DEFAULT_CFG = 'instances.yml'

DEFAULT_SCRIPT = 'init.lua'
DEFAULT_STATEBOARD_SCRIPT = 'stateboard.init.lua'

STATUS_NOT_STARTED = 'NOT STARTED'
STATUS_RUNNING = 'RUNNING'
STATUS_STOPPED = 'STOPPED'
STATUS_FAILED = 'FAILED'


# #######
# Helpers
# #######
def write_conf(path, conf):
    with open(path, 'w') as f:
        yaml.dump(conf, f, default_flow_style=False)


def get_instance_id(app_name, instance_name):
    return '{}.{}'.format(app_name, instance_name)


def get_stateboard_name(app_name):
    return '{}-stateboard'.format(app_name)


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


class Cli():
    def __init__(self, cartridge_cmd):
        self._cartridge_cmd = cartridge_cmd
        self._children = []
        self._instances = dict()

    def start(self, project, instances=[], daemonized=False, stateboard=False, stateboard_only=False,
              cfg=None, script=None, run_dir=None, data_dir=None):

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
        if data_dir is not None:
            cmd.extend(['--data-dir', data_dir])

        cmd.extend(instances)

        self._subprocess = subprocess.Popen(
            cmd, cwd=project.path,
            stdout=sys.stdout,
            stderr=sys.stderr,
        )

        self._pid = self._subprocess.pid
        self._process = psutil.Process(self._pid)

        time.sleep(1)  # let cli to start instances

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

            msg = logfmt.parse_line(line)['msg']
            m = re.match(r'^(\S+):\s+(.+)$', msg)
            assert m is not None

            instance_id = m.group(1)
            instance_status = m.group(2)

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


def check_running_instance(child_instances, app_path, app_name, instance_name,
                           cfg=DEFAULT_CFG,
                           script=DEFAULT_SCRIPT,
                           run_dir=DEFAULT_RUN_DIR,
                           data_dir=DEFAULT_DATA_DIR):
    instance_id = get_instance_id(app_name, instance_name)

    assert instance_id in child_instances
    instance = child_instances[instance_id]

    assert instance.is_running()
    assert instance.cmd == ["tarantool", os.path.join(app_path, script)]

    assert instance.getenv('TARANTOOL_APP_NAME') == app_name
    assert instance.getenv('TARANTOOL_INSTANCE_NAME') == instance_name
    assert instance.getenv('TARANTOOL_CFG') == os.path.join(app_path, cfg)
    assert instance.getenv('TARANTOOL_PID_FILE') == os.path.join(app_path, run_dir, '%s.pid' % instance_id)
    assert instance.getenv('TARANTOOL_CONSOLE_SOCK') == os.path.join(app_path, run_dir, '%s.control' % instance_id)
    assert instance.getenv('TARANTOOL_WORKDIR') == os.path.join(app_path, data_dir, instance_id)


def check_started_stateboard(child_instances, app_path, app_name,
                             cfg=DEFAULT_CFG,
                             script=DEFAULT_STATEBOARD_SCRIPT,
                             run_dir=DEFAULT_RUN_DIR,
                             data_dir=DEFAULT_DATA_DIR):
    stateboard_name = get_stateboard_name(app_name)

    assert stateboard_name in child_instances
    instance = child_instances[stateboard_name]

    assert instance.is_running()
    assert instance.cmd == ["tarantool",  os.path.join(app_path, script)]

    assert instance.getenv('TARANTOOL_APP_NAME') == stateboard_name
    assert instance.getenv('TARANTOOL_CFG') == os.path.join(app_path, cfg)
    assert instance.getenv('TARANTOOL_PID_FILE') == os.path.join(app_path, run_dir, '%s.pid' % stateboard_name)
    assert instance.getenv('TARANTOOL_CONSOLE_SOCK') == os.path.join(app_path, run_dir, '%s.control' % stateboard_name)
    assert instance.getenv('TARANTOOL_WORKDIR') == os.path.join(app_path, data_dir, stateboard_name)


def check_instances_running(cli, project, instances,
                            stateboard=False, stateboard_only=False,
                            daemonized=False,
                            cfg=DEFAULT_CFG,
                            script=DEFAULT_SCRIPT,
                            run_dir=DEFAULT_RUN_DIR,
                            data_dir=DEFAULT_DATA_DIR):
    child_instances = cli.get_child_instances()

    running_instances_count = len([
        instance
        for _, instance in child_instances.items()
        if instance.is_running()
    ])

    if stateboard_only:
        assert running_instances_count == 1
    elif stateboard:
        assert running_instances_count == len(instances) + 1
    else:
        assert running_instances_count == len(instances)

    if stateboard:
        check_started_stateboard(child_instances, project.path, project.name,
                                 cfg=cfg, run_dir=run_dir, data_dir=data_dir)
    if not stateboard_only:
        for instance in instances:
            check_running_instance(child_instances, project.path, project.name, instance,
                                   script=script, cfg=cfg, run_dir=run_dir, data_dir=data_dir)

    if not daemonized:
        assert cli.is_running()
    else:
        assert not cli.is_running()


def check_instances_stopped(cli, project, instances, run_dir=DEFAULT_RUN_DIR,
                            stateboard=False, stateboard_only=False):
    child_instances = cli.get_child_instances()

    if not stateboard_only:
        for instance_name in instances:
            instance_id = get_instance_id(project.name, instance_name)

            assert instance_id in child_instances
            instance = child_instances[instance_id]

            assert not instance.is_running()

    if stateboard:
        stateboard_name = get_stateboard_name(project.name)

        assert stateboard_name in child_instances
        instance = child_instances[stateboard_name]

        assert not instance.is_running()

    assert not cli.is_running()


# #####
# Tests
# #####
def test_start_interactive_by_id(cartridge_cmd, project_with_patched_init):
    project = project_with_patched_init
    cli = Cli(cartridge_cmd)

    INSTANCE1_NAME = 'instance-1'
    INSTANCE2_NAME = 'instance-2'

    # start instance-1 and instance-2
    cli.start(project, [INSTANCE1_NAME, INSTANCE2_NAME])
    check_instances_running(cli, project, [INSTANCE1_NAME, INSTANCE2_NAME])


def test_start_stop_by_id(cartridge_cmd, project_with_patched_init):
    project = project_with_patched_init
    cli = Cli(cartridge_cmd)

    INSTANCE1_NAME = 'instance-1'
    INSTANCE2_NAME = 'instance-2'

    # start instance-1 and instance-2
    cli.start(project, [INSTANCE1_NAME, INSTANCE2_NAME], daemonized=True)
    check_instances_running(cli, project, [INSTANCE1_NAME, INSTANCE2_NAME], daemonized=True)

    # stop instance-1
    cli.stop(project, [INSTANCE1_NAME])
    check_instances_running(cli, project, [INSTANCE2_NAME], daemonized=True)
    check_instances_stopped(cli, project, [INSTANCE1_NAME])


def test_start_interactive_by_id_with_stateboard(cartridge_cmd, project_with_patched_init):
    project = project_with_patched_init
    cli = Cli(cartridge_cmd)

    INSTANCE1_NAME = 'instance-1'
    INSTANCE2_NAME = 'instance-2'

    # start instance-1 and instance-2
    cli.start(project, [INSTANCE1_NAME, INSTANCE2_NAME], stateboard=True)
    check_instances_running(cli, project, [INSTANCE1_NAME, INSTANCE2_NAME], stateboard=True)


def test_start_stop_by_id_with_stateboard(cartridge_cmd, project_with_patched_init):
    project = project_with_patched_init
    cli = Cli(cartridge_cmd)

    INSTANCE1_NAME = 'instance-1'
    INSTANCE2_NAME = 'instance-2'

    # start instance-1 and instance-2
    cli.start(project, [INSTANCE1_NAME, INSTANCE2_NAME], daemonized=True, stateboard=True)
    check_instances_running(cli, project, [INSTANCE1_NAME, INSTANCE2_NAME], daemonized=True, stateboard=True)

    # stop instance-1 and stateboard
    cli.stop(project, [INSTANCE1_NAME], stateboard=True)
    check_instances_running(cli, project, [INSTANCE2_NAME], daemonized=True)
    check_instances_stopped(cli, project, [INSTANCE1_NAME], stateboard=True)


def test_start_interactive_from_conf(cartridge_cmd, project_with_patched_init):
    project = project_with_patched_init
    cli = Cli(cartridge_cmd)

    INSTANCE1_NAME = 'instance-1'
    INSTANCE2_NAME = 'instance-2'

    write_conf(os.path.join(project.path, DEFAULT_CFG), {
        get_instance_id(project.name, INSTANCE1_NAME): {},
        get_instance_id(project.name, INSTANCE2_NAME): {},
    })

    # start instances
    cli.start(project)
    check_instances_running(cli, project, [INSTANCE1_NAME, INSTANCE2_NAME])


def test_start_stop_from_conf(cartridge_cmd, project_with_patched_init):
    project = project_with_patched_init
    cli = Cli(cartridge_cmd)

    INSTANCE1_NAME = 'instance-1'
    INSTANCE2_NAME = 'instance-2'

    write_conf(os.path.join(project.path, DEFAULT_CFG), {
        get_instance_id(project.name, INSTANCE1_NAME): {},
        get_instance_id(project.name, INSTANCE2_NAME): {},
    })

    # start instances
    cli.start(project, daemonized=True)
    check_instances_running(cli, project, [INSTANCE1_NAME, INSTANCE2_NAME], daemonized=True)

    # stop instances
    cli.stop(project)
    check_instances_stopped(cli, project, [INSTANCE1_NAME, INSTANCE2_NAME])


def test_start_interactive_from_conf_with_stateboard(cartridge_cmd, project_with_patched_init):
    project = project_with_patched_init
    cli = Cli(cartridge_cmd)

    INSTANCE1_NAME = 'instance-1'
    INSTANCE2_NAME = 'instance-2'

    write_conf(os.path.join(project.path, DEFAULT_CFG), {
        get_instance_id(project.name, INSTANCE1_NAME): {},
        get_instance_id(project.name, INSTANCE2_NAME): {},
    })

    # start instances
    cli.start(project, stateboard=True)
    check_instances_running(cli, project, [INSTANCE1_NAME, INSTANCE2_NAME], stateboard=True)


def test_start_stop_from_conf_with_stateboard(cartridge_cmd, project_with_patched_init):
    project = project_with_patched_init
    cli = Cli(cartridge_cmd)

    INSTANCE1_NAME = 'instance-1'
    INSTANCE2_NAME = 'instance-2'

    write_conf(os.path.join(project.path, DEFAULT_CFG), {
        get_instance_id(project.name, INSTANCE1_NAME): {},
        get_instance_id(project.name, INSTANCE2_NAME): {},
    })

    # start instances
    cli.start(project, daemonized=True, stateboard=True)
    check_instances_running(cli, project, [INSTANCE1_NAME, INSTANCE2_NAME], daemonized=True, stateboard=True)

    # stop instances
    cli.stop(project, stateboard=True)
    check_instances_stopped(cli, project, [INSTANCE1_NAME, INSTANCE2_NAME], stateboard=True)


def test_status_by_id(cartridge_cmd, project_with_patched_init):
    project = project_with_patched_init
    cli = Cli(cartridge_cmd)

    INSTANCE1_NAME = 'instance-1'
    INSTANCE2_NAME = 'instance-2'

    INSTANCE1_ID = get_instance_id(project.name, INSTANCE1_NAME)
    INSTANCE2_ID = get_instance_id(project.name, INSTANCE2_NAME)
    STATEBOARD_ID = get_stateboard_name(project.name)

    # get status w/o stateboard
    status = cli.get_status(project, [INSTANCE1_NAME, INSTANCE2_NAME])
    assert status.get(INSTANCE1_ID) == STATUS_NOT_STARTED
    assert status.get(INSTANCE2_ID) == STATUS_NOT_STARTED

    # get status w/ stateboard
    status = cli.get_status(project, [INSTANCE1_NAME, INSTANCE2_NAME], stateboard=True)
    assert status.get(INSTANCE1_ID) == STATUS_NOT_STARTED
    assert status.get(INSTANCE2_ID) == STATUS_NOT_STARTED
    assert status.get(STATEBOARD_ID) == STATUS_NOT_STARTED

    # start instance-1 and stateboard
    cli.start(project, [INSTANCE1_NAME], stateboard=True, daemonized=True)

    # get status w/o stateboard
    status = cli.get_status(project, [INSTANCE1_NAME, INSTANCE2_NAME])
    assert status.get(INSTANCE1_ID) == STATUS_RUNNING
    assert status.get(INSTANCE2_ID) == STATUS_NOT_STARTED

    # get status w/ stateboard
    status = cli.get_status(project, [INSTANCE1_NAME, INSTANCE2_NAME], stateboard=True)
    assert status.get(INSTANCE1_ID) == STATUS_RUNNING
    assert status.get(INSTANCE2_ID) == STATUS_NOT_STARTED
    assert status.get(STATEBOARD_ID) == STATUS_RUNNING

    # stop instance-1
    cli.stop(project, [INSTANCE1_NAME])

    # get status w/o stateboard
    status = cli.get_status(project, [INSTANCE1_NAME, INSTANCE2_NAME])
    assert status.get(INSTANCE1_ID) == STATUS_STOPPED
    assert status.get(INSTANCE2_ID) == STATUS_NOT_STARTED

    # get status w/ stateboard
    status = cli.get_status(project, [INSTANCE1_NAME, INSTANCE2_NAME], stateboard=True)
    assert status.get(INSTANCE1_ID) == STATUS_STOPPED
    assert status.get(INSTANCE2_ID) == STATUS_NOT_STARTED
    assert status.get(STATEBOARD_ID) == STATUS_RUNNING


def test_status_from_conf(cartridge_cmd, project_with_patched_init):
    project = project_with_patched_init
    cli = Cli(cartridge_cmd)

    INSTANCE1_NAME = 'instance-1'
    INSTANCE2_NAME = 'instance-2'

    INSTANCE1_ID = get_instance_id(project.name, INSTANCE1_NAME)
    INSTANCE2_ID = get_instance_id(project.name, INSTANCE2_NAME)
    STATEBOARD_ID = get_stateboard_name(project.name)

    write_conf(os.path.join(project.path, DEFAULT_CFG), {
        INSTANCE1_ID: {},
        INSTANCE2_ID: {},
    })

    # get status w/o stateboard
    status = cli.get_status(project)
    assert status.get(INSTANCE1_ID) == STATUS_NOT_STARTED
    assert status.get(INSTANCE2_ID) == STATUS_NOT_STARTED

    # get status w/ stateboard
    status = cli.get_status(project, stateboard=True)
    assert status.get(INSTANCE1_ID) == STATUS_NOT_STARTED
    assert status.get(INSTANCE2_ID) == STATUS_NOT_STARTED
    assert status.get(STATEBOARD_ID) == STATUS_NOT_STARTED

    # start instance-1 and stateboard
    cli.start(project, [INSTANCE1_NAME], stateboard=True, daemonized=True)

    # get status w/o stateboard
    status = cli.get_status(project)
    assert status.get(INSTANCE1_ID) == STATUS_RUNNING
    assert status.get(INSTANCE2_ID) == STATUS_NOT_STARTED

    # get status w/ stateboard
    status = cli.get_status(project, stateboard=True)
    assert status.get(INSTANCE1_ID) == STATUS_RUNNING
    assert status.get(INSTANCE2_ID) == STATUS_NOT_STARTED
    assert status.get(STATEBOARD_ID) == STATUS_RUNNING

    # stop instance-1
    cli.stop(project, [INSTANCE1_NAME])

    # get status w/o stateboard
    status = cli.get_status(project)
    assert status.get(INSTANCE1_ID) == STATUS_STOPPED
    assert status.get(INSTANCE2_ID) == STATUS_NOT_STARTED

    # get status w/ stateboard
    status = cli.get_status(project, stateboard=True)
    assert status.get(INSTANCE1_ID) == STATUS_STOPPED
    assert status.get(INSTANCE2_ID) == STATUS_NOT_STARTED
    assert status.get(STATEBOARD_ID) == STATUS_RUNNING


def test_start_cfg(cartridge_cmd, project_with_patched_init):
    project = project_with_patched_init
    cli = Cli(cartridge_cmd)

    INSTANCE1_NAME = 'instance-1'
    INSTANCE2_NAME = 'instance-2'
    CFG = 'my-conf.yml'

    write_conf(os.path.join(project.path, CFG), {
        get_instance_id(project.name, INSTANCE1_NAME): {},
        get_instance_id(project.name, INSTANCE2_NAME): {},
    })

    cli.start(project, stateboard=True, cfg=CFG)
    check_instances_running(
        cli, project,
        [INSTANCE1_NAME, INSTANCE2_NAME],
        stateboard=True, cfg=CFG
    )


def test_start_stop_status_cfg(cartridge_cmd, project_with_patched_init):
    project = project_with_patched_init
    cli = Cli(cartridge_cmd)

    INSTANCE1_NAME = 'instance-1'
    INSTANCE2_NAME = 'instance-2'
    CFG = 'my-conf.yml'

    INSTANCE1_ID = get_instance_id(project.name, INSTANCE1_NAME)
    INSTANCE2_ID = get_instance_id(project.name, INSTANCE2_NAME)

    write_conf(os.path.join(project.path, CFG), {
        INSTANCE1_ID: {},
        INSTANCE2_ID: {},
    })

    status = cli.get_status(project, cfg=CFG)
    assert status.get(INSTANCE1_ID) == STATUS_NOT_STARTED
    assert status.get(INSTANCE2_ID) == STATUS_NOT_STARTED

    cli.start(project, stateboard=True, daemonized=True, cfg=CFG)
    check_instances_running(
        cli, project,
        [INSTANCE1_NAME, INSTANCE2_NAME],
        stateboard=True, cfg=CFG,
        daemonized=True,
    )

    status = cli.get_status(project, cfg=CFG)
    assert status.get(INSTANCE1_ID) == STATUS_RUNNING
    assert status.get(INSTANCE2_ID) == STATUS_RUNNING

    cli.stop(project, stateboard=True, cfg=CFG)
    check_instances_stopped(cli, project, [INSTANCE1_NAME, INSTANCE2_NAME])

    status = cli.get_status(project, cfg=CFG)
    assert status.get(INSTANCE1_ID) == STATUS_STOPPED
    assert status.get(INSTANCE2_ID) == STATUS_STOPPED


def test_start_run_dir(cartridge_cmd, project_with_patched_init):
    project = project_with_patched_init
    cli = Cli(cartridge_cmd)

    INSTANCE1_NAME = 'instance-1'
    INSTANCE2_NAME = 'instance-2'
    RUN_DIR = 'my-run'

    cli.start(project, [INSTANCE1_NAME, INSTANCE2_NAME], stateboard=True, run_dir=RUN_DIR)
    check_instances_running(
        cli, project,
        [INSTANCE1_NAME, INSTANCE2_NAME],
        stateboard=True, run_dir=RUN_DIR
    )


def test_start_stop_status_run_dir(cartridge_cmd, project_with_patched_init):
    project = project_with_patched_init
    cli = Cli(cartridge_cmd)

    INSTANCE1_NAME = 'instance-1'
    INSTANCE2_NAME = 'instance-2'
    RUN_DIR = 'my-run'

    INSTANCE1_ID = get_instance_id(project.name, INSTANCE1_NAME)
    INSTANCE2_ID = get_instance_id(project.name, INSTANCE2_NAME)

    status = cli.get_status(project, [INSTANCE1_NAME, INSTANCE2_NAME], run_dir=RUN_DIR)
    assert status.get(INSTANCE1_ID) == STATUS_NOT_STARTED
    assert status.get(INSTANCE2_ID) == STATUS_NOT_STARTED

    cli.start(project, [INSTANCE1_NAME, INSTANCE2_NAME], stateboard=True, daemonized=True, run_dir=RUN_DIR)
    check_instances_running(
        cli, project,
        [INSTANCE1_NAME, INSTANCE2_NAME],
        stateboard=True, run_dir=RUN_DIR,
        daemonized=True
    )

    status = cli.get_status(project, [INSTANCE1_NAME, INSTANCE2_NAME], run_dir=RUN_DIR)
    assert status.get(INSTANCE1_ID) == STATUS_RUNNING
    assert status.get(INSTANCE2_ID) == STATUS_RUNNING

    cli.stop(project, [INSTANCE1_NAME, INSTANCE2_NAME], stateboard=True, run_dir=RUN_DIR)
    check_instances_stopped(cli, project, [INSTANCE1_NAME, INSTANCE2_NAME], run_dir=RUN_DIR)

    status = cli.get_status(project, [INSTANCE1_NAME, INSTANCE2_NAME], run_dir=RUN_DIR)
    assert status.get(INSTANCE2_ID) == STATUS_STOPPED
    assert status.get(INSTANCE1_ID) == STATUS_STOPPED


def test_start_data_dir(cartridge_cmd, project_with_patched_init):
    project = project_with_patched_init
    cli = Cli(cartridge_cmd)

    INSTANCE1_NAME = 'instance-1'
    INSTANCE2_NAME = 'instance-2'
    DATA_DIR = 'my-data'

    cli.start(project, [INSTANCE1_NAME, INSTANCE2_NAME], stateboard=True, data_dir=DATA_DIR)
    check_instances_running(
        cli, project,
        [INSTANCE1_NAME, INSTANCE2_NAME],
        stateboard=True, data_dir=DATA_DIR
    )


def test_start_script(cartridge_cmd, project_with_patched_init):
    project = project_with_patched_init
    cli = Cli(cartridge_cmd)

    INSTANCE1_NAME = 'instance-1'
    INSTANCE2_NAME = 'instance-2'
    SCRIPT = 'my-init.lua'
    shutil.copyfile(os.path.join(project.path, DEFAULT_SCRIPT), os.path.join(project.path, SCRIPT))

    cli.start(project, [INSTANCE1_NAME, INSTANCE2_NAME], stateboard=True, script=SCRIPT)
    check_instances_running(
        cli, project,
        [INSTANCE1_NAME, INSTANCE2_NAME],
        stateboard=True, script=SCRIPT
    )
