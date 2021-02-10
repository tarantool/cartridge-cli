import os

from utils import write_conf
from utils import check_instances_running
from utils import check_instances_stopped


FILES_TO_BE_DELETED = ['log', 'workdir', 'console-sock', 'notify-sock', 'pid']


def get_instance_files(project, instance_name, log_dir=None, data_dir=None, run_dir=None):
    run_dir = project.get_run_dir(run_dir)

    return {
        'log': project.get_log_dir(instance_name, log_dir),
        'workdir': project.get_workdir(instance_name, data_dir),
        'console-sock': project.get_console_sock(instance_name, run_dir),
        'pid': project.get_pidfile(instance_name, run_dir),
        'notify-sock': project.get_notify_sock(instance_name, run_dir),
    }


def get_stateboard_files(project, log_dir=None, data_dir=None, run_dir=None):
    run_dir = project.get_run_dir(run_dir)

    return {
        'log': project.get_sb_log_dir(log_dir),
        'workdir': project.get_sb_workdir(data_dir),
        'console-sock': project.get_sb_console_sock(run_dir),
        'pid': project.get_sb_pidfile(run_dir),
        'notify-sock': project.get_sb_notify_sock(run_dir),
    }


def create_instances_files(project, instance_names=[], stateboard=False, log_dir=None, data_dir=None, run_dir=None):
    all_files = [
        get_instance_files(project, instance_name, log_dir, data_dir, run_dir)
        for instance_name in instance_names
    ]

    if stateboard:
        all_files.append(get_stateboard_files(project, log_dir, data_dir, run_dir))

    for instance_files in all_files:
        for _, path in instance_files.items():
            if not os.path.exists(path):
                basepath = os.path.dirname(path)
                os.makedirs(basepath, exist_ok=True)
                with open(path, 'w') as f:
                    f.write('')


def assert_clean_logs(logs, project, instance_names=[], stateboard=False, exp_res='OK'):
    instance_ids = [project.get_instance_id(instance_name) for instance_name in instance_names]
    if stateboard:
        instance_ids.append(project.get_stateboard_id())

    assert len(logs) == len(instance_ids)
    assert all([log_line.endswith(exp_res) for log_line in logs])
    assert all([
        any([log_line.startswith(instance_id) for log_line in logs])
        for instance_id in instance_ids
    ])


def assert_files_cleaned(project, instance_names=[], stateboard=False, logs=None, exp_res_log='OK',
                         log_dir=None, data_dir=None, run_dir=None):
    run_dir = project.get_run_dir(run_dir)

    instance_ids = [project.get_instance_id(instance_name) for instance_name in instance_names]
    if stateboard:
        instance_ids.append(project.get_stateboard_id())

    if logs is not None:
        assert_clean_logs(logs, project, instance_names, stateboard, exp_res_log)

    all_files = [
        get_instance_files(project, instance_name, log_dir, data_dir, run_dir)
        for instance_name in instance_names
    ]

    if stateboard:
        all_files.append(get_stateboard_files(project, log_dir, data_dir, run_dir))

    for instance_files in all_files:
        for file_type, path in instance_files.items():
            if file_type in FILES_TO_BE_DELETED:
                assert not os.path.exists(path)
            else:
                assert os.path.exists(path)


def assert_files_exists(project, instance_names=[], stateboard=False,
                        log_dir=None, data_dir=None, run_dir=None):
    run_dir = project.get_run_dir(run_dir)

    all_files = [
        get_instance_files(project, instance_name, log_dir, data_dir, run_dir)
        for instance_name in instance_names
    ]

    if stateboard:
        all_files.append(get_stateboard_files(project, log_dir, data_dir, run_dir))

    for instance_files in all_files:
        for _, path in instance_files.items():
            assert os.path.exists(path)


