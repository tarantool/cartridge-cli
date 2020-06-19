import os
import shutil
import pytest

from utils import get_instance_id, get_stateboard_name
from utils import check_instances_running, check_instances_stopped
from utils import DEFAULT_CFG
from utils import DEFAULT_SCRIPT
from utils import STATUS_NOT_STARTED, STATUS_RUNNING, STATUS_STOPPED
from utils import write_conf

from project import patch_init_to_send_statuses


CARTRIDGE_CONF = '.cartridge.yml'


# #####
# Tests
# #####
def test_start_interactive_by_id(start_stop_cli, project_with_patched_init):
    project = project_with_patched_init
    cli = start_stop_cli

    ID1 = get_instance_id(project.name, 'instance-1')

    # start instance-1
    cli.start(project, [ID1])
    check_instances_running(cli, project, [ID1])


def test_start_stop_by_id(start_stop_cli, project_with_patched_init):
    project = project_with_patched_init
    cli = start_stop_cli

    ID1 = get_instance_id(project.name, 'instance-1')
    ID2 = get_instance_id(project.name, 'instance-2')

    # start instance-1 and instance-2
    cli.start(project, [ID1, ID2], daemonized=True)
    check_instances_running(cli, project, [ID1, ID2], daemonized=True)

    # stop instance-1
    cli.stop(project, [ID1])
    check_instances_running(cli, project, [ID2], daemonized=True)
    check_instances_stopped(cli, project, [ID1])


def test_start_interactive_by_id_with_stateboard(start_stop_cli, project_with_patched_init):
    project = project_with_patched_init
    cli = start_stop_cli

    ID1 = get_instance_id(project.name, 'instance-1')
    ID2 = get_instance_id(project.name, 'instance-2')

    # start instance-1 and instance-2
    cli.start(project, [ID1, ID2], stateboard=True)
    check_instances_running(cli, project, [ID1, ID2], stateboard=True)


def test_start_interactive_stateboard_only(start_stop_cli, project_with_patched_init):
    project = project_with_patched_init
    cli = start_stop_cli

    # start with stateboard-only flag
    cli.start(project, stateboard_only=True)
    check_instances_running(cli, project, stateboard_only=True)


def test_start_stop_by_id_with_stateboard(start_stop_cli, project_with_patched_init):
    project = project_with_patched_init
    cli = start_stop_cli

    ID1 = get_instance_id(project.name, 'instance-1')
    ID2 = get_instance_id(project.name, 'instance-2')

    # start instance-1, instance-2 and stateboard
    cli.start(project, [ID1, ID2], daemonized=True, stateboard=True)
    check_instances_running(cli, project, [ID1, ID2], daemonized=True, stateboard=True)

    # stop instance-1 and stateboard
    cli.stop(project, [ID1], stateboard=True)
    check_instances_running(cli, project, [ID2], daemonized=True)
    check_instances_stopped(cli, project, [ID1], stateboard=True)


def test_start_stop_stateboard_only(start_stop_cli, project_with_patched_init):
    project = project_with_patched_init
    cli = start_stop_cli

    # start with stateboard-only flag
    cli.start(project, daemonized=True, stateboard_only=True)
    check_instances_running(cli, project, daemonized=True, stateboard_only=True)

    # stop stateboard
    cli.stop(project, stateboard_only=True)
    check_instances_stopped(cli, project, stateboard_only=True)


def test_start_interactive_from_conf(start_stop_cli, project_with_patched_init):
    project = project_with_patched_init
    cli = start_stop_cli

    ID1 = get_instance_id(project.name, 'instance-1')
    ID2 = get_instance_id(project.name, 'instance-2')

    write_conf(os.path.join(project.path, DEFAULT_CFG), {
        ID1: {},
        ID2: {},
    })

    # start instances
    cli.start(project)
    check_instances_running(cli, project, [ID1, ID2])


