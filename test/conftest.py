import py
import pytest
import tempfile
import docker
import os
import subprocess
import platform
import shutil

from project import Project
from project import remove_dependency
from project import add_dependency_submodule
from project import remove_all_dependencies
from project import replace_project_file
from project import patch_cartridge_proc_titile

from clusterwide_conf import ClusterwideConfig
from clusterwide_conf import get_srv_conf, get_expelled_srv_conf
from clusterwide_conf import get_rpl_conf
from clusterwide_conf import get_topology_conf, get_one_file_conf

from utils import Cli
from utils import start_instances
from utils import DEFAULT_RUN_DIR

from project import INIT_NO_CARTRIDGE_FILEPATH
from project import INIT_IGNORE_SIGTERM_FILEPATH
from project import INIT_ADMIN_FUNCS_FILEPATH


# ########
# Fixtures
# ########

def get_tmpdir(request):
    tmpdir = py.path.local(tempfile.mkdtemp())
    request.addfinalizer(lambda: tmpdir.remove(rec=1))
    return str(tmpdir)


@pytest.fixture(scope='module')
def module_tmpdir(request):
    return get_tmpdir(request)


@pytest.fixture(scope='session')
def session_tmpdir(request):
    return get_tmpdir(request)


def get_short_tmpdir(request):
    tmpbase = '/tmp'
    if platform.system() == 'Darwin':
        tmpbase = '/private/tmp'

    tmpdir = py.path.local(tempfile.mkdtemp(dir=tmpbase))
    request.addfinalizer(lambda: tmpdir.remove(rec=1))
    return str(tmpdir)


@pytest.fixture(scope='function')
def short_tmpdir(request):
    return get_short_tmpdir(request)


@pytest.fixture(scope='session')
def short_session_tmpdir(request):
    return get_short_tmpdir(request)


@pytest.fixture(scope="session")
def docker_client():
    client = docker.from_env()
    return client


@pytest.fixture(scope="session")
def cartridge_cmd(request, session_tmpdir):
    cli_base_path = os.path.realpath(os.path.join(os.path.dirname(__file__), '..'))
    cli_path = os.path.join(session_tmpdir, 'cartridge')

    cli_build_cmd = ['mage', '-v', 'build']

    build_env = os.environ.copy()
    build_env["CLIEXE"] = cli_path

    process = subprocess.run(cli_build_cmd, cwd=cli_base_path, env=build_env)
    assert process.returncode == 0, 'Failed to build cartridge-cli executable'

    return cli_path


@pytest.fixture(scope="function")
def start_stop_cli(cartridge_cmd, request):
    cli = Cli(cartridge_cmd)
    request.addfinalizer(lambda: cli.terminate())
    return cli


# ################
# Project fixtures
# ################
# There are three main types of projects:
# * light_project:
#   Cartridge CLI creates project with cartridge dependency by default.
#   It's known that installing cartridge rocks is a long operation,
#   so we don't want to perform it on every test.
#   These fixtures are used to decrease packing time.
#   They don't have a cartridge dependency,
#   but have dependency installed from submodule
#   (see add_dependency_submodule function for details)
#   In fact, we need to install cartridge dependency only
#   for e2e
#
# * project_with_cartridge:
#   This is a project with cartridge dependency installed.
#   Is used in `docker pack` tests. Test image is built once and then
#   it's used in all docker tests include e2e.
#   This project also have submodule dependency (see add_dependency_submodule)
#   to test pre and post build hooks
#
# * project_without_dependencies:
#   This is the empty project without dependencies.
#   It is used for error behavior tests and tests where
#   result package content doesn't matter
#
###############
# Light project
###############
@pytest.fixture(scope="function")
def light_project(cartridge_cmd, tmpdir):
    project = Project(cartridge_cmd, 'light-project', tmpdir, 'cartridge')

    remove_dependency(project, 'cartridge')

    add_dependency_submodule(project)

    return project


