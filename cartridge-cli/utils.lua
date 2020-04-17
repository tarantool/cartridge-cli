local fio = require('fio')
local errno = require('errno')
local digest = require('digest')
local yaml = require('yaml')
local fun = require('fun')

local utils = {}

-- * -------------------- Tables --------------------

function utils.array_contains(array, value)
    if not array then
        return false
    end

    for _, v in ipairs(array) do
        if v == value then
            return true
        end
    end

    return false
end

function utils.array_index_of(array, value)
    for i, v in ipairs(array) do
        if v == value then
            return i
        end
    end
end

function utils.dict_keys(dict)
    local keys = {}

    for key, _ in pairs(dict) do
        table.insert(keys, key)
    end
    return keys
end


function utils.array_slice(array, from, to)
    local result = {}

    if from == nil then
        from = 0
    end

    if to == nil then
        to = #array
    end

    for i = from,to do
        table.insert(result, array[i])
    end

    return result
end

function utils.merge_lists(...)
    return fun.chain(...):totable()
end

function utils.merge_tables(...)
    return fun.chain(...):tomap()
end

-- * ---------------------- Bytes ==---------------------

function utils.align(addr, bytes)
    return bit.band(addr + (bytes - 1), -bytes)
end

-- Pad the buffer with zeros so that its size is a multiple of 8 bytes
function utils.buf_pad_to_8_byte_boundary(buf)
    return buf .. string.rep('\0', utils.align(#buf, 8) - #buf)
end

-- * ---------------------- Strings ----------------------

function utils.remove_leading_dot(filename)
    if string.startswith(filename, '.') then
        return string.sub(filename, 2)
    end

    return filename
end

function utils.random_string()
    return digest.urandom(8):hex()
end

function utils.remove_leading_spaces(s, spaces_num)
    spaces_num = spaces_num or 8
    local REMOVE_PATTERN = string.format('^%s', string.rep(' ', spaces_num))

    local res_lines = {}
    for _, line in ipairs(s:split('\n')) do
        local res_line = line:gsub(REMOVE_PATTERN, '')
        table.insert(res_lines, res_line)
    end

    return table.concat(res_lines, '\n'):strip()
end

-- expand() allows to render a text template, expanding ${statement}
-- into the calculated value of that statement.
-- Roughly based on http://lua-users.org/wiki/TextTemplate
--
-- First argument is the template string, then arbitrary number of
-- tables with mappings of variable=value
function utils.expand(template, ...)
    assert(type(template)=='string', 'expecting string')
    local searchlist = {...}
    local estring,evar

    local statements = {'do', 'if', 'for', 'while', 'repeat'}

    function estring(str)
        local b,e,i
        b,i = string.find(str, '%$.')
        if not b then return str end

        local R, pos = {}, 1
        repeat
            b,e = string.find(str, '^%b{}', i)
            if b then
                table.insert(R, string.sub(str, pos, b-2))
                table.insert(R, evar(string.sub(str, b+1, e-1)))
                i = e+1
                pos = i
            else
                b,e = string.find(str, '^%b()', i)
                if b then
                    table.insert(R, string.sub(str, pos, b-2))
                    table.insert(R, evar(string.sub(str, b+1, e-1)))
                    i = e+1
                    pos = i
                elseif string.find(str, '^%a', i) then
                    table.insert(R, string.sub(str, pos, i-2))
                    table.insert(R, evar(string.sub(str, i, i)))
                    i = i+1
                    pos = i
                elseif string.find(str, '^%$', i) then
                    table.insert(R, string.sub(str, pos, i))
                    i = i+1
                    pos = i
                end
            end
            b,i = string.find(str, '%$.', i)
        until not b

        table.insert(R, string.sub(str, pos))
        return table.concat(R)
    end

    local function search(index)
        for _,symt in ipairs(searchlist) do
            local ts = type(symt)
            local value
            if     ts == 'function' then value = symt(index)
            elseif ts == 'table'
            or ts == 'userdata' then value = symt[index]
                if type(value)=='function' then value = value(symt) end
            else error'search item must be a function, table or userdata' end
            if value ~= nil then return value end
        end
        error('unknown variable: '.. index)
    end

    local function elist(var, v, str, sep)
        local tab = search(v)
        if tab then
            assert(type(tab)=='table', 'expecting table from: '.. var)
            local R = {}
            table.insert(searchlist, 1, tab)
            table.insert(searchlist, 1, false)
            for _,elem in ipairs(tab) do
                searchlist[1] = elem
                table.insert(R, estring(str))
            end
            table.remove(searchlist, 1)
            table.remove(searchlist, 1)
            return table.concat(R, sep)
        else
            return ''
        end
    end

    local function get(tab,index)
        for _,symt in ipairs(searchlist) do
            local ts = type(symt)
            local value
            if     ts == 'function' then value = symt(index)
            elseif ts == 'table'
            or ts == 'userdata' then value = symt[index]
            else error'search item must be a function, table or userdata' end
            if value ~= nil then
                tab[index] = value  -- caches value and prevents changing elements
                return value
            end
        end
    end

    function evar(var)
        if string.find(var, '^[_%a][_%w]*$') then -- ${vn}
            return estring(tostring(search(var)))
        end
        local _,e,cmd = string.find(var, '^(%a+)%s.')
        if cmd == 'foreach' then -- ${foreach vn xxx} or ${foreach vn/sep/xxx}
            local vn,s
            _,e,vn,s = string.find(var, '^([_%a][_%w]*)([%s%p]).', e)
            if vn then
                if string.find(s, '%s') then
                    return elist(var, vn, string.sub(var, e), '')
                end
                local b = string.find(var, s, e, true)
                if b then
                    return elist(var, vn, string.sub(var, b+1), string.sub(var,e,b-1))
                end
            end
            error('syntax error in: '.. var, 2)
        elseif cmd == 'when' then -- $(when vn xxx)
            local vn
            _,e,vn = string.find(var, '^([_%a][_%w]*)%s.', e)
            if vn then
                local t = search(vn)
                if not t then
                    return ''
                end
                local s = string.sub(var,e)
                if type(t)=='table' or type(t)=='userdata' then
                    table.insert(searchlist, 1, t)
                    s = estring(s)
                    table.remove(searchlist, 1)
                    return s
                else
                    return estring(s)
                end
            end
            error('syntax error in: '.. var, 2)
        else
            if statements[cmd] then -- do if for while repeat
                var = 'local OUT="" '.. var ..' return OUT'
            else  -- expression
                var = 'return '.. var
            end
            local f = assert(loadstring(var))
            local t = searchlist[1]
            assert(type(t)=='table' or type(t)=='userdata', 'expecting table')
            setfenv(f, setmetatable({}, {__index=get, __newindex=t}))
            return estring(tostring(f()))
        end
    end

    return estring(template)
end

-- * ------------------------ Files ------------------------

function utils.make_tree(path)
    local ok, err = fio.mktree(path)
    if not ok then
        return false, string.format("Failed to create path %s: %s", path, err)
    end
    return true
end

function utils.write_file(path, data, mode)
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

function utils.read_file(path)
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

function utils.yaml_decode(str)
    local ok, res = pcall(yaml.decode, str)
    if not ok then return nil, res end

    return res
end

function utils.remove_by_path(path)
    if fio.path.is_dir(path) then
        local ok, err = fio.rmtree(path)
        if not ok then
            return false, string.format("Failed to remove %s: %s", path, err)
        end
    else
        local ok, err = fio.unlink(path)
        if not ok then
            return false, string.format("Failed to remove %s: %s", path, err)
        end
    end

    return true
end

function utils.copyfile(path, new_path)
    local ok, err = fio.copyfile(path, new_path)
    if not ok then
        return false, string.format("Failed to copy %s to %s: %s", path, new_path, err)
    end
    return true
end

function utils.listdir(path)
    local res, err = fio.listdir(path)
    if res == nil then
        return nil, string.format("Failed to list directory %s: %s", path, err)
    end
    return res
end

function utils.copytree(from_path, to_path)
    local ok, err = fio.copytree(from_path, to_path)
    if not ok then
        return false, string.format("Failed to copy %s to %s: %s", from_path, to_path, err)
    end
    return true
end

function utils.file_md5_hex(filename)
    local data, err = utils.read_file(filename)
    if data == nil then return nil, err end

    return digest.md5_hex(data)
end

function utils.is_subdirectory(subdir, dir)
    subdir = fio.abspath(subdir)
    dir = fio.abspath(dir)

    if subdir == dir then
        return true
    end

    if string.startswith(subdir, string.format('%s/', dir)) then
        return true
    end

    return false
end

function utils.load_variables_from_file(filepath)
    local res = {}

    local file_content, err = utils.read_file(filepath)
    if file_content == nil then return nil, err end

    -- remove shebang
    file_content = file_content:gsub("^#![^\n]*\n", "")

    local chunk, load_err = load(file_content, filepath, "t", res)
    if not chunk then
        return nil, string.format('Failed to load file %s: %s', filepath, load_err)
    end

    local ok, err = pcall(chunk)
    if not ok then
        return nil, string.format('Failed to run file %s: %s', filepath, err)
    end

    return res
end

-- Returns a list of relative paths to files in directory `dir`
function utils.find_files(dir, options)
    options = options or {}
    local exclude = options.exclude or {}

    local function find_files_rec(base_dir, subdir)
        subdir = subdir or '.'
        local files = fio.listdir(fio.pathjoin(base_dir, subdir))
        table.sort(files)
        local res = {}

        for _, file in ipairs(files) do
            local fullpath = fio.pathjoin(base_dir, subdir, file)

            if not utils.array_contains(exclude, file) then
                if fio.path.is_dir(fullpath) then
                    if options.include_dirs then
                        table.insert(res, fio.pathjoin(subdir, file))
                    end

                    local subres = find_files_rec(base_dir, fio.pathjoin(subdir, file))
                    for _,v in pairs(subres) do table.insert(res, v) end
                elseif fio.path.is_file(fullpath) then
                    table.insert(res, fio.pathjoin(subdir, file))
                end
            end
        end

        return res
    end

    local res = find_files_rec(dir, nil)
    table.sort(res)
    return res
end

-- * --------------------- Utilities ---------------------

function utils.check_that_only_one_is_true(list)
    list = list or {}
    local true_values_count = 0

    for _, value in ipairs(list) do
        if value == true then
            true_values_count = true_values_count + 1
            if true_values_count > 1 then
                return false
            end
        end
    end

    return true_values_count == 1
end

-- * --------------- Stateboard ---------------

function utils.get_stateboard_name(app_name)
    assert(app_name ~= nil)
    return string.format('%s-stateboard', app_name)
end

return utils