def test_start_stop_from_conf(start_stop_cli, project_with_patched_init):
    project = project_with_patched_init
    cli = start_stop_cli

    ID1 = get_instance_id(project.name, 'instance-1')
    ID2 = get_instance_id(project.name, 'instance-2')

    write_conf(os.path.join(project.path, DEFAULT_CFG), {
        ID1: {},
        ID2: {},
    })

    # start instances
    cli.start(project, daemonized=True)
    check_instances_running(cli, project, [ID1, ID2], daemonized=True)

    # stop instances
    cli.stop(project)
    check_instances_stopped(cli, project, [ID1, ID2])


def test_start_interactive_from_conf_with_stateboard(start_stop_cli, project_with_patched_init):
    project = project_with_patched_init
    cli = start_stop_cli

    ID1 = get_instance_id(project.name, 'instance-1')
    ID2 = get_instance_id(project.name, 'instance-2')

    write_conf(os.path.join(project.path, DEFAULT_CFG), {
        ID1: {},
        ID2: {},
    })

    # start instances
    cli.start(project, stateboard=True)
    check_instances_running(cli, project, [ID1, ID2], stateboard=True)


def test_start_interactive_from_conf_stateboard_only(start_stop_cli, project_with_patched_init):
    project = project_with_patched_init
    cli = start_stop_cli

    ID1 = get_instance_id(project.name, 'instance-1')
    ID2 = get_instance_id(project.name, 'instance-2')

    write_conf(os.path.join(project.path, DEFAULT_CFG), {
        ID1: {},
        ID2: {},
    })

    # start instances
    cli.start(project, stateboard_only=True)
    check_instances_running(cli, project, stateboard_only=True)


def test_start_stop_from_conf_with_stateboard(start_stop_cli, project_with_patched_init):
    project = project_with_patched_init
    cli = start_stop_cli

    ID1 = get_instance_id(project.name, 'instance-1')
    ID2 = get_instance_id(project.name, 'instance-2')

    write_conf(os.path.join(project.path, DEFAULT_CFG), {
        ID1: {},
        ID2: {},
    })

    # start instances
    cli.start(project, daemonized=True, stateboard=True)
    check_instances_running(cli, project, [ID1, ID2], daemonized=True, stateboard=True)

    # stop instances
    cli.stop(project, stateboard=True)
    check_instances_stopped(cli, project, [ID1, ID2], stateboard=True)


def test_start_stop_from_conf_stateboard_only(start_stop_cli, project_with_patched_init):
    project = project_with_patched_init
    cli = start_stop_cli

    ID1 = get_instance_id(project.name, 'instance-1')
    ID2 = get_instance_id(project.name, 'instance-2')

    write_conf(os.path.join(project.path, DEFAULT_CFG), {
        ID1: {},
        ID2: {},
    })

    # start instances
    cli.start(project, daemonized=True, stateboard_only=True)
    check_instances_running(cli, project, daemonized=True, stateboard_only=True)

    # stop instances
    cli.stop(project, stateboard=True)
    check_instances_stopped(cli, project, stateboard_only=True)


