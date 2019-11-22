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

local luarocks_searcher = require('luarocks.search')
local luarocks_path = require('luarocks.path')

local self_name = fio.basename(arg[0])

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

local function die(fmt, ...)
    print("error: " .. string.format(fmt, ...))
    os.exit(1)
end

local function warn(fmt, ...)
    print("warning: " .. string.format(fmt, ...))
end

local function info(fmt, ...) -- luacheck: no unused
    print(string.format(fmt, ...))
end


local function read_file(path)
    local file = fio.open(path)
    if file == nil then
        die('Failed to open file %s: %s', path, errno.strerror())
    end
    local buf = {}
    while true do
        local val = file:read(1024)
        if val == nil then
            die('Failed to read from file %s: %s', path, errno.strerror())
        elseif val == '' then
            break
        end
        table.insert(buf, val)
    end
    file:close()
    return table.concat(buf, '')
end

local function write_file(path, data, mode)
    mode = mode or tonumber(644, 8)
    local file = fio.open(path, {'O_CREAT', 'O_WRONLY', 'O_TRUNC', 'O_SYNC'}, mode)
    if file == nil then
        die('Failed to open file %s: %s', path, errno.strerror())
    end

    local res = file:write(data)

    if not res then
        die('Failed to write to file %s: %s', path, errno.strerror())
    end

    file:close()
    return data
end

local function file_md5_hex(filename)
    local data = read_file(filename)

    return digest.md5_hex(data)
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
        for _, file in ipairs(fio.listdir(path) or {}) do
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

local function call(command, ...)
    local res, err = io.popen(string.format(command, ...))

    if res == nil then
        die("Failed to execute '%s': %s", command, err)
    end

    local result = res:read("*all")
    return result
end

local function tarantool_is_enterprise()
    local tarantool_dir = get_tarantool_dir()
    local tnt_version = fio.pathjoin(tarantool_dir, 'VERSION')
    return fio.path.exists(tnt_version)
end

-- * ---------------- Project-related functions ----------------

local function normalize_version(str)
    local patterns = {
        "(%d+)%.(%d+)%.(%d+)",
        "(%d+)%.(%d+)",
        "(%d+)"
    }

    for _, pattern in ipairs(patterns) do
        local major, minor, patch = string.match(str, pattern)

        if major ~= nil then
            minor = minor or '0'
            patch = patch or '0'

            return {major, minor, patch}
        end
    end
end

local function detect_version(source_dir)
    if which('git') == nil then
        return nil
    end

    if not fio.path.exists(fio.pathjoin(source_dir, '.git')) then
        return nil
    end

    local rc, raw_version = pcall(
        call,
        string.format('cd "%s" && git describe --tags --long', source_dir))

    if not rc then
        return nil
    end

    local version, release, commit = string.match(
        string.strip(raw_version), "^(.*)-(%d+)-(%g+)$")

    if version == nil then
        return nil
    end

    if normalize_version(version) == nil then

        print("warning: detected version '%s' ignored, " ..
              "because it doesn't look like proper " ..
              "version (major.minor.patch)", version)
    end

    version = normalize_version(version)

    return version, release, commit
end

local function find_rockspec(source_dir)
    for _, file in ipairs(fio.listdir(source_dir) or {}) do
        if string.endswith(file, '.rockspec') then
            return file
        end
    end
end

local function detect_name(source_dir)
    local rockspec = find_rockspec(source_dir)
    if rockspec ~= nil then
        return string.match(rockspec, '^(%g+)%-scm%-1%.rockspec$')
    end
end

local function detect_name_release_version(source_dir, raw_name, raw_version)
    local name = raw_name
    local release
    local version

    if name == nil then
        name = detect_name(source_dir)

        if name == nil then
            die("Failed to detect project name. Please pass it explicitly " ..
                    "via --name")
        end

        print(string.format("Detected project name: %s", name))
    end

    if raw_version then
        if not normalize_version(raw_version) then
            die("Passed version '%s' should be semantic (major.minor.patch)",
                raw_version)
        end
        version = normalize_version(raw_version)
        release = '0'
    else
        version, release = detect_version(source_dir)
        if version == nil then
            die("Failed to detect version from project in directory '%s'." ..
                    "Please pass it explicitly via --version", source_dir)
        end

        print(string.format("Detected project version: %s-%s",
                            table.concat(version, '.'), release))
    end

    if not fio.path.exists(fio.pathjoin(source_dir, 'init.lua')) then
        die("Application must have `init.lua` in its root directory")
    end

    return name, release, version
