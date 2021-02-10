import os
import shutil
import pytest
import re

from utils import check_instances_running, check_instances_stopped
from utils import STATUS_NOT_STARTED, STATUS_RUNNING, STATUS_STOPPED
from utils import write_conf

from project import patch_init_to_send_statuses
from project import patch_init_to_send_ready_after_timeout


# #####
# Tests
# #####
def test_start_interactive_by_name(start_stop_cli, project_without_dependencies):
    project = project_without_dependencies
    cli = start_stop_cli

    INSTANCE1 = 'instance-1'

    # start instance-1
    cli.start(project, [INSTANCE1])
    check_instances_running(cli, project, [INSTANCE1], stateboard=True)


def test_start_stop_by_name(start_stop_cli, project_without_dependencies):
    project = project_without_dependencies
    cli = start_stop_cli

    INSTANCE1 = 'instance-1'
    INSTANCE2 = 'instance-2'

    # start instance-1 and instance-2
    cli.start(project, [INSTANCE1, INSTANCE2], daemonized=True)
    check_instances_running(cli, project, [INSTANCE1, INSTANCE2], stateboard=True, daemonized=True)

    # stop instance-1
    cli.stop(project, [INSTANCE1])
    check_instances_running(cli, project, [INSTANCE2], daemonized=True)
    check_instances_stopped(cli, project, [INSTANCE1])


def test_start_interactive_by_name_with_stateboard(start_stop_cli, project_without_dependencies):
    project = project_without_dependencies
    cli = start_stop_cli

    INSTANCE1 = 'instance-1'
    INSTANCE2 = 'instance-2'

    # start instance-1 and instance-2
    cli.start(project, [INSTANCE1, INSTANCE2], stateboard=True)
    check_instances_running(cli, project, [INSTANCE1, INSTANCE2], stateboard=True)


def test_start_interactive_stateboard_only(start_stop_cli, project_without_dependencies):
    project = project_without_dependencies
    cli = start_stop_cli

    # start with stateboard-only flag
    cli.start(project, stateboard_only=True)
    check_instances_running(cli, project, stateboard_only=True)


def test_start_stop_by_name_with_stateboard(start_stop_cli, project_without_dependencies):
    project = project_without_dependencies
    cli = start_stop_cli

    INSTANCE1 = 'instance-1'
    INSTANCE2 = 'instance-2'

    # start instance-1, instance-2 and stateboard
    cli.start(project, [INSTANCE1, INSTANCE2], daemonized=True, stateboard=True)
    check_instances_running(cli, project, [INSTANCE1, INSTANCE2], daemonized=True, stateboard=True)

    # stop instance-1 and stateboard
    cli.stop(project, [INSTANCE1], stateboard=True)
    check_instances_running(cli, project, [INSTANCE2], daemonized=True)
    check_instances_stopped(cli, project, [INSTANCE1], stateboard=True)


def test_start_stop_stateboard_only(start_stop_cli, project_without_dependencies):
    project = project_without_dependencies
    cli = start_stop_cli

    # start with stateboard-only flag
    cli.start(project, daemonized=True, stateboard_only=True)
    check_instances_running(cli, project, daemonized=True, stateboard_only=True)

    # stop stateboard
    cli.stop(project, stateboard_only=True)
    check_instances_stopped(cli, project, stateboard_only=True)


def test_start_interactive_from_conf(start_stop_cli, project_without_dependencies):
    project = project_without_dependencies
    cli = start_stop_cli

    INSTANCE1 = 'instance-1'
    INSTANCE2 = 'instance-2'

    write_conf(project.get_cfg_path(), {
        project.get_instance_id(INSTANCE1): {},
        project.get_instance_id(INSTANCE2): {},
    })

    # start instances
    cli.start(project)
    check_instances_running(cli, project, [INSTANCE1, INSTANCE2], stateboard=True)