def test_clean_by_name(start_stop_cli, project_without_dependencies):
    project = project_without_dependencies
    cli = start_stop_cli

    INSTANCE1 = 'instance-1'
    INSTANCE2 = 'instance-2'

    # two instances
    create_instances_files(project, [INSTANCE1, INSTANCE2], stateboard=True)
    logs = cli.clean(project, [INSTANCE1, INSTANCE2])
    assert_files_cleaned(project, [INSTANCE1, INSTANCE2], stateboard=True, logs=logs)
    assert_files_exists(project, [])

    # one instance w/ stateboard
    create_instances_files(project, [INSTANCE1, INSTANCE2], stateboard=True)
    logs = cli.clean(project, [INSTANCE1], stateboard=True)
    assert_files_cleaned(project, [INSTANCE1],  stateboard=True, logs=logs)
    assert_files_exists(project, [INSTANCE2])

    # one instance stateboard-only
    create_instances_files(project, [INSTANCE1, INSTANCE2], stateboard=True)
    logs = cli.clean(project, stateboard_only=True)
    assert_files_cleaned(project, [], stateboard=True, logs=logs)
    assert_files_exists(project, [INSTANCE1, INSTANCE2])


def test_clean_from_conf(start_stop_cli, project_without_dependencies):
    project = project_without_dependencies
    cli = start_stop_cli

    INSTANCE1 = 'instance-1'
    INSTANCE2 = 'instance-2'

    write_conf(project.get_cfg_path(), {
        project.get_instance_id(INSTANCE1): {},
        project.get_instance_id(INSTANCE2): {},
    })

    # w/ stateboard
    create_instances_files(project, [INSTANCE1, INSTANCE2], stateboard=True)
    logs = cli.clean(project, stateboard=True)
    assert_files_cleaned(project, [INSTANCE1, INSTANCE2], stateboard=True, logs=logs)
    assert_files_exists(project)

    # stateboard-only
    create_instances_files(project, [INSTANCE1, INSTANCE2], stateboard=True)
    logs = cli.clean(project, stateboard_only=True)
    assert_files_cleaned(project,  stateboard=True, logs=logs)
    assert_files_exists(project, [INSTANCE1, INSTANCE2])


def test_clean_cfg(start_stop_cli, project_without_dependencies):
    project = project_without_dependencies
    cli = start_stop_cli

    INSTANCE1 = 'instance-1'
    INSTANCE2 = 'instance-2'

    cfg = 'my-conf.yml'

    write_conf(os.path.join(project.path, cfg), {
        project.get_instance_id(INSTANCE1): {},
        project.get_instance_id(INSTANCE2): {},
    })

    # --cfg
    create_instances_files(project, [INSTANCE1, INSTANCE2], stateboard=True)
    logs = cli.clean(project, stateboard=True, cfg=cfg)
    assert_files_cleaned(project, [INSTANCE1, INSTANCE2], stateboard=True, logs=logs)


def test_clean_by_name_with_paths(start_stop_cli, project_without_dependencies):
    project = project_without_dependencies
    cli = start_stop_cli

    INSTANCE1 = 'instance-1'
    INSTANCE2 = 'instance-2'

    # --log-dir
    log_dir = 'my-log'
    create_instances_files(project, [INSTANCE1, INSTANCE2], log_dir=log_dir, stateboard=True)
    logs = cli.clean(project, [INSTANCE1], stateboard=True, log_dir=log_dir)
    assert_files_cleaned(project, [INSTANCE1], stateboard=True, log_dir=log_dir, logs=logs)
    assert_files_exists(project, [INSTANCE2], log_dir=log_dir)

    # --run-dir
    run_dir = 'my-run'
    create_instances_files(project, [INSTANCE1, INSTANCE2], run_dir=run_dir, stateboard=True)
    logs = cli.clean(project, [INSTANCE1], stateboard=True, run_dir=run_dir)
    assert_files_cleaned(project, [INSTANCE1],  stateboard=True, run_dir=run_dir, logs=logs)
    assert_files_exists(project, [INSTANCE2], run_dir=run_dir)

    # --data-dir
    data_dir = 'my-data'
    create_instances_files(project, [INSTANCE1, INSTANCE2], data_dir=data_dir, stateboard=True)
    logs = cli.clean(project, [INSTANCE1], stateboard=True, data_dir=data_dir)
    assert_files_cleaned(project, [INSTANCE1],  stateboard=True, data_dir=data_dir, logs=logs)
    assert_files_exists(project, [INSTANCE2], data_dir=data_dir)


