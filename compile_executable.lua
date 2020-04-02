#!/usr/bin/env tarantool

local fio = require('fio')
local errno = require('errno')
local argparse_internal = require('internal.argparse')

local LUA_EXT = '.lua'

local MODULE_PRELOAD_TEMPLATE = [[
-- %s
package.preload['%s'] = load([==[
%s
]==], '@%s')
]]

local MODULE_RUN_MAIN_TEMPLATE = [[
local cli = require('%s')

xpcall(cli.main, function(err)
    if os.getenv('CARTRIDGE_CLI_DEBUG') then
        io.stderr:write(debug.traceback(tostring(err)))
    else
        io.stderr:write(tostring(err) .. '\\n')
    end
    os.exit(1)
end)
]]

local function merge_lists(...)
    local res = {}
    for i = 1, select('#', ...) do
        local t = select(i, ...)
        for _, v in ipairs(t) do
            res[#res + 1] = v
        end
    end
    return res
end

local function read_file(path)
    local file = fio.open(path)
    if file == nil then
        return nil, string.format('Failed to open file %s: %s', path, errno.strerror())
    end
    local buf = {}
    while true do
        local val = file:read(1024)
        if val == nil then
            pcall(function() file:close() end)
            return nil, string.format('Failed to read from file %s: %s', path, errno.strerror())
        elseif val == '' then
            break
        end
        table.insert(buf, val)
    end
    local ok, err = file:close()
    if not ok then return nil, err end

    return table.concat(buf, '')
end

local function write_file(path, data, mode)
    mode = mode or tonumber(644, 8)

    local file = fio.open(path, {'O_CREAT', 'O_WRONLY', 'O_TRUNC', 'O_SYNC'}, mode)
    if file == nil then
        return false, string.format('Failed to open file %s: %s', path, errno.strerror())
    end

    local res = file:write(data)

    if not res then
        return false, string.format('Failed to write to file %s: %s', path, errno.strerror())
    end

    file:close()

    return true
end

-- Function to list files in a specified directory
-- For example, for this dir:
-- /path/to/test-files
-- ├── cartridge-cli
-- │   ├── README.md
-- │   └── subcli.lua
-- └── cartridge-cli.lua
--
-- it will return:
--
-- test-files/cartridge-cli/subcli.lua
-- test-files/cartridge-cli/README.md
-- test-files/cartridge-cli.lua
local function get_directory_files(dir, basename)
    basename = basename or fio.basename(dir)
    local res = {}

    local dirfiles = fio.listdir(dir)
    for _, file in ipairs(dirfiles) do
        local filepath = fio.pathjoin(dir, file)
        local base_filepath = fio.pathjoin(basename, file)

        if fio.path.is_file(filepath) then
            table.insert(res, base_filepath)
        elseif fio.path.is_dir(filepath) then
            local subdir_files = get_directory_files(filepath, base_filepath)
            res = merge_lists(res, subdir_files)
        end
    end

    return res
end

local function is_lua_file(filepath)
    local basename = fio.basename(filepath)
    return string.endswith(basename, LUA_EXT)
end

-- Collects module files from source_dir:
-- - <module_name>.lua
-- - all lua files from <module_name>/ directory
local function collect_lua_module_files(source_dir, module_name)
    local module_files = {}

    local module_dir = fio.pathjoin(source_dir, module_name)
    local module_dir_files = get_directory_files(module_dir)

    for _, file in ipairs(module_dir_files) do
        if is_lua_file(file) then
            table.insert(module_files, file)
        end
    end

    local entrypoint_filename = string.format('%s.lua', module_name)
    local entrypoint_filepath = fio.pathjoin(source_dir, entrypoint_filename)

    if fio.path.exists(entrypoint_filepath) then
        table.insert(module_files, entrypoint_filename)
    end

    return module_files
end

local function get_module_name(filepath)
    -- aaa/bbb/ccc.lua -> aaa.bbb.bbb
    -- aaa/bbb/init.lua -> aaa.bbb
    local module_name
    if fio.basename(filepath) == 'init.lua' then
        local module_dir = fio.dirname(filepath)
        module_name = table.concat(string.split(module_dir, '/'), '.')
    else
        local module_path = string.gsub(filepath, LUA_EXT, '')
        module_name = table.concat(string.split(module_path, '/'), '.')
    end

    return module_name
end

local function generate_modules_preload(source_dir, files)
    local res = {}

    for _, filepath in ipairs(files) do
        local module_name = get_module_name(filepath)

        local full_filepath = fio.pathjoin(source_dir, filepath)
        local module_content, err = read_file(full_filepath)
        if module_content == nil then error(err) end

        local module_preload = string.format(
            MODULE_PRELOAD_TEMPLATE,
            module_name,
            module_name,
            module_content,
            filepath
        )
        table.insert(res, module_preload)
    end

    return res
end

local function compile_executable(source_dir, dest_file, module_name)
    source_dir = fio.abspath(source_dir)

    local shebang = '#!/usr/bin/env tarantool\n'

    local files = collect_lua_module_files(source_dir, module_name)
    local modules_preload = generate_modules_preload(source_dir, files)

    local module_run = string.format(MODULE_RUN_MAIN_TEMPLATE, module_name)

    local executable_strings = merge_lists(
        {shebang},
        modules_preload,
        {module_run}
    )
    local executable_content = table.concat(executable_strings, '\n')

    local ok, err = write_file(dest_file, executable_content, tonumber('0755', 8))
    if not ok then error(err) end

    print(string.format(
        'Executable %s compiled for %s module from %s',
        dest_file,
        module_name,
        source_dir
    ))
end

local args = argparse_internal.parse(arg, {
    {'source', 'string'},
    {'dest', 'string'},
    {'module', 'string'},
})

assert(args.source ~= nil, "{lease, specify source directory (--source)")
assert(args.dest ~= nil, "{lease, specify destination file name (--dest)")
assert(args.module ~= nil, "{lease, specify module name (--module)")

compile_executable(args.source, args.dest, args.module)