def test_start_stop_from_conf(start_stop_cli, project_without_dependencies):
    project = project_without_dependencies
    cli = start_stop_cli

    INSTANCE1 = 'instance-1'
    INSTANCE2 = 'instance-2'

    write_conf(project.get_cfg_path(), {
        project.get_instance_id(INSTANCE1): {},
        project.get_instance_id(INSTANCE2): {},
    })

    # start instances
    cli.start(project, daemonized=True)
    check_instances_running(cli, project, [INSTANCE1, INSTANCE2], stateboard=True, daemonized=True)

    # stop instances
    cli.stop(project)
    check_instances_stopped(cli, project, [INSTANCE1, INSTANCE2])


def test_start_interactive_from_conf_with_stateboard(start_stop_cli, project_without_dependencies):
    project = project_without_dependencies
    cli = start_stop_cli

    INSTANCE1 = 'instance-1'
    INSTANCE2 = 'instance-2'

    write_conf(project.get_cfg_path(), {
        project.get_instance_id(INSTANCE1): {},
        project.get_instance_id(INSTANCE2): {},
    })

    # start instances
    cli.start(project, stateboard=True)
    check_instances_running(cli, project, [INSTANCE1, INSTANCE2], stateboard=True)


def test_start_interactive_from_conf_stateboard_only(start_stop_cli, project_without_dependencies):
    project = project_without_dependencies
    cli = start_stop_cli

    INSTANCE1 = 'instance-1'
    INSTANCE2 = 'instance-2'

    write_conf(project.get_cfg_path(), {
        project.get_instance_id(INSTANCE1): {},
        project.get_instance_id(INSTANCE2): {},
    })

    # start instances
    cli.start(project, stateboard_only=True)
    check_instances_running(cli, project, stateboard_only=True)


def test_start_stop_from_conf_with_stateboard(start_stop_cli, project_without_dependencies):
    project = project_without_dependencies
    cli = start_stop_cli

    INSTANCE1 = 'instance-1'
    INSTANCE2 = 'instance-2'

    write_conf(project.get_cfg_path(), {
        project.get_instance_id(INSTANCE1): {},
        project.get_instance_id(INSTANCE2): {},
    })

    # start instances
    cli.start(project, daemonized=True, stateboard=True)
    check_instances_running(cli, project, [INSTANCE1, INSTANCE2], daemonized=True, stateboard=True)

    # stop instances
    cli.stop(project, stateboard=True)
    check_instances_stopped(cli, project, [INSTANCE1, INSTANCE2], stateboard=True)


def test_start_stop_from_conf_stateboard_only(start_stop_cli, project_without_dependencies):
    project = project_without_dependencies
    cli = start_stop_cli

    INSTANCE1 = 'instance-1'
    INSTANCE2 = 'instance-2'

    write_conf(project.get_cfg_path(), {
        project.get_instance_id(INSTANCE1): {},
        project.get_instance_id(INSTANCE2): {},
    })

    # start instances
    cli.start(project, daemonized=True, stateboard_only=True)
    check_instances_running(cli, project, daemonized=True, stateboard_only=True)

    # stop instances
    cli.stop(project, stateboard=True)
    check_instances_stopped(cli, project, stateboard_only=True)