end

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
${chown} -R tarantool:tarantool /var/lib/tarantool || :
${chown} -R tarantool:tarantool /usr/share/tarantool/${name} || :
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

-- * ---------------- Generic packing ----------------

local function get_rock_versions(project_dir)
    local query = luarocks_searcher.make_query("")
    query.exact_name = false

    local manifest_path = luarocks_path.rocks_dir(fio.pathjoin(project_dir, '.rocks'))

    local results = {}
    luarocks_searcher.manifest_search(results, manifest_path, query)

    for rock, versions in pairs(results) do
        results[rock] = ({ next(versions) })[1]
    end

    return results
end

local function generate_version_file(source_dir, dest_dir, app_name, app_version)
    local version_file, _ = fio.open(
        fio.pathjoin(dest_dir, 'VERSION'),
        {'O_TRUNC', 'O_WRONLY', 'O_CREAT'},
        tonumber(644, 8)
    )
    if not version_file then
        die("can't create new VERSION file. Version meta information can't be " ..
             "shipped to the resulting package. ")
    end

    if tarantool_is_enterprise() then
        local tarantool_dir = get_tarantool_dir()
        local tnt_version = fio.pathjoin(tarantool_dir, 'VERSION')
        if not fio.path.exists(tnt_version) then
            warn("can't open VERSION file from Tarantool SDK. SDK information can't be " ..
                "shipped to the resulting package. ")
        else
            version_file:write(fio.open(tnt_version):read())
        end
    end

    local _, _, app_commit = detect_version(source_dir)
    version_file:write(string.format("%s=%s-%s\n", app_name, table.concat(app_version, '.'), app_commit or ''))

    local ok, rocks_versions = pcall(get_rock_versions, dest_dir)
    if not ok then
        warn("can't process rocks manifest file. Dependency information can't be " ..
             "shipped to the resulting package. ")
    else
        local flat_rocks_versions = ""
        for rock, version in pairs(rocks_versions) do
            if rock ~= app_name then
                flat_rocks_versions = flat_rocks_versions .. string.format("%s=%s\n", rock, version)
            end
        end

        version_file:write(flat_rocks_versions)
    end

    version_file:close()
end

local function pattern_form(pattern)
    if pattern == '' or-- blank line
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

local function remove_by_path(path)
    if fio.path.is_dir(path) then
        fio.rmtree(path)
    else
        fio.unlink(path)
    end
end

local function remove_ignored(destdir)
    local ignore = fio.pathjoin(destdir, '.cartridge.ignore')
    if not fio.path.exists(ignore) then return end

    local files = find_files(destdir, { include_dirs = true })

    -- formatting all pattern and exclusion exception pattern
    local patterns, exceptions  = {}, {}
    for _, pattern in ipairs(string.split(read_file(ignore), '\n')) do
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
        print(pretty_file)
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
        remove_by_path(fio.pathjoin(destdir, f))
    end
end

