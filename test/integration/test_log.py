import os

from utils import get_instance_id, get_stateboard_name
from utils import check_instances_running
from utils import DEFAULT_CFG

from project import patch_init_to_log_lines
from utils import write_conf


DEFAULT_LAST_N_LINES = 15  # see cli/commands/log.go
LOG_LINES = ["I am log line number {}".format(i) for i in range(DEFAULT_LAST_N_LINES-1)]


def assert_instance_logs(logs, instance_id, expected_lines):
    assert instance_id in logs

    line_fmt = "{instance_id}: {log_line}"  # see patch_init_to_log_lines
    expected_lines_formatted = [
        line_fmt.format(instance_id=instance_id, log_line=line)
        for line in expected_lines
    ]

    assert logs[instance_id] == expected_lines_formatted


def test_log_by_name(start_stop_cli, project_with_patched_init):
    project = project_with_patched_init
    cli = start_stop_cli

    patch_init_to_log_lines(project, LOG_LINES)

    INSTANCE1 = 'instance-1'
    INSTANCE2 = 'instance-2'

    ID1 = get_instance_id(project.name, INSTANCE1)
    ID2 = get_instance_id(project.name, INSTANCE2)
    STATEBOARD_ID = get_stateboard_name(project.name)

    # start instance-1 and instance-2
    cli.start(project, [INSTANCE1, INSTANCE2], stateboard=True, daemonized=True)
    check_instances_running(cli, project, [INSTANCE1, INSTANCE2], stateboard=True, daemonized=True)

    # get logs w/o stateboard
    logs = cli.get_logs(project, [INSTANCE1, INSTANCE2])
    assert len(logs) == 2
    assert_instance_logs(logs, ID1, LOG_LINES)
    assert_instance_logs(logs, ID2, LOG_LINES)

    # get logs w/ stateboard
    logs = cli.get_logs(project, [INSTANCE1, INSTANCE2], stateboard=True)
    assert len(logs) == 3
    assert_instance_logs(logs, ID1, LOG_LINES)
    assert_instance_logs(logs, ID2, LOG_LINES)
    assert_instance_logs(logs, STATEBOARD_ID, LOG_LINES)

    # get logs stateboard-only
    logs = cli.get_logs(project, stateboard_only=True)
    assert len(logs) == 1
    assert_instance_logs(logs, STATEBOARD_ID, LOG_LINES)


def test_log_from_conf(start_stop_cli, project_with_patched_init):
    project = project_with_patched_init
    cli = start_stop_cli

    patch_init_to_log_lines(project, LOG_LINES)

    INSTANCE1 = 'instance-1'
    INSTANCE2 = 'instance-2'

    ID1 = get_instance_id(project.name, INSTANCE1)
    ID2 = get_instance_id(project.name, INSTANCE2)
    STATEBOARD_ID = get_stateboard_name(project.name)

    write_conf(os.path.join(project.path, DEFAULT_CFG), {
        ID1: {},
        ID2: {},
    })

    # start instance-1 and instance-2
    cli.start(project, [INSTANCE1, INSTANCE2], stateboard=True, daemonized=True)
    check_instances_running(cli, project, [INSTANCE1, INSTANCE2], stateboard=True, daemonized=True)

    # get logs w/o stateboard
    logs = cli.get_logs(project)
    assert len(logs) == 2
    assert_instance_logs(logs, ID1, LOG_LINES)
    assert_instance_logs(logs, ID2, LOG_LINES)

    # get logs w/ stateboard
    logs = cli.get_logs(project, stateboard=True)
    assert len(logs) == 3
    assert_instance_logs(logs, ID1, LOG_LINES)
    assert_instance_logs(logs, ID2, LOG_LINES)
    assert_instance_logs(logs, STATEBOARD_ID, LOG_LINES)

    # get logs stateboard-only
    logs = cli.get_logs(project, stateboard_only=True)
    assert len(logs) == 1
    assert_instance_logs(logs, STATEBOARD_ID, LOG_LINES)


def test_log_cfg(start_stop_cli, project_with_patched_init):
    project = project_with_patched_init
    cli = start_stop_cli

    patch_init_to_log_lines(project, LOG_LINES)

    INSTANCE1 = 'instance-1'
    INSTANCE2 = 'instance-2'

    ID1 = get_instance_id(project.name, INSTANCE1)
    ID2 = get_instance_id(project.name, INSTANCE2)

    CFG = 'my-conf.yml'

    write_conf(os.path.join(project.path, CFG), {
        ID1: {},
        ID2: {},
    })

    # start instance-1 and instance-2
    cli.start(project, cfg=CFG, daemonized=True)
    check_instances_running(cli, project, [INSTANCE1, INSTANCE2], cfg=CFG, daemonized=True)

    # get logs
    logs = cli.get_logs(project,  cfg=CFG)
    assert len(logs) == 2
    assert_instance_logs(logs, ID1, LOG_LINES)
    assert_instance_logs(logs, ID2, LOG_LINES)