def test_status_by_name(start_stop_cli, project_without_dependencies):
    project = project_without_dependencies
    cli = start_stop_cli

    INSTANCE1 = 'instance-1'
    INSTANCE2 = 'instance-2'

    ID1 = project.get_instance_id(INSTANCE1)
    ID2 = project.get_instance_id(INSTANCE2)
    STATEBOARD_ID = project.get_stateboard_id()

    # get status w/ stateboard
    status = cli.get_status(project, [INSTANCE1, INSTANCE2], stateboard=True)
    assert len(status) == 3
    assert status.get(ID1) == STATUS_NOT_STARTED
    assert status.get(ID2) == STATUS_NOT_STARTED
    assert status.get(STATEBOARD_ID) == STATUS_NOT_STARTED

    # get status stateboard-only
    status = cli.get_status(project, stateboard_only=True)
    assert len(status) == 1
    assert status.get(STATEBOARD_ID) == STATUS_NOT_STARTED

    # start instance-1 and stateboard
    cli.start(project, [INSTANCE1], stateboard=True, daemonized=True)
    check_instances_running(cli, project, [INSTANCE1], stateboard=True, daemonized=True)

    # get status w/o stateboard
    status = cli.get_status(project, [INSTANCE1, INSTANCE2])
    assert len(status) == 2
    assert status.get(ID1) == STATUS_RUNNING
    assert status.get(ID2) == STATUS_NOT_STARTED

    # get status w/ stateboard
    status = cli.get_status(project, [INSTANCE1, INSTANCE2], stateboard=True)
    assert len(status) == 3
    assert status.get(ID1) == STATUS_RUNNING
    assert status.get(ID2) == STATUS_NOT_STARTED
    assert status.get(STATEBOARD_ID) == STATUS_RUNNING

    # get status stateboard-only
    status = cli.get_status(project, stateboard_only=True)
    assert len(status) == 1
    assert status.get(STATEBOARD_ID) == STATUS_RUNNING

    # stop instance-1
    cli.stop(project, [INSTANCE1])

    # get status w/o stateboard
    status = cli.get_status(project, [INSTANCE1, INSTANCE2])
    assert len(status) == 2
    assert status.get(ID1) == STATUS_STOPPED
    assert status.get(ID2) == STATUS_NOT_STARTED

    # get status w/ stateboard
    status = cli.get_status(project, [INSTANCE1], stateboard=True)
    assert len(status) == 2
    assert status.get(ID1) == STATUS_STOPPED
    assert status.get(STATEBOARD_ID) == STATUS_RUNNING

    # get status stateboard-only
    status = cli.get_status(project, stateboard_only=True)
    assert len(status) == 1
    assert status.get(STATEBOARD_ID) == STATUS_RUNNING


def test_status_from_conf(start_stop_cli, project_without_dependencies):
    project = project_without_dependencies
    cli = start_stop_cli

    INSTANCE1 = 'instance-1'
    INSTANCE2 = 'instance-2'

    ID1 = project.get_instance_id(INSTANCE1)
    ID2 = project.get_instance_id(INSTANCE2)
    STATEBOARD_ID = project.get_stateboard_id()

    write_conf(project.get_cfg_path(), {
        ID1: {},
        ID2: {},
    })

    # get status w/ stateboard
    status = cli.get_status(project, stateboard=True)
    assert len(status) == 3
    assert status.get(ID1) == STATUS_NOT_STARTED
    assert status.get(ID2) == STATUS_NOT_STARTED
    assert status.get(STATEBOARD_ID) == STATUS_NOT_STARTED

    # get status stateboard-only
    status = cli.get_status(project, stateboard_only=True)
    assert len(status) == 1
    assert status.get(STATEBOARD_ID) == STATUS_NOT_STARTED

    # start instance-1 and stateboard
    cli.start(project, [INSTANCE1], stateboard=True, daemonized=True)
    check_instances_running(cli, project, [INSTANCE1], stateboard=True, daemonized=True)

    # get status w/o stateboard
    status = cli.get_status(project)
    assert len(status) == 2
    assert status.get(ID1) == STATUS_RUNNING
    assert status.get(ID2) == STATUS_NOT_STARTED

    # get status w/ stateboard
    status = cli.get_status(project, stateboard=True)
    assert len(status) == 3
    assert status.get(ID1) == STATUS_RUNNING
    assert status.get(ID2) == STATUS_NOT_STARTED
    assert status.get(STATEBOARD_ID) == STATUS_RUNNING

    # get status stateboard-only
    status = cli.get_status(project, stateboard_only=True)
    assert len(status) == 1
    assert status.get(STATEBOARD_ID) == STATUS_RUNNING

    # stop instance-1
    cli.stop(project, [INSTANCE1])

    # get status w/o stateboard
    status = cli.get_status(project)
    assert len(status) == 2
    assert status.get(ID1) == STATUS_STOPPED
    assert status.get(ID2) == STATUS_NOT_STARTED

    # get status w/ stateboard
    status = cli.get_status(project, stateboard=True)
    assert len(status) == 3
    assert status.get(ID1) == STATUS_STOPPED
    assert status.get(ID2) == STATUS_NOT_STARTED
    assert status.get(STATEBOARD_ID) == STATUS_RUNNING

    # get status stateboard-only
    status = cli.get_status(project, stateboard_only=True)
    assert len(status) == 1
    assert status.get(STATEBOARD_ID) == STATUS_RUNNING