local function form_distribution_dir(source_dir, destdir, app_name, app_version)
    assert(fio.copytree(source_dir, destdir))

    if tarantool_is_enterprise() then
        local tarantool_dir = get_tarantool_dir()
        assert(fio.copyfile(fio.pathjoin(tarantool_dir, 'tarantool'),
                            fio.pathjoin(destdir, 'tarantool')))
        assert(fio.copyfile(fio.pathjoin(tarantool_dir, 'tarantoolctl'),
                            fio.pathjoin(destdir, 'tarantoolctl')))
    end

    local rocks_dir = fio.pathjoin(destdir, '.rocks')
    if fio.path.exists(rocks_dir) then
        fio.rmtree(rocks_dir)
    end
    local git = which('git')
    if git ~= nil then
        -- Clean up all files explicitly ignored by git, to not accidentally
        -- ship development snaps, xlogs or other garbage to production.
        call("cd %q && %s clean -f -d -X", destdir, git)
    else
        warn("git not found. It is possible that some of the extra files " ..
                 "normally ignored are shipped to the resulting package. ")
    end

    if fio.path.exists(fio.pathjoin(destdir, '.cartridge.pre')) then
        print("Running .cartridge.pre")
        local ret = os.execute(
            "set -e\n" ..
            string.format("cd %q\n", destdir) ..
            ". ./.cartridge.pre"
        )
        if ret ~= 0 then
            die("Failed to execute pre-build stage")
        end
    end

    local rockspec = find_rockspec(destdir)
    if rockspec ~= nil then
        print("Running tarantoolctl rocks make")
        local ret = os.execute(
            string.format(
                "cd %q; exec tarantoolctl rocks make %q",
                destdir, rockspec
            )
        )
        if ret ~= 0 then
            die("Failed to install rocks")
        end
    end

    -- implicit uses rocks manifest, created after `tarantoolctl rocks make`
    generate_version_file(source_dir, destdir, app_name, app_version)

    -- deleting files matching patterns from .cartridge.ignore
    remove_ignored(destdir)

    remove_by_path(fio.pathjoin(destdir, '.git'))
    remove_by_path(fio.pathjoin(destdir, '.cartridge.pre'))
    remove_by_path(fio.pathjoin(destdir, '.cartridge.ignore'))
end

local function form_systemd_dir(dest_dir, name, opts)
    opts = opts or {}
    local unit_template = opts.unit_template or SYSTEMD_UNIT_FILE
    local instantiated_unit_template = opts.instantiated_unit_template or SYSTEMD_INSTANTIATED_UNIT_FILE

    fio.mktree(fio.pathjoin(dest_dir, '/etc/systemd/system/'))
    fio.mktree(fio.pathjoin(dest_dir, '/var/lib/tarantool/', name))
    fio.mktree(fio.pathjoin(dest_dir, '/usr/lib/tmpfiles.d'))

    local expand_params = {
        name = name,
        dir = fio.pathjoin('/usr/share/tarantool/', name),
        workdir = fio.pathjoin('/var/lib/tarantool/', name),
        mkdir = opts.mkdir,
    }

    if tarantool_is_enterprise() then
        expand_params.bindir = expand_params.dir
    else
        expand_params.bindir = '/usr/bin'
    end

    write_file(
        fio.pathjoin(dest_dir, string.format('/etc/systemd/system/%s.service', name)),
        expand(unit_template, expand_params)
    )
    write_file(
        fio.pathjoin(dest_dir, string.format('/etc/systemd/system/%s@.service', name)),
        expand(instantiated_unit_template, expand_params)
    )

    -- tmpfiles
    write_file(
        fio.pathjoin(dest_dir, string.format('/usr/lib/tmpfiles.d/%s.conf', name)),
        TMPFILES_CONFIG
    )
end

-- * ---------------- TAR.GZ packing ----------------

local function pack_tgz(source_dir, dest_dir, name, release, version)
    local tgz_file_name = string.format(
        "%s-%s-%s.tar.gz",
        name, table.concat(version, '.'), release)
    tgz_file_name = fio.pathjoin(dest_dir, tgz_file_name)

    print("Packing tar.gz file")

    local tar = which('tar')

    if tar == nil then
        die("tar binary is required to pack tar.gz")
    end

    local tmpdir = fio.tempdir()
    local destdir = fio.pathjoin(tmpdir, name)
    fio.mktree(destdir)

    print("Packing tar.gz in: " .. tmpdir)

    form_distribution_dir(source_dir, destdir, name, version)

    local data = call(string.format("cd %s && %s -cvzf - %s",
                                    tmpdir, tar, name))

    write_file(tgz_file_name, data)

    print("Resulting tar.gz saved as: " .. tgz_file_name)
end

