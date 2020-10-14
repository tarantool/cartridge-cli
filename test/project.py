import os
import re
import subprocess

from utils import create_project
from utils import recursive_listdir
from utils import tarantool_enterprise_is_used


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


def patch_cartridge_version(project, new_version):
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


def rewrite_project_file(project, project_filepath, filepath):
    with open(filepath) as file:
        with open(os.path.join(project.path, project_filepath), 'w') as project_file:
            project_file.write(file.read())
