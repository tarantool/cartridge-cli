import os
import re
import shutil
import subprocess

from utils import (create_project, recursive_listdir,
                   tarantool_enterprise_is_used)

FILES_DIR = 'test/files'
INIT_NO_CARTRIDGE_FILEPATH = os.path.join(FILES_DIR, 'init_no_cartridge.lua')
INIT_IGNORE_SIGTERM_FILEPATH = os.path.join(FILES_DIR, 'init_ignore_sigterm.lua')
INIT_ADMIN_FUNCS_FILEPATH = os.path.join(FILES_DIR, 'init_admin_funcs.lua')
INIT_PRINT_ENV_FILEPATH = os.path.join(FILES_DIR, 'init_print_environment.lua')
INIT_ROLES_RELOAD_ALLOWED_FILEPATH = os.path.join(FILES_DIR, 'init_roles_reload_allowed.lua')
INIT_CHECK_PASSED_PARAMS = os.path.join(FILES_DIR, 'init_check_passed_params.lua')
ROUTER_WITH_EVAL_FILEPATH = os.path.join(FILES_DIR, 'router_with_eval.lua')
RUNDIR_CLI_CONF = os.path.join(FILES_DIR, 'rundir.cartridge.yml')

CLI_CONF = '.cartridge.yml'

DEFAULT_CFG = 'instances.yml'
DEFAULT_REPLICASETS_CFG = 'replicasets.yml'
DEFAULT_RUN_DIR = 'tmp/run'
DEFAULT_DATA_DIR = 'tmp/data'
DEFAULT_LOG_DIR = 'tmp/log'

DEFAULT_SCRIPT = 'init.lua'
DEFAULT_STATEBOARD_SCRIPT = 'stateboard.init.lua'


CARTRIDGE_PACK_SPECIAL_FILES = {
    # pre and post build hooks
    'cartridge.pre-build',
    'cartridge.post-build',

    # deprecated files
    '.cartridge.ignore',
    '.cartridge.pre',
}


# ###############
# Helpers
# ###############
def get_base_project_rocks(project_name, rockspec_name):
    return {
        '.rocks',
        '.rocks/share/tarantool/rocks/manifest',
        os.path.join('.rocks/share/tarantool/rocks', project_name),
        os.path.join('.rocks/share/tarantool/rocks', project_name, 'scm-1'),
        os.path.join('.rocks/share/tarantool/rocks', project_name, 'scm-1/rock_manifest'),
        os.path.join('.rocks/share/tarantool/rocks', project_name, 'scm-1', rockspec_name),
    }


def filter_out_files_removed_on_pack(project_files):
    # remove .git files and special files (pre and post build hooks)
    def is_removed_on_pack(filename):
        if filename in CARTRIDGE_PACK_SPECIAL_FILES:
            return True

        if filename.startswith('.git') and filename != '.gitignore':
            return True

    return set(filter(
        lambda x: not is_removed_on_pack(x),
        project_files
    ))