########################
# Project with cartridge
########################
@pytest.fixture(scope="function")
def project_with_cartridge(cartridge_cmd, short_tmpdir):
    project = Project(cartridge_cmd, 'project-with-cartridge', short_tmpdir, 'cartridge')

    add_dependency_submodule(project)

    return project


##############################
# Project without dependencies
##############################
# This project is used in the `pack` and `running` tests
# It allows to build project faster and start an application
# that entering event loop and sends READY=q to notify socket
@pytest.fixture(scope="function")
def project_without_dependencies(cartridge_cmd, short_tmpdir):
    project = Project(cartridge_cmd, 'empty-project', short_tmpdir, 'cartridge')

    remove_all_dependencies(project)

    replace_project_file(project, 'init.lua', INIT_NO_CARTRIDGE_FILEPATH)
    replace_project_file(project, 'stateboard.init.lua', INIT_NO_CARTRIDGE_FILEPATH)

    return project


#############################
# Project getting-started-app
#############################
@pytest.fixture(scope="function")
def project_getting_started(cartridge_cmd, short_tmpdir):
    getting_started_path = os.path.realpath(
        os.path.join(os.path.dirname(__file__), '..', 'examples/getting-started-app'),
    )
    name = 'getting-started-app'

    def create_getting_started_app(basepath):
        path = os.path.join(basepath, name)
        shutil.copytree(getting_started_path, path, ignore=shutil.ignore_patterns(".rocks", "tmp"))
        return path

    project = Project(
        cartridge_cmd, name, short_tmpdir, 'cartridge',
        create_func=create_getting_started_app
    )

    return project


##############################
# Project that ignores SIGTERM
##############################
# This project is used in the `running` tests
# to check that `cartridge stop` sends SIGTERM,
# but `cartridge stop -f` sends SIGKILL
@pytest.fixture(scope="function")
def project_ignore_sigterm(cartridge_cmd, short_tmpdir):
    project = Project(cartridge_cmd, 'ignore-sigterm', short_tmpdir, 'cartridge')

    remove_all_dependencies(project)

    replace_project_file(project, 'init.lua', INIT_IGNORE_SIGTERM_FILEPATH)
    replace_project_file(project, 'stateboard.init.lua', INIT_IGNORE_SIGTERM_FILEPATH)

    return project


######################
# Custom admin project
######################
# This project is used to test `cartridge admin`
# with different commands added by user
@pytest.fixture(scope="function")
def custom_admin_project(cartridge_cmd, short_tmpdir):
    project = Project(cartridge_cmd, 'admin-project', short_tmpdir, 'cartridge')

    remove_dependency(project, 'cartridge')

    replace_project_file(project, 'init.lua', INIT_ADMIN_FUNCS_FILEPATH)

    return project


########################################
# Custom admin project running instances
########################################
# Running instances of custom_admin_project
@pytest.fixture(scope="function")
def custom_admin_running_instances(cartridge_cmd, start_stop_cli, custom_admin_project):
    project = custom_admin_project

    # build project
    cmd = [
        cartridge_cmd,
        "build",
    ]
    process = subprocess.run(cmd, cwd=project.path)
    assert process.returncode == 0, "Error during building the project"

    start_instances(cartridge_cmd, start_stop_cli, project)

    return {
        'project': project,
        'run_dir': os.path.join(project.path, DEFAULT_RUN_DIR),
    }


# ###########################
# Clusterwide config fixtures
# ###########################
@pytest.fixture(scope="function")
def clusterwide_conf_non_existent_instance():
    REPLICASET_UUID = 'rpl-1'
    NON_EXISTENT_INSTANCE_UUID = 'srv-non-existent'

    conf = get_topology_conf(
        instances=[get_srv_conf('srv-1', rpl_uuid=REPLICASET_UUID)],
        replicasets=[get_rpl_conf(REPLICASET_UUID, leaders=['srv-1'])]
    )

    return ClusterwideConfig(conf, instance_uuid=NON_EXISTENT_INSTANCE_UUID,
                             replicaset_uuid=REPLICASET_UUID)


