import subprocess
import os
import yaml
import time
import sys
import psutil
import signal
import atexit
import shutil


DEFAULT_RUN_DIR = 'tmp/run'
DEFAULT_DATA_DIR = 'tmp/data'
DEFAULT_CFG = 'instances.yml'

DEFAULT_SCRIPT = 'init.lua'
DEFAULT_STATEBOARD_SCRIPT = 'stateboard.init.lua'


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
        return self._process.status() == psutil.STATUS_RUNNING

    def getenv(self, name):
        return self._env.get(name)


class CliProcess():
    def __init__(self, cmd, cwd=None):
        self.cmd = cmd
        self._subprocess = subprocess.Popen(
            cmd, cwd=cwd,
            stdout=sys.stdout,
            stderr=sys.stderr,
        )

        self._pid = self._subprocess.pid
        self._process = psutil.Process(self._pid)

        time.sleep(1)  # let cli to start child processes

        self._children = [
            psutil.Process(child.pid)
            for child in self._process.children()
        ]

        atexit.register(self.stop)

    def get_started_instances(self):
        instances = {}

        for child in self._children:
            instance = InstanceProcess(child)

            assert instance.id not in instances
            instances[instance.id] = instance

        return instances

    def is_running(self):
        return self._process.status() == psutil.STATUS_RUNNING

    def terminate(self):
        os.kill(self._pid, signal.SIGTERM)

    def stop(self):
        self._subprocess.terminate()
        for child in self._children:
            child.terminate()


def check_started_instance(started_instances, app_path, app_name, instance_name,
                           cfg=DEFAULT_CFG,
                           script=DEFAULT_SCRIPT,
                           run_dir=DEFAULT_RUN_DIR,
                           data_dir=DEFAULT_DATA_DIR):
    instance_id = get_instance_id(app_name, instance_name)

    assert instance_id in started_instances
    instance = started_instances[instance_id]

    assert instance.is_running()
    assert instance.cmd == ["tarantool", os.path.join(app_path, script)]

    assert instance.getenv('TARANTOOL_APP_NAME') == app_name
    assert instance.getenv('TARANTOOL_INSTANCE_NAME') == instance_name
    assert instance.getenv('TARANTOOL_CFG') == os.path.join(app_path, cfg)
    assert instance.getenv('TARANTOOL_PID_FILE') == os.path.join(app_path, run_dir, '%s.pid' % instance_id)
    assert instance.getenv('TARANTOOL_CONSOLE_SOCK') == os.path.join(app_path, run_dir, '%s.control' % instance_id)
    assert instance.getenv('TARANTOOL_WORKDIR') == os.path.join(app_path, data_dir, instance_id)


def check_started_stateboard(started_instances, app_path, app_name,
                             cfg=DEFAULT_CFG,
                             script=DEFAULT_STATEBOARD_SCRIPT,
                             run_dir=DEFAULT_RUN_DIR,
                             data_dir=DEFAULT_DATA_DIR):
    stateboard_name = get_stateboard_name(app_name)

    assert stateboard_name in started_instances
    instance = started_instances[stateboard_name]

    assert instance.is_running()
    assert instance.cmd == ["tarantool",  os.path.join(app_path, script)]

    assert instance.getenv('TARANTOOL_APP_NAME') == stateboard_name
    assert instance.getenv('TARANTOOL_CFG') == os.path.join(app_path, cfg)
    assert instance.getenv('TARANTOOL_PID_FILE') == os.path.join(app_path, run_dir, '%s.pid' % stateboard_name)
    assert instance.getenv('TARANTOOL_CONSOLE_SOCK') == os.path.join(app_path, run_dir, '%s.control' % stateboard_name)
    assert instance.getenv('TARANTOOL_WORKDIR') == os.path.join(app_path, data_dir, stateboard_name)


def check_instances_started(cmd, project, instances,
                            stateboard=False, stateboard_only=False,
                            daemonized=False,
                            cfg=DEFAULT_CFG,
                            script=DEFAULT_SCRIPT,
                            run_dir=DEFAULT_RUN_DIR,
                            data_dir=DEFAULT_DATA_DIR):
    cli = CliProcess(cmd, project.path)
    started_instances = cli.get_started_instances()

    if stateboard_only:
        assert len(started_instances) == 1
    elif stateboard:
        assert len(started_instances) == len(instances) + 1
    else:
        assert len(started_instances) == len(instances)

    if stateboard:
        check_started_stateboard(started_instances, project.path, project.name,
                                 cfg=cfg, run_dir=run_dir, data_dir=data_dir)
    if not stateboard_only:
        for instance in instances:
            check_started_instance(started_instances, project.path, project.name, instance,
                                   script=script, cfg=cfg, run_dir=run_dir, data_dir=data_dir)

    if not daemonized:
        assert cli.is_running()
    else:
        assert not cli.is_running()