###############
# Class Project
###############
class Project:
    def __init__(self, cartridge_cmd, name, basepath, template='cartridge', create_func=None):
        self.name = name
        self.basepath = basepath
        self.template = template
        self.deprecated_flow_is_used = False
        self.vshard_groups_names = None
        self.custom_roles = None

        if create_func is None:
            # create project and save its path
            self.path = create_project(cartridge_cmd, basepath, name, template)
        else:
            self.path = create_func(basepath)

        # save tarantool_enterprise_is_used() result to variable
        tarantool_is_enterprise = tarantool_enterprise_is_used()

        # files that should be delivered in the result package
        project_files = recursive_listdir(self.path)
        self.distribution_files = filter_out_files_removed_on_pack(project_files)
        self.distribution_files.add('VERSION')
        self.distribution_files.add('VERSION.lua')
        if tarantool_is_enterprise:
            self.distribution_files.update({'tarantool', 'tarantoolctl'})

        # project rockspec name and path
        self.rockspec_name = '{}-scm-1.rockspec'.format(self.name)
        self.rockspec_path = os.path.join(self.path, self.rockspec_name)

        # rocks that should be delivered in the result package
        self.rocks_content = get_base_project_rocks(self.name, self.rockspec_name)

        # keys that should be mentioned in the package VERSION file
        self.version_file_keys = {
            'TARANTOOL',
            self.name,
            # default application dependencies
            'cartridge',
        }
        if tarantool_is_enterprise:
            self.version_file_keys.add('TARANTOOL_SDK')

        self.image_runtime_requirements_filepath = None

    # IDs
    def get_instance_id(self, instance_name):
        return '%s.%s' % (self.name, instance_name)

    def get_stateboard_id(self):
        return '{}-stateboard'.format(self.name)

    # cfg files
    def get_cli_cfg_path(self):
        return os.path.join(self.path, CLI_CONF)

    def get_cfg_path(self, specified_cfg=None):
        cfg_path = specified_cfg if specified_cfg is not None else DEFAULT_CFG
        return os.path.join(self.path, cfg_path)

    def get_replicasets_cfg_path(self, specified_cfg=None):
        replicasets_cfg_path = specified_cfg if specified_cfg is not None else DEFAULT_REPLICASETS_CFG
        return os.path.join(self.path, replicasets_cfg_path)

    # data dir
    def get_data_dir(self, specified_data_dir=None):
        data_dir = specified_data_dir if specified_data_dir is not None else DEFAULT_DATA_DIR
        return os.path.join(self.path, data_dir)

    # - work dirs
    def get_workdir(self, instance_name, specified_data_dir=None):
        data_dir = self.get_data_dir(specified_data_dir)
        instance_id = self.get_instance_id(instance_name)
        return os.path.join(data_dir, instance_id)

    def get_sb_workdir(self, specified_data_dir=None):
        data_dir = self.get_data_dir(specified_data_dir)
        stateboard_id = self.get_stateboard_id()
        return os.path.join(data_dir, stateboard_id)

    # log dir
    def get_log_dir(self, instance_name, specified_path=None):
        log_dir = specified_path if specified_path is not None else DEFAULT_LOG_DIR
        instance_id = self.get_instance_id(instance_name)
        return os.path.join(self.path, log_dir, '%s.log' % instance_id)

    def get_sb_log_dir(self, specified_path=None):
        log_dir = specified_path if specified_path is not None else DEFAULT_LOG_DIR
        stateboard_id = self.get_stateboard_id()
        return os.path.join(self.path, log_dir, '%s.log' % stateboard_id)

    # run dir
    def get_run_dir(self, specified_path=None):
        run_dir = specified_path if specified_path is not None else DEFAULT_RUN_DIR
        return os.path.join(self.path, run_dir)

    # - PID files
    def get_pidfile(self, instance_name, specified_run_dir=None):
        run_dir = self.get_run_dir(specified_run_dir)
        instance_id = self.get_instance_id(instance_name)
        return os.path.join(run_dir, '%s.pid' % instance_id)

    def get_sb_pidfile(self, specified_run_dir=None):
        run_dir = self.get_run_dir(specified_run_dir)
        stateboard_id = self.get_stateboard_id()
        return os.path.join(run_dir, '%s.pid' % stateboard_id)

    # - control sockets
    def get_console_sock(self, instance_name, specified_run_dir=None):
        run_dir = self.get_run_dir(specified_run_dir)
        instance_id = self.get_instance_id(instance_name)
        return os.path.join(run_dir, '%s.control' % instance_id)

    def get_sb_console_sock(self, specified_run_dir=None):
        run_dir = self.get_run_dir(specified_run_dir)
        stateboard_id = self.get_stateboard_id()
        return os.path.join(run_dir, '%s.control' % stateboard_id)

    # - notify sockets
    def get_notify_sock(self, instance_name, specified_run_dir=None):
        run_dir = self.get_run_dir(specified_run_dir)
        instance_id = self.get_instance_id(instance_name)
        return os.path.join(run_dir, '%s.notify' % instance_id)

    def get_sb_notify_sock(self, specified_run_dir=None):
        run_dir = self.get_run_dir(specified_run_dir)
        stateboard_id = self.get_stateboard_id()
        return os.path.join(run_dir, '%s.notify' % stateboard_id)

    # scripts
    def get_script(self, specified_script=None):
        script_path = specified_script if specified_script is not None else DEFAULT_SCRIPT
        return os.path.join(self.path, script_path)

    def get_sb_script(self):
        return os.path.join(self.path, DEFAULT_STATEBOARD_SCRIPT)


