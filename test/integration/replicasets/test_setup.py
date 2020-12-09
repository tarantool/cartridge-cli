import os
import subprocess
import yaml

from utils import run_command_and_get_output
from utils import get_log_lines
from utils import DEFAULT_RPL_CFG
from utils import write_conf
from utils import get_replicasets
from utils import is_vshard_bootstrapped

from integration.replicasets.utils import get_list_from_log_lines

from project import patch_cartridge_proc_titile


def assert_setup_logs(output, rpl_conf_path, created_rpls=[], updated_rpls=[], ok_rpls=[], vshard_bootstrapped=False):
    log_lines = get_log_lines(output)
    assert log_lines[:1] == [
        "• Set up replicasets described in %s" % rpl_conf_path
    ]

    set_replicasets_logs = log_lines
    if vshard_bootstrapped:
        bootstrap_vshard_logs = log_lines[-1:]
        assert bootstrap_vshard_logs == [
            "• Vshard is bootstrapped successfully",
        ]

        set_replicasets_logs = log_lines[:-1]

    replicasets_list = get_list_from_log_lines(set_replicasets_logs[1:-1])

    exp_replicasets_list = []
    exp_replicasets_list.extend(['%s... CREATED' % rpl_name for rpl_name in created_rpls])
    exp_replicasets_list.extend(['%s... UPDATED' % rpl_name for rpl_name in updated_rpls])
    exp_replicasets_list.extend(['%s... OK' % rpl_name for rpl_name in ok_rpls])
    assert set(replicasets_list) == set(exp_replicasets_list)

    assert set_replicasets_logs[-1:] == [
        "• Replicasets are set up successfully",
    ]


def assert_replicasets(rpl_cfg, admin_api_url):
    replicasets = get_replicasets(admin_api_url)
    replicasets = {r['alias']: r for r in replicasets}

    assert replicasets.keys() == rpl_cfg.keys()

    for rpl_alias, configured_rpl in rpl_cfg.items():
        rpl = replicasets[rpl_alias]

        assert set(rpl['roles']) == set(configured_rpl['roles'])

        rpl_instances = [i['alias'] for i in rpl['servers']]
        assert rpl_instances == configured_rpl['instances']

        assert rpl.get('weight') == configured_rpl.get('weight')
        assert rpl.get('vshard_group') == configured_rpl.get('vshard_group')

        if 'all_rw' in configured_rpl:
            assert rpl.get('all_rw') == configured_rpl.get('all_rw')


def test_default_application(cartridge_cmd, start_stop_cli, project_with_cartridge):
    cli = start_stop_cli
    project = project_with_cartridge

    # build project
    cmd = [
        cartridge_cmd,
        "build",
    ]
    process = subprocess.run(cmd, cwd=project.path)
    assert process.returncode == 0, "Error during building the project"

    # don't change process title
    patch_cartridge_proc_titile(project)

    # start instances
    cli.start(project, daemonized=True)

    # setup replicasets
    cmd = [
        cartridge_cmd, 'replicasets', 'setup',
        '--bootstrap-vshard',
    ]

    rc, output = run_command_and_get_output(cmd, cwd=project.path)
    assert rc == 0

    admin_api_url = 'http://localhost:%s/admin/api' % '8081'

    rpl_cfg_path = os.path.join(project.path, DEFAULT_RPL_CFG)
    with open(rpl_cfg_path) as f:
        rpl_cfg = yaml.load(f, Loader=yaml.FullLoader)

    assert_replicasets(rpl_cfg, admin_api_url)
    assert is_vshard_bootstrapped(admin_api_url)

    assert_setup_logs(output, rpl_cfg_path, created_rpls=['router', 's-1', 's-2'], vshard_bootstrapped=True)

    # check that save without topology changing produces exactly the same file
    with open(rpl_cfg_path) as f:
        old_rpl_cfg_content = f.read()

    os.remove(rpl_cfg_path)

    cmd = [
        cartridge_cmd, 'replicasets', 'save',
    ]

    rc, output = run_command_and_get_output(cmd, cwd=project.path)
    assert rc == 0

    assert os.path.exists(rpl_cfg_path)
    with open(rpl_cfg_path) as f:
        new_rpl_cfg_content = f.read()

    assert new_rpl_cfg_content == old_rpl_cfg_content