def test_start_stop_status_cfg(start_stop_cli, project_without_dependencies):
    project = project_without_dependencies
    cli = start_stop_cli

    INSTANCE1 = 'instance-1'
    INSTANCE2 = 'instance-2'

    ID1 = project.get_instance_id(INSTANCE1)
    ID2 = project.get_instance_id(INSTANCE2)

    CFG = 'my-conf.yml'

    write_conf(project.get_cfg_path(CFG), {
        ID1: {},
        ID2: {},
    })

    status = cli.get_status(project, cfg=CFG)
    assert status.get(ID1) == STATUS_NOT_STARTED
    assert status.get(ID2) == STATUS_NOT_STARTED

    cli.start(project, stateboard=True, daemonized=True, cfg=CFG)
    check_instances_running(
        cli, project,
        [INSTANCE1, INSTANCE2],
        stateboard=True, cfg=CFG,
        daemonized=True,
    )

    status = cli.get_status(project, cfg=CFG)
    assert status.get(ID1) == STATUS_RUNNING
    assert status.get(ID2) == STATUS_RUNNING

    cli.stop(project, stateboard=True, cfg=CFG)
    check_instances_stopped(cli, project, [INSTANCE1, INSTANCE2])

    status = cli.get_status(project, cfg=CFG)
    assert status.get(ID1) == STATUS_STOPPED
    assert status.get(ID2) == STATUS_STOPPED


def test_start_stop_status_run_dir(start_stop_cli, project_without_dependencies):
    project = project_without_dependencies
    cli = start_stop_cli

    INSTANCE1 = 'instance-1'
    INSTANCE2 = 'instance-2'

    ID1 = project.get_instance_id(INSTANCE1)
    ID2 = project.get_instance_id(INSTANCE2)
    STATEBOARD_ID = project.get_stateboard_id()

    RUN_DIR = 'my-run'

    status = cli.get_status(project, [INSTANCE1, INSTANCE2], stateboard=True, run_dir=RUN_DIR)
    assert status.get(ID1) == STATUS_NOT_STARTED
    assert status.get(ID2) == STATUS_NOT_STARTED

    cli.start(project, [INSTANCE1], stateboard=True, daemonized=True, run_dir=RUN_DIR)
    check_instances_running(cli, project, [INSTANCE1], stateboard=True, run_dir=RUN_DIR, daemonized=True)

    status = cli.get_status(project, [INSTANCE1, INSTANCE2], stateboard=True, run_dir=RUN_DIR)
    assert len(status) == 3
    assert status.get(ID1) == STATUS_RUNNING
    assert status.get(ID2) == STATUS_NOT_STARTED
    assert status.get(STATEBOARD_ID) == STATUS_RUNNING

    cli.stop(project, [INSTANCE1], stateboard=True, run_dir=RUN_DIR)
    check_instances_stopped(cli, project, [INSTANCE1], stateboard=True, run_dir=RUN_DIR)

    status = cli.get_status(project, [INSTANCE1, INSTANCE2], stateboard=True, run_dir=RUN_DIR)
    assert len(status) == 3
    assert status.get(ID1) == STATUS_STOPPED
    assert status.get(ID2) == STATUS_NOT_STARTED
    assert status.get(STATEBOARD_ID) == STATUS_STOPPED