@pytest.fixture(scope="function")
def clusterwide_conf_simple():
    INSTANCE_UUID = 'srv-3'
    INSTANCE_URI = 'srv-3:3303'
    REPLICASET_UUID = 'rpl-1'

    conf = get_topology_conf(
        instances=[
            get_srv_conf('srv-1', rpl_uuid=REPLICASET_UUID),
            get_srv_conf('srv-2', rpl_uuid=REPLICASET_UUID),
            get_srv_conf(INSTANCE_UUID, uri=INSTANCE_URI, rpl_uuid=REPLICASET_UUID),
            get_srv_conf('srv-4', rpl_uuid='rpl-2'),
        ],
        replicasets=[
            get_rpl_conf(REPLICASET_UUID, leaders=[
                'srv-1', 'srv-2', INSTANCE_UUID,
            ]),
            get_rpl_conf('rpl-2', leaders=['srv-4']),
        ]
    )

    return ClusterwideConfig(conf, instance_uuid=INSTANCE_UUID,
                             instance_uri=INSTANCE_URI,
                             replicaset_uuid=REPLICASET_UUID)


@pytest.fixture(scope="function")
def clusterwide_conf_one_file():
    INSTANCE_UUID = 'srv-3'
    INSTANCE_URI = 'srv-3:3303'
    REPLICASET_UUID = 'rpl-1'

    conf = get_one_file_conf(
        instances=[
            get_srv_conf('srv-1', rpl_uuid=REPLICASET_UUID),
            get_srv_conf('srv-2', rpl_uuid=REPLICASET_UUID),
            get_srv_conf(INSTANCE_UUID, uri=INSTANCE_URI, rpl_uuid=REPLICASET_UUID),
            get_srv_conf('srv-4', rpl_uuid='rpl-2'),
        ],
        replicasets=[
            get_rpl_conf(REPLICASET_UUID, leaders=[
                'srv-1', 'srv-2', INSTANCE_UUID,
            ]),
            get_rpl_conf('rpl-2', leaders=['srv-4']),
        ]
    )

    return ClusterwideConfig(conf, instance_uuid=INSTANCE_UUID,
                             instance_uri=INSTANCE_URI,
                             replicaset_uuid=REPLICASET_UUID,
                             one_file=True)


@pytest.fixture(scope="function")
def clusterwide_conf_other_leader_is_string():
    INSTANCE_UUID = 'srv-3'
    REPLICASET_UUID = 'rpl-1'

    conf = get_topology_conf(
        instances=[
            get_srv_conf('srv-1', rpl_uuid=REPLICASET_UUID),
            get_srv_conf('srv-2', rpl_uuid=REPLICASET_UUID),
            get_srv_conf(INSTANCE_UUID, rpl_uuid=REPLICASET_UUID),
            get_srv_conf('srv-4', rpl_uuid='rpl-2'),
        ],
        replicasets=[
            get_rpl_conf(REPLICASET_UUID, leaders='srv-1'),
            get_rpl_conf('rpl-2', leaders='srv-4'),
        ]
    )

    return ClusterwideConfig(conf, instance_uuid=INSTANCE_UUID,
                             replicaset_uuid=REPLICASET_UUID)


@pytest.fixture(scope="function")
def clusterwide_conf_current_leader_is_string():
    INSTANCE_UUID = 'srv-current-leader'
    REPLICASET_UUID = 'rpl-1'

    conf = get_topology_conf(
        instances=[
            get_srv_conf('srv-1', rpl_uuid=REPLICASET_UUID),
            get_srv_conf('srv-2', rpl_uuid=REPLICASET_UUID),
            get_srv_conf(INSTANCE_UUID, rpl_uuid=REPLICASET_UUID),
            get_srv_conf('srv-4', rpl_uuid='rpl-2'),
        ],
        replicasets=[
            get_rpl_conf(REPLICASET_UUID, leaders=INSTANCE_UUID),
            get_rpl_conf('rpl-2', leaders='srv-4'),
        ]
    )

    return ClusterwideConfig(conf, instance_uuid=INSTANCE_UUID,
                             replicaset_uuid=REPLICASET_UUID)


