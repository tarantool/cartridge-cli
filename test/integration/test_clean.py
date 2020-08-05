import os

from utils import get_instance_id, get_stateboard_name
from utils import DEFAULT_CFG, DEFAULT_DATA_DIR, DEFAULT_LOG_DIR, DEFAULT_RUN_DIR
from utils import write_conf
from utils import check_instances_running
from utils import check_instances_stopped


CARTRIDGE_CONF = '.cartridge.yml'

FILES_TO_BE_DELETED = ['log', 'workdir', 'console-sock', 'notify-sock', 'pid']


def get_instance_files(project, instance_id, log_dir=DEFAULT_LOG_DIR, data_dir=DEFAULT_DATA_DIR,
                       run_dir=DEFAULT_RUN_DIR):
    return {
        'log': os.path.join(project.path, log_dir, '%s.log' % instance_id),
        'workdir': os.path.join(project.path, data_dir, instance_id),
        'console-sock': os.path.join(project.path, run_dir, '%s.control' % instance_id),
        'pid': os.path.join(project.path, run_dir, '%s.pid' % instance_id),
        'notify-sock': os.path.join(project.path, run_dir, '%s.notify' % instance_id),
    }


def create_instances_files(project, instance_names=[],
                           log_dir=DEFAULT_LOG_DIR, data_dir=DEFAULT_DATA_DIR, run_dir=DEFAULT_RUN_DIR):
    instance_ids = [get_instance_id(project.name, instance_name) for instance_name in instance_names]
    instance_ids.append(get_stateboard_name(project.name))

    for instance_id in instance_ids:
        instance_files = get_instance_files(project, instance_id, log_dir, data_dir, run_dir)

        for _, path in instance_files.items():
            if not os.path.exists(path):
                basepath = os.path.dirname(path)
                os.makedirs(basepath, exist_ok=True)
                with open(path, 'w') as f:
                    f.write('')


def assert_clean_logs(logs, project, instance_names=[], stateboard=False, exp_res='OK'):
    instance_ids = [get_instance_id(project.name, instance_name) for instance_name in instance_names]
    if stateboard:
        instance_ids.append(get_stateboard_name(project.name))

    assert len(logs) == len(instance_ids)
    assert all([log_line.endswith(exp_res) for log_line in logs])
    assert all([
        any([log_line.startswith(instance_id) for log_line in logs])
        for instance_id in instance_ids
    ])


def assert_files_cleaned(project, instance_names=[], stateboard=False, logs=None, exp_res_log='OK',
                         log_dir=DEFAULT_LOG_DIR, data_dir=DEFAULT_DATA_DIR, run_dir=DEFAULT_RUN_DIR):
    instance_ids = [get_instance_id(project.name, instance_name) for instance_name in instance_names]
    if stateboard:
        instance_ids.append(get_stateboard_name(project.name))

    if logs is not None:
        assert_clean_logs(logs, project, instance_names, stateboard, exp_res_log)

    for instance_id in instance_ids:
        instance_files = get_instance_files(project, instance_id, log_dir, data_dir, run_dir)

        for file_type, path in instance_files.items():
            if file_type in FILES_TO_BE_DELETED:
                assert not os.path.exists(path)
            else:
                assert os.path.exists(path)


def assert_files_exists(project, instance_names=[], stateboard=False,
                        log_dir=DEFAULT_LOG_DIR, data_dir=DEFAULT_DATA_DIR, run_dir=DEFAULT_RUN_DIR):
    instance_ids = [get_instance_id(project.name, instance_name) for instance_name in instance_names]
    if stateboard:
        instance_ids.append(get_stateboard_name(project.name))

    for instance_id in instance_ids:
        instance_files = get_instance_files(project, instance_id, log_dir, data_dir, run_dir)

        for _, path in instance_files.items():
            assert os.path.exists(path)