def test_start_stop_status_run_dir_from_conf(start_stop_cli, project_without_dependencies):
    project = project_without_dependencies
    cli = start_stop_cli

    INSTANCE1 = 'instance-1'
    INSTANCE2 = 'instance-2'

    ID1 = project.get_instance_id(INSTANCE1)
    ID2 = project.get_instance_id(INSTANCE2)
    STATEBOARD_ID = project.get_stateboard_id()

    RUN_DIR = 'my-run'

    write_conf(project.get_cli_cfg_path(), {
        'run-dir': RUN_DIR,
    })

    status = cli.get_status(project, [INSTANCE1, INSTANCE2], stateboard=True)
    assert status.get(ID1) == STATUS_NOT_STARTED
    assert status.get(ID2) == STATUS_NOT_STARTED

    cli.start(project, [INSTANCE1], stateboard=True, daemonized=True)
    check_instances_running(cli, project, [INSTANCE1], stateboard=True, run_dir=RUN_DIR, daemonized=True)

    status = cli.get_status(project, [INSTANCE1, INSTANCE2], stateboard=True)
    assert len(status) == 3
    assert status.get(ID1) == STATUS_RUNNING
    assert status.get(ID2) == STATUS_NOT_STARTED
    assert status.get(STATEBOARD_ID) == STATUS_RUNNING

    cli.stop(project, [INSTANCE1], stateboard=True)
    check_instances_stopped(cli, project, [INSTANCE1], stateboard=True, run_dir=RUN_DIR)

    status = cli.get_status(project, [INSTANCE1, INSTANCE2], stateboard=True)
    assert len(status) == 3
    assert status.get(ID1) == STATUS_STOPPED
    assert status.get(ID2) == STATUS_NOT_STARTED
    assert status.get(STATEBOARD_ID) == STATUS_STOPPED


def test_start_stop_status_stateboard_from_conf(start_stop_cli, project_without_dependencies):
    project = project_without_dependencies
    cli = start_stop_cli

    INSTANCE1 = 'instance-1'
    INSTANCE2 = 'instance-2'

    ID1 = project.get_instance_id(INSTANCE1)
    ID2 = project.get_instance_id(INSTANCE2)
    STATEBOARD_ID = project.get_stateboard_id()

    write_conf(project.get_cli_cfg_path(), {
        'stateboard': 'true',
    })

    status = cli.get_status(project, [INSTANCE1, INSTANCE2], stateboard=True)
    assert status.get(ID1) == STATUS_NOT_STARTED
    assert status.get(ID2) == STATUS_NOT_STARTED

    cli.start(project, [INSTANCE1], stateboard=True, daemonized=True)
    check_instances_running(cli, project, [INSTANCE1], stateboard=True, run_dir=RUN_DIR, daemonized=True)

    status = cli.get_status(project, [INSTANCE1, INSTANCE2], stateboard=True)
    assert len(status) == 3
    assert status.get(ID1) == STATUS_RUNNING
    assert status.get(ID2) == STATUS_NOT_STARTED
    assert status.get(STATEBOARD_ID) == STATUS_RUNNING

    cli.stop(project, [INSTANCE1], stateboard=True)
    check_instances_stopped(cli, project, [INSTANCE1], stateboard=True, run_dir=RUN_DIR)

    status = cli.get_status(project, [INSTANCE1, INSTANCE2], stateboard=True)
    assert len(status) == 3
    assert status.get(ID1) == STATUS_STOPPED
    assert status.get(ID2) == STATUS_NOT_STARTED
    assert status.get(STATEBOARD_ID) == STATUS_STOPPED


def test_start_data_dir(start_stop_cli, project_without_dependencies):
    project = project_without_dependencies
    cli = start_stop_cli

    INSTANCE1 = 'instance-1'
    INSTANCE2 = 'instance-2'

    DATA_DIR = 'my-data'

    cli.start(project, [INSTANCE1, INSTANCE2], stateboard=True, data_dir=DATA_DIR)
    check_instances_running(
        cli, project,
        [INSTANCE1, INSTANCE2],
        stateboard=True, data_dir=DATA_DIR
    )