# ###############
# Project helpers
# ###############
def remove_dependency(project, dependency_name):
    with open(project.rockspec_path, 'r') as f:
        current_rockspec = f.read()

    new_rockspec = re.sub(
        r"'{}\s+==\s+\S+,\n".format(dependency_name),
        '',
        current_rockspec
    )

    with open(project.rockspec_path, 'w') as f:
        f.write(new_rockspec)

    project.version_file_keys.difference_update({dependency_name})


def add_dependency(project, dependency_name, dependency_version='scm-1'):
    with open(project.rockspec_path, 'r') as f:
        current_rockspec = f.read()
    new_rockspec = re.sub(
        r"dependencies\s+=\s+{\s*\n",
        '\n'.join([
            "dependencies = {",
            "   '{} == {}',\n".format(dependency_name, dependency_version)
        ]),
        current_rockspec
    )

    with open(project.rockspec_path, 'w') as f:
        f.write(new_rockspec)

    project.version_file_keys.update({dependency_name})


def add_dependency_submodule(project):
    SUBMODULE_NAME = 'custom-module'

    # create submodule itself
    submodule_path = os.path.join(project.path, 'third_party', SUBMODULE_NAME)
    os.makedirs(submodule_path)
    with open(os.path.join(submodule_path, '{}-scm-1.rockspec'.format(SUBMODULE_NAME)), 'w') as f:
        rockspec_lines = [
            "package = '{}'".format(SUBMODULE_NAME),
            "version = 'scm-1'",
            "source  = { url = '/dev/null' }",
            "build = { type = 'none'}",
        ]
        f.write('\n'.join(rockspec_lines))

    # init git repo and add to project as a submodule
    process = subprocess.run(['git', 'init'], cwd=submodule_path)
    assert process.returncode == 0, "Failed to init git repo for project submodule"

    process = subprocess.run(['git', 'add', '-A'], cwd=submodule_path)
    assert process.returncode == 0, "Failed to add project files to git"
    process = subprocess.run(['git', 'commit', '-m', '"Init"'], cwd=submodule_path)
    assert process.returncode == 0, "Failed to add initial commin"

    submodule_relpath = os.path.join('.', os.path.relpath(submodule_path, project.path))
    process = subprocess.run(
        ['git', 'submodule', 'add', submodule_relpath, submodule_relpath],
        cwd=project.path
    )
    assert process.returncode == 0, "Failed to add a submodule"

    project.distribution_files.add('.gitmodules')

    # add third-party module dependency to the rockspec
    add_dependency(project, SUBMODULE_NAME)

    # add submodule to rocks content
    project.rocks_content.add('.rocks/share/tarantool/rocks/{}'.format(SUBMODULE_NAME))

    # add cartridge.pre-build file to install submodule dependency
    with open(os.path.join(project.path, 'cartridge.pre-build'), 'w') as f:
        prebuild_script_lines = [
            "#!/bin/sh",
            "set -xe",
            "tarantoolctl rocks make --chdir ./third_party/{}".format(SUBMODULE_NAME),
        ]
        f.write('\n'.join(prebuild_script_lines))

    # add cartridge.post-build file to remove test/, tmp/ and third_party/ contents
    # and remove test/ and tmp/ from project.distribution_files
    with open(os.path.join(project.path, 'cartridge.post-build'), 'w') as f:
        postbuild_script_lines = [
            "#!/bin/sh",
            "rm -rf test tmp third_party"
        ]
        f.write('\n'.join(postbuild_script_lines))

        project.distribution_files = set(filter(
            lambda x: not any([x.startswith(p) for p in ['test', 'tmp']]),
            project.distribution_files
        ))

    # add custom-project to version_file_keys
    project.version_file_keys.add(SUBMODULE_NAME)


def remove_all_dependencies(project):
    with open(project.rockspec_path, 'w') as f:
        f.write('''
                package = '{}'
                version = 'scm-1'
                source  = {{ url = '/dev/null' }}
                dependencies = {{ 'tarantool' }}
                build = {{ type = 'none' }}
            '''.format(project.name))