def test_clean_by_name_with_paths_from_conf(start_stop_cli, project_without_dependencies):
    project = project_without_dependencies
    cli = start_stop_cli

    INSTANCE1 = 'instance-1'
    INSTANCE2 = 'instance-2'

    # --log-dir
    log_dir = 'my-log'
    write_conf(project.get_cli_cfg_path(), {
        'log-dir': log_dir,
    })
    create_instances_files(project, [INSTANCE1, INSTANCE2], stateboard=True, log_dir=log_dir)
    logs = cli.clean(project, [INSTANCE1], stateboard=True)
    assert_files_cleaned(project, [INSTANCE1], log_dir=log_dir, logs=logs)
    assert_files_exists(project, [INSTANCE2], log_dir=log_dir)

    # --run-dir
    run_dir = 'my-run'
    write_conf(project.get_cli_cfg_path(), {
        'run-dir': run_dir,
    })
    create_instances_files(project, [INSTANCE1, INSTANCE2], stateboard=True, run_dir=run_dir)
    logs = cli.clean(project, [INSTANCE1], stateboard=True)
    assert_files_cleaned(project, [INSTANCE1], run_dir=run_dir, logs=logs)
    assert_files_exists(project, [INSTANCE2], run_dir=run_dir)

    # --data-dir
    data_dir = 'my-data'
    write_conf(project.get_cli_cfg_path(), {
        'data-dir': data_dir,
    })
    create_instances_files(project, [INSTANCE1, INSTANCE2], stateboard=True, data_dir=data_dir)
    logs = cli.clean(project, [INSTANCE1], stateboard=True)
    assert_files_cleaned(project, [INSTANCE1], data_dir=data_dir, logs=logs)
    assert_files_exists(project, [INSTANCE2], data_dir=data_dir)


def test_skipped(start_stop_cli, project_without_dependencies):
    project = project_without_dependencies
    cli = start_stop_cli

    INSTANCE1 = 'instance-1'
    INSTANCE2 = 'instance-2'

    # clean once
    create_instances_files(project, [INSTANCE1, INSTANCE2], stateboard=True)
    logs = cli.clean(project, [INSTANCE1, INSTANCE2], stateboard=True)
    assert_files_cleaned(project, [INSTANCE1, INSTANCE2], stateboard=True, logs=logs)

    # clean again
    logs = cli.clean(project, [INSTANCE1, INSTANCE2], stateboard=True)
    assert_files_cleaned(project, [INSTANCE1, INSTANCE2], stateboard=True, logs=logs, exp_res_log='SKIPPED')


def test_for_running(start_stop_cli, project_without_dependencies):
    project = project_without_dependencies
    cli = start_stop_cli

    INSTANCE1 = 'instance-1'
    INSTANCE2 = 'instance-2'

    ID1 = project.get_instance_id(INSTANCE1)
    ID2 = project.get_instance_id(INSTANCE2)

    # start two instances
    cli.start(project, [INSTANCE1, INSTANCE2], daemonized=True)
    check_instances_running(cli, project, [INSTANCE1, INSTANCE2], stateboard=True, daemonized=True)

    # stop one
    cli.stop(project, [INSTANCE1])
    check_instances_stopped(cli, project, [INSTANCE1])
    check_instances_running(cli, project, [INSTANCE2], daemonized=True)

    # instance-1 is stopped, instance-2 is running
    logs = cli.clean(project, [INSTANCE1, INSTANCE2], exp_rc=1)
    assert_files_cleaned(project, [INSTANCE1])

    assert any([line.endswith('OK') and line.startswith(ID1) for line in logs])
    assert any([line.endswith('FAILED') and line.startswith(ID2) for line in logs])
    assert any(["%s: Instance is running" % ID2 in line for line in logs])