# #####
# Tests
# #####
def test_start_interactive_by_id(cartridge_cmd, project_with_patched_init):
    project = project_with_patched_init

    INSTANCE1_NAME = 'instance-1'
    INSTANCE2_NAME = 'instance-2'

    # start instance-1 and instance-2
    cmd = [cartridge_cmd, 'start', INSTANCE1_NAME, INSTANCE2_NAME]
    check_instances_started(cmd, project, [INSTANCE1_NAME, INSTANCE2_NAME])


def test_start_interactive_by_id_with_stateboard(cartridge_cmd, project_with_patched_init):
    project = project_with_patched_init

    INSTANCE1_NAME = 'instance-1'
    INSTANCE2_NAME = 'instance-2'

    # start instance-1 and instance-2
    cmd = [cartridge_cmd, 'start', INSTANCE1_NAME, INSTANCE2_NAME, "--stateboard"]
    check_instances_started(cmd, project, [INSTANCE1_NAME, INSTANCE2_NAME], stateboard=True)


def test_start_interactive_from_conf(cartridge_cmd, project_with_patched_init):
    project = project_with_patched_init

    INSTANCE1_NAME = 'instance-1'
    INSTANCE2_NAME = 'instance-2'

    write_conf(os.path.join(project.path, DEFAULT_CFG), {
        get_instance_id(project.name, INSTANCE1_NAME): {},
        get_instance_id(project.name, INSTANCE2_NAME): {},
    })

    # start instances
    cmd = [cartridge_cmd, 'start']
    check_instances_started(cmd, project, [INSTANCE1_NAME, INSTANCE2_NAME])


def test_start_interactive_from_conf_with_stateboard(cartridge_cmd, project_with_patched_init):
    project = project_with_patched_init

    INSTANCE1_NAME = 'instance-1'
    INSTANCE2_NAME = 'instance-2'

    write_conf(os.path.join(project.path, DEFAULT_CFG), {
        get_instance_id(project.name, INSTANCE1_NAME): {},
        get_instance_id(project.name, INSTANCE2_NAME): {},
    })

    # start instances
    cmd = [cartridge_cmd, 'start', '--stateboard']
    check_instances_started(cmd, project, [INSTANCE1_NAME, INSTANCE2_NAME], stateboard=True)


def test_start_cfg(cartridge_cmd, project_with_patched_init):
    project = project_with_patched_init

    INSTANCE1_NAME = 'instance-1'
    INSTANCE2_NAME = 'instance-2'
    CFG = 'my-conf.yml'

    cmd = [
        cartridge_cmd, 'start', '--stateboard',
        '--cfg', CFG,
        INSTANCE1_NAME, INSTANCE2_NAME
    ]

    check_instances_started(
        cmd, project,
        [INSTANCE1_NAME, INSTANCE2_NAME],
        stateboard=True, cfg=CFG
    )


def test_start_run_dir(cartridge_cmd, project_with_patched_init):
    project = project_with_patched_init

    INSTANCE1_NAME = 'instance-1'
    INSTANCE2_NAME = 'instance-2'
    RUN_DIR = 'my-run'

    cmd = [
        cartridge_cmd, 'start',
        '--stateboard',
        '--run-dir', RUN_DIR,
        INSTANCE1_NAME, INSTANCE2_NAME
    ]

    check_instances_started(
        cmd, project,
        [INSTANCE1_NAME, INSTANCE2_NAME],
        stateboard=True, run_dir=RUN_DIR
    )


def test_start_data_dir(cartridge_cmd, project_with_patched_init):
    project = project_with_patched_init

    INSTANCE1_NAME = 'instance-1'
    INSTANCE2_NAME = 'instance-2'
    DATA_DIR = 'my-data'

    cmd = [
        cartridge_cmd, 'start',
        '--stateboard',
        '--data-dir', DATA_DIR,
        INSTANCE1_NAME, INSTANCE2_NAME
    ]

    check_instances_started(
        cmd, project,
        [INSTANCE1_NAME, INSTANCE2_NAME],
        stateboard=True, data_dir=DATA_DIR
    )


def test_start_script(cartridge_cmd, project_with_patched_init):
    project = project_with_patched_init

    INSTANCE1_NAME = 'instance-1'
    INSTANCE2_NAME = 'instance-2'
    SCRIPT = 'my-init.lua'
    shutil.copyfile(os.path.join(project.path, DEFAULT_SCRIPT), os.path.join(project.path, SCRIPT))

    cmd = [
        cartridge_cmd, 'start',
        '--stateboard',
        '--script', SCRIPT,
        INSTANCE1_NAME, INSTANCE2_NAME
    ]

    check_instances_started(
        cmd, project,
        [INSTANCE1_NAME, INSTANCE2_NAME],
        stateboard=True, script=SCRIPT
    )