def rewrite_project_file(project, project_filepath, filepath):
    with open(filepath) as file:
        with open(os.path.join(project.path, project_filepath), 'w') as project_file:
            project_file.write(file.read())


def remove_project_file(project, filepath):
    fullpath = os.path.join(project.path, filepath)
    if os.path.exists(fullpath):
        os.remove(fullpath)


# patches init to send specified statuses one by one
# to the NOTIFY_SOCKET
def patch_init_to_send_statuses(project, statuses):
    patched_init_fmt = '''#!/usr/bin/env tarantool
local notify_socket = os.getenv('NOTIFY_SOCKET')
assert(notify_socket ~= nil)

local socket = require('socket')
local sock = assert(socket('AF_UNIX', 'SOCK_DGRAM', 0), 'Can not create socket')

{send_statuses}

local fiber = require('fiber')
fiber.create(function()
    fiber.sleep(1)
end)

require('log').info('I am starting...')

-- Send READY=1
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
end'''

    send_status_fmt = '''sock:sendto('unix/', notify_socket, [=[STATUS={status}]=])
require('fiber').sleep(1)
'''

    send_statuses = '\n'.join([
        send_status_fmt.format(status=status)
        for status in statuses
    ])

    patched_init = patched_init_fmt.format(send_statuses=send_statuses)

    with open(os.path.join(project.path, 'init.lua'), 'w') as f:
        f.write(patched_init)

    with open(os.path.join(project.path, 'stateboard.init.lua'), 'w') as f:
        f.write(patched_init)


# pathes init to wait for a specified timeout (in seconds)
# before sending READY=1 to NOTIFY_SOCKET
def patch_init_to_send_ready_after_timeout(project, timeout):
    patched_init_fmt = '''#!/usr/bin/env tarantool
local notify_socket = os.getenv('NOTIFY_SOCKET')
assert(notify_socket ~= nil)

require('log').info('I am starting...')

local fiber = require('fiber')

{wait}

fiber.create(function()
    fiber.sleep(1)
end)

-- Send READY=1
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

    patched_init = patched_init_fmt.format(wait="fiber.sleep({})".format(timeout))

    with open(os.path.join(project.path, 'init.lua'), 'w') as f:
        f.write(patched_init)

    with open(os.path.join(project.path, 'stateboard.init.lua'), 'w') as f:
        f.write(patched_init)


# This function replaces init.lua and stateboard.init.lua with a script that
# logs specified messages.
# The instance ID is appended to log messages to identify them in tests
# This script doesn't enter the event loop and sends READY=1
# to NOTIFY_SOCKET after logging all messages
def patch_init_to_log_lines(project, lines):
    patched_init_fmt = '''#!/usr/bin/env tarantool
local socket = require('socket')
local sock = assert(socket('AF_UNIX', 'SOCK_DGRAM', 0), 'Can not create socket')

local instance_id

local app_name = assert(os.getenv("TARANTOOL_APP_NAME"))
if app_name:endswith("stateboard") then
    instance_id = app_name
else
    local instance_name = assert(os.getenv("TARANTOOL_INSTANCE_NAME"))
    instance_id = string.format("%s.%s", app_name, instance_name)
end

local log = require('log')
{log_lines}

-- Send READY=1
-- Copied from cartridge.cfg to provide support for NOTIFY_SOCKET
-- without entering the event loop
local notify_socket = os.getenv('NOTIFY_SOCKET')
if notify_socket then
    local socket = require('socket')
    local sock = assert(socket('AF_UNIX', 'SOCK_DGRAM', 0), 'Can not create socket')
    sock:sendto('unix/', notify_socket, 'READY=1')
end

