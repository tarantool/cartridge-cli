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

from clusterwide_conf import ClusterwideConfig
from clusterwide_conf import get_expelled_srv_conf
from clusterwide_conf import get_srv_conf
from clusterwide_conf import get_rpl_conf
from clusterwide_conf import get_conf

from utils import Cli


# ########
# Fixtures
# ########
@pytest.fixture(scope='module')
def module_tmpdir(request):
    tmpdir = py.path.local(tempfile.mkdtemp())
    request.addfinalizer(lambda: tmpdir.remove(rec=1))
    return str(tmpdir)


@pytest.fixture(scope='function')
def short_tmpdir(request):
    tmpbase = '/tmp'
    if platform.system() == 'Darwin':
        tmpbase = '/private/tmp'

    tmpdir = py.path.local(tempfile.mkdtemp(dir=tmpbase))
    request.addfinalizer(lambda: tmpdir.remove(rec=1))
    return str(tmpdir)


@pytest.fixture(scope="session")
def docker_client():
    client = docker.from_env()
    return client


@pytest.fixture(scope="module")
def cartridge_cmd(request, module_tmpdir):
    cli_base_path = os.path.realpath(os.path.join(os.path.dirname(__file__), '..'))
    cli_path = os.path.join(module_tmpdir, 'cartridge')

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
################
# Light projects
################
@pytest.fixture(scope="function")
def light_project(cartridge_cmd, tmpdir):
    project = Project(cartridge_cmd, 'light-project', tmpdir, 'cartridge')

    remove_dependency(project, 'cartridge')
    remove_dependency(project, 'luatest')

    add_dependency_submodule(project)

    return project


#########################
# Projects with cartridge
#########################
@pytest.fixture(scope="function")
def project_with_cartridge(cartridge_cmd, tmpdir):
    project = Project(cartridge_cmd, 'project-with-cartridge', tmpdir, 'cartridge')
    remove_dependency(project, 'luatest')

    add_dependency_submodule(project)

    return project


##############################
# Project without dependencies
##############################
@pytest.fixture(scope="function")
def project_without_dependencies(cartridge_cmd, tmpdir):
    project = Project(cartridge_cmd, 'empty-project', tmpdir, 'cartridge')

    remove_all_dependencies(project)
    return project


#######################################################
# Project with patched init.lua and stateboard.init.lua
#######################################################
# This project is used in the `running` tests
# It doesn't require cartridge, but sends READY=1 signal
# in old Tarantool versions just like `cartridge.cfg` does
# It creates a simple fiber to start an event loop
@pytest.fixture(scope="function")
def project_with_patched_init(cartridge_cmd, short_tmpdir):
    project = Project(cartridge_cmd, 'patched-project', short_tmpdir, 'cartridge')

    remove_all_dependencies(project)

    patched_init = '''#!/usr/bin/env tarantool
local fiber = require('fiber')
fiber.create(function()
    fiber.sleep(1)
end)

require('log').info('I am starting...')

-- Copied from cartridge.cfg to provide support for NOTIFY_SOCKET in old tarantool
local tnt_version = string.split(_TARANTOOL, '.')
local tnt_major = tonumber(tnt_version[1])
local tnt_minor = tonumber(tnt_version[2])
if tnt_major < 2 or (tnt_major == 2 and tnt_minor < 2) then
  local notify_socket = os.getenv('NOTIFY_SOCKET')
  if notify_socket then
      local socket = require('socket')
      local sock = assert(socket('AF_UNIX', 'SOCK_DGRAM', 0), 'Can not create socket')
      sock:sendto('unix/', notify_socket, 'READY=1')
  end
end
'''

    with open(os.path.join(project.path, 'init.lua'), 'w') as f:
        f.write(patched_init)

    with open(os.path.join(project.path, 'stateboard.init.lua'), 'w') as f:
        f.write(patched_init)

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

    patched_init = '''#!/usr/bin/env tarantool
local fiber = require('fiber')
fiber.create(function()
    fiber.sleep(1)
end)

require('log').info('I am starting...')

-- ignore SIGTERM
local ffi = require('ffi')
local SIG_IGN = 1
local SIGTERM = 15
ffi.cdef[[
    void (*signal(int sig, void (*func)(int)))(int);
]]
local ignore_handler = ffi.cast("void (*)(int)", SIG_IGN)
ffi.C.signal(SIGTERM, ignore_handler)

-- Copied from cartridge.cfg to provide support for NOTIFY_SOCKET in old tarantool
local tnt_version = string.split(_TARANTOOL, '.')
local tnt_major = tonumber(tnt_version[1])
local tnt_minor = tonumber(tnt_version[2])
if tnt_major < 2 or (tnt_major == 2 and tnt_minor < 2) then
  local notify_socket = os.getenv('NOTIFY_SOCKET')
  if notify_socket then
      local socket = require('socket')
      local sock = assert(socket('AF_UNIX', 'SOCK_DGRAM', 0), 'Can not create socket')
      sock:sendto('unix/', notify_socket, 'READY=1')
  end
end
'''

    with open(os.path.join(project.path, 'init.lua'), 'w') as f:
        f.write(patched_init)

    with open(os.path.join(project.path, 'stateboard.init.lua'), 'w') as f:
        f.write(patched_init)

    return project