def test_status_by_id(start_stop_cli, project_with_patched_init):
    project = project_with_patched_init
    cli = start_stop_cli

    ID1 = get_instance_id(project.name, 'instance-1')
    ID2 = get_instance_id(project.name, 'instance-2')
    STATEBOARD_ID = get_stateboard_name(project.name)

    # get status w/o stateboard
    status = cli.get_status(project, [ID1, ID2])
    assert len(status) == 2
    assert status.get(ID1) == STATUS_NOT_STARTED
    assert status.get(ID2) == STATUS_NOT_STARTED

    # get status w/ stateboard
    status = cli.get_status(project, [ID1, ID2], stateboard=True)
    assert len(status) == 3
    assert status.get(ID1) == STATUS_NOT_STARTED
    assert status.get(ID2) == STATUS_NOT_STARTED
    assert status.get(STATEBOARD_ID) == STATUS_NOT_STARTED

    # get status stateboard-only
    status = cli.get_status(project, stateboard_only=True)
    assert len(status) == 1
    assert status.get(STATEBOARD_ID) == STATUS_NOT_STARTED

    # start instance-1 and stateboard
    cli.start(project, [ID1], stateboard=True, daemonized=True)
    check_instances_running(cli, project, [ID1], stateboard=True, daemonized=True)

    # get status w/o stateboard
    status = cli.get_status(project, [ID1, ID2])
    assert len(status) == 2
    assert status.get(ID1) == STATUS_RUNNING
    assert status.get(ID2) == STATUS_NOT_STARTED

    # get status w/ stateboard
    status = cli.get_status(project, [ID1, ID2], stateboard=True)
    assert len(status) == 3
    assert status.get(ID1) == STATUS_RUNNING
    assert status.get(ID2) == STATUS_NOT_STARTED
    assert status.get(STATEBOARD_ID) == STATUS_RUNNING

    # get status stateboard-only
    status = cli.get_status(project, stateboard_only=True)
    assert len(status) == 1
    assert status.get(STATEBOARD_ID) == STATUS_RUNNING

    # stop instance-1
    cli.stop(project, [ID1])

    # get status w/o stateboard
    status = cli.get_status(project, [ID1, ID2])
    assert len(status) == 2
    assert status.get(ID1) == STATUS_STOPPED
    assert status.get(ID2) == STATUS_NOT_STARTED

    # get status w/ stateboard
    status = cli.get_status(project, [ID1], stateboard=True)
    assert len(status) == 2
    assert status.get(ID1) == STATUS_STOPPED
    assert status.get(STATEBOARD_ID) == STATUS_RUNNING

    # get status stateboard-only
    status = cli.get_status(project, stateboard_only=True)
    assert len(status) == 1
    assert status.get(STATEBOARD_ID) == STATUS_RUNNING


def test_status_from_conf(start_stop_cli, project_with_patched_init):
    project = project_with_patched_init
    cli = start_stop_cli

    ID1 = get_instance_id(project.name, 'instance-1')
    ID2 = get_instance_id(project.name, 'instance-2')
    STATEBOARD_ID = get_stateboard_name(project.name)

    write_conf(os.path.join(project.path, DEFAULT_CFG), {
        ID1: {},
        ID2: {},
    })

    # get status w/o stateboard
    status = cli.get_status(project)
    assert len(status) == 2
    assert status.get(ID1) == STATUS_NOT_STARTED
    assert status.get(ID2) == STATUS_NOT_STARTED

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
    cli.start(project, [ID1], stateboard=True, daemonized=True)
    check_instances_running(cli, project, [ID1], stateboard=True, daemonized=True)

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
    cli.stop(project, [ID1])

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


def test_start_stop_status_cfg(start_stop_cli, project_with_patched_init):
    project = project_with_patched_init
    cli = start_stop_cli

    ID1 = get_instance_id(project.name, 'instance-1')
    ID2 = get_instance_id(project.name, 'instance-2')
    CFG = 'my-conf.yml'

    write_conf(os.path.join(project.path, CFG), {
        ID1: {},
        ID2: {},
    })

    status = cli.get_status(project, cfg=CFG)
    assert status.get(ID1) == STATUS_NOT_STARTED
    assert status.get(ID2) == STATUS_NOT_STARTED

    cli.start(project, stateboard=True, daemonized=True, cfg=CFG)
    check_instances_running(
        cli, project,
        [ID1, ID2],
        stateboard=True, cfg=CFG,
        daemonized=True,
    )

    status = cli.get_status(project, cfg=CFG)
    assert status.get(ID1) == STATUS_RUNNING
    assert status.get(ID2) == STATUS_RUNNING

    cli.stop(project, stateboard=True, cfg=CFG)
    check_instances_stopped(cli, project, [ID1, ID2])

    status = cli.get_status(project, cfg=CFG)
    assert status.get(ID1) == STATUS_STOPPED
    assert status.get(ID2) == STATUS_STOPPED