@pytest.fixture(scope="function")
def clusterwide_conf_srv_disabled():
    DISABLED_INSTANCE_UUID = 'srv-disabled'
    REPLICASET_UUID = 'rpl-1'
    INSTANCE_URI = 'srv-disabled:3303'

    conf = get_topology_conf(
        instances=[
            get_srv_conf('srv-1', rpl_uuid=REPLICASET_UUID),
            get_srv_conf('srv-2', rpl_uuid=REPLICASET_UUID),
            get_srv_conf(
                DISABLED_INSTANCE_UUID, uri=INSTANCE_URI,
                rpl_uuid=REPLICASET_UUID, disabled=True),
            get_srv_conf('srv-4', rpl_uuid='rpl-2'),
        ],
        replicasets=[
            get_rpl_conf(REPLICASET_UUID, leaders=[
                'srv-1', 'srv-2',
            ]),
            get_rpl_conf('rpl-2', leaders=['srv-4']),
        ]
    )

    return ClusterwideConfig(conf, instance_uuid=DISABLED_INSTANCE_UUID,
                             instance_uri=INSTANCE_URI,
                             replicaset_uuid=REPLICASET_UUID)


@pytest.fixture(scope="function")
def clusterwide_conf_srv_expelled():
    EXPELLED_INSTANCE_UUID = 'srv-expelled'
    REPLICASET_UUID = 'rpl-1'

    conf = get_topology_conf(
        instances=[
            get_srv_conf('srv-1', rpl_uuid=REPLICASET_UUID),
            get_srv_conf('srv-2', rpl_uuid=REPLICASET_UUID),
            get_srv_conf('srv-4', rpl_uuid='rpl-2'),
            get_expelled_srv_conf(EXPELLED_INSTANCE_UUID),
        ],
        replicasets=[
            get_rpl_conf(REPLICASET_UUID, leaders=[
                'srv-1', 'srv-3',
            ]),
            get_rpl_conf('rpl-2', leaders=['srv-4']),
        ]
    )

    return ClusterwideConfig(conf, instance_uuid=EXPELLED_INSTANCE_UUID,
                             replicaset_uuid=REPLICASET_UUID)


@pytest.fixture(scope="function")
def clusterwide_conf_srv_not_in_leaders():
    INSTANCE_NOT_IN_LEADERS_UUID = 'srv-not-in-leaders'
    REPLICASET_UUID = 'rpl-1'

    conf = get_topology_conf(
        instances=[
            get_srv_conf('srv-1', rpl_uuid=REPLICASET_UUID),
            get_srv_conf('srv-2', rpl_uuid=REPLICASET_UUID),
            get_srv_conf(INSTANCE_NOT_IN_LEADERS_UUID, rpl_uuid=REPLICASET_UUID),
            get_srv_conf('srv-4', rpl_uuid='rpl-2'),
        ],
        replicasets=[
            get_rpl_conf(REPLICASET_UUID, leaders=[
                'srv-1', 'srv-2',
            ]),
            get_rpl_conf('rpl-2', leaders=['srv-4']),
        ]
    )

    return ClusterwideConfig(conf, instance_uuid=INSTANCE_NOT_IN_LEADERS_UUID,
                             replicaset_uuid=REPLICASET_UUID)