def test_clean_by_name(start_stop_cli, project_with_patched_init):
    project = project_with_patched_init
    cli = start_stop_cli

    INSTANCE1 = 'instance-1'
    INSTANCE2 = 'instance-2'

    # two instances
    create_instances_files(project, [INSTANCE1, INSTANCE2])
    logs = cli.clean(project, [INSTANCE1, INSTANCE2])
    assert_files_cleaned(project, [INSTANCE1, INSTANCE2], logs=logs)
    assert_files_exists(project, [], stateboard=True)

    # one instance w/ stateboard
    create_instances_files(project, [INSTANCE1, INSTANCE2])
    logs = cli.clean(project, [INSTANCE1], stateboard=True)
    assert_files_cleaned(project, [INSTANCE1],  stateboard=True, logs=logs)
    assert_files_exists(project, [INSTANCE2])

    # one instance stateboard-only
    create_instances_files(project, [INSTANCE1, INSTANCE2])
    logs = cli.clean(project, stateboard_only=True)
    assert_files_cleaned(project, [],  stateboard=True, logs=logs)
    assert_files_exists(project, [INSTANCE1, INSTANCE2])


def test_clean_from_conf(start_stop_cli, project_with_patched_init):
    project = project_with_patched_init
    cli = start_stop_cli

    INSTANCE1 = 'instance-1'
    INSTANCE2 = 'instance-2'

    write_conf(os.path.join(project.path, DEFAULT_CFG), {
        get_instance_id(project.name, INSTANCE1): {},
        get_instance_id(project.name, INSTANCE2): {},
    })

    # only instances
    create_instances_files(project, [INSTANCE1, INSTANCE2])
    logs = cli.clean(project)
    assert_files_cleaned(project, [INSTANCE1, INSTANCE2], logs=logs)
    assert_files_exists(project, stateboard=True)

    # w/ stateboard
    create_instances_files(project, [INSTANCE1, INSTANCE2])
    logs = cli.clean(project, stateboard=True)
    assert_files_cleaned(project, [INSTANCE1, INSTANCE2], stateboard=True, logs=logs)

    # stateboard-only
    create_instances_files(project, [INSTANCE1, INSTANCE2])
    logs = cli.clean(project, stateboard_only=True)
    assert_files_cleaned(project,  stateboard=True, logs=logs)
    assert_files_exists(project, [INSTANCE1, INSTANCE2])


def test_clean_cfg(start_stop_cli, project_with_patched_init):
    project = project_with_patched_init
    cli = start_stop_cli

    INSTANCE1 = 'instance-1'
    INSTANCE2 = 'instance-2'

    cfg = 'my-conf.yml'

    write_conf(os.path.join(project.path, cfg), {
        get_instance_id(project.name, INSTANCE1): {},
        get_instance_id(project.name, INSTANCE2): {},
    })

    # --cfg
    create_instances_files(project, [INSTANCE1, INSTANCE2])
    logs = cli.clean(project, stateboard=True, cfg=cfg)
    assert_files_cleaned(project, [INSTANCE1, INSTANCE2], stateboard=True, logs=logs)


def test_clean_by_name_with_paths(start_stop_cli, project_with_patched_init):
    project = project_with_patched_init
    cli = start_stop_cli

    INSTANCE1 = 'instance-1'
    INSTANCE2 = 'instance-2'

    # --log-dir
    log_dir = 'my-log'
    create_instances_files(project, [INSTANCE1, INSTANCE2], log_dir=log_dir)
    logs = cli.clean(project, [INSTANCE1], stateboard=True, log_dir=log_dir)
    assert_files_cleaned(project, [INSTANCE1],  stateboard=True, log_dir=log_dir, logs=logs)
    assert_files_exists(project, [INSTANCE2], log_dir=log_dir)

    # --run-dir
    run_dir = 'my-run'
    create_instances_files(project, [INSTANCE1, INSTANCE2], run_dir=run_dir)
    logs = cli.clean(project, [INSTANCE1], stateboard=True, run_dir=run_dir)
    assert_files_cleaned(project, [INSTANCE1],  stateboard=True, run_dir=run_dir, logs=logs)
    assert_files_exists(project, [INSTANCE2], run_dir=run_dir)

    # --data-dir
    data_dir = 'my-data'
    create_instances_files(project, [INSTANCE1, INSTANCE2], data_dir=data_dir)
    logs = cli.clean(project, [INSTANCE1], stateboard=True, data_dir=data_dir)
    assert_files_cleaned(project, [INSTANCE1],  stateboard=True, data_dir=data_dir, logs=logs)
    assert_files_exists(project, [INSTANCE2], data_dir=data_dir)