def test_start_stop_status_run_dir(start_stop_cli, project_with_patched_init):
    project = project_with_patched_init
    cli = start_stop_cli

    ID1 = get_instance_id(project.name, 'instance-1')
    ID2 = get_instance_id(project.name, 'instance-2')
    STATEBOARD_ID = get_stateboard_name(project.name)
    RUN_DIR = 'my-run'

    status = cli.get_status(project, [ID1, ID2], stateboard=True, run_dir=RUN_DIR)
    assert status.get(ID1) == STATUS_NOT_STARTED
    assert status.get(ID2) == STATUS_NOT_STARTED

    cli.start(project, [ID1], stateboard=True, daemonized=True, run_dir=RUN_DIR)
    check_instances_running(cli, project, [ID1], stateboard=True, run_dir=RUN_DIR, daemonized=True)

    status = cli.get_status(project, [ID1, ID2], stateboard=True, run_dir=RUN_DIR)
    assert len(status) == 3
    assert status.get(ID1) == STATUS_RUNNING
    assert status.get(ID2) == STATUS_NOT_STARTED
    assert status.get(STATEBOARD_ID) == STATUS_RUNNING

    cli.stop(project, [ID1], stateboard=True, run_dir=RUN_DIR)
    check_instances_stopped(cli, project, [ID1], stateboard=True, run_dir=RUN_DIR)

    status = cli.get_status(project, [ID1, ID2], stateboard=True, run_dir=RUN_DIR)
    assert len(status) == 3
    assert status.get(ID1) == STATUS_STOPPED
    assert status.get(ID2) == STATUS_NOT_STARTED
    assert status.get(STATEBOARD_ID) == STATUS_STOPPED


def test_start_stop_status_run_dir_from_conf(start_stop_cli, project_with_patched_init):
    project = project_with_patched_init
    cli = start_stop_cli

    ID1 = get_instance_id(project.name, 'instance-1')
    ID2 = get_instance_id(project.name, 'instance-2')
    STATEBOARD_ID = get_stateboard_name(project.name)
    RUN_DIR = 'my-run'

    write_conf(os.path.join(project.path, CARTRIDGE_CONF), {
        'run-dir': RUN_DIR,
    })

    status = cli.get_status(project, [ID1, ID2], stateboard=True)
    assert status.get(ID1) == STATUS_NOT_STARTED
    assert status.get(ID2) == STATUS_NOT_STARTED

    cli.start(project, [ID1], stateboard=True, daemonized=True)
    check_instances_running(cli, project, [ID1], stateboard=True, run_dir=RUN_DIR, daemonized=True)

    status = cli.get_status(project, [ID1, ID2], stateboard=True)
    assert len(status) == 3
    assert status.get(ID1) == STATUS_RUNNING
    assert status.get(ID2) == STATUS_NOT_STARTED
    assert status.get(STATEBOARD_ID) == STATUS_RUNNING

    cli.stop(project, [ID1], stateboard=True)
    check_instances_stopped(cli, project, [ID1], stateboard=True, run_dir=RUN_DIR)

    status = cli.get_status(project, [ID1, ID2], stateboard=True)
    assert len(status) == 3
    assert status.get(ID1) == STATUS_STOPPED
    assert status.get(ID2) == STATUS_NOT_STARTED
    assert status.get(STATEBOARD_ID) == STATUS_STOPPED


def test_start_data_dir(start_stop_cli, project_with_patched_init):
    project = project_with_patched_init
    cli = start_stop_cli

    ID1 = get_instance_id(project.name, 'instance-1')
    ID2 = get_instance_id(project.name, 'instance-2')
    DATA_DIR = 'my-data'

    cli.start(project, [ID1, ID2], stateboard=True, data_dir=DATA_DIR)
    check_instances_running(
        cli, project,
        [ID1, ID2],
        stateboard=True, data_dir=DATA_DIR
    )


def test_start_data_dir_from_conf(start_stop_cli, project_with_patched_init):
    project = project_with_patched_init
    cli = start_stop_cli

    ID1 = get_instance_id(project.name, 'instance-1')
    ID2 = get_instance_id(project.name, 'instance-2')
    DATA_DIR = 'my-data'

    write_conf(os.path.join(project.path, CARTRIDGE_CONF), {
        'data-dir': DATA_DIR,
    })

    cli.start(project, [ID1, ID2], stateboard=True)
    check_instances_running(
        cli, project,
        [ID1, ID2],
        stateboard=True, data_dir=DATA_DIR
    )