-- * ---------------- ROCK packing ----------------

local function pack_rock(source_dir, dest_dir, name, release, version)
    local tmpdir = fio.tempdir()
    local destdir = fio.pathjoin(
        tmpdir, name)
    fio.mktree(destdir)

    dest_dir = fio.abspath(dest_dir)

    print("Packing binary rock in: " .. tmpdir)

    form_distribution_dir(source_dir, destdir)

    fio.chdir(tmpdir)

    local rockspec = find_rockspec(destdir)
    local content = ''
    if rockspec then
        content = read_file(fio.pathjoin(destdir, rockspec))
        content = string.gsub(content, "(.-version%s-=%s-['\"])(.-)(['\"].*)",
                '%1' .. string.format('%s-%s', table.concat(version, '.'), release) .. '%3')
        if not content then
            die('Rockspec %s is not valid! Version not found!')
        end
    end

    local name_of_rockspec = string.format('%s-%s-%s.rockspec', name, table.concat(version, '.'),
                    release)

    local new_rockspec = fio.pathjoin(destdir, name_of_rockspec)

    write_file(new_rockspec, content)

    fio.chdir(destdir)

    local rock_filename = string.format('%s-%s-%s.*.rock', name, table.concat(version, '.'),
                                        release)

    print(call('tarantoolctl rocks pack %s ', new_rockspec))

    rock_filename = fio.glob(fio.pathjoin(destdir, rock_filename))[1]

    local dest_rock_filename = fio.pathjoin(dest_dir, fio.basename(rock_filename))

    fio.copyfile(rock_filename, dest_rock_filename)

    print('Resulting rock saved as: ' .. dest_rock_filename)
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
}

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

local function rpm_get_file_owner(filename)
    -- We know that PREIN script has created user 'tarantool', and
    -- that in order to allow tarantool-based systemd services to
    -- write to their working directory, that directory should be
    -- owned by tarantool. Everything else in RPM package should be
    -- owned by root
    if string.startswith(filename, '/var/lib/tarantool') then
        return 'tarantool', 'tarantool'
    end
    if string.startswith(filename, '/var/run/tarantool') then
        return 'tarantool', 'tarantool'
    end

    return 'root', 'root'
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
        local fileuser, filegroup = rpm_get_file_owner(file)
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
            table.insert(result.fileflags, bit.lshift(1, 4))
            table.insert(result.filedigests, file_md5_hex(fullpath))
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

local function pack_cpio(source_dir, name, version, opts)
    -- The resulting CPIO structure should look like it will be
    -- extracted to /
    -- So it contains /usr/share/tarantool/<app>, socket
    -- directories, working directories, systemd unit files, etc...
    local cpio = which('cpio')

    if cpio == nil then
        die("cpio binary is required to build rpm packages")
    end

    local gzip = which('gzip')

    if gzip == nil then
        die("gzip binary is required to build rpm packages")
    end

    local tmpdir = fio.tempdir()

    opts = opts or {}
    opts.mkdir = '/usr/bin/mkdir'
    form_systemd_dir(tmpdir, name, opts)

    print("Packing CPIO in: " .. tmpdir)

    local destdir = fio.pathjoin(tmpdir, '/usr/share/tarantool/', name)
    form_distribution_dir(source_dir, destdir, name, version)

    local files = find_files(tmpdir, {include_dirs=true, exclude={'.git'}})
    files = filter_out_known_files(files)

    write_file(fio.pathjoin(tmpdir, 'files'), table.concat(files, '\n'))

    call(string.format("cd %s && cat files | %s -o -H newc > unpacked",
                       tmpdir, cpio))
    local payloadsize = fio.stat(fio.pathjoin(tmpdir, 'unpacked')).size
    local archive = call(string.format("cd %s && cat unpacked | %s -9",
                                       tmpdir, gzip))

    call(string.format("rm '%s'", fio.pathjoin(tmpdir, 'unpacked')))
    call(string.format("rm '%s'", fio.pathjoin(tmpdir, 'files')))

    local fileinfo = generate_fileinfo(tmpdir)

    fio.rmtree(tmpdir)

    return archive, fileinfo, payloadsize