def test_clean_by_name_with_paths_from_conf(start_stop_cli, project_with_patched_init):
    project = project_with_patched_init
    cli = start_stop_cli

    INSTANCE1 = 'instance-1'
    INSTANCE2 = 'instance-2'

    # --log-dir
    log_dir = 'my-log'
    write_conf(os.path.join(project.path, CARTRIDGE_CONF), {
        'log-dir': log_dir,
    })
    create_instances_files(project, [INSTANCE1, INSTANCE2], log_dir=log_dir)
    logs = cli.clean(project, [INSTANCE1], stateboard=True)
    assert_files_cleaned(project, [INSTANCE1],  stateboard=True, log_dir=log_dir, logs=logs)
    assert_files_exists(project, [INSTANCE2], log_dir=log_dir)

    # --run-dir
    run_dir = 'my-run'
    write_conf(os.path.join(project.path, CARTRIDGE_CONF), {
        'run-dir': run_dir,
    })
    create_instances_files(project, [INSTANCE1, INSTANCE2], run_dir=run_dir)
    logs = cli.clean(project, [INSTANCE1], stateboard=True)
    assert_files_cleaned(project, [INSTANCE1],  stateboard=True, run_dir=run_dir, logs=logs)
    assert_files_exists(project, [INSTANCE2], run_dir=run_dir)

    # --data-dir
    data_dir = 'my-data'
    write_conf(os.path.join(project.path, CARTRIDGE_CONF), {
        'data-dir': data_dir,
    })
    create_instances_files(project, [INSTANCE1, INSTANCE2], data_dir=data_dir)
    logs = cli.clean(project, [INSTANCE1], stateboard=True)
    assert_files_cleaned(project, [INSTANCE1],  stateboard=True, data_dir=data_dir, logs=logs)
    assert_files_exists(project, [INSTANCE2], data_dir=data_dir)


def test_skipped(start_stop_cli, project_with_patched_init):
    project = project_with_patched_init
    cli = start_stop_cli

    INSTANCE1 = 'instance-1'
    INSTANCE2 = 'instance-2'

    # clean once
    create_instances_files(project, [INSTANCE1, INSTANCE2])
    logs = cli.clean(project, [INSTANCE1, INSTANCE2])
    assert_files_cleaned(project, [INSTANCE1, INSTANCE2], logs=logs)
    assert_files_exists(project, [], stateboard=True)

    # clean again
    logs = cli.clean(project, [INSTANCE1, INSTANCE2])
    assert_files_cleaned(project, [INSTANCE1, INSTANCE2], logs=logs, exp_res_log='SKIPPED')
    assert_files_exists(project, [], stateboard=True)


def test_for_running(start_stop_cli, project_with_patched_init):
    project = project_with_patched_init
    cli = start_stop_cli

    INSTANCE1 = 'instance-1'
    INSTANCE2 = 'instance-2'

    ID1 = get_instance_id(project.name, INSTANCE1)
    ID2 = get_instance_id(project.name, INSTANCE2)

    # start two instances
    cli.start(project, [INSTANCE1, INSTANCE2], daemonized=True)
    check_instances_running(cli, project, [INSTANCE1, INSTANCE2], daemonized=True)

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
