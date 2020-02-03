local argparse = require('internal.argparse').parse
local digest = require('digest')
local errno = require('errno')
local ffi = require('ffi')
local fiber = require('fiber')
local fio = require('fio')
local fun = require('fun')
local log = require('log')
local socket = require('socket')
local yaml = require('yaml')

local self_name = fio.basename(arg[0])

local _TARANTOOL = _G._TARANTOOL
local function VERSION()
    if package.search('cartridge-cli.VERSION') then
        return require('cartridge-cli.VERSION')
    end
    return 'unknown'
end

-- * ---------------- Utility functions ----------------

-- box.NULL, custom and cdata errors aware assert
function assert(val, message, ...) -- luacheck: no global
    if not val or val == nil then
        error(tostring(message), 2)
    end
    return val, message, ...
end

local function get_cartridgecli_dir()
    local str = debug.getinfo(1, "S").source:sub(2)
    return str:match("(.*/)") or '.'
end

local function get_tarantool_dir()
    return fio.abspath(fio.dirname(arg[-1]))
end

local function get_template_dir()
    local template_dir = fio.pathjoin(get_cartridgecli_dir(), 'templates')
    if fio.path.exists(template_dir) then
        return template_dir
    end
    error('Templates not found neither in base dir nor in .rocks')
end

local function array_contains(array, value)
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

local function array_index_of(array, value)
    for i, v in ipairs(array) do
        if v == value then
            return i
        end
    end
end

local function dict_keys(dict)
    local keys = {}

    for key, _ in pairs(dict) do
        table.insert(keys, key)
    end
    return keys
end


local function array_slice(array, from, to)
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

local function align(addr, bytes)
    return bit.band(addr + (bytes - 1), -bytes)
end