end

local function pack_rpm(source_dir, dest_dir, name, release, version, opts)
    opts = opts or {}
    local rpm_file_name = fio.pathjoin(
        dest_dir,
        string.format(
            "%s-%s-%s.rpm",
            name, table.concat(version, '.'), release))

    print("Packing rpm file")
    local lead = gen_lead(name)

    local cpio, fileinfo, payloadsize = pack_cpio(source_dir, name, version, opts)

    local create_user_script_rpm = expand(CREATE_USER_SCRIPT, {
        groupadd = '/usr/sbin/groupadd',
        useradd = '/usr/sbin/useradd',
        mkdir = '/usr/bin/mkdir',
        chown = '/usr/bin/chown',
    })

    local header_tags = {
        {'NAME', 'STRING', name},
        {'VERSION', 'STRING', table.concat(version, '.')},
        {'RELEASE', 'STRING', release},
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
        {'RPMVERSION', 'STRING', '4.11.3'},
        {'SIZE', 'INT32', payloadsize}
    }

    if not tarantool_is_enterprise() then
        --- Append RPM dependency flags for Tarantool
        --- See Dependency Tags section of
        --- - https://docs.fedoraproject.org/ro/Fedora_Draft_Documentation/0.1/html/RPM_Guide/ch-package-structure.html

        local major, minor, patch = unpack(normalize_version(_TARANTOOL))
        local min_version = ('%s.%s.%s'):format(major, minor, patch)
        local max_version = ('%s'):format(tonumber(major) + 1)

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

    local body = header .. cpio
    local md5 = digest.md5(body)
    local sha1 = digest.sha1_hex(header)
    local sig_size = #body
    local signature_header = gen_header(
        {
            {'SHA1', 'STRING', sha1},
            {'SIG_SIZE', 'INT32', sig_size},
            {'PAYLOADSIZE', 'INT32', payloadsize},
            {'MD5', 'BIN', md5},
        },
        SIGNATURE_TAG_TABLE,
        HEADERSIGNATURES
    )

    body = lead .. buf_pad_to_8_byte_boundary(signature_header) .. body

    write_file(rpm_file_name, body)

    print("Resulting rpm saved as: " .. rpm_file_name)
end


-- * ---------------- DEB packing ----------------

-- DEB package is an ar archive contains debian-binary, control.tar.gz and data.tar.gz files
--
-- debian-binary  : contains format version string (2.0)
-- data.tar.xz    : package files
-- control.tar.xz : control files (control, preinst etc.)
--
local function form_deb_control_dir(dest_dir, name, release, version)
    fio.mktree(dest_dir)

    -- control
    local control_filepath = fio.pathjoin(dest_dir, 'control')
    local control_params = {
        name = name,
        version = ('%s-%s'):format(table.concat(version, '.'), release),
        maintainer = 'Tarantool Cartridge Developer',
        arch = 'all',
        desc = ('Tarantool Cartridge app %s'):format(name),
        deps = ''
    }

    if not tarantool_is_enterprise() then
        -- Add tarantool dependency
        local major, minor, patch = unpack(normalize_version(_TARANTOOL))
        local min_version = ('%s.%s.%s'):format(major, minor, patch)
        local max_version = ('%s'):format(tonumber(major) + 1)

        control_params.deps = string.format('tarantool (>= %s), tarantool (<< %s)',
                                            min_version, max_version)
    end

    write_file(
        control_filepath,
        expand(DEBIAN_CONTROL_FILE, control_params)
    )

    -- preinst
    local preinst_filepath = fio.pathjoin(dest_dir, 'preinst')
    write_file(
        preinst_filepath,
        expand(CREATE_USER_SCRIPT, {
            groupadd = '/usr/sbin/groupadd',
            useradd = '/usr/sbin/useradd',
            mkdir = '/bin/mkdir',
            chown = '/bin/chown',
        })
    )
    fio.chmod(preinst_filepath, tonumber('0755', 8))

    -- postinst
    local postinst_filepath = fio.pathjoin(dest_dir, 'postinst')
    write_file(
        postinst_filepath,
        expand(SET_OWNER_SCRIPT, {
            groupadd = '/usr/sbin/groupadd',
            useradd = '/usr/sbin/useradd',
            mkdir = '/bin/mkdir',
            chown = '/bin/chown',
            name = name,
        })
    )
    fio.chmod(postinst_filepath, tonumber('0755', 8))