def test_setup(project_with_instances, cartridge_cmd):
    project = project_with_instances.project
    instances = project_with_instances.instances

    router = instances['router']
    s1_master = instances['s1-master']
    s1_replica = instances['s1-replica']
    s1_replica2 = instances['s1-replica-2']

    admin_api_url = router.get_admin_api_url()

    rpl_cfg_path = os.path.join(project.path, DEFAULT_RPL_CFG)

    # create router replicaset
    rpl_cfg = {
        'router': {
            'roles': ['vshard-router', 'app.roles.custom', 'failover-coordinator'],
            'instances': [router.name],
        }
    }

    write_conf(rpl_cfg_path, rpl_cfg)

    cmd = [
        cartridge_cmd, 'replicasets', 'setup',
    ]

    rc, output = run_command_and_get_output(cmd, cwd=project.path)
    assert rc == 0

    assert_replicasets(rpl_cfg, admin_api_url)
    assert_setup_logs(output, rpl_cfg_path, created_rpls=['router'])

    # create s-1 replicaset
    rpl_cfg.update({
        's-1': {
            'roles': ['vshard-storage'],
            'instances': [s1_master.name, s1_replica.name],
            'weight': 1.234,
            'vshard_group': 'hot',
        }
    })

    write_conf(rpl_cfg_path, rpl_cfg)

    cmd = [
        cartridge_cmd, 'replicasets', 'setup',
    ]

    rc, output = run_command_and_get_output(cmd, cwd=project.path)
    assert rc == 0

    assert_replicasets(rpl_cfg, admin_api_url)
    assert_setup_logs(output, rpl_cfg_path, created_rpls=['s-1'], ok_rpls=['router'])

    # add one more instance to s-1
    rpl_cfg['s-1']['instances'].append(s1_replica2.name)
    write_conf(rpl_cfg_path, rpl_cfg)

    cmd = [
        cartridge_cmd, 'replicasets', 'setup',
    ]

    rc, output = run_command_and_get_output(cmd, cwd=project.path)
    assert rc == 0

    assert_replicasets(rpl_cfg, admin_api_url)
    assert_setup_logs(output, rpl_cfg_path, updated_rpls=['s-1'], ok_rpls=['router'])

    # change s-1 failover priority
    rpl_cfg['s-1']['instances'] = [
        s1_replica2.name, s1_replica.name, s1_master.name,
    ]
    write_conf(rpl_cfg_path, rpl_cfg)

    cmd = [
        cartridge_cmd, 'replicasets', 'setup',
    ]

    rc, output = run_command_and_get_output(cmd, cwd=project.path)
    assert rc == 0

    assert_replicasets(rpl_cfg, admin_api_url)
    assert_setup_logs(output, rpl_cfg_path, updated_rpls=['s-1'], ok_rpls=['router'])

    # set s-1 all_rw
    rpl_cfg['s-1']['all_rw'] = True
    write_conf(rpl_cfg_path, rpl_cfg)

    cmd = [
        cartridge_cmd, 'replicasets', 'setup',
    ]

    rc, output = run_command_and_get_output(cmd, cwd=project.path)
    assert rc == 0

    assert_replicasets(rpl_cfg, admin_api_url)
    assert_setup_logs(output, rpl_cfg_path, updated_rpls=['s-1'], ok_rpls=['router'])

    # change s-1 weight
    rpl_cfg['s-1']['weight'] = 2.345
    write_conf(rpl_cfg_path, rpl_cfg)

    cmd = [
        cartridge_cmd, 'replicasets', 'setup',
    ]

    rc, output = run_command_and_get_output(cmd, cwd=project.path)
    assert rc == 0

    assert_replicasets(rpl_cfg, admin_api_url)
    assert_setup_logs(output, rpl_cfg_path, updated_rpls=['s-1'], ok_rpls=['router'])


def test_setup_bootstrap_vshard(project_with_instances, cartridge_cmd):
    project = project_with_instances.project
    instances = project_with_instances.instances

    router = instances['router']
    s1_master = instances['s1-master']
    s1_replica = instances['s1-replica']
    s2_master = instances['s2-master']

    admin_api_url = router.get_admin_api_url()

    rpl_cfg_path = os.path.join(project.path, DEFAULT_RPL_CFG)

    # create router replicaset
    # vshard bootstrapping will fail
    rpl_cfg = {
        'router': {
            'roles': ['vshard-router', 'app.roles.custom', 'failover-coordinator'],
            'instances': [router.name],
        }
    }

    write_conf(rpl_cfg_path, rpl_cfg)

    cmd = [
        cartridge_cmd, 'replicasets', 'setup',
        '--bootstrap-vshard'
    ]

    rc, output = run_command_and_get_output(cmd, cwd=project.path)
    assert rc == 1

    assert "router... CREATED" in output
    assert "Bootstrapping vshard failed: Sharding config is empty" in output

    assert_replicasets(rpl_cfg, admin_api_url)
    assert not is_vshard_bootstrapped(admin_api_url)

    # create s-1 and s-2 replicasets
    rpl_cfg.update({
        's-1': {
            'roles': ['vshard-storage'],
            'instances': [s1_master.name, s1_replica.name],
            'weight': 1.234,
            'vshard_group': 'hot',
        },
        's-2': {
            'roles': ['vshard-storage'],
            'instances': [s2_master.name],
            'weight': 1.234,
            'vshard_group': 'cold',
        },
    })

    write_conf(rpl_cfg_path, rpl_cfg)

    cmd = [
        cartridge_cmd, 'replicasets', 'setup',
        '--bootstrap-vshard'
    ]

    rc, output = run_command_and_get_output(cmd, cwd=project.path)
    assert rc == 0

    assert_replicasets(rpl_cfg, admin_api_url)
    assert_setup_logs(output, rpl_cfg_path, created_rpls=['s-1', 's-2'], ok_rpls=['router'], vshard_bootstrapped=True)