-- Pad the buffer with zeros so that its size is a multiple of 8 bytes
local function buf_pad_to_8_byte_boundary(buf)
    return buf .. string.rep('\0', align(#buf, 8) - #buf)
end

local function remove_leading_dot(filename)
    if string.startswith(filename, '.') then
        return string.sub(filename, 2)
    end

    return filename
end


-- Returns a list of relative paths to files in directory `dir`
local function find_files(dir, options)
    options = options or {}
    local exclude = options.exclude or {}

    local function find_files_rec(base_dir, subdir)
        subdir = subdir or '.'
        local files = fio.listdir(fio.pathjoin(base_dir, subdir))
        table.sort(files)
        local res = {}

        for _, file in ipairs(files) do
            local fullpath = fio.pathjoin(base_dir, subdir, file)

            if not array_contains(exclude, file) then
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

local function globtopattern(g)
    -- glob pattern to lua format
    -- source: https://github.com/davidm/lua-glob-pattern

    local p = "^"  -- pattern being built
    local i = 0    -- index in g
    local c        -- char at index i in g.

    -- unescape glob char
    local function unescape()
        if c == '\\' then
            i = i + 1; c = g:sub(i,i)
            if c == '' then
                p = '[^]'
                return false
            end
        end
        return true
    end

    -- escape pattern char
    local function escape(str)
        return str:match("^%w$") and str or '%' .. str
    end

    -- Convert tokens at end of charset.
    local function charset_end()
        while 1 do
            if c == '' then
                p = '[^]'
                return false
            elseif c == ']' then
                p = p .. ']'
                break
            else
                if not unescape() then break end
                local c1 = c
                i = i + 1; c = g:sub(i,i)
                if c == '' then
                    p = '[^]'
                    return false
                elseif c == '-' then
                    i = i + 1; c = g:sub(i,i)
                    if c == '' then
                        p = '[^]'
                        return false
                    elseif c == ']' then
                        p = p .. escape(c1) .. '%-]'
                        break
                    else
                        if not unescape() then break end
                        p = p .. escape(c1) .. '-' .. escape(c)
                    end
                elseif c == ']' then
                    p = p .. escape(c1) .. ']'
                    break
                else
                    p = p .. escape(c1)
                    i = i - 1 -- put back
                end
            end
            i = i + 1; c = g:sub(i,i)
        end
        return true
    end

    -- Convert tokens in charset.
    local function charset()
        i = i + 1; c = g:sub(i,i)
        if c == '' or c == ']' then
            p = '[^]'
            return false
        elseif c == '^' or c == '!' then
            i = i + 1; c = g:sub(i,i)
            if c ~= ']' then
                p = p .. '[^'
                if not charset_end() then return false end
            end
        else
            p = p .. '['
            if not charset_end() then return false end
        end
        return true
    end

    -- Convert tokens.
    while 1 do
        i = i + 1; c = g:sub(i,i)
        if c == '' then
            p = p .. '$'
            break
        elseif c == '?' then
            p = p .. '.'
        elseif c == '*' then
            -- if double asterisk
            if i + 1 <= #g and g:sub(i + 1, i + 1) == '*' then
                p = p .. '.*'
            else
                p = p .. '[^/]*'
            end
        elseif c == '[' then
            if not charset() then break end
        elseif c == '\\' then
            i = i + 1; c = g:sub(i,i)
            if c == '' then
                p = p .. '\\$'
                break
            end
            p = p .. escape(c)
        else
            p = p .. escape(c)
        end
    end
    return p
end

-- * --------------------------- Color helpers ---------------------------

local RESET_TERM = '\x1B[0m'
local COLORS = {
    {'magenta', '\x1B[35m'},
    {'blue', '\x1B[34m'},
    {'cyan', '\x1B[36m'},
    {'green', '\x1B[32m'},
    {'bright_magenta', '\x1B[95m'},
    {'bright_cyan', '\x1B[96m'},
    {'bright_blue', '\x1B[94m'},
    {'bright_green', '\x1B[92m'},
}
local COLORS_ITER = fun.iter(COLORS):map(function(x) return x[2] end):cycle()
local NEXT_COLOR = 0
local function next_color_code()
    NEXT_COLOR =  NEXT_COLOR + 1
    return COLORS_ITER:nth(NEXT_COLOR)
end

local ERROR_COLOR_CODE = '\x1B[31m' -- red
local WARN_COLOR_CODE = '\x1B[33m' -- yellow
local INFO_COLOR_CODE = '\x1B[36m' -- cyan
local DEBUG_COLOR_CODE = '\x1B[35m' -- magneta

-- Map of `log_level_letter => color_code`.
local COLOR_CODE_BY_LOG_LEVEL = fun.iter({
    S_FATAL = ERROR_COLOR_CODE,
    S_SYSERROR = ERROR_COLOR_CODE,
    S_ERROR = ERROR_COLOR_CODE,
    S_CRIT = ERROR_COLOR_CODE,
    S_WARN = WARN_COLOR_CODE,
    S_INFO = RESET_TERM,
    S_VERBOSE = RESET_TERM,
    S_DEBUG = RESET_TERM,
}):map(function(k, v) return k:sub(3, 3), v end):tomap()
local ERROR_LOG_LINE_PATTERN = ' (%u)> '

local function colored_msg(msg, color_code)
    return color_code .. msg .. RESET_TERM
end

-- * ------------------------------ Messages ------------------------------

local function die(fmt, ...)
    local msg = "ERROR: " .. string.format(fmt, ...)
    print(colored_msg(msg, ERROR_COLOR_CODE))
    os.exit(1)
end

local function warn(fmt, ...)
    local msg = "WARNING: " .. string.format(fmt, ...)
    print(colored_msg(msg, WARN_COLOR_CODE))
end

local function info(fmt, ...) -- luacheck: no unused
    local msg = string.format(fmt, ...)
    print(colored_msg(msg, INFO_COLOR_CODE))
end

local function debug(fmt, ...) -- luacheck: no unused
    local msg = string.format(fmt, ...)
    print(colored_msg(msg, DEBUG_COLOR_CODE))
end

-- * ------------------------------ Files ------------------------------
-- `fio` functions wrappers
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

local function file_md5_hex(filename)
    local data, err = read_file(filename)
    if data == nil then return false, err end

    return digest.md5_hex(data)
end

local function remove_by_path(path)
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

local function make_tree(path)
    local ok, err = fio.mktree(path)
    if not ok then
        return false, string.format("Failed to create path %s: %s", path, err)
    end
    return true
end

local function copyfile(path, new_path)
    local ok, err = fio.copyfile(path, new_path)
    if not ok then
        return false, string.format("Failed to copy %s to %s: %s", path, new_path, err)
    end
    return true
end

local function listdir(path)
    local res, err = fio.listdir(path)
    if res == nil then
        return nil, string.format("Failed to list directory %s: %s", path, err)
    end
    return res
end

local function copytree(from_path, to_path)
    local ok, err = fio.copytree(from_path, to_path)
    if not ok then
        return false, string.format("Failed to copy %s to %s: %s", from_path, to_path, err)
    end
    return true
end

local function is_subdirectory(subdir, dir)
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

local function load_variables_from_file(filepath)
    local res = {}

    local file_content, err = read_file(filepath)
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

-- expand() allows to render a text template, expanding ${statement}
-- into the calculated value of that statement.
-- Roughly based on http://lua-users.org/wiki/TextTemplate
--
-- First argument is the template string, then arbitrary number of
-- tables with mappings of variable=value
local function expand(template, ...)
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


-- pack() allows to pack a number of values to a binary string
-- in a printf-like manner
local function pack(format, ...)
  local stream = {}
  local vars = {...}
  local endianness = true

  local i = 1
  while i <= format:len() do
    local opt = format:sub(i, i)

    if opt == '<' then
      endianness = true
    elseif opt == '>' then
      endianness = false
    elseif opt:find('[bBhHiIlL]') then
      local n = opt:find('[hH]') and 2 or opt:find('[iI]') and 4 or opt:find('[lL]') and 8 or 1
      local val = tonumber(table.remove(vars, 1))

      local bytes = {}
      for _ = 1, n do
        table.insert(bytes, string.char(val % (2 ^ 8)))
        val = math.floor(val / (2 ^ 8))
      end

      if not endianness then
        table.insert(stream, string.reverse(table.concat(bytes)))
      else
        table.insert(stream, table.concat(bytes))
      end
    elseif opt:find('[fd]') then
      local val = tonumber(table.remove(vars, 1))
      local sign = 0

      if val < 0 then
        sign = 1
        val = -val
      end

      local mantissa, exponent = math.frexp(val)
      if val == 0 then
        mantissa = 0
        exponent = 0
      else
        mantissa = (mantissa * 2 - 1) * math.ldexp(0.5, (opt == 'd') and 53 or 24)
        exponent = exponent + ((opt == 'd') and 1022 or 126)
      end

      local bytes = {}
      if opt == 'd' then
        val = mantissa
        for _ = 1, 6 do
          table.insert(bytes, string.char(math.floor(val) % (2 ^ 8)))
          val = math.floor(val / (2 ^ 8))
        end
      else
        table.insert(bytes, string.char(math.floor(mantissa) % (2 ^ 8)))
        val = math.floor(mantissa / (2 ^ 8))
        table.insert(bytes, string.char(math.floor(val) % (2 ^ 8)))
        val = math.floor(val / (2 ^ 8))
      end

      table.insert(bytes, string.char(math.floor(exponent * ((opt == 'd') and 16 or 128) + val) % (2 ^ 8)))
      val = math.floor((exponent * ((opt == 'd') and 16 or 128) + val) / (2 ^ 8))
      table.insert(bytes, string.char(math.floor(sign * 128 + val) % (2 ^ 8)))

      if not endianness then
        table.insert(stream, string.reverse(table.concat(bytes)))
      else
        table.insert(stream, table.concat(bytes))
      end
    elseif opt == 's' then
      table.insert(stream, tostring(table.remove(vars, 1)))
      table.insert(stream, string.char(0))
    elseif opt == 'c' then
      local n = format:sub(i + 1):match('%d+')
      local length = tonumber(n)

      if length > 0 then
        local str = tostring(table.remove(vars, 1))
        if length - str:len() > 0 then
          str = str .. string.rep(' ', length - str:len())
        end
        table.insert(stream, str:sub(1, length))
      end
      i = i + n:len()
    end
    i = i + 1
  end

  return table.concat(stream)
end


local function prompt(text, default)
    if default == nil then
        io.write(string.format("%s: ", text))
    elseif type(default) == 'string' then
        io.write(string.format("%s [%s]: ", text, default))
    end

    local entry = io.read()

    if string.strip(entry) == "" then
        return default
    end

    return entry
end

local function is_executable(path)
    local S_IEXEC = 64
    return bit.band(fio.stat(path).mode, S_IEXEC) ~= 0
end

local function which(binary)
    for _, path in ipairs(string.split(os.getenv("PATH"), ':') or {}) do
        local files, err = listdir(path)
        if files == nil then return nil, err end

        for _, file in ipairs(files) do
            local full_path = fio.pathjoin(path, file)
            if file == binary and
                fio.path.exists(full_path) and
                fio.path.is_file(full_path) and
                is_executable(full_path) then
                    return full_path
            end
        end
    end
end

-- * ---------------------- Running commands ----------------------

-- Runs command using `os.execute`
-- Returns:
-- - true        in case of success
-- - false, err  otherwise
local function call(command, ...)
    local cmd = string.format(command, ...)

    local rc = os.execute(cmd)
    if rc == 0 then
        return true
    end

    local err = string.format(
        'Failed to execute "%s". Command returned non-zero code: %s',
        cmd, rc
    )
    return false, err
end

-- Runs command using `io.popen` and returns output
-- Command stderr is redirected to /dev/null
-- - output        in case of success
-- - nil, err      otherwise
local function check_output(command, ...)
    local cmd = string.format(command, ...)
    local res, popen_err = io.popen(string.format('((%s) 2>/dev/null) && echo OK', cmd))

    if res == nil then
        return nil, popen_err
    end

    local output = res:read("*all")
    if output:endswith('OK\n') then
        output = output:gsub('OK\n$', '')
        return output
    end

    local cmd_err = string.format('Failed to execute "%s": %s', cmd, output)
    return nil, cmd_err
end

local function tarantool_is_enterprise()
    local tarantool_dir = get_tarantool_dir()
    local tnt_version = fio.pathjoin(tarantool_dir, 'VERSION')
    return fio.path.exists(tnt_version)
end

-- * ---------------- Project-related functions ----------------

local function format_version(major, minor, patch)
    major = major or 0
    minor = minor or 0
    patch = patch or 0
    return string.format('%s.%s.%s', major, minor, patch)
end

local function normalize_version(str)
    local patterns = {
        "(%d+)%.(%d+)%.(%d+)-(%d+)-(%g+)",
        "(%d+)%.(%d+)%.(%d+)-(%d+)",
        "(%d+)%.(%d+)%.(%d+)-(%g+)",
        "(%d+)%.(%d+)%.(%d+)",
        "(%d+)%.(%d+)",
        "(%d+)"
    }

    for _, pattern in ipairs(patterns) do
        local major, minor, patch, count, hash = string.match(str, pattern)
        if major ~= nil then
            local release = '0'
            if count ~= nil and hash ~= nil then
                release = string.format('%s-%s', count, hash)
            elseif count ~= nil then
                release = tostring(count)
            elseif hash ~= nil then
                release = tostring(hash)
            end
            return format_version(major, minor, patch), release
        end
    end
end

local function is_git_project(dir)
    local git_path = fio.pathjoin(dir, '.git')
    return fio.path.exists(git_path) and fio.path.is_dir(git_path)
end

local function detect_git_version(source_dir)
    if which('git') == nil then
        return nil
    end

    if not is_git_project(source_dir) then
        return nil
    end

    local raw_version, err = check_output('cd "%s" && git describe --tags --long', source_dir)
    if raw_version == nil then
        warn('Failed to getect version from git: %s', err)
        return nil
    end

    local version, release = normalize_version(raw_version)
    if version == nil then
        warn("Detected version '%s' ignored, " ..
              "because it doesn't look like proper " ..
              "version (major.minor.patch[-count][-commit])", version)
    end

    return version, release
end

local function find_rockspec(source_dir)
    local files, err = listdir(source_dir)
    if files == nil then return nil, err end

    for _, file in ipairs(files) do
        if string.endswith(file, '.rockspec') then
            return file
        end
    end
end

local function detect_name(source_dir)
    local rockspec_filename, err = find_rockspec(source_dir)
    if rockspec_filename == nil then return nil, err end

    local rockspec_filepath = fio.pathjoin(source_dir, rockspec_filename)
    local rockspec, err = load_variables_from_file(rockspec_filepath)
    if rockspec == nil then
        return nil, string.format('Failed to load rockspec %s: %s', rockspec_filepath, err)
    end

    local name = rockspec.package
    if name == nil then
        return nil, string.format("Rockspec %s doesn't contain required field 'package'", rockspec_filepath)
    end

    return name
end

local function detect_name_version_release(source_dir, raw_name, raw_version)
    local name
    local release
    local version

    if raw_name ~= nil then
        name = raw_name
    else
        local detected_name, err = detect_name(source_dir)

        if detected_name == nil then
            die(
                "Failed to detect project name: %s.\n" ..
                "Please pass project name explicitly via --name",
                err
            )
        end

        info("Detected project name: %s", detected_name)
        name = detected_name
    end

    if raw_version then
        version, release = normalize_version(raw_version)
        if version == nil then
            die("Passed version '%s' should be semantic (major.minor.patch[-count][-commit])",
                raw_version)
        end
        release = release or '0'
    else
        version, release = detect_git_version(source_dir)
        if version == nil then
            die("Failed to detect version from project in directory '%s'. " ..
                    "Please pass it explicitly via --version", source_dir)
        end

        info("Detected project version: %s-%s", version, release)
    end

    if not fio.path.exists(fio.pathjoin(source_dir, 'init.lua')) then
        die("Application must have `init.lua` in its root directory")
    end

    return name, version, release
end

-- * ----------- Distribution types -----------

local distribution_types = {
    TGZ = 'tgz',
    ROCK = 'rock',
    RPM = 'rpm',
    DEB = 'deb',
    DOCKER = 'docker',
}

local available_distribution_types = {}
for _, t in pairs(distribution_types) do
    table.insert(available_distribution_types, t)
end

-- * ----------- Special filenames ------------

local PREBUILD_SCRIPT_NAME = 'cartridge.pre-build'
local POSTBUILD_SCRIPT_NAME = 'cartridge.post-build'

-- deprecated files

local DEP_PREBUILD_SCRIPT_NAME = '.cartridge.pre'
local DEP_IGNORE_FILE_NAME = '.cartridge.ignore'

-- build directory

local HOME_DIR = os.getenv('HOME') or '/home'
local DEFAULT_BUILD_DIRECTORY_NAME = 'build.cartridge'
local CARTRIDGE_TMP_PATH = fio.pathjoin(HOME_DIR, '.cartridge/tmp')
local BUILD_DIRECTORY_NAME_TEMPLATE = 'cartridge-build-%s'

-- * --------------- Preinstall ---------------

local CREATE_USER_SCRIPT = [[
${groupadd} -r tarantool > /dev/null 2>&1 || :
${useradd} -M -N -g tarantool -r -d /var/lib/tarantool -s /sbin/nologin\
    -c "Tarantool Server" tarantool > /dev/null 2>&1 || :
${mkdir} -p /etc/tarantool/conf.d/ --mode 755 2>&1 || :
${mkdir} -p /var/lib/tarantool/ --mode 755 2>&1 || :
${chown} tarantool:tarantool /var/lib/tarantool 2>&1 || :
${mkdir} -p /var/run/tarantool/ --mode 755 2>&1 || :
${chown} tarantool:tarantool /var/run/tarantool 2>&1 || :
]]

-- * -------------- Postinstall --------------

local SET_OWNER_SCRIPT = [[
${chown} -R root:root /usr/share/tarantool/${name}
${chown} root:root /etc/systemd/system/${name}.service
${chown} root:root /etc/systemd/system/${name}@.service
${chown} root:root /usr/lib/tmpfiles.d/${name}.conf
]]

-- * ---------------- Systemd ----------------

local SYSTEMD_UNIT_FILE = [[
[Unit]
Description=Tarantool Cartridge app ${name}.default
After=network.target

[Service]
Type=simple
ExecStartPre=${mkdir} -p ${workdir}.default
ExecStart=${bindir}/tarantool ${dir}/init.lua
Restart=on-failure
RestartSec=2
User=tarantool
Group=tarantool

Environment=TARANTOOL_WORKDIR=${workdir}.default
Environment=TARANTOOL_CFG=/etc/tarantool/conf.d/
Environment=TARANTOOL_PID_FILE=/var/run/tarantool/${name}.default.pid
Environment=TARANTOOL_CONSOLE_SOCK=/var/run/tarantool/${name}.default.control

LimitCORE=infinity
# Disable OOM killer
OOMScoreAdjust=-1000
# Increase fd limit for Vinyl
LimitNOFILE=65535

# Systemd waits until all xlogs are recovered
TimeoutStartSec=86400s
# Give a reasonable amount of time to close xlogs
TimeoutStopSec=10s

[Install]
WantedBy=multi-user.target
Alias=${name}
]]

local SYSTEMD_INSTANTIATED_UNIT_FILE = [[
[Unit]
Description=Tarantool Cartridge app ${name}@%i
After=network.target

[Service]
Type=simple
ExecStartPre=${mkdir} -p ${workdir}.%i
ExecStart=${bindir}/tarantool ${dir}/init.lua
Restart=on-failure
RestartSec=2
User=tarantool
Group=tarantool

Environment=TARANTOOL_WORKDIR=${workdir}.%i
Environment=TARANTOOL_CFG=/etc/tarantool/conf.d/
Environment=TARANTOOL_PID_FILE=/var/run/tarantool/${name}.%i.pid
Environment=TARANTOOL_CONSOLE_SOCK=/var/run/tarantool/${name}.%i.control
Environment=TARANTOOL_INSTANCE_NAME=%i

LimitCORE=infinity
# Disable OOM killer
OOMScoreAdjust=-1000
# Increase fd limit for Vinyl
LimitNOFILE=65535

# Systemd waits until all xlogs are recovered
TimeoutStartSec=86400s
# Give a reasonable amount of time to close xlogs
TimeoutStopSec=10s

[Install]
WantedBy=multi-user.target
Alias=${name}.%i
]]

-- * --------------------- Debian --------------------

local DEBIAN_CONTROL_FILE = [[
Package: ${name}
Version: ${version}
Maintainer: ${maintainer}
Architecture: ${arch}
Description: ${desc}
Depends: ${deps}

]]

-- * ------------------- Tmpfiles --------------------

local TMPFILES_CONFIG = [[
d /var/run/tarantool 0755 tarantool tarantool
]]

-- * ------------------- Dockerfile -------------------

local DOCKERFILE_FROM_DEFAULT = 'FROM centos:8'

local DOCKERFILE_TAIL_TEMPLATE = [[
SHELL ["/bin/bash", "-c"]

RUN yum install -y git gcc make cmake unzip

# create user and directories
RUN groupadd -r tarantool \
    && useradd -M -N -g tarantool -r -d /var/lib/tarantool -s /sbin/nologin \
        -c "Tarantool Server" tarantool \
    &&  mkdir -p /var/lib/tarantool/ --mode 755 \
    && chown tarantool:tarantool /var/lib/tarantool \
    && mkdir -p /var/run/tarantool/ --mode 755 \
    && chown tarantool:tarantool /var/run/tarantool

${install_tarantool}

RUN echo 'd /var/run/tarantool 644 tarantool tarantool' > /usr/lib/tmpfiles.d/${name}.conf \
    && chmod 644 /usr/lib/tmpfiles.d/${name}.conf

# copy application source code
COPY . ${dir}

WORKDIR ${dir}

RUN : "----------- pre-build -----------" \
    && \
    if [ -f ${prebuild_script_name} ]; then \
       set -e && . ${prebuild_script_name} \
       && rm ${prebuild_script_name}; \
    fi \
    && \
    : "------------- build -------------" \
    && \
    if ls *.rockspec 1> /dev/null 2>&1; then \
        tarantoolctl rocks make; \
    fi \
    && \
    : "----------- post-build -----------" \
    && \
    if [ -f ${postbuild_script_name} ]; then \
        set -e && . ${postbuild_script_name} \
        && rm ${postbuild_script_name}; \
    fi

USER tarantool:tarantool

CMD TARANTOOL_WORKDIR=${workdir}.${instance_name} \
    TARANTOOL_PID_FILE=/var/run/tarantool/${name}.${instance_name}.pid \
    TARANTOOL_CONSOLE_SOCK=/var/run/tarantool/${name}.${instance_name}.control \
    tarantool ${dir}/init.lua
]]

local DOCKER_INSTALL_OPENSOURCE_TARANTOOL_TEMPLATE = [[
# install opensource Tarantool
RUN curl -s \
        https://packagecloud.io/install/repositories/tarantool/${tarantool_repo_version}/script.rpm.sh | bash \
    && yum -y install tarantool tarantool-devel
]]

local DOCKER_INSTALL_ENTERPRISE_TARANTOOL_TEMPLATE = [[
ARG DOWNLOAD_TOKEN
WORKDIR /usr/share/tarantool

RUN DOWNLOAD_URL=https://tarantool:${"$"}{DOWNLOAD_TOKEN}@download.tarantool.io \
    && curl -O -L ${"$"}{DOWNLOAD_URL}/enterprise/tarantool-enterprise-bundle-${sdk_version}.tar.gz \
    && tar -xzf tarantool-enterprise-bundle-${sdk_version}.tar.gz \
    && rm -rf tarantool-enterprise-bundle-${sdk_version}.tar.gz

ENV PATH="/usr/share/tarantool/tarantool-enterprise:${"$"}{PATH}"
]]

-- * ----------- Packing flow global state -----------

local pack_state = {
    -- Here will be stored general application info to be used on
    --   application packing, for example, application name or
    --   flag detects if application uses deprecated packing flow
}

-- * ---------------- Generic packing ----------------

local function get_rock_versions(project_dir)
    local dependencies = {}
    -- XXX: fix manifest filepath compution
    local manifest_filepath = fio.pathjoin(project_dir, '.rocks/share/tarantool/rocks/manifest')

    if fio.path.exists(manifest_filepath) then
        if not fio.path.is_file(manifest_filepath) then
            local err = string.format('Manifest is not a file: %s', manifest_filepath)
            return nil, err
        end
        -- parse manifest file
        local manifest, err = load_variables_from_file(manifest_filepath)
        if manifest == nil then
            return nil, err
        end

        for module, versions in pairs(manifest.dependencies) do
            for version, _ in pairs(versions) do
                if dependencies[module] ~= nil then
                    warn(
                        'Found multiple versions for %s dependency in rocks manifest %s',
                        module, manifest_filepath
                    )
                    break
                end
                dependencies[module] = version
            end
        end
    end

    return dependencies
end

local function generate_version_file(distribution_dir)
    info('Generate VERSION file')

    -- collect VERSION file lines
    local version_file_lines = {}

    if pack_state.tarantool_is_enterprise then
        -- copy TARANTOOL and TARANTOOL_SDK versions from SDK version file
        local tarantool_dir = get_tarantool_dir()
        local tnt_version = fio.pathjoin(tarantool_dir, 'VERSION')
        if not fio.path.exists(tnt_version) then
            warn("can't open VERSION file from Tarantool SDK. SDK information can't be " ..
                "shipped to the resulting package. ")
        else
            local tnt_versions_content, err = read_file(tnt_version)
            if tnt_versions_content == nil then return false, err end

            local tnt_version_lines = tnt_versions_content:split()
            for _, line in ipairs(tnt_version_lines) do
                table.insert(version_file_lines, line)
            end
        end
    else
        -- TARANTOOL version
        local tnt_version_line = string.format('TARANTOOL=%s', _TARANTOOL)
        table.insert(version_file_lines, tnt_version_line)
    end

    -- application version
    local app_version_line = string.format(
        "%s=%s",
        pack_state.name,
        pack_state.version_release
    )
    table.insert(version_file_lines, app_version_line)

    -- rocks versions
    local rocks_versions, err = get_rock_versions(distribution_dir)
    if rocks_versions == nil then
        warn("can't process rocks manifest file. Dependency information can't be " ..
             "shipped to the resulting package: %s", err)
    else
        for rock, version in pairs(rocks_versions) do
            if rock ~= pack_state.name then
                local rock_version_line = string.format("%s=%s", rock, version)
                table.insert(version_file_lines, rock_version_line)
            end
        end
    end

    -- write collected info to VERSION file
    local version_filepath = fio.pathjoin(distribution_dir, 'VERSION')
    local version_file_content = table.concat(version_file_lines, '\n') .. '\n'
    local ok, err = write_file(version_filepath, version_file_content, tonumber(644, 8))
    if not ok then return false, err end

    return true
end

local function pattern_form(pattern)
    if pattern == '' or -- blank line
            string.startswith(pattern, '#') then -- comment
        return nil, false
    end

    if string.startswith(pattern, '\\#') then -- escape #
        pattern = pattern:sub(2, #pattern)
    end
     -- trim space
    pattern = pattern:gsub("%s+", "")

    local negative = false
    if string.startswith(pattern, '!') then
        pattern = pattern:sub(2, #pattern)
        negative = true
    end
    return pattern, negative
end

local function path_form(path, destdir)
    if string.startswith(path, './') then
        path = path:sub(3, #path)
    end
    -- if this is a folder, then added to end /
    if fio.path.is_dir(fio.pathjoin(destdir, path)) then
        path = path .. '/'
    end


    return path
end

local function matching(str, pattern)
    -- case: pattern <simple>, str [<simple> | <simple/>]
    if not string.endswith(pattern, '/') and string.endswith(str, '/') then
        str = str:sub(1, #str - 1)
    end

    -- pattern <simple> --> <**/simple>
    -- str     <simple> --> </simple>
    if not string.startswith(pattern, '/') then
        pattern = '**/' .. pattern
    end
    if not string.startswith(str, '/') then
        str = '/' .. str
    end

    local matched = string.match(str, globtopattern(pattern))

    if matched ~= nil and #matched == #str then
        return true
    else
        return false
    end
end

local function remove_ignored(destdir)
    local ignore = fio.pathjoin(destdir, DEP_IGNORE_FILE_NAME)
    if not fio.path.exists(ignore) then return true end

    local files = find_files(destdir, { include_dirs = true })

    -- formatting all pattern and exclusion exception pattern
    local patterns, exceptions  = {}, {}

    local ignore_file_content, err = read_file(ignore)
    if ignore_file_content == nil then return false, err end

    for _, pattern in ipairs(string.split(ignore_file_content, '\n')) do
        local pretty_pattern, negative = pattern_form(pattern)
        if pretty_pattern then
            if negative then
                table.insert(exceptions, pretty_pattern)
            else
                table.insert(patterns, pretty_pattern)
            end
        end
    end

    local matched = {}
    for _, file in ipairs(files) do
        local pretty_file = path_form(file, destdir)
        for _, ignore_glob in ipairs(patterns) do
            if matching(pretty_file, ignore_glob) then
                local except = false
                for _, e in ipairs(exceptions) do
                    if matching(pretty_file, e) then
                        except = true
                        break
                    end
                end
                if not except then
                    table.insert(matched, pretty_file)
                end
            end
        end
    end

    for _, f in ipairs(matched) do
        local path = fio.pathjoin(destdir, f)
        if fio.path.exists(path) then
            local ok, err = remove_by_path(path)
            if not ok then return false, err end
        end
    end

    return true
end

local function check_filemodes(dir)
    local FILE_REQURED_BITS = tonumber('444', 8)
    local DIR_REQUIRED_BITS = tonumber('555', 8)

    local function has_bits(mode, bits)
        return bit.band(mode, bits) == bits
    end

    local files, err = listdir(dir)
    if files == nil then return false, err end

    for _, filename in ipairs(files) do
        local filepath = fio.pathjoin(dir, filename)
        local filemode = fio.stat(filepath).mode

        if fio.path.is_file(filepath) then
            if not has_bits(filemode, FILE_REQURED_BITS) then
                local err = string.format(
                    'File %s has invalid mode: %o. ' ..
                        'It should have read permissions for all',
                    filepath, filemode
                )
                return false, err
            end
        elseif fio.path.is_dir(filepath) then
            if not has_bits(filemode, DIR_REQUIRED_BITS) then
                local err =string.format(
                    'Directory %s has invalid mode: %o. ' ..
                        'It should have read and execute permissions for all',
                    filepath, filemode
                )
                return false, err
            end

            if not fio.path.is_link(filepath) then
                local ok, err = check_filemodes(filepath)
                if not ok then return false, err end
            end
        end
    end

    return true
end

local function form_distribution_dir(dest_dir)
    local ok, err = copytree(pack_state.path, dest_dir)
    if not ok then return false, err end

    local rocks_dir = fio.pathjoin(dest_dir, '.rocks')
    if fio.path.exists(rocks_dir) then
        local ok, err = remove_by_path(rocks_dir)
        if not ok then return false, err end
    end
    local git = which('git')
    if git == nil then
        warn(
            "git not found. It is possible that some of the extra files " ..
            "normally ignored are shipped to the resulting package. "
        )
    elseif not is_git_project(dest_dir) then
        warn(
            "Directory %s is not a git project. It is possible that some of the extra files " ..
                "normally ignored are shipped to the resulting package. ",
            dest_dir
        )
    else
        info('Running `git clean`')
        -- Clean up all files explicitly ignored by git, to not accidentally
        -- ship development snaps, xlogs or other garbage to production.
        local ok, err = call("cd %q && %s clean -f -d -X", dest_dir, git)
        if not ok then return false, err end

        info('Running `git clean` for submodules')
        -- Recursively cleanup all submodules
        local ok, err = call(
            "cd %q && %s submodule foreach --recursive %s clean -f -d -X",
            dest_dir, git, git
        )
        if not ok then return false, err end
    end

    if not pack_state.deprecated_flow then
        local git_dir = fio.pathjoin(dest_dir, '.git')
        if fio.path.exists(git_dir) then
            info('Remove .git directory')
            local ok, err = remove_by_path(git_dir)
            if not ok then return false, err end
        end
        -- check application files mode
        info('Check application file modes')
        local ok, err = check_filemodes(dest_dir)
        if not ok then return false, err end
    end

    return true
end

local function run_hook(dir, filename)
    info('Running %s', filename)
    assert(fio.path.exists(fio.pathjoin(dir, filename)))

    local ret = os.execute(
        'set -e\n' ..
        string.format('cd %q\n', dir) ..
        string.format('. ./%s', filename)
    )

    if ret ~= 0 then
        return false, string.format('Failed to execute %s', filename)
    end

    local ok, err = remove_by_path(fio.pathjoin(dir, filename))
    if not ok then return false, err end

    return true
end

local function build_application(dir)
    -- pre build
    if pack_state.deprecated_flow then
        if fio.path.exists(fio.pathjoin(dir, DEP_PREBUILD_SCRIPT_NAME)) then
            local ok, err = run_hook(dir, DEP_PREBUILD_SCRIPT_NAME)
            if not ok then return false, err end
        end
    else  -- new build flow
        if fio.path.exists(fio.pathjoin(dir, PREBUILD_SCRIPT_NAME)) then
            local ok, err = run_hook(dir, PREBUILD_SCRIPT_NAME)
            if not ok then return false, err end
        end
    end

    -- build
    local rockspec = find_rockspec(dir)
    if rockspec ~= nil then
        info('Running tarantoolctl rocks make')
        local ret = os.execute(
            string.format(
                'cd %q; exec tarantoolctl rocks make %q',
                dir, rockspec
            )
        )
        if ret ~= 0 then
            return false, 'Failed to install rocks'
        end
    end

    -- apply .cartridge.ignore (DEPRECATED)
    if pack_state.deprecated_flow then
        -- deleting files matching patterns from .cartridge.ignore
        info('Remove files matching patterns from %s', DEP_IGNORE_FILE_NAME)
        local ok, err = remove_ignored(dir)
        if not ok then return false, err end

        -- remove special files
        for _, filename in ipairs({DEP_IGNORE_FILE_NAME, DEP_PREBUILD_SCRIPT_NAME}) do
            local filepath = fio.pathjoin(dir, filename)
            if fio.path.exists(filepath) then
                info('Remove %s', filename)
                local ok, err = remove_by_path(filepath)
                if not ok then return false, err end
            end
        end

        -- remove git dir
        local git_dir = fio.pathjoin(dir, '.git')
        if fio.path.exists(git_dir) then
            info('Remove .git directory')
            local ok, err = remove_by_path(git_dir)
            if not ok then return false, err end
        end
    else  -- new build flow
        if fio.path.exists(fio.pathjoin(dir, POSTBUILD_SCRIPT_NAME)) then
            local ok, err = run_hook(dir, POSTBUILD_SCRIPT_NAME)
            if not ok then return false, err end
        end
    end

    return true
end

local function copy_taranool_binaries(dir)
    info('Copy Tarantool Enterprise binaries')
    assert(pack_state.tarantool_is_enterprise)

    local tarantool_dir = get_tarantool_dir()

    for _, binary in ipairs({'tarantool', 'tarantoolctl'}) do
        local path_from = fio.pathjoin(tarantool_dir, binary)
        local path_to = fio.pathjoin(dir, binary)

        local ok, err = copyfile(path_from, path_to)
        if not ok then return false, err end
    end

    return true
end

local function form_systemd_dir(base_dir, opts)
    opts = opts or {}
    info('Form application systemd dir')

    local unit_template = opts.unit_template or SYSTEMD_UNIT_FILE
    local instantiated_unit_template = opts.instantiated_unit_template or SYSTEMD_INSTANTIATED_UNIT_FILE

    local systemd_dir = fio.pathjoin(base_dir, '/etc/systemd/system')
    local ok, err = make_tree(systemd_dir)
    if not ok then return false, err end

    local expand_params = {
        name = pack_state.name,
        dir = fio.pathjoin('/usr/share/tarantool/', pack_state.name),
        workdir = fio.pathjoin('/var/lib/tarantool/', pack_state.name),
        mkdir = opts.mkdir,
    }

    if pack_state.tarantool_is_enterprise then
        expand_params.bindir = expand_params.dir
    else
        expand_params.bindir = '/usr/bin'
    end

    local unit_template_filepath = fio.pathjoin(systemd_dir, string.format('%s.service', pack_state.name))
    local instantiated_unit_template_filepath = fio.pathjoin(systemd_dir, string.format('%s@.service', pack_state.name))
    local ok, err = write_file(
        unit_template_filepath,
        expand(unit_template, expand_params)
    )
    if not ok then return false, err end

    local ok, err = write_file(
        instantiated_unit_template_filepath,
        expand(instantiated_unit_template, expand_params)
    )
    if not ok then return false, err end

    return true
end

local function write_tmpfiles_conf(base_dir)
    info('Write application tmpfiles configuration')

    local tmpfiles_dir = fio.pathjoin(base_dir, '/usr/lib/tmpfiles.d')
    local ok, err = make_tree(tmpfiles_dir)
    if not ok then return false, err end

    local tmpfiles_conf_filepath = fio.pathjoin(
        tmpfiles_dir,
        string.format('%s.conf', pack_state.name)
    )
    local ok, err = write_file(
        tmpfiles_conf_filepath,
        TMPFILES_CONFIG,
        tonumber('0644', 8)  -- filemode
    )
    if not ok then return false, err end

    return true
end

-- * ---------------- TAR.GZ packing ----------------

local function pack_tgz()
    local tgz_file_name = string.format(
        "%s-%s.tar.gz",
        pack_state.name,
        pack_state.version_release
    )
    tgz_file_name = fio.pathjoin(pack_state.dest_dir, tgz_file_name)

    info("Packing tar.gz file")

    local tar = which('tar')

    if tar == nil then
        return false, "tar binary is required to pack tar.gz"
    end

    local distribution_dir = fio.pathjoin(pack_state.build_dir, pack_state.name)
    local ok, err = make_tree(distribution_dir)
    if not ok then return false, err end

    info("Packing tar.gz in: %s", pack_state.build_dir)

    local ok, err = form_distribution_dir(distribution_dir)
    if not ok then return false, err end

    local ok, err = build_application(distribution_dir)
    if not ok then return false, err end

    local ok, err = generate_version_file(distribution_dir)
    if not ok then return false, err end

    if pack_state.tarantool_is_enterprise then
        local ok, err = copy_taranool_binaries(distribution_dir)
        if not ok then return false, err end
    end

    local data, err = check_output(
        "cd %s && %s -cvzf - %s",
        pack_state.build_dir, tar, pack_state.name
    )
    if data == nil then
        return false, string.format("Failed to pack tgz: %s", err)
    end

    local ok, err = write_file(tgz_file_name, data)
    if not ok then return false, err end

    info("Resulting tar.gz saved as: %s", tgz_file_name)

    return true
end

-- * ---------------- ROCK packing ----------------

local function pack_rock()
    local distribution_dir = fio.pathjoin(pack_state.build_dir, pack_state.name)
    local ok, err = make_tree(distribution_dir)
    if not ok then return false, err end

    info("Packing binary rock in: %s", pack_state.build_dir)

    local ok, err = form_distribution_dir(distribution_dir)
    if not ok then return false, err end

    local ok, err = build_application(distribution_dir)
    if not ok then return false, err end

    local ok, err = generate_version_file(distribution_dir)
    if not ok then return false, err end

    if pack_state.tarantool_is_enterprise then
        local ok, err = copy_taranool_binaries(distribution_dir)
        if not ok then return false, err end
    end

    fio.chdir(pack_state.build_dir)

    local rockspec = find_rockspec(distribution_dir)
    local content = ''
    if rockspec then
        local err
        content, err = read_file(fio.pathjoin(distribution_dir, rockspec))
        if content == nil then return false, err end

        content = string.gsub(content, "(.-version%s-=%s-['\"])(.-)(['\"].*)",
                '%1' .. pack_state.version_release .. '%3')
        if not content then
            return false, string.format('Rockspec %s is not valid! Version not found!', rockspec)
        end
    end

    local name_of_rockspec = string.format(
        '%s-%s.rockspec',
        pack_state.name,
        pack_state.version_release
    )

    local new_rockspec = fio.pathjoin(distribution_dir, name_of_rockspec)

    local ok, err = write_file(new_rockspec, content)
    if not ok then return false, err end

    fio.chdir(distribution_dir)

    local rock_filename = string.format(
        '%s-%s.*.rock',
        pack_state.name,
        pack_state.version_release
    )

    local ok, err = call('tarantoolctl rocks pack %s ', new_rockspec)
    if not ok then return false, err end

    rock_filename = fio.glob(fio.pathjoin(distribution_dir, rock_filename))[1]

    local dest_rock_filename = fio.pathjoin(pack_state.dest_dir, fio.basename(rock_filename))

   local ok, err = copyfile(rock_filename, dest_rock_filename)
    if not ok then return false, err end

    info('Resulting rock saved as: %s', dest_rock_filename)

    return true
end

-- * ---------------- RPM packing ----------------

-- RPM file is a binary format, consisting of metadata in form of
-- key-value pairs and then a gzipped cpio archive (of SVR-4 variety).
--
-- Documentation on the binary format can be found here:
-- - http://ftp.rpm.org/max-rpm/s1-rpm-file-format-rpm-file-format.html
-- - https://docs.fedoraproject.org/ro/Fedora_Draft_Documentation/0.1/html/RPM_Guide/ch-package-structure.html
--
-- Also I've found this explanatory blog post to be of great help:
-- - https://blog.bethselamin.de/posts/argh-pm.html

-- Here's how the layout looks like:
--
-- +-----------------------+
-- |                       |
-- |     Lead (legacy)     |
-- |                       |
-- +-----------------------+
-- |                       |
-- |   Signature Header    |
-- |                       |
-- +-----------------------+
-- |                       |
-- |        Header         |
-- |                       |
-- +-----------------------+
-- |                       |
-- |                       |
-- |    Data (cpio.gz)     |
-- |                       |
-- |                       |
-- +-----------------------+
--
-- Both signature sections have the same format: a set of typed
-- key-value pairs.
--
-- While debugging, I've used rpm-dissecting tool from mkrepo:
-- - https://github.com/tarantool/mkrepo/blob/master/mkrepo.py

local RPM_MAGIC = 0xedabeedb
local RPM_VER = {3, 0}

local HEADERSIGNATURES=62
local HEADERIMMUTABLE=63


-- There are way more tags in the spec than what I've included here
-- both for signature header and regular header. Most of them are
-- optional.
--
-- Explanation and values for most of this tags can be found in documentation:
-- - http://ftp.rpm.org/max-rpm/s1-rpm-file-format-rpm-file-format.html
-- - https://docs.fedoraproject.org/ro/Fedora_Draft_Documentation/0.1/html/RPM_Guide/ch-package-structure.html
--
-- But I didn't find documentation for some tags (PAYLOADDIGEST, PAYLOADDIGESTALGO),
-- so, I got these values from the rpm repo:
-- - https://github.com/rpm-software-management/rpm/blob/master/lib/rpmtag.h
-- payload digest explanation can be found here:
-- - https://github.com/rpm-software-management/rpm/issues/163
--
local SIGNATURE_TAG_TABLE = {
    SIG_SIZE = 1000,
    MD5 = 1004,
    PAYLOADSIZE = 1007,
    SHA1 = 269
}

local HEADER_TAG_TABLE = {
    NAME = 1000,
    VERSION = 1001,
    RELEASE = 1002,
    EPOCH = 1003,
    SUMMARY = 1004,
    DESCRIPTION = 1005,
    BUILDTIME = 1006,
    SIZE = 1009,
    OS = 1021,
    ARCH = 1022,
    PAYLOADFORMAT = 1124,
    LICENSE = 1014,
    GROUP = 1016,
    PAYLOADCOMPRESSOR = 1125,
    PAYLOADFLAGS = 1126,
    PREIN = 1023,
    PREINPROG = 1085,
    DIRNAMES = 1118,
    BASENAMES = 1117,
    DIRINDEXES = 1116,
    FILEUSERNAME = 1039,
    FILEGROUPNAME = 1040,
    FILESIZES = 1028,
    FILEMODES = 1030,
    FILEINODES = 1096,
    FILEDEVICES = 1095,
    RPMVERSION = 1064,
    FILEMTIMES = 1034,
    FILEFLAGS = 1037,
    FILELANGS = 1097,
    FILERDEVS = 1033,
    FILEDIGESTS = 1035,
    FILELINKTOS = 1036,
    REQUIREFLAGS = 1048,
    REQUIRENAME = 1049,
    REQUIREVERSION = 1050,
    PAYLOADDIGEST = 5092,
    PAYLOADDIGESTALGO = 5093,
}

local PGPHASHALGO_SHA256 =  8  -- used for PAYLOADDIGEST

local RPMSENSE_FLAGS = {
    LESS =  0x02,
    GREATER = 0x04,
    EQUAL = 0x08,
    PREREQ = 0x40,
    INTERP = 0x100,
    SCRIPT_PRE = 0x200,
    SCRIPT_POST = 0x400,
    SCRIPT_PREUN = 0x800,
    SCRIPT_POSTUN = 0x1000,
}

local function gen_lead(name)
    -- The Lead is a legacy structure that used to describe RPM files
    -- before header sections were introduced.
    --
    -- struct rpmlead {
    --   unsigned char magic[4];
    --   unsigned char major, minor;
    --   short type;
    --   short archnum;
    --   char name[66];
    --   short osnum;
    --   short signature_type;
    --   char reserved[16];
    -- } ;
    name = name .. string.rep('\0', 66-#name)
    local lead = pack(
        '>IBBHHc66HHc16',
        RPM_MAGIC, RPM_VER[1], RPM_VER[2],
        0, 1, name,
        1, 5, string.rep('\0', 16))

    return lead
end

local function gen_header(tags, tag_table, region_tag)
    -- Pack a value to a binary form, and align it to a required
    -- address boundary. Since rpm headers are mmap-ed, numeric
    -- types need to have proper alignment (say, 32-bit integer
    -- addresses should be aligned to 4-byte boundary)
    --
    -- Parameters:
    -- value - value to pack, or an array of values
    -- val_type - expected type of value, e.g. INT32, STRING, etc...
    -- offset - Since we can't calculate alignment "in a vacuum", we
    --          need to know the absolute address of the value in the
    --          resulting buffer. That'd be a basis for calculating
    --          alignment
    --
    -- Return values: {tag, num_elements, buffer, padding}
    -- tag -- type tag, as per the RPM spec (e.g. 5 for int64 data)
    -- num_elements -- 1 in case of single values, otherwise the size
    --                 of packed array
    -- buffer -- packed binary data
    -- padding -- how many zero bytes were added to the beginning of
    --            buffer to properly align its contents
    local function pack_value(value, val_type, offset)
        if val_type == 'STRING_ARRAY' then
            local buf = ""
            for _, v in ipairs(value) do
                buf = buf .. pack(">s", v)
            end
            return 8, #value, buf, 0
        end

        if val_type == 'BIN' then
            return 7, #value, pack(string.format(">c%d", #value), value), 0
        end

        if type(value) ~= 'table' then
            value = {value}
        end

        local ret_val_type = 0
        local pad = 0
        local buf = ""
        for _, v in ipairs(value) do
            if val_type == 'NULL' then
                ret_val_type = 0
                buf = buf .. ''
            elseif val_type == 'CHAR' then
                ret_val_type = 1
                buf = buf .. pack('>B', v)
            elseif val_type == 'INT8' then
                ret_val_type = 2
                buf = buf .. pack('>b', v)
            elseif val_type == 'INT16' then
                ret_val_type = 3
                pad = align(offset, 2) - offset
                buf = buf .. pack('>h', v)
            elseif val_type == 'INT32' then
                ret_val_type = 4
                pad = align(offset, 4) - offset
                buf = buf .. pack('>i', v)
            elseif val_type == 'INT64' then
                ret_val_type = 5
                pad = align(offset, 8) - offset
                buf = buf .. pack('>L', v)
            elseif val_type == 'STRING' then
                ret_val_type = 6
                buf = buf .. pack('>s', v)
            end
        end
        buf = string.rep('\0', pad) .. buf
        return ret_val_type, #value, buf, pad
    end

    local function gen_index_header(num_index_entries, num_data_bytes)
        local buf = pack('>BBB', 0x8e, 0xad, 0xe8) -- header magic number
        buf = buf .. pack('>BI', 0x01, 0x0) -- version and "reserved"
        buf = buf .. pack('>II', num_index_entries, num_data_bytes)

        return buf
    end

    local function header_index_record(tag, val_type, offset, count)
        return pack(">iiii", tag, val_type, offset, count)
    end

    -- This function generates the binary header
    local function gen_header_body()
        local num_index_entries = 0

        local indexes = ""
        local store = ""
        local offset = 0

        for _, tg in ipairs(tags) do
            local key = tg[1]
            local value_type = tg[2]
            local value = tg[3]

            local tag = tag_table[key]
            if tag == nil then
                error("No such tag: " .. key)
            end

            local val_type, count, buf, pad = pack_value(value, value_type, offset)
            local index_entry = header_index_record(tag, val_type, offset+pad, count)

            indexes = indexes .. index_entry
            store = store .. buf
            offset = #store

            num_index_entries = num_index_entries + 1
        end

        -- Region tag is INSANE. Basically, there's a SHA1 digest tag
        -- in the signature header, to check the consistency of cpio
        -- arhive and other headers. So far, so good. But someone long
        -- ago decided it would be a nice idea to have some mutable
        -- tags that can be binary-patched in the rpm file by some
        -- utility and not affect the SHA1 signature. (I know, right?)
        -- So the region tag is a special tag that says how large is
        -- the area of tag space that is immutable. It should be
        -- calculated exactly as written below, with size value itself
        -- negative (sic!).
        indexes = header_index_record(region_tag, 7, #store, 16) .. indexes
        num_index_entries = num_index_entries + 1
        store = store .. header_index_record(region_tag, 7, -num_index_entries*16, 16)

        local body = gen_index_header(num_index_entries, #store)
        body = body .. indexes
        body = body .. store

        return body
    end


    return gen_header_body()
end

local function filter_out_known_files(files)
    -- RPM metadata shouldn't contain some known directories, so that
    -- it doesn't update their mtime during package installation.
    local result = {}

    local RPM_DIRNAME_BLACKLIST = {
        './',
        './bin',
        './usr',
        './usr/bin',
        './usr/local',
        './usr/local/bin',
        './usr/share',
        './usr/share/tarantool',
        './usr/lib',
        './usr/lib/tmpfiles.d',
        './var',
        './var/lib',
        './var/lib/tarantool',
        './var/run',
        './var/log',
        './etc',
        './etc/tarantool',
        './etc/tarantool/conf.d',
        './etc/systemd',
        './etc/systemd/system'
    }

    for _, file in ipairs(files) do
        if file ~= '' and not array_contains(RPM_DIRNAME_BLACKLIST, file) then
            table.insert(result, file)
        end
    end

    return result
end

local function generate_fileinfo(source_dir)
    local function gen_dirnames(files)
        local dirnames = {}

        for _, file in ipairs(files) do
            file = remove_leading_dot(file)
            local dirname = fio.dirname(file)
            dirnames[dirname..'/'] = true
        end

        return dict_keys(dirnames)
    end

    local files = find_files(
        source_dir,
        {include_dirs=true})

    files = filter_out_known_files(files)

    local dirnames = gen_dirnames(files)
    table.sort(dirnames)

    local result = {
        dirnames = dirnames,
        basenames = {},
        dirindexes = {},
        filegroupnames = {},
        fileusernames = {},
        filesizes = {},
        filemodes = {},
        fileinodes = {},
        filedevices = {},
        filemtimes = {},
        fileflags = {},
        filelangs = {},
        filerdevs = {},
        filedigests = {},
        filelinktos = {}
    }


    for _, file in ipairs(files) do
        file = remove_leading_dot(file)

        local fullpath = fio.pathjoin(source_dir, file)
        local dirname = fio.dirname(file)
        local basename = fio.basename(file)
        local fileuser, filegroup = 'root', 'root'
        local filesize = fio.stat(fullpath).size
        local filemode = fio.stat(fullpath).mode
        local fileinode = fio.stat(fullpath).inode
        local filedevice = fio.stat(fullpath).dev
        local filemtime = fio.stat(fullpath).mtime

        if fio.path.is_dir(fullpath) then
            table.insert(result.fileflags, 0)
            table.insert(result.filedigests, '')
            filesize = 4096
        elseif fio.path.is_link(fullpath) then
            table.insert(result.fileflags, 0)
            table.insert(result.filedigests, '')
        else
            local filedigest, err = file_md5_hex(fullpath)
            if filedigest == nil then return false, err end

            table.insert(result.fileflags, bit.lshift(1, 4))
            table.insert(result.filedigests, filedigest)
        end

        table.insert(result.basenames, basename)
        table.insert(result.dirindexes, array_index_of(dirnames, dirname..'/')-1)
        table.insert(result.filegroupnames, filegroup)
        table.insert(result.fileusernames, fileuser)
        table.insert(result.filesizes, filesize)
        table.insert(result.filemodes, filemode)
        table.insert(result.fileinodes, fileinode)
        table.insert(result.filedevices, filedevice)
        table.insert(result.filemtimes, filemtime)
        table.insert(result.filelangs, '')
        table.insert(result.filerdevs, 0)

        table.insert(result.filelinktos, '')
    end

    return result
end

local function pack_cpio(opts)
    -- The resulting CPIO structure should look like it will be
    -- extracted to /
    -- So it contains /usr/share/tarantool/<app>, systemd unit files and tmpfiles conf
    local cpio = which('cpio')

    if cpio == nil then
        return nil, "cpio binary is required to build rpm packages"
    end

    local gzip = which('gzip')

    if gzip == nil then
        return nil, "gzip binary is required to build rpm packages"
    end

    opts = opts or {}
    opts.mkdir = '/usr/bin/mkdir'

    local distribution_dir = fio.pathjoin(pack_state.build_dir, '/usr/share/tarantool/', pack_state.name)
    local ok, err = form_distribution_dir(distribution_dir)
    if not ok then return nil, err end

    local ok, err = build_application(distribution_dir)
    if not ok then return nil, err end

    local ok, err = generate_version_file(distribution_dir)
    if not ok then return nil, err end

    local ok, err = form_systemd_dir(pack_state.build_dir, opts)
    if not ok then return nil, err end

    local ok, err = write_tmpfiles_conf(pack_state.build_dir)
    if not ok then return nil, err end

    if pack_state.tarantool_is_enterprise then
        local ok, err = copy_taranool_binaries(distribution_dir)
        if not ok then return nil, err end
    end

    local files = find_files(pack_state.build_dir, {include_dirs=true, exclude={'.git'}})
    files = filter_out_known_files(files)

    local ok, err = write_file(fio.pathjoin(pack_state.build_dir, 'files'), table.concat(files, '\n'))
    if not ok then return nil, err end

    local ok, pack_err = call("cd %s && cat files | %s -o -H newc > unpacked", pack_state.build_dir, cpio)
    if not ok then
        return nil, string.format("Failed to pack CPIO: %s", pack_err)
    end

    local payloadsize = fio.stat(fio.pathjoin(pack_state.build_dir, 'unpacked')).size
    local archive, read_err = check_output("cd %s && cat unpacked | %s -9", pack_state.build_dir, gzip)
    if archive == nil then
        return nil, string.format("Failed to pack CPIO: %s", read_err)
    end

    for _, f in ipairs({'unpacked', 'files'}) do
        local filepath = fio.pathjoin(pack_state.build_dir, f)
        local ok, err = remove_by_path(filepath)
        if not ok then return nil, err end
    end

    local fileinfo = generate_fileinfo(pack_state.build_dir)

    return {
        archive = archive,
        fileinfo = fileinfo,
        payloadsize = payloadsize,
    }
end

local function tarantool_version()
    return _TARANTOOL:split('-', 1)[1]
end

local function tarantool_next_major_version()
    return tostring(_TARANTOOL:split('.', 1)[1] + 1)
end

local function pack_rpm(opts)
    opts = opts or {}
    local rpm_file_name = fio.pathjoin(
        pack_state.dest_dir,
        string.format(
            "%s-%s.rpm",
            pack_state.name,
            pack_state.version_release
        )
    )

    info("Packing rpm file")
    local lead = gen_lead(pack_state.name)

    local cpio, err = pack_cpio(opts)
    if cpio == nil then return false, err end

    -- compute payload digest
    local payloaddigest_algo = PGPHASHALGO_SHA256
    local payloaddigest = digest.sha256_hex(cpio.archive)

    local create_user_script_rpm = expand(CREATE_USER_SCRIPT, {
        groupadd = '/usr/sbin/groupadd',
        useradd = '/usr/sbin/useradd',
        mkdir = '/usr/bin/mkdir',
        chown = '/usr/bin/chown',
    })

    local fileinfo = cpio.fileinfo
    local header_tags = {
        {'NAME', 'STRING', pack_state.name},
        {'VERSION', 'STRING', pack_state.version},
        {'RELEASE', 'STRING', pack_state.release},
        {'SUMMARY', 'STRING', ''},
        {'DESCRIPTION', 'STRING', ''},
        {'PAYLOADFORMAT', 'STRING', 'cpio'},
        {'LICENSE', 'STRING', 'N/A'},
        {'GROUP', 'STRING', 'None'},
        {'OS', 'STRING', 'linux'},
        {'ARCH', 'STRING', 'x86_64'},
        {'PAYLOADCOMPRESSOR', 'STRING', 'gzip'},
        {'PAYLOADFLAGS', 'STRING', '5'},
        {'PREIN', 'STRING', create_user_script_rpm},
        {'PREINPROG', 'STRING', '/bin/sh'},
        {'DIRNAMES', 'STRING_ARRAY', fileinfo.dirnames},
        {'BASENAMES', 'STRING_ARRAY', fileinfo.basenames},
        {'DIRINDEXES', 'INT32', fileinfo.dirindexes},
        {'FILEUSERNAME', 'STRING_ARRAY', fileinfo.fileusernames},
        {'FILEGROUPNAME', 'STRING_ARRAY', fileinfo.filegroupnames},
        {'FILESIZES', 'INT32', fileinfo.filesizes},
        {'FILEMODES', 'INT16', fileinfo.filemodes},
        {'FILEINODES', 'INT32', fileinfo.fileinodes},
        {'FILEDEVICES', 'INT32', fileinfo.filedevices},
        {'FILERDEVS', 'INT16', fileinfo.filerdevs},
        {'FILEMTIMES', 'INT32', fileinfo.filemtimes},
        {'FILEFLAGS', 'INT32', fileinfo.fileflags},
        {'FILELANGS', 'STRING_ARRAY', fileinfo.filelangs},
        {'FILEDIGESTS', 'STRING_ARRAY', fileinfo.filedigests},
        {'FILELINKTOS', 'STRING_ARRAY', fileinfo.filelinktos},
        {'SIZE', 'INT32', cpio.payloadsize},
        {'PAYLOADDIGEST', 'STRING_ARRAY', {payloaddigest}},
        {'PAYLOADDIGESTALGO', 'INT32', payloaddigest_algo},
    }

    if not pack_state.tarantool_is_enterprise then
        --- Append RPM dependency flags for Tarantool
        --- See Dependency Tags section of
        --- - https://docs.fedoraproject.org/ro/Fedora_Draft_Documentation/0.1/html/RPM_Guide/ch-package-structure.html

        local min_version = tarantool_version()
        local max_version = tarantool_next_major_version()

        -- tarantool >= min_version < max_version
        local tarantool_dependency_tags = {
            {'REQUIREFLAGS', 'INT32',
                { bit.bor(RPMSENSE_FLAGS.GREATER, RPMSENSE_FLAGS.EQUAL), RPMSENSE_FLAGS.LESS }},
            {'REQUIRENAME', 'STRING_ARRAY', { 'tarantool', 'tarantool' }},
            {'REQUIREVERSION', 'STRING_ARRAY', { min_version, max_version }}
        }

        for _, tag in ipairs(tarantool_dependency_tags) do
            table.insert(header_tags, tag)
        end
    end

    local header = gen_header(
        header_tags,
        HEADER_TAG_TABLE,
        HEADERIMMUTABLE
    )

    local body = header .. cpio.archive
    local md5 = digest.md5(body)
    local sha1 = digest.sha1_hex(header)
    local sig_size = #body
    local signature_header = gen_header(
        {
            {'SHA1', 'STRING', sha1},
            {'SIG_SIZE', 'INT32', sig_size},
            {'PAYLOADSIZE', 'INT32', cpio.payloadsize},
            {'MD5', 'BIN', md5},
        },
        SIGNATURE_TAG_TABLE,
        HEADERSIGNATURES
    )

    body = lead .. buf_pad_to_8_byte_boundary(signature_header) .. body

    local ok, err = write_file(rpm_file_name, body)
    if not ok then return false, err end

    info("Resulting rpm saved as: %s", rpm_file_name)

    return true
end


-- * ---------------- DEB packing ----------------

-- DEB package is an ar archive contains debian-binary, control.tar.gz and data.tar.gz files
--
-- debian-binary  : contains format version string (2.0)
-- data.tar.xz    : package files
-- control.tar.xz : control files (control, preinst etc.)
--
local function form_deb_control_dir(dest_dir, name, version)
    local ok, err = make_tree(dest_dir)
    if not ok then return false, err end

    -- control
    local control_filepath = fio.pathjoin(dest_dir, 'control')
    local control_params = {
        name = name,
        version = version,
        maintainer = 'Tarantool Cartridge Developer',
        arch = 'all',
        desc = ('Tarantool Cartridge app %s'):format(name),
        deps = ''
    }

    if not pack_state.tarantool_is_enterprise then
        -- Add tarantool dependency
        local min_version = tarantool_version()
        local max_version = tarantool_next_major_version()

        control_params.deps = string.format('tarantool (>= %s), tarantool (<< %s)',
                                            min_version, max_version)
    end

    local ok, err = write_file(
        control_filepath,
        expand(DEBIAN_CONTROL_FILE, control_params)
    )
    if not ok then return false, err end

    -- preinst
    local preinst_filepath = fio.pathjoin(dest_dir, 'preinst')
    local ok, err = write_file(
        preinst_filepath,
        expand(CREATE_USER_SCRIPT, {
            groupadd = '/usr/sbin/groupadd',
            useradd = '/usr/sbin/useradd',
            mkdir = '/bin/mkdir',
            chown = '/bin/chown',
        }),
        tonumber('0755', 8)  -- filemode
    )
    if not ok then return false, err end

    -- postinst
    local postinst_filepath = fio.pathjoin(dest_dir, 'postinst')
    local ok, err = write_file(
        postinst_filepath,
        expand(SET_OWNER_SCRIPT, {
            chown = '/bin/chown',
            name = name,
        }),
        tonumber('0755', 8)  -- filemode
    )
    if not ok then return false, err end

    return true
end

local function pack_deb(opts)
    local deb_file_name = string.format(
        "%s-%s.deb",
        pack_state.name,
        pack_state.version_release
    )

    local tar = which('tar')

    if tar == nil then
       return false, "tar binary is required to pack deb"
    end

    local ar = which('ar')

    if ar == nil then
        return false, "ar binary is required to pack deb"
    end

    opts = opts or {}
    opts.mkdir = '/bin/mkdir'

    info("Packing deb in: %s", pack_state.build_dir)

    -- debian-binary
    local debian_binary_path = fio.pathjoin(pack_state.build_dir, 'debian-binary')
    local ok, err = write_file(debian_binary_path, '2.0\n')
    if not ok then return false, err end

    -- control.tar.xz
    local control_dir = fio.pathjoin(pack_state.build_dir, 'control')
    local control_tgz_path = fio.pathjoin(pack_state.build_dir, 'control.tar.xz')
    local ok, err = form_deb_control_dir(control_dir, pack_state.name, pack_state.version_release)
    if not ok then return false, err end

    local control_data, pack_control_err = check_output("cd %s && %s -cvJf - .", control_dir, tar)
    if control_data == nil then
        die('Failed to pack deb control files: %s', pack_control_err)
    end
    local ok, err = write_file(control_tgz_path, control_data)
    if not ok then return false, err end

    -- data.tar.xz
    local data_dir = fio.pathjoin(pack_state.build_dir, 'data')
    local data_tgz_path = fio.pathjoin(pack_state.build_dir, 'data.tar.xz')
    local ok, err = make_tree(data_dir)
    if not ok then return false, err end

    local distribution_dir = fio.pathjoin(data_dir, '/usr/share/tarantool/', pack_state.name)
    local ok, err = form_distribution_dir(distribution_dir)
    if not ok then return false, err end

    local ok, err = build_application(distribution_dir)
    if not ok then return false, err end

    local ok, err = generate_version_file(distribution_dir)
    if not ok then return false, err end

    local ok, err = form_systemd_dir(data_dir, opts)
    if not ok then return false, err end

    local ok, err = write_tmpfiles_conf(data_dir)
    if not ok then return false, err end

    if pack_state.tarantool_is_enterprise then
        local ok, err = copy_taranool_binaries(distribution_dir)
        if not ok then return false, err end
    end

    local data, pack_data_err = check_output("cd %s && %s -cvJf - .", data_dir, tar)
    if data == nil then
        die('Failed to pack deb package files: %s', pack_data_err)
    end
    local ok, err = write_file(data_tgz_path, data)
    if not ok then return false, err end

    -- pack .deb
    local ok, pack_deb_err = call(
        "cd %s && %s r %s debian-binary control.tar.xz data.tar.xz",
        pack_state.build_dir, ar, deb_file_name
    )
    if not ok then
        die('Failed to pack DEB package: %s', pack_deb_err)
    end

    local ok, err = copyfile(fio.pathjoin(pack_state.build_dir, deb_file_name), pack_state.dest_dir)
    if not ok then return false, err end

    return true
end

local function validate_from_dockerfile(dockerfile_content)
    local from_line

    for _, line in ipairs(dockerfile_content:split('\n')) do
        line = line:strip()
        -- skip comments and empty lines
        if not (line == '' or line:startswith('#')) then
            if not line:strip():lower():startswith('from') then
                die('Base Dockerfile should be started with `FROM centos:8`')
            end

            from_line = line:strip()
            break
        end
    end

    if from_line == nil then
        return false, 'Base Dockerfile should be started with `FROM centos:8`'
    end

    if from_line:lower() ~= 'from centos:8' then
        return false, 'The base image must be centos:8'
    end

    return true
end

local function tarantool_repo_version()
    local parts = _TARANTOOL:split('.')
    return string.format('%s_%s', parts[1], parts[2])
end

local function construct_dockerfile(filepath, from)
    local expand_params = {
        name = pack_state.name,
        instance_name = '${"$"}{TARANTOOL_INSTANCE_NAME:-default}',
        workdir = fio.pathjoin('/var/lib/tarantool/', pack_state.name),
        dir = fio.pathjoin('/usr/share/tarantool/', pack_state.name),
        prebuild_script_name = PREBUILD_SCRIPT_NAME,
        postbuild_script_name = POSTBUILD_SCRIPT_NAME,
    }

    if pack_state.deprecated_flow then
        expand_params.prebuild_script_name = DEP_PREBUILD_SCRIPT_NAME
    end

    if pack_state.tarantool_is_enterprise then
        local tnt_version_filepath = fio.pathjoin(get_tarantool_dir(), 'VERSION')
        local tnt_version, err = read_file(tnt_version_filepath)
        if tnt_version == nil then return false, err end

        local sdk_version = string.match(tnt_version, 'TARANTOOL_SDK=(%S+)\n')
        if sdk_version == nil then
            return false, string.format('Failed to get SDK version from %s file', tnt_version_filepath)
        end
        sdk_version = sdk_version:gsub('-macos', '')

        expand_params.install_tarantool = DOCKER_INSTALL_ENTERPRISE_TARANTOOL_TEMPLATE
        expand_params.sdk_version = sdk_version
    else
        expand_params.install_tarantool = DOCKER_INSTALL_OPENSOURCE_TARANTOOL_TEMPLATE

        expand_params.tarantool_repo_version = tarantool_repo_version()
    end

    -- dockerfile tail is expanded separately to prevent errors
    -- in case of using environment variables in from Dockerfile
    local dockerfile_content = string.format(
        '%s\n\n%s', from, expand(DOCKERFILE_TAIL_TEMPLATE, expand_params)
    )
    local ok, err = write_file(filepath, dockerfile_content)
    if not ok then return false, err end

    return true
end

local function pack_docker(opts)
    opts = opts or {}

    local docker = which('docker')
    if docker == nil then
        return false, "docker binary is required to pack docker image"
    end

    local from = DOCKERFILE_FROM_DEFAULT
    if pack_state.from ~= nil then
        if not fio.path.exists(pack_state.from) then
            die('Specified base dockerfile does not exists: %s', pack_state.from)
        end

        info('Detected base Dockerfile %s', pack_state.from)

        local dockerfile_content, err = read_file(pack_state.from)
        if dockerfile_content == nil then return false, err end

        local ok, err = validate_from_dockerfile(dockerfile_content)
        if not ok then return false, err end

        info('Base Dockerfile is OK')
        from = dockerfile_content
    end

    info("Packing docker in: %s", pack_state.build_dir)

    local distribution_dir = fio.pathjoin(pack_state.build_dir, pack_state.name)

    local ok, err = form_distribution_dir(distribution_dir)
    if not ok then return false, err end

    local ok, err = generate_version_file(distribution_dir)
    if not ok then return false, err end

    local dockerfile_path = fio.pathjoin(pack_state.build_dir, 'Dockerfile')
    local ok, err = construct_dockerfile(dockerfile_path, from)
    if not ok then return false, err end

    local image_fullname
    if opts.tag ~= nil then
        image_fullname = opts.tag
    else
        image_fullname = string.format(
            '%s:%s',
            pack_state.name,
            pack_state.version_release
        )
    end
    info('Building docker image: %s', image_fullname)

    local download_token_arg = ''
    if pack_state.tarantool_is_enterprise then
        download_token_arg = string.format('--build-arg DOWNLOAD_TOKEN=%s', pack_state.download_token)
    end

    local ok, docker_build_err = call(
        "cd %s && docker build -t %s -f %s %s %s .",
        distribution_dir, image_fullname,
        dockerfile_path, download_token_arg, pack_state.docker_build_args
    )
    if not ok then
        return false, "Failed to create application image: %s", docker_build_err
    end

    info('Resulting image tagged as: %s', image_fullname)

    return true
end

-- * -------------------------- Packing handlers -------------------------

local pack_handlers = {
    [distribution_types.TGZ] = pack_tgz,
    [distribution_types.ROCK] = pack_rock,
    [distribution_types.RPM] = pack_rpm,
    [distribution_types.DEB] = pack_deb,
    [distribution_types.DOCKER] = pack_docker,
}

-- * -------------------------- Packing helpers --------------------------

local function check_if_deprecated_build_flow_is_ised(app_path)
    local dep_build_flow_files = {
        fio.pathjoin(app_path, DEP_IGNORE_FILE_NAME),
        fio.pathjoin(app_path, DEP_PREBUILD_SCRIPT_NAME)
    }

    local new_build_flow_files = {
        fio.pathjoin(app_path, PREBUILD_SCRIPT_NAME),
        fio.pathjoin(app_path, POSTBUILD_SCRIPT_NAME)
    }

    local deprecated_build_flow_is_ised = false
    local forbidden_files = dep_build_flow_files

    -- check if deprecated build flow files are exists
    for _, f in ipairs(dep_build_flow_files) do
        if fio.path.exists(f) then
            deprecated_build_flow_is_ised = true
            forbidden_files = new_build_flow_files
            break
        end
    end

    for _, f in ipairs(forbidden_files) do
        if fio.path.exists(f) then
            die(
                "You use deprecated `.cartridge.ignore + .cartridge.pre` files " ..
                       "and `cartridge.pre-build + cartridge.post-build` files at the same time`. " ..
                "You can use any of these approaches (just take care not to mix them)."
            )
        end
    end

    return deprecated_build_flow_is_ised
end

-- * ------------------- Build dir --------------------

local function detect_and_create_build_dir(app_dir)
    -- By default, application is built in the <app_dir>/build.cartridge.
    -- User can specify build directory in CARTRIDGE_BUILDDIR env variable.
    -- There are two cases:
    -- - specified directory doesn't exists: we just create it and remove after the build
    -- - directory already exists:
    --   - ${CARTRIDGE_BUILDDIR}/build.cartridge will be the build directory
    --   - sub-directory build.cartridge is (re)created and used for application build
    --   - after the build, ${CARTRIDGE_BUILDDIR}/build.cartridge  is removed

    local specified_build_dir = os.getenv('CARTRIDGE_BUILDDIR')

    local build_dir
    if specified_build_dir == nil then
        local build_dir_name = string.format(
            BUILD_DIRECTORY_NAME_TEMPLATE,
            digest.urandom(8):hex()
        )
        build_dir = fio.pathjoin(CARTRIDGE_TMP_PATH, build_dir_name)
    else
        specified_build_dir = fio.abspath(specified_build_dir)
        -- specified build directory can't be project subdirectory
        if is_subdirectory(specified_build_dir, app_dir) then
            die("Build directory can't be project subdirectory, specified: %s", specified_build_dir)
        end

        if not fio.path.exists(specified_build_dir) then
            build_dir = specified_build_dir
        else
            -- This little hack is used to prevent deletion user files
            -- from specified build directory on cleanup.
            -- Moreover, this subdirectory is defenitely clean,
            -- so we wouldn't have any problems
            if not fio.path.is_dir(specified_build_dir) then
                die("Specified build directory is not a directory: %s", specified_build_dir)
            end

            build_dir = fio.pathjoin(specified_build_dir, DEFAULT_BUILD_DIRECTORY_NAME)
        end
    end

    info('Build directory is set to %s', build_dir)

    if fio.path.exists(build_dir) then
        info('Build irectory is already exists. Cleanning it.')
        remove_by_path(build_dir)
    end

    make_tree(build_dir)

    return build_dir
end

local function remove_build_dir()
    info('Remove build directory %s', pack_state.build_dir)
    local ok, err = remove_by_path(pack_state.build_dir)
    if not ok then
        warn('Failed to clean up build directory %s: %s', pack_state.build_dir, err)
    end
end

-- * --------------- Application packing ---------------

local function check_pack_state(state)
    local required_params = {
        'path', 'name', 'version', 'release', 'version', 'version_release',
        'dest_dir', 'deprecated_flow', 'tarantool_is_enterprise', 'build_dir',
    }

    for _, p in ipairs(required_params) do
        if state[p] == nil then
            local err = string.format('Missed reqiured pack_state parameter: %s', p)
            return false, err
        end
    end

    return true
end

local function app_pack(args)
    if not fio.path.exists(args.path) then
        die("Specified path %s doesn't exist", args.path)
    end

    if not fio.path.is_dir(args.path) then
        die("Specified path %s is not a directory", args.path)
    end

    local name, version, release = detect_name_version_release(args.path, args.name, args.version)

    -- collect general application info
    pack_state.path = fio.abspath(args.path)
    pack_state.name = name
    pack_state.version = version
    pack_state.release = release
    pack_state.version_release = string.format('%s-%s', version, release)

    -- collect pack-specific application info
    pack_state.dest_dir = fio.abspath('.')
    pack_state.from = args.from
    pack_state.download_token = args.download_token
    pack_state.docker_build_args = args.docker_build_args
    pack_state.deprecated_flow = check_if_deprecated_build_flow_is_ised(pack_state.path)
    pack_state.tarantool_is_enterprise = tarantool_is_enterprise()
    pack_state.build_dir = detect_and_create_build_dir(pack_state.path)

    local ok_state, err_state = check_pack_state(pack_state)
    if not ok_state then
        die(
            "Whoops! It looks like something is wrong with this version of Cartridge CLI. " ..
            "Please, report a bug on https://github.com/tarantool/cartridge-cli/issues/new. " ..
            "The error is: %s.", err_state
        )
    end

    local instantiated_unit_template
    if args.instantiated_unit_template then
        local err
        instantiated_unit_template, err = read_file(args.instantiated_unit_template)
        if instantiated_unit_template == nil then return false, err end
    end

    local unit_template
    if args.unit_template then
        local err
        unit_template, err = read_file(args.unit_template)
        if unit_template == nil then return false, err end
    end

    if pack_state.deprecated_flow then
        warn(
            "Using `.cartridge.ignore` and `.cartridge.pre` files is deprecated in 1.3.0 " ..
                "and will be removed in 2.0.0"
        )

        if args.type ==  distribution_types.DOCKER then
            die(
                "Using `.cartridge.ignore` and `.cartridge.pre` files is forbidden for " ..
                    "`docker` distribution type. " ..
                "Please, use `cartridge.pre-build` + `cartridge.post-build` approach instead."
            )
        end
    end

    local opts
    if array_contains({distribution_types.RPM, distribution_types.DEB}, args.type) then
        opts = {
            unit_template = unit_template,
            instantiated_unit_template = instantiated_unit_template
        }
    elseif args.type == distribution_types.DOCKER then
        opts = {
            tag = args.tag,
            from = args.from,
            download_token = args.download_token,
            docker_build_args = args.docker_build_args,
        }
    end

    local pack_handler = pack_handlers[args.type]
    if pack_handler == nil then
        local handler_err = string.format("Pack handler for %s distribution type not found", args.type)
        die(
            "Whoops! It looks like something is wrong with this version of Cartridge CLI. " ..
            "Please, report a bug on https://github.com/tarantool/cartridge-cli/issues/new. " ..
            "The error is: %s.", handler_err
        )
    end

    local ok_pack, err_pack = pack_handler(opts)
    if not ok_pack then
        warn('Failed to pack application')
        remove_build_dir()
        die('Failed to pack application: %s', err_pack)
    end

    -- clean build directory
    remove_build_dir()
    info('Packing application succeded!')
end

local function app_pack_parse(arg)
    local args = {}

    local parameters = argparse(
            arg, {
                { 'name', 'string' },
                { 'version', 'string' },
                { 'instantiated_unit_template', 'string' },
                { 'unit_template', 'string' },
                { 'download_token', 'string'},
                { 'tag', 'string' },
                { 'from', 'string' },
            }
    )

    args.name = parameters.name
    args.version = parameters.version
    args.unit_template = parameters.unit_template
    args.instantiated_unit_template = parameters.instantiated_unit_template
    args.download_token = parameters.download_token or os.getenv('TARANTOOL_DOWNLOAD_TOKEN')
    args.docker_build_args = os.getenv('TARANTOOL_DOCKER_BUILD_ARGS') or ''
    args.tag = parameters.tag
    args.from = parameters.from
    args.type = parameters[1]
    args.path = parameters[2]

    if not array_contains(available_distribution_types, args.type) then
        die("Package type should be one of: %s",
                table.concat(available_distribution_types, ', '))
    end

    if pack_state.tarantool_is_enterprise and args.type == distribution_types.DOCKER then
        if not args.download_token then
            die(
                'Tarantool download token is required to pack enterprise Tarantool app in docker. ' ..
                'Please, specify it using --download_token option or TARANTOOL_DOWNLOAD_TOKEN env variable'
            )
        end
    end

    if args.type ==  distribution_types.DOCKER then
        if args.version ~= nil and args.tag ~= nil then
            die(
                'You can specify only one of --version and --tag options. ' ..
                'Run `cartridge pack --help` for details.'
            )
        end

        if args.from == nil then
            local default_dockerfile_path = fio.pathjoin(args.path, 'Dockerfile.cartridge')
            if fio.path.exists(default_dockerfile_path) then
                args.from = default_dockerfile_path
            end
        end
    end

    if args.path == nil then
        die("Path to application is required")
    end

    return args
end


local function app_pack_usage()
    print(string.format("Usage: %s pack [--name <name>] [<type>] [<path>]\n", self_name))

    print("Arguments")
    print("   type                                           Distribution type to create (rpm, tgz, rock, deb, docker)")
    print("   path                                           Directory with app source code")
    print()

    print("Options:")
    print("   --name <name>                                  Name of the app to pack")
    print("   --version <version>                            App version")
    print()

    print("Options specific for rpm and deb types:")
    print("   --unit_template <path to file>                 Path to the template for systemd unit file")
    print("   --instantiated_unit_template <path to file>    Path to the template for systemd instantiated unit file")
    print()

    print("Options specific for docker type:")
    print("   --tag <tag>                                    Resulting image tag")
    print("   --download_token <download_token>              Tarantool Enterprise download token")
    print()

    print("Docker image is tagged:")
    print("    <name>:<detected_version>     By default")
    print("    <name>:<version>              If --version parameter is specified")
    print("    <tag>                         If --tag parameter is specified")
    print("<name> can be specified in --name parameter, otherwise it will be auto-detected from application rockspec.")
    print()

    print(
        "If you use Tarantool Enterprise, it's required to specify a Tarantool Enterprise download token. " ..
        "You can also specify it in TARANTOOL_DOWNLOAD_TOKEN environment variable " ..
        "(has lower priority than --download_token option)"
    )
    print("You can pass additional arguments to `docker build` command using TARANTOOL_DOCKER_BUILD_ARGS env variable.")
end

-- * ---------------- Application templating ----------------

local GITIGNORE = [[
.rocks
.swo
.swp
CMakeCache.txt
CMakeFiles
cmake_install.cmake
*.dylib
*.idea
__pycache__
*pyc
.cache
.pytest_cache
.vagrant
.DS_Store
*.xlog
*.snap
*.rpm
*.deb
*.tar.gz
node_modules
/tmp/*
!/tmp/.keep
]]

local function instantiate_template(template_dir, dest_dir, app_name)
    local files, err = find_files(template_dir)
    if files == nil then return false, err end

    local context = {project_name=app_name,
                     project_name_lower=string.lower(app_name)}

    for _, file in ipairs(files) do
        local srcname = fio.pathjoin(template_dir, file)
        local content = expand(read_file(srcname), context)

        local mode = fio.stat(srcname).mode
        local destname = fio.pathjoin(dest_dir, expand(file, context))
        local destdir = fio.dirname(destname)

        if not fio.path.exists(destdir) then
            local ok, err = make_tree(destdir)
            if not ok then return false, err end
        end

        local ok, err = write_file(destname, content, mode)
        if not ok then return false, err end
    end

    return true
end

local function create_app_directory_and_init_git(dest_dir, template, name)
    assert(fio.path.exists(dest_dir))

    local template_dir = fio.pathjoin(get_template_dir(), template)

    if not fio.path.exists(template_dir) then
        return false, string.format("Template '%s' doesn't exist", template_dir)
    end

    local ok, err = instantiate_template(template_dir, dest_dir, name)
    if not ok then
        return false, string.format('Failed to instantiate application template: %s', err)
    end

    local git = which('git')

    if git ~= nil then
        info("Initializing git repo in: %s", dest_dir)

        repeat  -- until true
            local ok, err = call("cd %s && %s init .", dest_dir, git)
            if not ok then
                warn('Failed to initialize git repo: %s', err)
                break
            end

            local ok, err = write_file(fio.pathjoin(dest_dir, '.gitignore'), GITIGNORE)
            if not ok then
                warn('Failed to create .gitignore file: %s', err)
                break
            end

            local ok, err = call("cd %s && %s add -A", dest_dir, git)
            if not ok then
                warn('Failed to add files to git: %s', err)
                break
            end

            local ok, err = call('cd %s && %s commit -m "Initial commit"', dest_dir, git)
            if not ok then
                warn('Failed to create initial commit: %s', err)
                break
            end

            local ok, err = call('cd %s && %s tag 0.1.0', dest_dir, git)
            if not ok then
                warn('Failed to create initial commit: %s', err)
                break
            end
        until true
    else
        warn("git not found. You'll need to add the app "..
                  "to version control yourself later.")
    end

    return true
end

local function app_create(args)
    local path = args.path or "."

    if not fio.path.exists(path) then
        die("Directory doesn't exist: '%s'", path)
    end

    local name = args.name
    if name == nil then
        name = prompt("Enter project name", "myproject")
    end

    local template = args.template or 'cartridge'
    local dest_dir = fio.pathjoin(path, name)

    if fio.path.exists(dest_dir) then
        die("Can't create app: directory '%s' already exists", dest_dir)
    end

    local ok, err = make_tree(dest_dir)
    if not ok then
        die("Failed to create application directory: %s", err)
    end

    local ok_create, err_create = create_app_directory_and_init_git(dest_dir, template, name)
    if not ok_create then
        warn("Failed to create application...")

        -- clean application directory
        info('Clean destination sirectory %s', dest_dir)
        local ok, err = remove_by_path(dest_dir)
        if not ok then
            warn('Failed to clean up  destination sirectory %s: %s', dest_dir, err)
        end

        die('Failed to create application: %s', err_create)
    end

    info("Application successfully created in '%s'", dest_dir)
end

local function app_create_parse(arg)
    local args = {}

    local parameters = argparse(
        arg,
        {{'name',     'string'},
            {'template', 'string'}}
    )

    args.name = parameters.name
    args.template = parameters.template
    args.path = parameters[1]

    return args
end


local function app_create_usage()
    print(string.format("Usage: %s create [--name <name>] [<path>]\n", self_name))

    print("Arguments")
    print("   path                   Directory to create the app in\n")

    print("Options:")
    print("   --name <name>          Name of the app to create")
    print("   --template <template>  Name of template to use")
end

-- * ---------------- Instance management ----------------

local cmd_start = {
    name = 'start',
    doc = 'Start a Tarantool instance(s)',
    usage = ([=[
        %s start [APP_NAME[.INSTANCE_NAME]] [options]

        Default APP_NAME is is parsed from ./*.rockspec filename.
        When INSTANCE_NAME is not provided it reads `cfg` file and starts all
        defined instances.

        Options
            --script FILE       Application's entry point.
                                Default to TARANTOOL_SCRIPT,
                                or ./init.lua when running from app's directory,
                                or :apps_path/:app_name/init.lua in multi-app env.

            --apps_path PATH    Path to apps direcrory when running in multi-app env.
                                Default to /usr/share/tarantool

            --run_dir DIR       Directory with pid and sock files
                                Default to TARANTOOL_RUN_DIR or /var/run/tarantool

            --cfg FILE          Cartridge instances config file.
                                Default to TARANTOOL_CFG or ./instances.yml

            --daemonize / -d    Start in background

        Default options can be overriden in ./.cartridge.yml or ~/.cartridge.yml,
        also options from .cartridge.yml can be overriden by corresponding to
        them environment variables TARANTOOL_*.
    ]=]):format(self_name):gsub('(\n?)' .. (' '):rep(8), '%1'),
}

-- Fetches app_name from .rockspec file.
-- Extracted from cartridge.argparse, but searches for rockspec only in current
-- directory.
local function get_app_name_from_rockspec()
    local rockspecs = fio.glob('*-scm-1.rockspec')
    if #rockspecs == 1 then
        return string.match(fio.basename(rockspecs[1]), '^(%g+)%-scm%-1%.rockspec$')
    end
end

local function read_cartridge_defaults()
    local cwd = fio.cwd()
    local defaults = {
        cfg = fio.pathjoin(cwd, 'instances.yml'),
        run_dir = '/var/run/tarantool',
        apps_path = '/usr/share/tarantool',
    }

    local from_file = {}
    local paths = {
        '.cartridge.yml',
        fio.pathjoin(os.getenv('HOME') or '',  '.cartridge.yml'),
    }
    for _, path in pairs(paths) do
        if fio.stat(path) then
            from_file = yaml.decode(read_file(path))
            break
        end
    end

    local env_vars = {
        cfg = os.getenv('TARANTOOL_CFG'),
        script = os.getenv('TARANTOOL_SCRIPT'),
        run_dir = os.getenv('TARANTOOL_RUN_DIR'),
        apps_path = os.getenv('TARANTOOL_APPS_PATH')
    }
    return fun.chain(defaults, from_file, env_vars):tomap()
end

function cmd_start.parse(args)
    local result = argparse(args, {
        {'script', 'string'},
        {'apps_path', 'string'},
        {'run_dir', 'string'},
        {'cfg', 'string'},
        'daemonize', 'd',
        'verbose', -- Do not close standard FDs for child process. Private flag for debugging.
    })
    if result.daemonize == nil then result.daemonize = result.d end
    local defaults = read_cartridge_defaults()
    for k, v in pairs(defaults) do
        result[k] = result[k] or v
    end

    local instance_id = (result[1] or ''):split('.')
    local app_name = get_app_name_from_rockspec()
    result.app_name = #instance_id[1] > 0 and instance_id[1] or app_name
    assert(result.app_name and #result.app_name > 0, 'APP_NAME is required')
    result.instance_name = instance_id[2]

    if result.script == nil then
        if app_name then -- cartridge is called inside app directory
            result.script = 'init.lua'
        else
            result.script = fio.pathjoin(result.apps_path, result.app_name, 'init.lua')
        end
    end
    result.script = fio.abspath(result.script)

    result.cfg = fio.abspath(result.cfg)
    result.run_dir = fio.abspath(result.run_dir)
    return result
end

function cmd_start.finalize_args(args)
    assert(args.instance_name and #args.instance_name > 0, 'INSTANCE_NAME is required')
    local basename = args.app_name .. '.' .. args.instance_name
    args.pid_file = fio.pathjoin(args.run_dir,  basename .. '.pid')
    args.console_sock = fio.pathjoin(args.run_dir, basename .. '.sock')
end

ffi.cdef([[
    pid_t fork(void);
    int execve(const char *pathname, char *const argv[], char *const envp[]);
    int dup2(int oldfd, int newfd);
    int fileno(struct FILE *stream);
    int kill(pid_t pid, int sig);

    int open(const char *path, int oflag, ...);
    int close(int fildes);

    int pipe(int fildes[2]);
]])

local function ffi_table_to_const_char(input)
    local result = ffi.new('char const*[?]', #input + 1, input)
    result[#input] = nil
    return ffi.cast('char *const*', result)
end

-- Starts process and returns immediately, not waiting until process is finished.
-- @param path Executable path.
-- @param[opt] args
-- @param[opt] env
local function execve(path, args, env)
    path = fio.abspath(path)
    args = args or {}
    env = env or {}
    table.insert(args, 1, path)
    local argv = ffi_table_to_const_char(args)
    local env_list = fun.iter(env):map(function(k, v) return k .. '=' .. v end):totable()
    local envp = ffi_table_to_const_char(env_list)
    ffi.C.execve(path, argv, envp)
    io.stderr:write('execve failed: ' .. path ..  ' - ' .. errno.strerror() .. '\n')
    os.exit(1)
end

local function check_pid_running(pid)
    return ffi.C.kill(tonumber(pid), 0) == 0
end

local function read_configuration(path)
    if fio.path.is_dir(path) then
        local configs = fun.chain(
            fio.glob(fio.pathjoin(path, '*.yml')),
            fio.glob(fio.pathjoin(path, '*.yaml'))
        ):map(function(x) return read_configuration(x) end):totable()
        return fun.chain(unpack(configs)):tomap()
    else
        return yaml.decode(read_file(path))
    end
end

-- Read configuration at `path` and fetch instance names.
local function get_configured_isntances(path, app_name)
    local config = read_configuration(path)
    local result = {}
    for name, _ in pairs(config) do
        -- instance id must be `app_name.instance_name`
        local parts = name:split('.', 1)
        if #parts == 2 and parts[1] == app_name then
            table.insert(result, parts[2])
        end
    end
    assert(#result > 0, 'No configured instances found for app ' .. app_name)
    return result
end

local function start_all(args)
    local instance_names = get_configured_isntances(args.cfg, args.app_name)
    for _, instance_name in pairs(instance_names) do
        local instance_args = table.copy(args)
        instance_args.instance_name = instance_name
        cmd_start.callback(instance_args)
    end
end

local Process = {}

-- Runs tarantool script with several enforced env vars.
-- If `daemonize` option is set then new processes are started in background.
--
-- Otherwise it creates UPD socket and passes it's name in NOTIFY_SOCKET
-- to the forked instance. This makes it possible to wait until child process
-- is successfully bootstraped: after tarantool executes main script
-- it sends `READY=1` message to the NOTIFY_SOCKET.
--
-- It also creates pid file, because app does not create it until box.cfg is called.
-- However it does not lock the file to let box.cfg lock and overwrite it.
function cmd_start.callback(args)
    if args.instance_name == nil then
        args.multiple = true
        return start_all(args)
    end

    cmd_start.finalize_args(args)

    log.info('Starting %s...', args.instance_name)
    local process = Process:new(args)
    process:check_pid_file()

    if args.daemonize then
        process:start_and_wait()
    elseif args.multiple then
        process:start_with_decorated_output()
    else
        process:start_in_foreground()
    end
end

function Process:new(object)
    setmetatable(object, self)
    self.__index = self
    object:initialize()
    return object
end

function Process:initialize()
    fio.mktree(self.run_dir)

    self.env = table.copy(os.environ())
    self.env.TARANTOOL_APP_NAME = self.app_name
    self.env.TARANTOOL_INSTANCE_NAME = self.instance_name
    self.env.TARANTOOL_CFG = self.cfg
    self.env.TARANTOOL_PID_FILE = self.pid_file
    self.env.TARANTOOL_CONSOLE_SOCK = self.console_sock
end

function Process:check_pid_file()
    if fio.stat(self.pid_file) then
        local pid = tonumber(read_file(self.pid_file))
        if pid == nil or pid <= 0 then
            error('Pid file exists with unknown format: ' .. self.pid_file)
        elseif check_pid_running(pid) then
            error('Process is already running with pid_file: ' .. self.pid_file)
        else
            assert(fio.unlink(self.pid_file))
        end
    end
end

function Process:start_in_foreground()
    write_file(self.pid_file, require('tarantool').pid(), tonumber('644', 8))
    execve(arg[-1], {self.script}, self.env) -- stops execution
end

function Process:build_notify_socket()
    local sock = assert(socket('AF_UNIX', 'SOCK_DGRAM', 0), 'Can not create socket')
    local basename = self.app_name .. '.' .. self.instance_name .. '-notify.sock'
    local sock_name = fio.pathjoin(self.run_dir, basename)
    if fio.stat(sock_name) then
        assert(fio.unlink(sock_name))
    end
    local ok = sock:bind('unix/', sock_name)
    assert(ok, sock:error())
    fio.chmod(sock_name, tonumber('0666', 8))
    self.notify_socket = sock
    self.env.NOTIFY_SOCKET = sock_name
end

function Process:start()
    local pid = tonumber(ffi.C.fork())
    if pid == -1 then
        error('fork failed: ' .. errno.strerror())
    elseif pid == 0 then
        if not self.verbose then
            local null_fd = ffi.C.open('/dev/null', fio.c.flag.O_RDONLY)
            if null_fd == -1 then
                io.stdout:write('Failed to open /dev/null\n')
                os.exit(1)
            end
            ffi.C.dup2(null_fd, ffi.C.fileno(io.stdout))
            ffi.C.dup2(null_fd, ffi.C.fileno(io.stderr))
            if ffi.C.close(null_fd) == -1 then
                os.exit(1)
            end
        end
        execve(arg[-1], {self.script}, self.env)
    end
    self.pid = pid
    write_file(self.pid_file, pid, tonumber('644', 8))
end

function Process:start_and_wait()
    self:build_notify_socket()
    self:start()
    self:wait_started()
end

Process.POLL_SOCKET_TIMEOUT = 1 -- sec.
function Process:wait_started()
    while true do
        if not check_pid_running(self.pid) then
            fio.unlink(self.env.NOTIFY_SOCKET)
            error('Child process exited unexpectedly')
        end
        -- check that child process is still alive
        if self.notify_socket:readable(self.POLL_SOCKET_TIMEOUT) then
            local str = self.notify_socket:recv()
            if str:match('READY=1') then
                fio.unlink(self.env.NOTIFY_SOCKET)
                return
            elseif not (str:find('^STATUS=running$') or str:find('^STATUS=loading$')) then
                log.info(str)
            end
        end
    end
end

Process.PIPE_READ_BUFFER_SIZE = 4096

-- Read fd into chunks array while it's readable.
local function read_fd(fd, chunks)
    chunks = chunks or {}
    local buffer = nil
    while socket.iowait(fd, 'R', 0) ~= '' do
        buffer = buffer or ffi.new('char[?]', Process.PIPE_READ_BUFFER_SIZE)
        local count = ffi.C.read(fd, buffer, Process.PIPE_READ_BUFFER_SIZE)
        if count < 0 then
            error('read pipe failed')
        end
        table.insert(chunks, ffi.string(buffer, count))
    end
    return chunks
end

-- Returns new color code for line or nil if it should not be changed.
local function color_for_line(line)
    local mark = line:match(ERROR_LOG_LINE_PATTERN)
    return mark and COLOR_CODE_BY_LOG_LEVEL[mark]
end

-- Returns function which reads all data from fd and print each line with prefix.
-- Every line with log level mark (` X> `) changes the color for all following
-- lines until the next one with the mark.
local function fd_forwarder_fn(fd, prefix)
    local line_color_code = RESET_TERM
    return function()
        local chunks = read_fd(fd)
        if #chunks > 0 then
            local lines = table.concat(chunks):split('\n')
            if lines[#lines] == '' then
                table.remove(lines)
            end
            for _, line in pairs(lines) do
                line_color_code = color_for_line(line) or line_color_code
                io.stdout:write(table.concat({prefix, line_color_code, line, '\n'}))
            end
        end
    end
end

-- Takes multiple pipes and prefix string.
-- It reads from pipes' outputs and prints lines prefixed with given value.
-- Prefix is colored with color_code (which is random by default) and
-- error lines are printed in red.
local function log_pipes_forwarder(pipes, prefix, color_code)
    color_code = color_code or next_color_code()
    local forwarders = {}
    for _, pipe in pairs(pipes) do
        table.insert(forwarders, fd_forwarder_fn(pipe[0], color_code .. prefix .. ' | '))
    end
    while true do
        for _, forwarder in pairs(forwarders) do
            forwarder()
        end
        fiber.testcancel()
        fiber.sleep(0.5)
    end
end

function Process:start_with_decorated_output()
    local pipes = {stdout = ffi.new('int[?]', 2), stderr = ffi.new('int[?]', 2)}
    if ffi.C.pipe(pipes.stdout) ~= 0 or ffi.C.pipe(pipes.stderr) ~= 0 then
        error('pipe call failed')
    end
    local pid = tonumber(ffi.C.fork())
    if pid == -1 then
        error('fork failed: ' .. errno.strerror())
    elseif pid == 0 then
        ffi.C.dup2(pipes.stdout[1], ffi.C.fileno(io.stdout))
        ffi.C.dup2(pipes.stderr[1], ffi.C.fileno(io.stderr))
        execve(arg[-1], {self.script}, self.env)
    end
    self.pid = pid
    write_file(self.pid_file, pid, tonumber('644', 8))
    fiber.create(log_pipes_forwarder, pipes, self.instance_name)
end

local cmd_stop = {
    name = 'stop',
    doc = 'Stop a Tarantool instance(s)',
    usage = ([=[
        %s stop [APP_NAME[.INSTANCE_NAME]] [options]

        When INSTANCE_NAME is not provided it reads `cfg` file and stops all
        defined instances.

        These options from `start` command are supported
            --run_dir DIR
            --cfg FILE
    ]=]):format(self_name):gsub('(\n?)' .. (' '):rep(8), '%1'),
    parse = cmd_start.parse,
}

local function stop_all(args)
    local instance_names = get_configured_isntances(args.cfg, args.app_name)
    for _, instance_name in pairs(instance_names) do
        local instance_args = table.copy(args)
        instance_args.instance_name = instance_name
        cmd_stop.callback(instance_args)
    end
end

function cmd_stop.callback(args)
    if args.instance_name == nil then
        return stop_all(args)
    end

    cmd_start.finalize_args(args)
    log.info('Stopping %s...', args.instance_name)

    local pid_file = args.pid_file
    if fio.stat(pid_file) == nil then
        log.error('Process is not running (pid_file: %s)', pid_file)
        return
    end

    local pid = tonumber(read_file(pid_file))
    if pid == nil or pid <= 0 then
        log.error('Broken pid file %s. Check it and remove manually if required.', pid_file)
        os.exit(1)
    end

    if not check_pid_running(pid) then
        log.error('Process is not running, removing stale pid_file (%s)', pid_file)
        assert(fio.unlink(pid_file))
        return
    end

    if os.execute('ps -p ' .. pid .. ' | grep tarantool > /dev/null') ~= 0 then
        log.error('Process %d does not seem to be tarantool. Skipping.', pid, errno.strerror())
        os.exit(1)
    end

    if ffi.C.kill(pid, 15) < 0 then
        log.error('Can not kill process %d: %s', pid, errno.strerror())
        os.exit(1)
    end

    -- Don't remove pid_file until process is terminated to prevent warnings
    -- from tarantool trying to remove absent pid file.
    while check_pid_running(pid) do
        fiber.sleep(0.1)
    end
    if fio.stat(pid_file) then
        assert(fio.unlink(pid_file))
    end
    if fio.stat(args.console_sock) then
        assert(fio.unlink(args.console_sock))
    end
end

-- * ---------------- Processing commands ----------------

local commands = {
    {
        name = "create",
        doc = "Create a new app from template",
        callback = app_create, parse = app_create_parse, usage = app_create_usage,
    },
    {
        name = "pack",
        doc = "Pack application into a distributable bundle",
        callback = app_pack, parse = app_pack_parse, usage = app_pack_usage,
    },
    cmd_start,
    cmd_stop,
}

-- * ---------------- Entry point ----------------

local function print_usage()
    print(string.format("Usage: %s [--help] <command> [<args>]\n", self_name))

    print("Supported commands:")
    for _, command in pairs(commands) do
        print(string.format("\t%s\t%s", command.name, command.doc))
    end
end

local function main()
    if #arg < 1 then
        print_usage()
        os.exit(1)
    end

    if arg[1] == "--version" or arg[1] == "-v" then
        print("Tarantool cartridge-cli v" .. VERSION())
        os.exit(0)
    end

    if arg[1] == "--help" then
        print_usage()
        os.exit(0)
    end

    local command = fun.iter(commands):filter(function(x) return x.name == arg[1] end):totable()[1]
    if command == nil then
        print_usage()
        os.exit(1)
    end

    if array_contains(arg, "--help") then
        if type(command.usage) == "string" then
            print(command.usage)
        else
            command.usage()
        end
        os.exit(0)
    end

    local args = command.parse(array_slice(arg, 2))

    command.callback(args)
end

return {
   matching = matching,
   main = main
}