end

local function pack_deb(source_dir, dest_dir, name, release, version, opts)
    local deb_file_name = string.format(
            "%s-%s-%s.deb",
            name, table.concat(version, '.'), release)

    local tar = which('tar')

    if tar == nil then
        die("tar binary is required to pack deb")
    end

    local ar = which('ar')

    if ar == nil then
        die("ar binary is required to pack deb")
    end

    local tmpdir = fio.tempdir()
    print("Packing deb in: " .. tmpdir)

    -- debian-binary
    local debian_binary_path = fio.pathjoin(tmpdir, 'debian-binary')
    write_file(debian_binary_path, '2.0\n')

    -- control.tar.xz
    local control_dir = fio.pathjoin(tmpdir, 'control')
    local control_tgz_path = fio.pathjoin(tmpdir, 'control.tar.xz')
    form_deb_control_dir(control_dir, name, release, version)

    local control_data = call(string.format("cd %s && %s -cvJf - .", control_dir, tar))
    write_file(control_tgz_path, control_data)

    -- data.tar.xz
    local data_dir = fio.pathjoin(tmpdir, 'data')
    local data_tgz_path = fio.pathjoin(tmpdir, 'data.tar.xz')
    fio.mktree(data_dir)

    opts = opts or {}
    opts.mkdir = '/bin/mkdir'
    form_systemd_dir(data_dir, name, opts)

    local distribution_dir = fio.pathjoin(data_dir, '/usr/share/tarantool/', name)
    form_distribution_dir(source_dir, distribution_dir, name, version)

    local data = call(string.format("cd %s && %s -cvJf - .", data_dir, tar))
    write_file(data_tgz_path, data)

    -- pack .deb
    call(string.format("cd %s && %s r %s debian-binary control.tar.xz data.tar.xz",
         tmpdir, ar, deb_file_name))
    fio.copyfile(fio.pathjoin(tmpdir, deb_file_name), dest_dir)
end

local function app_pack(args)
    local name, release, version = detect_name_release_version(args.path, args.name, args.version)
    local instantiated_unit_template
    if args.instantiated_unit_template then
        instantiated_unit_template = read_file(args.instantiated_unit_template)
    end

    local unit_template
    if args.unit_template then
        unit_template = read_file(args.unit_template)
    end

    if args.type == 'rpm' then
        pack_rpm(args.path, '.', name, release, version, {
            unit_template = unit_template,
            instantiated_unit_template = instantiated_unit_template
        })
    elseif args.type == 'deb' then
        pack_deb(args.path, '.', name, release, version, {
            unit_template = unit_template,
            instantiated_unit_template = instantiated_unit_template
        })
    elseif args.type == 'tgz' then
        pack_tgz(args.path, '.', name, release, version)
    elseif args.type == 'rock' then
        pack_rock(args.path, '.', name, release, version)
    else
        die("Unknown package type: %s", args.type)
    end
end

local function app_pack_parse(arg)
    local args = {}

    local parameters = argparse(
            arg, {
                { 'name', 'string' },
                { 'version', 'string' },
                { 'instantiated_unit_template', 'string'},
                { 'unit_template', 'string'}
            }
    )

    args.name = parameters.name
    args.version = parameters.version
    args.unit_template = parameters.unit_template
    args.instantiated_unit_template = parameters.instantiated_unit_template
    args.type = parameters[1]
    args.path = parameters[2]

    local available_package_types = { 'rpm', 'tgz', 'rock', 'deb' }
    if not array_contains(available_package_types, args.type) then
        die("Package type should be one of: %s",
                table.concat(available_package_types, ', '))
    end

    if args.path == nil then
        die("Path to application is required")
    end

    return args