local fiber = require('fiber')
fiber.sleep(3)
'''

    log_line_fmt = 'log.info(string.format("%s: {line}", instance_id))'

    log_lines = '\n'.join([
        log_line_fmt.format(line=line)
        for line in lines
    ])

    patched_init = patched_init_fmt.format(log_lines=log_lines)

    with open(os.path.join(project.path, 'init.lua'), 'w') as f:
        f.write(patched_init)

    with open(os.path.join(project.path, 'stateboard.init.lua'), 'w') as f:
        f.write(patched_init)


# `cartridge.cfg` changes process title to <appname>@<instance_name>
# It turned out that psutil can't get environ of the process with
# changed title.
# This function can be useful for testing start/stop with
# application that calls `cartridge.cfg`
def patch_cartridge_proc_titile(project):
    filepath = os.path.join(project.path, '.rocks/share/tarantool/cartridge.lua')
    with open(filepath) as f:
        data = f.read()

    patched_data = data.replace(
        'title.update(box_opts.custom_proc_title)',
        '-- title.update(box_opts.custom_proc_title)'
    )

    with open(filepath, 'w') as f:
        f.write(patched_data)


def patch_cartridge_version_in_rockspec(project, new_version):
    with open(project.rockspec_path, 'r') as f:
        old_rockspec = f.read()

    new_rockspec = re.sub(
        r"'cartridge == [\d\.]+'",
        "'cartridge == %s'" % new_version,
        old_rockspec
    )

    with open(project.rockspec_path, 'w') as f:
        f.write(new_rockspec)


def patch_cartridge_returned_version(project, new_version):
    if new_version is None:
        new_version_str = 'nil'
    else:
        new_version_str = "'%s'" % new_version

    new_version_code = '''
    -- patch cartridge version
    if not pcall(function() require('cartridge') end) then
        package.preload.cartridge = function() return {{ VERSION = {new_version_str} }} end
    else
        require('cartridge').VERSION = {new_version_str}
    end
'''.format(new_version_str=new_version_str)

    with open(os.path.join(project.path, 'init.lua'), 'a') as f:
        f.write(new_version_code)


def replace_project_file(project, project_file_rel_path, new_file_path):
    shutil.copy(new_file_path, os.path.join(project.path, project_file_rel_path))


def configure_vshard_groups(project, vshard_groups_names):
    project.vshard_groups_names = vshard_groups_names
    init_filepath = os.path.join(project.path, 'init.lua')

    with open(init_filepath) as f:
        old_init = f.read()

    vshard_groups_str = ', '.join([
        '%s = {}' % group_name for group_name in vshard_groups_names
    ])

    new_init = re.sub(
        r"cartridge.cfg\({\s*\n",
        '\n'.join([
            "cartridge.cfg({",
            "vshard_groups = {%s}," % vshard_groups_str,
        ]),
        old_init
    )

    with open(init_filepath, 'w') as f:
        f.write(new_init)


def add_custom_roles(project, roles):
    project.custom_roles = roles
    init_filepath = os.path.join(project.path, 'init.lua')

    with open(init_filepath) as f:
        old_init = f.read()

    new_init = re.sub(
        r"roles\s*=\s*{\s*\n",
        '\n'.join([
            "roles = {",
            ",\n".join(["        'app.roles.%s'" % r['name'] for r in roles]) + ",\n",
        ]),
        old_init
    )

    role_content_fmt = """
        return {{
            role_name = 'app.roles.{role_name}',
            init = function() end,
            stop = function() end,
            validate_config = function() return true end,
            apply_config = function() return true end,
            dependencies = {role_dependencies},
        }}
    """

    for role in roles:
        role_filepath = os.path.join(project.path, "app", "roles", "%s.lua" % role['name'])
        role_dependencies = role.get('dependencies', [])

        role_content = role_content_fmt.format(
            role_name=role['name'],
            role_dependencies='{%s}' % ', '.join(["'app.roles.%s'" % dep for dep in role_dependencies])
        )

        with open(role_filepath, 'w') as f:
            f.write(role_content)

    with open(init_filepath, 'w') as f:
        f.write(new_init)


def set_and_return_whoami_on_build(rockspec_path, project_name, version):
    who_am_i = 'I am %s' % rockspec_path

    with open(rockspec_path, 'w') as f:
        f.write('''
                package = '{}'
                version = '{}'
                source  = {{ url = '/dev/null' }}
                dependencies = {{ 'tarantool' }}
                build = {{ type = 'command',
                          build_command = 'echo {}'}}
            '''.format(project_name, version, who_am_i))

    return who_am_i


def copy_project(project_name, project):
    dir = os.getenv("CC_TEST_PREBUILT_PROJECTS")
    shutil.copytree(dir + "/" + project_name, project.path + "/built")
    project.path = project.path + "/built"