def test_start_data_dir_from_conf(start_stop_cli, project_without_dependencies):
    project = project_without_dependencies
    cli = start_stop_cli

    INSTANCE1 = 'instance-1'
    INSTANCE2 = 'instance-2'

    DATA_DIR = 'my-data'

    write_conf(project.get_cli_cfg_path(), {
        'data-dir': DATA_DIR,
    })

    cli.start(project, [INSTANCE1, INSTANCE2], stateboard=True)
    check_instances_running(
        cli, project,
        [INSTANCE1, INSTANCE2],
        stateboard=True, data_dir=DATA_DIR
    )


def test_start_script(start_stop_cli, project_without_dependencies):
    project = project_without_dependencies
    cli = start_stop_cli

    INSTANCE1 = 'instance-1'
    INSTANCE2 = 'instance-2'

    SCRIPT = 'my-init.lua'
    shutil.copyfile(project.get_script(), os.path.join(project.path, SCRIPT))

    cli.start(project, [INSTANCE1, INSTANCE2], stateboard=True, script=SCRIPT)
    check_instances_running(
        cli, project,
        [INSTANCE1, INSTANCE2],
        stateboard=True, script=SCRIPT
    )


def test_start_script_from_conf(start_stop_cli, project_without_dependencies):
    project = project_without_dependencies
    cli = start_stop_cli

    INSTANCE1 = 'instance-1'
    INSTANCE2 = 'instance-2'

    SCRIPT = 'my-init.lua'
    shutil.copyfile(project.get_script(), os.path.join(project.path, SCRIPT))

    write_conf(project.get_cli_cfg_path(), {
        'script': SCRIPT,
    })

    cli.start(project, [INSTANCE1, INSTANCE2], stateboard=True)
    check_instances_running(
        cli, project,
        [INSTANCE1, INSTANCE2],
        stateboard=True, script=SCRIPT
    )


def test_start_log_dir(start_stop_cli, project_without_dependencies):
    project = project_without_dependencies
    cli = start_stop_cli

    INSTANCE1 = 'instance-1'
    INSTANCE2 = 'instance-2'

    LOG_DIR = 'my-log-dir'

    cli.start(project, [INSTANCE1, INSTANCE2], daemonized=True, stateboard=True, log_dir=LOG_DIR)
    check_instances_running(
        cli, project,
        [INSTANCE1, INSTANCE2],
        daemonized=True,
        stateboard=True, log_dir=LOG_DIR
    )


def test_start_log_dir_from_conf(start_stop_cli, project_without_dependencies):
    project = project_without_dependencies
    cli = start_stop_cli

    INSTANCE1 = 'instance-1'
    INSTANCE2 = 'instance-2'

    LOG_DIR = 'my-log-dir'

    write_conf(project.get_cli_cfg_path(), {
        'log-dir': LOG_DIR,
    })

    cli.start(project, [INSTANCE1, INSTANCE2], daemonized=True, stateboard=True)
    check_instances_running(
        cli, project,
        [INSTANCE1, INSTANCE2],
        daemonized=True,
        stateboard=True, log_dir=LOG_DIR
    )


def test_notify_status_failed(start_stop_cli, project_without_dependencies):
    project = project_without_dependencies
    cli = start_stop_cli

    HORRIBLE_ERR = "SOME HORRIBLE ERROR"
    patch_init_to_send_statuses(project, ["Failed: %s" % HORRIBLE_ERR])

    INSTANCE1 = 'instance-1'

    logs = cli.start(project, [INSTANCE1], daemonized=True, capture_output=True, exp_rc=1)
    assert any([HORRIBLE_ERR in msg for msg in logs])

    logs = cli.start(project, stateboard_only=True, daemonized=True, capture_output=True, exp_rc=1)
    assert any([HORRIBLE_ERR in msg for msg in logs])