end


local function app_pack_usage()
    print(string.format("Usage: %s pack [--name <name>] [<type>] [<path>]\n", self_name))

    print("Arguments")
    print("   type                                           Distribution type to create (rpm, tgz, rock or deb)")
    print("   path                                           Directory with app source code in\n")

    print("Options:")
    print("   --name <name>                                  Name of the app to pack")
    print("   --version <version>                            App version")
    print("   --unit_template <path to file>                 Path of the template for systemd unit file")
    print("   --instantiated_unit_template <path to file>    Path of the template for systemd instantiated unit file")
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
    local files = find_files(template_dir)

    local context = {project_name=app_name,
                     project_name_lower=string.lower(app_name)}

    for _, file in ipairs(files) do
        local srcname = fio.pathjoin(template_dir, file)
        local content = expand(read_file(srcname), context)

        local mode = fio.stat(srcname).mode
        local destname = fio.pathjoin(dest_dir, expand(file, context))
        local destdir = fio.dirname(destname)

        if not fio.path.exists(destdir) then
            fio.mktree(destdir)
        end

        write_file(destname, content, mode)
    end
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

    local template_dir = fio.pathjoin(get_template_dir(), template)

    if not fio.path.exists(template_dir) then
        die("Template '%s' doesn't exist", template_dir)
    end

    local dest_dir = fio.pathjoin(path, name)

    if fio.path.exists(dest_dir) then
        die("Can't create app: directory '%s' already exists", dest_dir)
    end

    fio.mktree(dest_dir)

    instantiate_template(template_dir, dest_dir, name)
    local git = which('git')

    if git ~= nil then
        print("Initializing git repo in: " .. dest_dir)
        call("cd %s && %s init .", dest_dir, git)
        write_file(fio.pathjoin(dest_dir, '.gitignore'), GITIGNORE)
        call("cd %s && %s add -A", dest_dir, git)

        -- git commit on centos 7 fails with cryptic error when called with
        -- io.popen:
        -- error: unable to create temporary file: File exists
        -- this happens only on git commit for some reason
        -- so here we replace io.popen with os.execute
        os.execute(
            string.format(
                'cd %s && %s commit -m "Initial commit"', dest_dir, git))
        call('cd %s && %s tag 0.1.0', dest_dir, git)
    else
        print("warning: git not found. You'll need to add the app "..
                  "to version control yourself later.")
    end

    print(string.format("Application successfully created in '%s'", dest_dir))
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

        Default options can be overriden in ./.cartridge.yml or ~/.cartridge.yml.
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
        cfg = os.getenv('TARANTOOL_CFG') or fio.pathjoin(cwd, 'instances.yml'),
        run_dir = os.getenv('TARANTOOL_RUN_DIR') or '/var/run/tarantool',
        apps_path = '/usr/share/tarantool',
    }
    local paths = {
        '.cartridge.yml',
        fio.pathjoin(os.getenv('HOME') or '',  '.cartridge.yml'),
    }
    for _, path in pairs(paths) do
        if fio.stat(path) then
            local from_file = yaml.decode(read_file(path))
            return fun.chain(defaults, from_file):tomap()
        end
    end
    return defaults
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
    local instance_id = (result[1] or ''):split('.')
    local app_name = get_app_name_from_rockspec()
    result.app_name = #instance_id[1] > 0 and instance_id[1] or app_name
    assert(result.app_name and #result.app_name > 0, 'APP_NAME is required')
    result.instance_name = instance_id[2]

    if result.script == nil then
        local env_script = os.getenv('TARANTOOL_SCRIPT')
        if env_script then
            result.script = env_script
        elseif app_name then -- cartridge is called inside app directory
            result.script = 'init.lua'
        else
            result.script = fio.pathjoin(defaults.apps_path, result.app_name, 'init.lua')
        end
    end
    result.script = fio.abspath(result.script)

    result.cfg = fio.abspath(result.cfg or defaults.cfg)
    result.run_dir = fio.abspath(result.run_dir or defaults.run_dir)
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