def test_start_script(start_stop_cli, project_with_patched_init):
    project = project_with_patched_init
    cli = start_stop_cli

    ID1 = get_instance_id(project.name, 'instance-1')
    ID2 = get_instance_id(project.name, 'instance-2')

    SCRIPT = 'my-init.lua'
    shutil.copyfile(os.path.join(project.path, DEFAULT_SCRIPT), os.path.join(project.path, SCRIPT))

    cli.start(project, [ID1, ID2], stateboard=True, script=SCRIPT)
    check_instances_running(
        cli, project,
        [ID1, ID2],
        stateboard=True, script=SCRIPT
    )


def test_start_script_from_conf(start_stop_cli, project_with_patched_init):
    project = project_with_patched_init
    cli = start_stop_cli

    ID1 = get_instance_id(project.name, 'instance-1')
    ID2 = get_instance_id(project.name, 'instance-2')

    SCRIPT = 'my-init.lua'
    shutil.copyfile(os.path.join(project.path, DEFAULT_SCRIPT), os.path.join(project.path, SCRIPT))

    write_conf(os.path.join(project.path, CARTRIDGE_CONF), {
        'script': SCRIPT,
    })

    cli.start(project, [ID1, ID2], stateboard=True)
    check_instances_running(
        cli, project,
        [ID1, ID2],
        stateboard=True, script=SCRIPT
    )


def test_start_log_dir(start_stop_cli, project_with_patched_init):
    project = project_with_patched_init
    cli = start_stop_cli

    ID1 = get_instance_id(project.name, 'instance-1')
    ID2 = get_instance_id(project.name, 'instance-2')

    LOG_DIR = 'my-log-dir'

    cli.start(project, [ID1, ID2], daemonized=True, stateboard=True, log_dir=LOG_DIR)
    check_instances_running(
        cli, project,
        [ID1, ID2],
        daemonized=True,
        stateboard=True, log_dir=LOG_DIR
    )


def test_start_log_dir_from_conf(start_stop_cli, project_with_patched_init):
    project = project_with_patched_init
    cli = start_stop_cli

    ID1 = get_instance_id(project.name, 'instance-1')
    ID2 = get_instance_id(project.name, 'instance-2')

    LOG_DIR = 'my-log-dir'

    write_conf(os.path.join(project.path, CARTRIDGE_CONF), {
        'log-dir': LOG_DIR,
    })

    cli.start(project, [ID1, ID2], daemonized=True, stateboard=True)
    check_instances_running(
        cli, project,
        [ID1, ID2],
        daemonized=True,
        stateboard=True, log_dir=LOG_DIR
    )


def test_notify_status_failed(start_stop_cli, project_with_patched_init):
    project = project_with_patched_init
    cli = start_stop_cli

    HORRIBLE_ERR = "SOME\nMULTILINE\nHORRIBLE ERROR"
    patch_init_to_send_statuses(project, ["Failed: %s" % HORRIBLE_ERR])

    ID1 = get_instance_id(project.name, 'instance-1')

    logs = cli.start(project, [ID1], daemonized=True, capture_output=True, exp_rc=1)
    assert any([HORRIBLE_ERR in log_entry.msg for log_entry in logs])

    logs = cli.start(project, stateboard_only=True, daemonized=True, capture_output=True, exp_rc=1)
    assert any([HORRIBLE_ERR in log_entry.msg for log_entry in logs])


@pytest.mark.parametrize('status', ['running', 'loading', 'orphan', 'hot_standby'])
def test_notify_status_allowed(start_stop_cli, project_with_patched_init, status):
    project = project_with_patched_init
    cli = start_stop_cli

    patch_init_to_send_statuses(project, [status])

    ID1 = get_instance_id(project.name, 'instance-1')

    cli.start(project, [ID1], daemonized=True, stateboard=True)
    check_instances_running(cli, project, [ID1], daemonized=True, stateboard=True)


def test_project_with_non_exitent_script(start_stop_cli, project_with_patched_init):
    project = project_with_patched_init
    cli = start_stop_cli

    os.remove(os.path.join(project.path, DEFAULT_SCRIPT))

    ID1 = get_instance_id(project.name, 'instance-1')

    logs = cli.start(project, [ID1], daemonized=True, capture_output=True, exp_rc=1)
    assert any(["Can't use instance entrypoint" in log_entry.msg for log_entry in logs])