def test_log_log_dir(start_stop_cli, project_with_patched_init):
    project = project_with_patched_init
    cli = start_stop_cli

    patch_init_to_log_lines(project, LOG_LINES)

    INSTANCE1 = 'instance-1'
    INSTANCE2 = 'instance-2'

    ID1 = get_instance_id(project.name, INSTANCE1)
    ID2 = get_instance_id(project.name, INSTANCE2)

    LOG_DIR = 'my-log-dir'

    write_conf(os.path.join(project.path, DEFAULT_CFG), {
        ID1: {},
        ID2: {},
    })

    # start instance-1 and instance-2
    cli.start(project, log_dir=LOG_DIR, daemonized=True)
    check_instances_running(cli, project, [INSTANCE1, INSTANCE2], log_dir=LOG_DIR, daemonized=True)

    # get logs
    logs = cli.get_logs(project,  log_dir=LOG_DIR)
    assert len(logs) == 2
    assert_instance_logs(logs, ID1, LOG_LINES)
    assert_instance_logs(logs, ID2, LOG_LINES)


def test_log_run_dir(start_stop_cli, project_with_patched_init):
    project = project_with_patched_init
    cli = start_stop_cli

    patch_init_to_log_lines(project, LOG_LINES)

    INSTANCE1 = 'instance-1'
    INSTANCE2 = 'instance-2'

    ID1 = get_instance_id(project.name, INSTANCE1)
    ID2 = get_instance_id(project.name, INSTANCE2)

    RUN_DIR = 'my-log-dir'

    write_conf(os.path.join(project.path, DEFAULT_CFG), {
        ID1: {},
        ID2: {},
    })

    # start instance-1 and instance-2
    cli.start(project, run_dir=RUN_DIR, daemonized=True)
    check_instances_running(cli, project, [INSTANCE1, INSTANCE2], run_dir=RUN_DIR, daemonized=True)

    # get logs
    logs = cli.get_logs(project, run_dir=RUN_DIR)
    assert len(logs) == 2
    assert_instance_logs(logs, ID1, LOG_LINES)
    assert_instance_logs(logs, ID2, LOG_LINES)


def test_log_last_n_lines(start_stop_cli, project_with_patched_init):
    project = project_with_patched_init
    cli = start_stop_cli

    log_lines = ["I am log line number {}".format(i) for i in range(DEFAULT_LAST_N_LINES+5)]

    patch_init_to_log_lines(project, log_lines)

    INSTANCE1 = 'instance-1'
    INSTANCE2 = 'instance-2'

    ID1 = get_instance_id(project.name, INSTANCE1)
    ID2 = get_instance_id(project.name, INSTANCE2)
    STATEBOARD_ID = get_stateboard_name(project.name)

    # start instance-1 and instance-2
    cli.start(project, [INSTANCE1, INSTANCE2], stateboard=True, daemonized=True)
    check_instances_running(cli, project, [INSTANCE1, INSTANCE2], stateboard=True, daemonized=True)

    # get logs w/o -n
    logs = cli.get_logs(project, [INSTANCE1, INSTANCE2], stateboard=True)
    assert len(logs) == 3
    assert_instance_logs(logs, ID1, log_lines[-DEFAULT_LAST_N_LINES:])
    assert_instance_logs(logs, ID2, log_lines[-DEFAULT_LAST_N_LINES:])
    assert_instance_logs(logs, STATEBOARD_ID, log_lines[-DEFAULT_LAST_N_LINES:])

    # get logs w/ -n > log lines count
    logs = cli.get_logs(project, [INSTANCE1, INSTANCE2], stateboard=True, n=len(log_lines)*2)
    assert len(logs) == 3
    assert_instance_logs(logs, ID1, log_lines)
    assert_instance_logs(logs, ID2, log_lines)
    assert_instance_logs(logs, STATEBOARD_ID, log_lines)

    # get logs w/ -n0
    logs = cli.get_logs(project, [INSTANCE1, INSTANCE2], stateboard=True, n=0)
    assert len(logs) == 3
    assert_instance_logs(logs, ID1, log_lines)
    assert_instance_logs(logs, ID2, log_lines)
    assert_instance_logs(logs, STATEBOARD_ID, log_lines)

    # get logs w/ -n1
    logs = cli.get_logs(project, [INSTANCE1, INSTANCE2], stateboard=True, n=1)
    assert len(logs) == 3
    assert_instance_logs(logs, ID1, log_lines[-1:])
    assert_instance_logs(logs, ID2, log_lines[-1:])
    assert_instance_logs(logs, STATEBOARD_ID, log_lines[-1:])

    # get logs w/ -n5
    logs = cli.get_logs(project, [INSTANCE1, INSTANCE2], stateboard=True, n=5)
    assert len(logs) == 3
    assert_instance_logs(logs, ID1, log_lines[-5:])
    assert_instance_logs(logs, ID2, log_lines[-5:])
    assert_instance_logs(logs, STATEBOARD_ID, log_lines[-5:])