def test_save(project_with_vshard_replicasets, cartridge_cmd):
    project = project_with_vshard_replicasets.project
    instances = project_with_vshard_replicasets.instances

    router = instances['router']
    admin_api_url = router.get_admin_api_url()

    rpl_cfg_path = os.path.join(project.path, DEFAULT_RPL_CFG)

    cmd = [
        cartridge_cmd, 'replicasets', 'save',
    ]

    rc, output = run_command_and_get_output(cmd, cwd=project.path)
    assert rc == 0
    assert get_log_lines(output) == [
        "• Save current replicasets to %s" % rpl_cfg_path,
    ]

    assert os.path.exists(rpl_cfg_path)

    with open(rpl_cfg_path) as f:
        rpl_cfg = yaml.load(f, Loader=yaml.FullLoader)

    assert_replicasets(rpl_cfg, admin_api_url)


def test_setup_file_not_exists(project_with_vshard_replicasets, cartridge_cmd):
    project = project_with_vshard_replicasets.project

    cmd = [
        cartridge_cmd, 'replicasets', 'setup',
        '--file', 'non-existent-file',
    ]

    rc, output = run_command_and_get_output(cmd, cwd=project.path)
    assert rc == 1

    assert "Failed to use replicasets configuration file:" in output
    assert "no such file or directory" in output


def test_bad_rpl_conf_format(project_with_instances, cartridge_cmd):
    project = project_with_instances.project

    rpl_cfg_path = os.path.join(project.path, DEFAULT_RPL_CFG)

    # create router replicaset
    rpl_cfg = {
        'router': {
            'roles': 'vshard-router',  # should be a list
            'instances': [],
        }
    }

    write_conf(rpl_cfg_path, rpl_cfg)

    cmd = [
        cartridge_cmd, 'replicasets', 'setup',
    ]

    rc, output = run_command_and_get_output(cmd, cwd=project.path)
    assert rc == 1


def test_setup_file_specified(project_with_instances, cartridge_cmd):
    project = project_with_instances.project
    instances = project_with_instances.instances

    router = instances['router']

    admin_api_url = router.get_admin_api_url()

    rpl_cfg_path = os.path.join(project.path, 'my-replicasets.yml')

    # create router replicaset
    rpl_cfg = {
        'router': {
            'roles': ['vshard-router', 'app.roles.custom', 'failover-coordinator'],
            'instances': [router.name],
        }
    }

    write_conf(rpl_cfg_path, rpl_cfg)

    cmd = [
        cartridge_cmd, 'replicasets', 'setup',
        '--file', rpl_cfg_path,
    ]

    rc, output = run_command_and_get_output(cmd, cwd=project.path)
    assert rc == 0

    assert_replicasets(rpl_cfg, admin_api_url)
    assert_setup_logs(output, rpl_cfg_path, created_rpls=['router'])


def test_save_file_specified(project_with_vshard_replicasets, cartridge_cmd):
    project = project_with_vshard_replicasets.project
    instances = project_with_vshard_replicasets.instances

    router = instances['router']
    admin_api_url = router.get_admin_api_url()

    rpl_cfg_path = os.path.join(project.path, 'my-replicasets.yml')

    cmd = [
        cartridge_cmd, 'replicasets', 'save',
        '--file', rpl_cfg_path,
    ]

    rc, output = run_command_and_get_output(cmd, cwd=project.path)
    assert rc == 0
    assert get_log_lines(output) == [
        "• Save current replicasets to %s" % rpl_cfg_path,
    ]

    assert os.path.exists(rpl_cfg_path)

    with open(rpl_cfg_path) as f:
        rpl_cfg = yaml.load(f, Loader=yaml.FullLoader)

    assert_replicasets(rpl_cfg, admin_api_url)