@pytest.fixture(scope="function")
def clusterwide_conf_srv_last_in_rpl():
    INSTANCE_LAST_IN_RPL_UUID = 'srv-last-in-rpl'
    REPLICASET_UUID = 'rpl-1'

    conf = get_topology_conf(
        instances=[
            get_srv_conf(INSTANCE_LAST_IN_RPL_UUID, rpl_uuid=REPLICASET_UUID),
            get_srv_conf('srv-4', rpl_uuid='rpl-2'),
        ],
        replicasets=[
            get_rpl_conf(REPLICASET_UUID, leaders=[INSTANCE_LAST_IN_RPL_UUID]),
            get_rpl_conf('rpl-2', leaders=['srv-4']),
        ]
    )

    return ClusterwideConfig(conf, instance_uuid=INSTANCE_LAST_IN_RPL_UUID,
                             replicaset_uuid=REPLICASET_UUID)


@pytest.fixture(scope="function")
def clusterwide_conf_srv_last_in_leaders():
    INSTANCE_LAST_IN_LEADERS_UUID = 'srv-last-in-leaders'
    REPLICASET_UUID = 'rpl-1'

    conf = get_topology_conf(
        instances=[
            get_srv_conf(INSTANCE_LAST_IN_LEADERS_UUID, rpl_uuid=REPLICASET_UUID),
            get_srv_conf('srv-2', rpl_uuid=REPLICASET_UUID),
            get_srv_conf('srv-4', rpl_uuid='rpl-2'),
        ],
        replicasets=[
            get_rpl_conf(REPLICASET_UUID, leaders=[INSTANCE_LAST_IN_LEADERS_UUID]),
            get_rpl_conf('rpl-2', leaders=['srv-4']),
        ]
    )

    return ClusterwideConfig(conf, instance_uuid=INSTANCE_LAST_IN_LEADERS_UUID,
                             replicaset_uuid=REPLICASET_UUID)


@pytest.fixture(scope="function")
def clusterwide_conf_non_existent_rpl():
    NON_EXISTENT_RPL_UUID = 'non-existent-rpl'
    INSTANCE_UUID = 'srv-from-non-existent-rpl'

    conf = get_topology_conf(
        instances=[
            get_srv_conf('srv-1', rpl_uuid='rpl-1'),
            get_srv_conf('srv-2', rpl_uuid='rpl-1'),
            get_srv_conf(INSTANCE_UUID, rpl_uuid=NON_EXISTENT_RPL_UUID),
            get_srv_conf('srv-4', rpl_uuid='rpl-2'),
        ],
        replicasets=[
            get_rpl_conf('rpl-1', leaders=[
                'srv-1', 'srv-3',
            ]),
            get_rpl_conf('rpl-2', leaders=['srv-4']),
        ]
    )

    return ClusterwideConfig(conf, instance_uuid=INSTANCE_UUID,
                             replicaset_uuid=NON_EXISTENT_RPL_UUID)


@pytest.fixture(scope="function")
def clusterwide_conf_srv_from_other_rpl():
    RPL_UUID = 'rpl-1'
    INSTANCE_UUID = 'srv-from-other-rpl'

    conf = get_topology_conf(
        instances=[
            get_srv_conf('srv-1', rpl_uuid='rpl-1'),
            get_srv_conf('srv-2', rpl_uuid='rpl-1'),
            get_srv_conf(INSTANCE_UUID, rpl_uuid='rpl-2'),
        ],
        replicasets=[
            get_rpl_conf('rpl-1', leaders=[
                'srv-1', 'srv-3',
            ]),
            get_rpl_conf('rpl-2', leaders=[INSTANCE_UUID]),
        ]
    )

    return ClusterwideConfig(conf, instance_uuid=INSTANCE_UUID,
                             replicaset_uuid=RPL_UUID)


@pytest.fixture(scope="session")
def built_default_project(cartridge_cmd, short_session_tmpdir):
    project = Project(cartridge_cmd, 'default-project', short_session_tmpdir, 'cartridge')

    # build project
    cmd = [
        cartridge_cmd,
        "build",
    ]
    process = subprocess.run(cmd, cwd=project.path)
    assert process.returncode == 0, "Error during building the project"

    # don't change process title
    patch_cartridge_proc_titile(project)

    return project