@pytest.mark.parametrize('status', ['running', 'loading', 'orphan', 'hot_standby'])
def test_notify_status_allowed(start_stop_cli, project_without_dependencies, status):
    project = project_without_dependencies
    cli = start_stop_cli

    patch_init_to_send_statuses(project, [status])

    INSTANCE1 = 'instance-1'

    cli.start(project, [INSTANCE1], daemonized=True, stateboard=True)
    check_instances_running(cli, project, [INSTANCE1], daemonized=True, stateboard=True)


def test_project_with_non_existent_script(start_stop_cli, project_without_dependencies):
    project = project_without_dependencies
    cli = start_stop_cli

    os.remove(project.get_script())

    INSTANCE1 = 'instance-1'

    logs = cli.start(project, [INSTANCE1], daemonized=True, capture_output=True, exp_rc=1)
    assert any(["Can't use instance entrypoint" in msg for msg in logs])


def test_start_with_timeout(start_stop_cli, project_without_dependencies):
    project = project_without_dependencies
    cli = start_stop_cli

    TIMEOUT_SECONDS = 2

    patch_init_to_send_ready_after_timeout(project, TIMEOUT_SECONDS)
    # patch_init_to_log_signals(project)

    INSTANCE1 = 'instance-1'
    INSTANCE2 = 'instance-2'

    ID1 = project.get_instance_id(INSTANCE1)
    ID2 = project.get_instance_id(INSTANCE2)
    STATEBOARD_ID = project.get_stateboard_id()

    # start w/ timeout > TIMEOUT_SECONDS
    cli.terminate()
    cli.start(
        project, [INSTANCE1, INSTANCE2], daemonized=True, stateboard=True,
        timeout="{}s".format(TIMEOUT_SECONDS+1)
    )
    check_instances_running(cli, project, [INSTANCE1, INSTANCE2], daemonized=True, stateboard=True)

    # start w/ timeout < TIMEOUT_SECONDS
    cli.terminate()
    logs = cli.start(
        project, [INSTANCE1, INSTANCE2], daemonized=True, stateboard=True,
        timeout="{}s".format(TIMEOUT_SECONDS-1),
        capture_output=True, exp_rc=1,
    )
    check_instances_stopped(cli, project, [INSTANCE1, INSTANCE2], stateboard=True)
    for instance_id in [ID1, ID2, STATEBOARD_ID]:
        assert any([re.search(r"%s:.+Timeout was reached" % instance_id, msg) is not None for msg in logs])

    # start w/ timeout 0s
    cli.terminate()
    logs = cli.start(
        project, [INSTANCE1, INSTANCE2], daemonized=True, stateboard=True,
        timeout="{}s".format(0),
        capture_output=True,
    )
    check_instances_running(cli, project, [INSTANCE1, INSTANCE2], daemonized=True, stateboard=True)
    for instance_id in [ID1, ID2, STATEBOARD_ID]:
        assert all([re.search(r"%s:.+Timeout was reached" % instance_id, msg) is None for msg in logs])


def test_stop_signals(start_stop_cli, project_ignore_sigterm):
    project = project_ignore_sigterm
    cli = start_stop_cli

    INSTANCE1 = 'instance-1'
    INSTANCE2 = 'instance-2'

    # start instances
    cli.start(project, [INSTANCE1, INSTANCE2], stateboard=True, daemonized=True)
    check_instances_running(cli, project, [INSTANCE1, INSTANCE2], stateboard=True, daemonized=True)

    # try to stop instaces using `cartridge stop`
    # since it sends SIGTERM and instances ignore this signal,
    # instances are still running
    cli.stop(project, [INSTANCE1, INSTANCE2], stateboard=True,)
    check_instances_running(cli, project, [INSTANCE1, INSTANCE2], stateboard=True, daemonized=True)

    # now, use `cartridge stop -d`
    # it sends SIGKILL that can't be ignored,
    # so instances are stopped
    cli.stop(project, [INSTANCE1, INSTANCE2], stateboard=True, force=True)
    check_instances_stopped(cli, project, [INSTANCE1, INSTANCE2], stateboard=True)