# ###########################
# Clusterwide config fixtures
# ###########################
@pytest.fixture(scope="function")
def clusterwide_conf_non_existent_instance():
    REPLICASET_UUID = 'rpl-1'
    NON_EXISTENT_INSTANCE_UUID = 'srv-non-existent'

    conf = get_conf(
        instances=[get_srv_conf('srv-1', rpl_uuid=REPLICASET_UUID)],
        replicasets=[get_rpl_conf(REPLICASET_UUID, leaders=['srv-1'])]
    )

    return ClusterwideConfig(conf, instance_uuid=NON_EXISTENT_INSTANCE_UUID,
                             replicaset_uuid=REPLICASET_UUID)


@pytest.fixture(scope="function")
def clusterwide_conf_non_existent_uri():
    NON_EXISTENT_INSTANCE_URI = 'non-existent-uri'
    REPLICASET_UUID = 'rpl-1'

    conf = get_conf(
        instances=[get_srv_conf('srv-1', rpl_uuid=REPLICASET_UUID)],
        replicasets=[get_rpl_conf(REPLICASET_UUID, leaders=['srv-1'])]
    )

    return ClusterwideConfig(conf, instance_uri=NON_EXISTENT_INSTANCE_URI,
                             replicaset_uuid=REPLICASET_UUID)


@pytest.fixture(scope="function")
def clusterwide_conf_simple():
    INSTANCE_UUID = 'srv-3'
    INSTANCE_URI = 'srv-3:3303'
    REPLICASET_UUID = 'rpl-1'

    conf = get_conf(
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
def clusterwide_conf_srv_disabled():
    DISABLED_INSTANCE_UUID = 'srv-disabled'
    REPLICASET_UUID = 'rpl-1'
    INSTANCE_URI = 'srv-disabled:3303'

    conf = get_conf(
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
    # for set-uri check that expelled instance doesn't cause an error
    INSTANCE_URI = 'srv-1:3303'

    conf = get_conf(
        instances=[
            get_srv_conf('srv-1', uri=INSTANCE_URI, rpl_uuid=REPLICASET_UUID),
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
                             instance_uri=INSTANCE_URI,
                             replicaset_uuid=REPLICASET_UUID)


@pytest.fixture(scope="function")
def clusterwide_conf_srv_not_in_leaders():
    INSTANCE_NOT_IN_LEADERS_UUID = 'srv-not-in-leaders'
    REPLICASET_UUID = 'rpl-1'

    conf = get_conf(
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

    return ClusterwideConfig(conf, instance_uuid=INSTANCE_NOT_IN_LEADERS_UUID, replicaset_uuid=REPLICASET_UUID)


@pytest.fixture(scope="function")
def clusterwide_conf_non_existent_rpl():
    NON_EXISTENT_RPL_UUID = 'non-existent-rpl'
    INSTANCE_UUID = 'srv-from-non-existent-rpl'

    conf = get_conf(
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

    conf = get_conf(
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
