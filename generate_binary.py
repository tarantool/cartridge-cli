#! /usr/bin/env python

import os
import argparse

CLI_MODULE_NAME = 'cartridge-cli'
LUA_EXT = '.lua'

MODULE_PRELOAD_TEMPLATE = ('''
-- {module_name}
package.preload['{module_name}'] = load([==[
{module_content}
]==])
''')


MODULE_RUN_MAIN_TEMPLATE = ('''
local cli = require('{module_name}')

xpcall(cli.main, function(err)
    if os.getenv('CARTRIDGE_CLI_DEBUG') then
        io.stderr:write(debug.traceback(tostring(err)))
    else
        io.stderr:write(tostring(err) .. '\\n')
    end
    os.exit(1)
end)
''')


def is_lua_file(filename):
    _, ext = os.path.splitext(filename)
    return ext == LUA_EXT


def collect_lua_module_files(source_dir, modulename):
    files = []

    module_dir = os.path.join(source_dir, modulename)
    for root, _, filenames in os.walk(module_dir):
        for filename in filenames:
            if is_lua_file(filename):
                files.append(os.path.join(os.path.relpath(root, source_dir), filename))

    entrypoint_filename = '{}.lua'.format(modulename)
    entrypoint_filepath = os.path.join(source_dir, entrypoint_filename)
    if os.path.exists(entrypoint_filepath):
        files.append(os.path.relpath(entrypoint_filepath, source_dir))

    return files


def get_module_name(filename):
    # aaa/bbb/ccc.lua -> aaa.bbb.bbb
    # aaa/bbb/init.lua -> aaa.bbb
    module_name = None
    if os.path.basename(filename) == 'init.lua':
        module_path = os.path.split(filename)[0]
        module_name = '.'.join(module_path.split('/'))
    else:
        filepath, _ = os.path.splitext(filename)
        module_name = '.'.join(filepath.split('/'))

    return module_name


def generate_modules_preload(source_dir, files, module_name):
    res = []

    for filename in files:
        with open(os.path.join(source_dir, filename), 'r') as f:
            module_name = get_module_name(filename)
            module_content = f.read()

            module_preload = MODULE_PRELOAD_TEMPLATE.format(
                module_name=module_name,
                module_content=module_content
            )
            res.append(module_preload)

    return res


def generate_binary(source_dir, dest_file, module_name):
    res = ['#!/usr/bin/env tarantool']

    files = collect_lua_module_files(source_dir, CLI_MODULE_NAME)
    modules_preload = generate_modules_preload(source_dir, files, CLI_MODULE_NAME)
    res += modules_preload

    module_run = MODULE_RUN_MAIN_TEMPLATE.format(module_name=module_name)
    res.append(module_run)

    with open(dest_file, 'w') as f:
        f.write(''.join(res))

    os.chmod(dest_file, 0o755)

    print('Generated {} binary for {} module from {} directory'.format(
        dest_file,
        module_name,
        source_dir
    ))


if __name__ == "__main__":
    parser = argparse.ArgumentParser(description='Generate module binary')
    parser.add_argument('--source-dir', '-s', type=str, help="Module source dir")
    parser.add_argument('--dest-file', '-d', type=str, help="Destination binary path")
    parser.add_argument('--module-name', '-m', type=str, help="Module name")

    args = parser.parse_args()

    generate_binary(args.source_dir, args.dest_file, args.module_name)
