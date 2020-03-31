local digest = require('digest')
local errno = require('errno')
local ffi = require('ffi')
local fiber = require('fiber')
local fio = require('fio')
local fun = require('fun')
local log = require('log')
local socket = require('socket')
local yaml = require('yaml')

local argparse = require('cartridge-cli.argparse')
local templates = require('cartridge-cli.templates')
local utils = require('cartridge-cli.utils')

local self_name = fio.basename(arg[0])

local _TARANTOOL = _G._TARANTOOL

local function VERSION()
    return require('cartridge-cli.VERSION')
end

-- box.NULL, custom and cdata errors aware assert
function assert(val, message, ...) -- luacheck: no global
    if not val or val == nil then
        error(tostring(message), 2)
    end
    return val, message, ...
end

local function get_tarantool_dir()
    return fio.abspath(fio.dirname(arg[-1]))
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
local DEBUG_COLOR_CODE = '\x1B[35m' -- magenta

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

local function print_and_flush(msg)
    print(msg)
    io.flush()
end

local function die(fmt, ...)
    local msg = "ERROR: " .. string.format(fmt, ...)
    print_and_flush(colored_msg(msg, ERROR_COLOR_CODE))
    os.exit(1)
end

local function warn(fmt, ...)
    local msg = "WARNING: " .. string.format(fmt, ...)
    print_and_flush(colored_msg(msg, WARN_COLOR_CODE))
end

local function info(fmt, ...) -- luacheck: no unused
    local msg = string.format(fmt, ...)
    print_and_flush(colored_msg(msg, INFO_COLOR_CODE))
end

local function debug(fmt, ...) -- luacheck: no unused
    local msg = string.format(fmt, ...)
    print_and_flush(colored_msg(msg, DEBUG_COLOR_CODE))
end

local function format_internal_error(err)
    local formatted_error = string.format(
        "Whoops! It looks like something is wrong with this version of Cartridge CLI. " ..
        "Please, report a bug at https://github.com/tarantool/cartridge-cli/issues/new. " ..
        "The error is: %s.", err
    )
    return formatted_error
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
        local files, _ = utils.listdir(path)  -- ignore utils.listdir error

        for _, file in ipairs(files or {}) do
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
        warn('Failed to detect version from git: %s', err)
        return nil
    end

    local version, release = normalize_version(raw_version)
    if version == nil then
        warn("Detected version '%s' is ignored, " ..
              "because it doesn't look like a proper " ..
              "version (major.minor.patch[-count][-commit])", version)
    end

    return version, release
end

local function find_rockspec(source_dir)
    local files, err = utils.listdir(source_dir)
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
    local rockspec, err = utils.load_variables_from_file(rockspec_filepath)
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
                "Please pass it explicitly via --name",
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

-- Tarantool Enterprise directory

local IMAGE_SDK_DIRNAME = 'tarantool-enterprise'
local APPFILES_DIRNAME = 'app-files'

-- DEB

local DEBIAN_BINARY_FILENAME = 'debian-binary'
local DEBIAN_CONTROL_ARCHIVE_NAME = 'control.tar.xz'
local DEBIAN_DATA_ARCHIVE_NAME = 'data.tar.xz'

-- * --------------- Preinstall ---------------

local CREATE_USER_SCRIPT = [[
/bin/sh -c 'groupadd -r tarantool > /dev/null 2>&1 || :'
/bin/sh -c 'useradd -M -N -g tarantool -r -d /var/lib/tarantool -s /sbin/nologin \
    -c "Tarantool Server" tarantool > /dev/null 2>&1 || :'
/bin/sh -c 'mkdir -p /etc/tarantool/conf.d/ --mode 755 2>&1 || :'
/bin/sh -c 'mkdir -p /var/lib/tarantool/ --mode 755 2>&1 || :'
/bin/sh -c 'chown tarantool:tarantool /var/lib/tarantool 2>&1 || :'
/bin/sh -c 'mkdir -p /var/run/tarantool/ --mode 755 2>&1 || :'
/bin/sh -c 'chown tarantool:tarantool /var/run/tarantool 2>&1 || :'
]]

-- * -------------- Postinstall --------------

local SET_OWNER_SCRIPT = [[
/bin/sh -c 'chown -R root:root /usr/share/tarantool/${name}'
/bin/sh -c 'chown root:root /etc/systemd/system/${name}.service'
/bin/sh -c 'chown root:root /etc/systemd/system/${name}@.service'
/bin/sh -c 'chown root:root /usr/lib/tmpfiles.d/${name}.conf'
]]

-- * ---------------- Systemd ----------------

local SYSTEMD_UNIT_FILE = [[
[Unit]
Description=Tarantool Cartridge app ${name}.default
After=network.target

[Service]
Type=simple
ExecStartPre=/bin/sh -c 'mkdir -p ${workdir}.default'
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
ExecStartPre=/bin/sh -c 'mkdir -p ${workdir}.%i'
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

local TMPFILES_CONFIG = 'd /var/run/tarantool 0755 tarantool tarantool'

-- * ------------------- Dockerfile -------------------

local DEFAULT_RUNTIME_BASE_DOCKERFILE_NAME = 'Dockerfile.cartridge'
local DEFAULT_BUILD_BASE_DOCKERFILE_NAME = 'Dockerfile.build.cartridge'

local DEFAULT_BUILD_BASE_DOCKERFILE_LAYERS = 'FROM centos:8\n'
local DEFAULT_RUNTIME_BASE_DOCKERFILE_LAYERS = 'FROM centos:8\n'

-- Don't forget to edit Dockerfile.cache when change this layers
local DOCKERFILE_PREPARE = [[
# Create Tarantool user and directories
RUN groupadd -r tarantool \
    && useradd -M -N -g tarantool -r -d /var/lib/tarantool -s /sbin/nologin \
        -c "Tarantool Server" tarantool \
    &&  mkdir -p /var/lib/tarantool/ --mode 755 \
    && chown tarantool:tarantool /var/lib/tarantool \
    && mkdir -p /var/run/tarantool/ --mode 755 \
    && chown tarantool:tarantool /var/run/tarantool
]]

-- Don't forget to edit Dockerfile.cache when change this layers
local DOCKERFILE_INSTALL_PACKAGES_REQUIRED_FOR_BUILD = [[
### Install packages required for build
RUN yum install -y git-core gcc make cmake unzip
]]

-- Don't forget to edit Dockerfile.cache when change this layers
local DOCKER_INSTALL_OPENSOURCE_TARANTOOL_TEMPLATE = [[
### Install opensource Tarantool
RUN curl -s \
        https://packagecloud.io/install/repositories/tarantool/${tarantool_repo_version}/script.rpm.sh | bash \
    && yum -y install tarantool tarantool-devel
]]

local DOCKER_COPY_ENTERPRISE_TARANTOOL_TEMPLATE = [[
### Copy Tarantool Enterprise
COPY ${sdk_dirname} /usr/share/tarantool/${image_sdk_dirname}

ENV PATH="/usr/share/tarantool/${image_sdk_dirname}:${"$"}{PATH}"
]]

-- This part of Dockerfile is very important
-- Without it, the owner of .rocks in distribution dir is root
-- and build directory can't be removed
-- (of course, if you aren't root)
-- Be careful changing it and always test it
-- Note that Docker Desktop for Mac isn't affected by this bug
--
-- We need to create a user with the same UID on the image
-- and use it for application building
-- Username doesn't matter, but it's used by `usermod` command
-- So, if a user with the desired UID already exists on the image,
-- the user's name is used, otherwise user `cartridge` with the desired UID
-- is created (we assume that a user with this name doesn't exist on the image)
--
local DOCKERFILE_WRAP_USER = [[
### Wrap user
RUN if id -u ${user_id} 2>/dev/null; then \
        USERNAME=${"$"}(id -nu ${user_id}); \
    else \
        USERNAME=cartridge; \
        useradd -u ${user_id} ${"$"}{USERNAME}; \
    fi \
    && (usermod -a -G sudo ${"$"}{USERNAME} 2>/dev/null || :) \
    && (usermod -a -G wheel ${"$"}{USERNAME} 2>/dev/null || :) \
    && (usermod -a -G adm ${"$"}{USERNAME} 2>/dev/null || :)

USER ${user_id}
]]

local DOCKERFILE_COPY_APPLICATION_CODE_TEMPLATE = [[
COPY . /usr/share/tarantool/${name}
]]

-- In case of using Tarantool Enterprise, application directory
-- contains SDK directory.
-- It shouldn't be delivered to the runtime image, so it's removed
-- after copying application files
local REMOVE_BUILD_SDK_FROM_APP_DIR = [[
RUN rm -rf /usr/share/tarantool/${name}/${sdk_dirname}
]]

local DOCKERFILE_SET_PATH = 'ENV PATH="/usr/share/tarantool/${name}:${"$"}{PATH}"\n'

local DOCKERFILE_RUNTIME_TEMPLATE = [[
### Application runtime
RUN echo '${tmpfiles_config}' > /usr/lib/tmpfiles.d/${name}.conf \
&& chmod 644 /usr/lib/tmpfiles.d/${name}.conf
USER tarantool:tarantool
CMD TARANTOOL_WORKDIR=${workdir}.${instance_name} \
    TARANTOOL_PID_FILE=/var/run/tarantool/${name}.${instance_name}.pid \
    TARANTOOL_CONSOLE_SOCK=/var/run/tarantool/${name}.${instance_name}.control \
    tarantool ${dir}/init.lua
]]

-- * ---------- Application build commands ----------

-- Image should always be built from the project root directory
-- It's important to allow users to use COPY instruction correctly
local BUILD_IMAGE_COMMAND_TEMPLATE = [[
    cd ${build_dir} \
    && ${docker} build -t ${image_fullname} \
                    -f ${dockerfile_name} \
                    ${docker_build_args} \
                    . \
                    1>&2
]]

local BUILD_APPLICATION_ON_IMAGE_COMMAND = [[
    ${docker} run \
        --volume ${dir}:/opt/tarantool \
        --rm \
        ${image_fullname} \
        /bin/bash -c '
            cd /opt/tarantool \
            && \
            : "----------- pre-build -----------" \
            && \
            if [ -f ${prebuild_script_name} ]; then \
                echo "Running pre-build script..." \
                && . ${prebuild_script_name}; \
            fi \
            && \
            : "------------- build -------------" \
            && \
            if ls *.rockspec 1> /dev/null 2>&1; then \
                echo "Running tarantoolctl rocks make..." \
                && tarantoolctl rocks make; \
            fi \
            && ${copy_tarantool_binaries}
    '
]]

local COPY_TARANTOOL_BINARIES_COMMAND_TEMPLATE = [[
        : "----- copy Tarantool binaries ----" \
        && echo "Copy Tarantool Enterprise binaries..." \
        && [ -d /usr/share/tarantool/${image_sdk_dirname} ] \
        && \
        cp /usr/share/tarantool/${image_sdk_dirname}/{tarantool,tarantoolctl} \
                        /opt/tarantool
]]

-- * ----------- Packing flow global state -----------

local app_state = {
    -- Here will be stored general application info to be used for
    --   application building and packing, for example, application name
    --   or flag that detects if the application uses deprecated packing flow
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
        local manifest, err = utils.load_variables_from_file(manifest_filepath)
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

    if app_state.tarantool_is_enterprise then
        -- copy TARANTOOL and TARANTOOL_SDK versions from SDK version file
        local tarantool_dir = get_tarantool_dir()
        local tnt_version = fio.pathjoin(tarantool_dir, 'VERSION')
        if not fio.path.exists(tnt_version) then
            warn("can't open VERSION file from Tarantool SDK. SDK information can't be " ..
                "shipped to the resulting package. ")
        else
            local tnt_versions_content, err = utils.read_file(tnt_version)
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
        app_state.name,
        app_state.version_release
    )
    table.insert(version_file_lines, app_version_line)

    -- rocks versions
    local rocks_versions, err = get_rock_versions(distribution_dir)
    if rocks_versions == nil then
        warn("can't process rocks manifest file. Dependency information can't be " ..
             "shipped to the resulting package: %s", err)
    else
        for rock, version in pairs(rocks_versions) do
            if rock ~= app_state.name then
                local rock_version_line = string.format("%s=%s", rock, version)
                table.insert(version_file_lines, rock_version_line)
            end
        end
    end

    -- write collected info to VERSION file
    local version_filepath = fio.pathjoin(distribution_dir, 'VERSION')
    local version_file_content = table.concat(version_file_lines, '\n') .. '\n'
    local ok, err = utils.write_file(version_filepath, version_file_content, tonumber(644, 8))
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

    local files = utils.find_files(destdir, { include_dirs = true })

    -- formatting all pattern and exclusion exception pattern
    local patterns, exceptions  = {}, {}

    local ignore_file_content, err = utils.read_file(ignore)
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
            local ok, err = utils.remove_by_path(path)
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

    local files, err = utils.listdir(dir)
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
                local err = string.format(
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

local function cleanup_distribution_files(dest_dir)
    local rocks_dir = fio.pathjoin(dest_dir, '.rocks')
    if fio.path.exists(rocks_dir) then
        local ok, err = utils.remove_by_path(rocks_dir)
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
        if not ok then
            warn(
                "Failed to run `git clean` in the project root. " ..
                "It is possible that some of the extra files " ..
                "normally ignored are shipped to the resulting package. " ..
                "The error is: %s",
                err
            )
        end

        info('Running `git clean` for submodules')
        -- Recursively cleanup all submodules
        local ok, err = call(
            "cd %q && %s submodule foreach --recursive %s clean -f -d -X",
            dest_dir, git, git
        )
        if not ok then
            warn(
                "Failed to run `git clean` for submodules. " ..
                "It is possible that some of the extra files " ..
                "normally ignored are shipped to the resulting package. " ..
                "The error is: %s",
                err
            )
        end
    end

    if not app_state.deprecated_flow then
        local git_dir = fio.pathjoin(dest_dir, '.git')
        if fio.path.exists(git_dir) then
            info('Remove .git directory')
            local ok, err = utils.remove_by_path(git_dir)
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

    return true
end

-- * --------------------------- Dockerfile parts --------------------------

local function get_tarantool_repo_version(tarantool_version)
    local parts = tarantool_version:split('.')
    local version = string.format('%s_%s', parts[1], parts[2])

    -- For Tarantool 2.1 tarantool/2x repo is used
    -- (not tarantool/2_1)
    if version == '2_1' then
        version = '2x'
    end

    return version
end

local function construct_install_tarantool_dockerfile_part()
    local install_tarantool_dockerfile_part
    if app_state.tarantool_is_enterprise then
        if app_state.sdk_path ~= nil then
            install_tarantool_dockerfile_part = utils.expand(
                DOCKER_COPY_ENTERPRISE_TARANTOOL_TEMPLATE, {
                    sdk_dirname = app_state.build_sdk_dirname,
                    image_sdk_dirname = IMAGE_SDK_DIRNAME,
                }
            )
        else
            return nil, format_internal_error('app_state.sdk_path is not set')
        end
    else
        install_tarantool_dockerfile_part = utils.expand(
            DOCKER_INSTALL_OPENSOURCE_TARANTOOL_TEMPLATE, {
                tarantool_repo_version = get_tarantool_repo_version(app_state.tarantool_version),
            }
        )
    end

    return install_tarantool_dockerfile_part
end

local function construct_build_image_dockerfile()
    -- The application build dockerfile consists of these parts:
    -- - build_base_dockerfile_layers: the base image
    -- - prepare: install packages required for build (git gcc make cmake unzip)
    --            and create tarantool user and directories
    -- - install_tarantool: install Tarantool to the image
    -- - wrap_user: add user with the same UID as host user

    if app_state.build_base_dockerfile_layers == nil then
        return nil, format_internal_error('Build base dockerfile layers should be set')
    end

    local instal_tarantool_part, err = construct_install_tarantool_dockerfile_part()
    if instal_tarantool_part == nil then
        return nil, err
    end

    local user_id, err = check_output('id -u')
    if user_id == nil then
        return nil, string.format('Failed to get user ID: %s', err)
    end

    user_id = user_id:strip()
    local wrap_user_part = utils.expand(DOCKERFILE_WRAP_USER, { user_id = user_id })

    -- Dockerfile parts
    local dockerfile_parts = {
        app_state.build_base_dockerfile_layers,
        DOCKERFILE_INSTALL_PACKAGES_REQUIRED_FOR_BUILD,
        DOCKERFILE_PREPARE,
        instal_tarantool_part,
        wrap_user_part,
    }

    -- Concatenate all parts together
    local dockerfile_content = table.concat(dockerfile_parts, '\n')
    return dockerfile_content
end

local function construct_runtime_image_dockerfile()
    -- The application runtime dockerfile consists of these parts:
    -- - runtime_base_dockerfile_layers: the base image
    -- - prepare: install packages required for build (git gcc make cmake unzip)
    --            and create tarantool user and directories
    -- - install_tarantool: install opensource Tarantool to the image
    -- - application_code: copy application code
    -- - set_path: set PATH for Tarantool Enterprise
    -- - runtime: tmpfiles configuration, CMS and USER directives

    if app_state.runtime_base_dockerfile_layers == nil then
        return nil, format_internal_error('Runtime base dockerfile layers should be set')
    end

    -- Dockerfile parts
    local dockerfile_parts = {
        app_state.runtime_base_dockerfile_layers,
        DOCKERFILE_PREPARE,
    }

    -- install opensource Tarantool
    if not app_state.tarantool_is_enterprise then
        local install_tarantool_part, err = construct_install_tarantool_dockerfile_part()
        if install_tarantool_part == nil then
            return nil, err
        end

        table.insert(dockerfile_parts, install_tarantool_part)
    end

    -- copy application files
    local application_code_part = utils.expand(
        DOCKERFILE_COPY_APPLICATION_CODE_TEMPLATE,
        { name = app_state.name }
    )
    table.insert(dockerfile_parts, application_code_part)

    -- set PATH for Tarantool Enterprise
    if app_state.tarantool_is_enterprise then
        local remove_build_sdk = utils.expand(REMOVE_BUILD_SDK_FROM_APP_DIR, {
            name = app_state.name,
            sdk_dirname = app_state.build_sdk_dirname,
        })
        table.insert(dockerfile_parts, remove_build_sdk)

        local set_path_part = utils.expand(DOCKERFILE_SET_PATH, { name = app_state.name })
        table.insert(dockerfile_parts, set_path_part)
    end

    -- runtime layers
    local runtime_part = utils.expand(DOCKERFILE_RUNTIME_TEMPLATE, {
        name = app_state.name,
        dir = fio.pathjoin('/usr/share/tarantool/', app_state.name),
        instance_name = '${"$"}{TARANTOOL_INSTANCE_NAME:-default}',
        workdir = fio.pathjoin('/var/lib/tarantool/', app_state.name),
        tmpfiles_config = TMPFILES_CONFIG,
    })
    table.insert(dockerfile_parts, runtime_part)

    -- Concatenate all parts together
    local dockerfile_content = table.concat(dockerfile_parts, '\n')
    return dockerfile_content
end

local function get_docker_build_args_string()
    local docker_build_args = { app_state.docker_build_args or '' }

    -- Use base image as a cache
    local cache_from_base_arg = string.format(
        '--cache-from %s',
        app_state.base_image_fullname
    )
    table.insert(docker_build_args, cache_from_base_arg)

    return table.concat(docker_build_args, ' ')
end

-- * ---------------------- Building application itself ----------------------

local function copy_tarantool_binaries(dir)
    info('Copy Tarantool Enterprise binaries')
    assert(app_state.tarantool_is_enterprise)

    local tarantool_dir = get_tarantool_dir()

    for _, binary in ipairs({'tarantool', 'tarantoolctl'}) do
        local path_from = fio.pathjoin(tarantool_dir, binary)
        local path_to = fio.pathjoin(dir, binary)

        local ok, err = utils.copyfile(path_from, path_to)
        if not ok then return false, err end
    end

    return true
end

local function build_application_in_docker(dir)
    assert(app_state.build_in_docker)
    -- Application is built in a docker container:
    -- - First, the base docker image <app-name>-base is created.
    -- - Then, build commands are run on this image
    --   with volume in the local dir
    -- As a result, we have the application with rocks modules
    --   specific for the target platform in the local dir

    local docker = which('docker')
    if docker == nil then
        return false, 'docker binary is required to build application in docker'
    end

    if app_state.tarantool_is_enterprise then
        -- copy Tarantool SDK
        local build_sdk_dirpath = fio.pathjoin(dir, app_state.build_sdk_dirname)
        local ok, err = utils.copytree(app_state.sdk_path, build_sdk_dirpath)
        if not ok then return false, err end
    end

    -- Build the base image
    info('Building docker image: %s', app_state.base_image_fullname)

    -- - Write base image Dockerfile
    local build_image_dockerfile_name = string.format('Dockerfile.build.%s', app_state.build_id)
    local build_image_dockerfile_path = fio.pathjoin(dir, build_image_dockerfile_name)
    local build_image_dockerfile_content, err = construct_build_image_dockerfile()
    if build_image_dockerfile_content == nil then return false, err end

    local ok, err = utils.write_file(build_image_dockerfile_path, build_image_dockerfile_content)
    if not ok then return false, err end

    -- - Build the base docker image
    local create_build_image_command = utils.expand(BUILD_IMAGE_COMMAND_TEMPLATE, {
        docker = docker,
        build_dir = dir,
        image_fullname = app_state.base_image_fullname,
        dockerfile_name = build_image_dockerfile_name,
        docker_build_args = get_docker_build_args_string(),
    })
    local ok, err = call(create_build_image_command)
    if not ok then
        return false, string.format('Failed to build image: %s', err)
    end

    info('Base image tagged as: %s', app_state.base_image_fullname)

    -- Build application in the base image
    info('Build application in %s', app_state.base_image_fullname)

    -- - Construct application build command (`docker run <base-image> <build-commands>`)
    local build_app_command_params = {
        docker = docker,
        dir = dir,
        image_fullname = app_state.base_image_fullname,
        prebuild_script_name = PREBUILD_SCRIPT_NAME,
        copy_tarantool_binaries = ':',  -- XXX: refactor it
    }

    if app_state.deprecated_flow then
        build_app_command_params.prebuild_script_name = DEP_PREBUILD_SCRIPT_NAME
    end

    if app_state.tarantool_is_enterprise then
        build_app_command_params.copy_tarantool_binaries = utils.expand(
            COPY_TARANTOOL_BINARIES_COMMAND_TEMPLATE, {
                image_sdk_dirname = IMAGE_SDK_DIRNAME,
            }
        )
    end

    local build_app_command = utils.expand(
        BUILD_APPLICATION_ON_IMAGE_COMMAND,
        build_app_command_params
    )

    -- - Build application
    local ok, err = call(build_app_command)
    if not ok then
        return false, string.format('Failed to build application: %s', err)
    end

    local ok, err = utils.remove_by_path(build_image_dockerfile_path)
    if not ok then
        warn('Failed to remove build base image Dockerfile %s: %s', build_image_dockerfile_name, err)
    end

    if app_state.tarantool_is_enterprise then
        local build_sdk_dirpath = fio.pathjoin(dir, app_state.build_sdk_dirname)
        local ok, err = utils.remove_by_path(build_sdk_dirpath)
        if not ok then
            return false, string.format('Failed to remove build SDK: %s', err)
        end
    end

    info('Application build succeeded')

    return true
end

local function build_application_locally(dir)
    assert(not app_state.build_in_docker)
    -- pre build
    if app_state.deprecated_flow then
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

    info('Application build succeeded')

    return true
end

local function build_application(dir)
    -- build application
    if app_state.build_in_docker then
        local ok, err = build_application_in_docker(dir)
        if not ok then return false, err end
    else
        local ok, err = build_application_locally(dir)
        if not ok then return false, err end
    end

    return true
end

local function cleanup_after_build(dir)
    -- apply .cartridge.ignore (DEPRECATED)
    if app_state.deprecated_flow then
        -- deleting files matching patterns from .cartridge.ignore
        info('Remove files matching patterns from %s', DEP_IGNORE_FILE_NAME)
        local ok, err = remove_ignored(dir)
        if not ok then return false, err end

        -- remove special files
        for _, filename in ipairs({DEP_IGNORE_FILE_NAME, DEP_PREBUILD_SCRIPT_NAME}) do
            local filepath = fio.pathjoin(dir, filename)
            if fio.path.exists(filepath) then
                info('Remove %s', filename)
                local ok, err = utils.remove_by_path(filepath)
                if not ok then return false, err end
            end
        end

        -- remove git dir
        local git_dir = fio.pathjoin(dir, '.git')
        if fio.path.exists(git_dir) then
            info('Remove .git directory')
            local ok, err = utils.remove_by_path(git_dir)
            if not ok then return false, err end
        end
    else  -- new build flow
        if fio.path.exists(fio.pathjoin(dir, POSTBUILD_SCRIPT_NAME)) then
            local ok, err = run_hook(dir, POSTBUILD_SCRIPT_NAME)
            if not ok then return false, err end
        end
    end

    -- remove special files
    local special_files = {
        PREBUILD_SCRIPT_NAME,
        DEP_PREBUILD_SCRIPT_NAME,
        POSTBUILD_SCRIPT_NAME,
        DEP_IGNORE_FILE_NAME
    }
    for _, filename in ipairs(special_files) do
        local filepath = fio.pathjoin(dir, filename)
        if fio.path.exists(filepath) then
            info('Remove %s', filename)
            local ok, err = utils.remove_by_path(filepath)
            if not ok then return false, err end
        end
    end

    return true
end

-- * ----------------- Distribution dir -----------------

local function form_distribution_dir(dest_dir)
    local ok, err = utils.copytree(app_state.path, dest_dir)
    if not ok then return false, err end

    local ok, err = cleanup_distribution_files(dest_dir)
    if not ok then return false, err end

    local ok, err = build_application(dest_dir)
    if not ok then return false, err end

    local ok, err = cleanup_after_build(dest_dir)
    if not ok then return false, err end

    local ok, err = generate_version_file(dest_dir)
    if not ok then return false, err end

    if not app_state.build_in_docker and app_state.tarantool_is_enterprise then
        local ok, err = copy_tarantool_binaries(dest_dir)
        if not ok then return false, err end
    end

    return true
end

-- * -------------------- Systemd dir --------------------

local function form_systemd_dir(base_dir, opts)
    opts = opts or {}
    info('Form application systemd dir')

    local unit_template = opts.unit_template or SYSTEMD_UNIT_FILE
    local instantiated_unit_template = opts.instantiated_unit_template or SYSTEMD_INSTANTIATED_UNIT_FILE

    local systemd_dir = fio.pathjoin(base_dir, '/etc/systemd/system')
    local ok, err = utils.make_tree(systemd_dir)
    if not ok then return false, err end

    local expand_params = {
        name = app_state.name,
        dir = fio.pathjoin('/usr/share/tarantool/', app_state.name),
        workdir = fio.pathjoin('/var/lib/tarantool/', app_state.name),
    }

    if app_state.tarantool_is_enterprise then
        expand_params.bindir = expand_params.dir
    else
        expand_params.bindir = '/usr/bin'
    end

    local unit_template_filepath = fio.pathjoin(systemd_dir, string.format('%s.service', app_state.name))
    local instantiated_unit_template_filepath = fio.pathjoin(systemd_dir, string.format('%s@.service', app_state.name))
    local ok, err = utils.write_file(
        unit_template_filepath,
        utils.expand(unit_template, expand_params)
    )
    if not ok then return false, err end

    local ok, err = utils.write_file(
        instantiated_unit_template_filepath,
        utils.expand(instantiated_unit_template, expand_params)
    )
    if not ok then return false, err end

    return true
end

-- * ---------------- Tmpfiles configuration ----------------

local function write_tmpfiles_conf(base_dir)
    info('Write application tmpfiles configuration')

    local tmpfiles_dir = fio.pathjoin(base_dir, '/usr/lib/tmpfiles.d')
    local ok, err = utils.make_tree(tmpfiles_dir)
    if not ok then return false, err end

    local tmpfiles_conf_filepath = fio.pathjoin(
        tmpfiles_dir,
        string.format('%s.conf', app_state.name)
    )
    local ok, err = utils.write_file(
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
        app_state.name,
        app_state.version_release
    )
    tgz_file_name = fio.pathjoin(app_state.dest_dir, tgz_file_name)

    info("Packing tar.gz file")

    local tar = which('tar')

    if tar == nil then
        return false, "tar binary is required to pack tar.gz"
    end

    info("Packing tar.gz in: %s", app_state.build_dir)

    local distribution_dir = fio.pathjoin(app_state.appfiles_dir, app_state.name)
    local ok, err = utils.make_tree(distribution_dir)
    if not ok then return false, err end

    local ok, err = form_distribution_dir(distribution_dir)
    if not ok then return false, err end

    info('Create archive')
    local data, err = check_output(
        "cd %s && %s -cvzf - %s",
        app_state.appfiles_dir, tar, app_state.name
    )
    if data == nil then
        return false, string.format("Failed to pack tgz: %s", err)
    end

    local ok, err = utils.write_file(tgz_file_name, data)
    if not ok then return false, err end

    info("Resulting tar.gz saved as: %s", tgz_file_name)

    return true
end

-- * ---------------- ROCK packing ----------------

local function pack_rock()
    local distribution_dir = fio.pathjoin(app_state.appfiles_dir, app_state.name)
    local ok, err = utils.make_tree(distribution_dir)
    if not ok then return false, err end

    info("Packing binary rock in: %s", app_state.build_dir)

    local ok, err = form_distribution_dir(distribution_dir)
    if not ok then return false, err end

    fio.chdir(app_state.build_dir)

    local rockspec = find_rockspec(distribution_dir)
    local content = ''
    if rockspec then
        local err
        content, err = utils.read_file(fio.pathjoin(distribution_dir, rockspec))
        if content == nil then return false, err end

        content = string.gsub(content, "(.-version%s-=%s-['\"])(.-)(['\"].*)",
                '%1' .. app_state.version_release .. '%3')
        if not content then
            return false, string.format('Rockspec %s is not valid! Version not found!', rockspec)
        end
    end

    local name_of_rockspec = string.format(
        '%s-%s.rockspec',
        app_state.name,
        app_state.version_release
    )

    local new_rockspec = fio.pathjoin(distribution_dir, name_of_rockspec)

    local ok, err = utils.write_file(new_rockspec, content)
    if not ok then return false, err end

    fio.chdir(distribution_dir)

    local rock_filename = string.format(
        '%s-%s.*.rock',
        app_state.name,
        app_state.version_release
    )

    local ok, err = call('tarantoolctl rocks pack %s ', new_rockspec)
    if not ok then return false, err end

    rock_filename = fio.glob(fio.pathjoin(distribution_dir, rock_filename))[1]

    local dest_rock_filename = fio.pathjoin(app_state.dest_dir, fio.basename(rock_filename))

   local ok, err = utils.copyfile(rock_filename, dest_rock_filename)
    if not ok then return false, err end

    info('Resulting rock saved as: %s', dest_rock_filename)

    return true
end

-- * ---------------- RPM packing ----------------

-- RPM file is a binary format, consisting of metadata in the form of
-- key-value pairs and then a gzipped cpio archive (of SVR-4 variety).
--
-- Documentation on the binary format can be found here:
-- - http://ftp.rpm.org/max-rpm/s1-rpm-file-format-rpm-file-format.html
-- - https://docs.fedoraproject.org/ro/Fedora_Draft_Documentation/0.1/html/RPM_Guide/ch-package-structure.html
--
-- Also I've found this explanatory blog post to be of great help:
-- - https://blog.bethselamin.de/posts/argh-pm.html

-- Here's what the layout looks like:
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
-- While debugging, I used rpm-dissecting tool from mkrepo:
-- - https://github.com/tarantool/mkrepo/blob/master/mkrepo.py

local RPM_MAGIC = 0xedabeedb
local RPM_VER = {3, 0}

local HEADERSIGNATURES=62
local HEADERIMMUTABLE=63


-- There are way more tags in the spec than what I've included here
-- both for signature header and regular header. Most of them are
-- optional.
--
-- Explanation and values for most of these tags can be found in documentation:
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
    -- Pack a value to the binary form, and align it to a required
    -- address boundary. Since rpm headers are mmap-ed, numeric
    -- types need to have proper alignment (say, 32-bit integer
    -- addresses should be aligned to 4-byte boundary)
    --
    -- Parameters:
    -- value - value to pack, or an array of values
    -- val_type - expected type of value, e.g. INT32, STRING, etc...
    -- offset - Since we can't calculate alignment "in vacuum", we
    --          need to know the absolute address of the value in the
    --          resulting buffer. That'd be the basis for calculating
    --          the alignment
    --
    -- Return values: {tag, num_elements, buffer, padding}
    -- tag -- type tag, as per the RPM spec (e.g. 5 for int64 data)
    -- num_elements -- 1 in case of single values, otherwise the size
    --                 of packed array
    -- buffer -- packed binary data
    -- padding -- how many zero bytes were added to the beginning of
    --            the buffer to properly align its contents
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
                pad = utils.align(offset, 2) - offset
                buf = buf .. pack('>h', v)
            elseif val_type == 'INT32' then
                ret_val_type = 4
                pad = utils.align(offset, 4) - offset
                buf = buf .. pack('>i', v)
            elseif val_type == 'INT64' then
                ret_val_type = 5
                pad = utils.align(offset, 8) - offset
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
        -- archive and other headers. So far, so good. But long ago
        -- someone decided it would be a nice idea to have some mutable
        -- tags that can be binary-patched in the rpm file by some
        -- utility and do not affect the SHA1 signature. (I know, right?)
        -- So the region tag is a special tag that says how large
        -- the immutable area of tag space is. It should be
        -- calculated exactly as written below, with size value itself
        -- being negative (sic!).
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
        if file ~= '' and not utils.array_contains(RPM_DIRNAME_BLACKLIST, file) then
            table.insert(result, file)
        end
    end

    return result
end

local function generate_fileinfo(source_dir)
    local function gen_dirnames(files)
        local dirnames = {}

        for _, file in ipairs(files) do
            file = utils.remove_leading_dot(file)
            local dirname = fio.dirname(file)
            dirnames[dirname..'/'] = true
        end

        return utils.dict_keys(dirnames)
    end

    local files = utils.find_files(
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
        file = utils.remove_leading_dot(file)

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
            local filedigest, err = utils.file_md5_hex(fullpath)
            if filedigest == nil then return false, err end

            table.insert(result.fileflags, bit.lshift(1, 4))
            table.insert(result.filedigests, filedigest)
        end

        table.insert(result.basenames, basename)
        table.insert(result.dirindexes, utils.array_index_of(dirnames, dirname..'/')-1)
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
    -- The resulting CPIO structure should look like what it will be
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

    local distribution_dir = fio.pathjoin(app_state.appfiles_dir, '/usr/share/tarantool/', app_state.name)
    local ok, err = form_distribution_dir(distribution_dir)
    if not ok then return nil, err end

    local ok, err = form_systemd_dir(app_state.appfiles_dir, opts)
    if not ok then return nil, err end

    local ok, err = write_tmpfiles_conf(app_state.appfiles_dir)
    if not ok then return nil, err end

    local files = utils.find_files(app_state.appfiles_dir, {include_dirs=true, exclude={'.git'}})
    files = filter_out_known_files(files)

    local ok, err = utils.write_file(fio.pathjoin(app_state.appfiles_dir, 'files'), table.concat(files, '\n'))
    if not ok then return nil, err end

    info('Create CPIO archive')
    local ok, pack_err = call("cd %s && cat files | %s -o -H newc > unpacked", app_state.appfiles_dir, cpio)
    if not ok then
        return nil, string.format("Failed to pack CPIO: %s", pack_err)
    end

    info('Compress it using GZIP')
    local payloadsize = fio.stat(fio.pathjoin(app_state.appfiles_dir, 'unpacked')).size
    local archive, read_err = check_output("cd %s && cat unpacked | %s -9", app_state.appfiles_dir, gzip)
    if archive == nil then
        return nil, string.format("Failed to pack CPIO: %s", read_err)
    end

    for _, f in ipairs({'unpacked', 'files'}) do
        local filepath = fio.pathjoin(app_state.appfiles_dir, f)
        local ok, err = utils.remove_by_path(filepath)
        if not ok then return nil, err end
    end

    local fileinfo = generate_fileinfo(app_state.appfiles_dir)

    return {
        archive = archive,
        fileinfo = fileinfo,
        payloadsize = payloadsize,
    }
end

local function get_tarantool_version()
    return _TARANTOOL:split('-', 1)[1]
end

local function tarantool_next_major_version(tarantool_version)
    return tostring(tarantool_version:split('.', 1)[1] + 1)
end

local function pack_rpm(opts)
    opts = opts or {}
    local rpm_file_name = fio.pathjoin(
        app_state.dest_dir,
        string.format(
            "%s-%s.rpm",
            app_state.name,
            app_state.version_release
        )
    )

    info("Packing rpm file")
    local lead = gen_lead(app_state.name)

    local cpio, err = pack_cpio(opts)
    if cpio == nil then return false, err end

    info('Construct RPM header')
    -- compute payload digest
    local payloaddigest_algo = PGPHASHALGO_SHA256
    local payloaddigest = digest.sha256_hex(cpio.archive)

    local create_user_script_rpm = CREATE_USER_SCRIPT

    local fileinfo = cpio.fileinfo
    local header_tags = {
        {'NAME', 'STRING', app_state.name},
        {'VERSION', 'STRING', app_state.version},
        {'RELEASE', 'STRING', app_state.release},
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

    if not app_state.tarantool_is_enterprise then
        --- Append RPM dependency flags for Tarantool
        --- See Dependency Tags section of
        --- - https://docs.fedoraproject.org/ro/Fedora_Draft_Documentation/0.1/html/RPM_Guide/ch-package-structure.html

        local min_version = app_state.tarantool_version
        local max_version = tarantool_next_major_version(app_state.tarantool_version)

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

    body = lead .. utils.buf_pad_to_8_byte_boundary(signature_header) .. body

    info('Write RPM file')
    local ok, err = utils.write_file(rpm_file_name, body)
    if not ok then return false, err end

    info("Resulting rpm saved as: %s", rpm_file_name)

    return true
end


-- * ---------------- DEB packing ----------------

-- DEB package is an ar archive that contains debian-binary, control.tar.gz and data.tar.gz files
--
-- debian-binary  : contains format version string (2.0)
-- data.tar.xz    : package files
-- control.tar.xz : control files (control, preinst etc.)
--
local function form_deb_control_dir(dest_dir, name, version)
    local ok, err = utils.make_tree(dest_dir)
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

    if not app_state.tarantool_is_enterprise then
        -- Add tarantool dependency
        local min_version = app_state.tarantool_version
        local max_version = tarantool_next_major_version(app_state.tarantool_version)

        control_params.deps = string.format('tarantool (>= %s), tarantool (<< %s)',
                                            min_version, max_version)
    end

    local ok, err = utils.write_file(
        control_filepath,
        utils.expand(DEBIAN_CONTROL_FILE, control_params)
    )
    if not ok then return false, err end

    -- preinst
    local preinst_filepath = fio.pathjoin(dest_dir, 'preinst')
    local ok, err = utils.write_file(
        preinst_filepath,
        CREATE_USER_SCRIPT,
        tonumber('0755', 8)  -- filemode
    )
    if not ok then return false, err end

    -- postinst
    local postinst_filepath = fio.pathjoin(dest_dir, 'postinst')
    local ok, err = utils.write_file(
        postinst_filepath,
        utils.expand(SET_OWNER_SCRIPT, {
            name = name,
        }),
        tonumber('0755', 8)  -- filemode
    )
    if not ok then return false, err end

    return true
end

local function pack_deb(opts)
    opts = opts or {}

    local deb_file_name = string.format(
        "%s-%s.deb",
        app_state.name,
        app_state.version_release
    )

    local tar = which('tar')

    if tar == nil then
       return false, "tar binary is required to pack deb"
    end

    local ar = which('ar')

    if ar == nil then
        return false, "ar binary is required to pack deb"
    end

    info("Packing deb in: %s", app_state.build_dir)

    -- debian-binary
    info('Write debian-binary')
    local debian_binary_path = fio.pathjoin(app_state.appfiles_dir, DEBIAN_BINARY_FILENAME)
    local ok, err = utils.write_file(debian_binary_path, '2.0\n')
    if not ok then return false, err end

    -- control.tar.xz
    info('Generate deb control data')
    local control_dir = fio.pathjoin(app_state.appfiles_dir, 'control')
    local control_tgz_path = fio.pathjoin(app_state.appfiles_dir, DEBIAN_CONTROL_ARCHIVE_NAME)
    local ok, err = form_deb_control_dir(control_dir, app_state.name, app_state.version_release)
    if not ok then return false, err end

    info('Archive deb control data')
    local control_data, pack_control_err = check_output("cd %s && %s -cvJf - .", control_dir, tar)
    if control_data == nil then
        die('Failed to pack deb control files: %s', pack_control_err)
    end
    local ok, err = utils.write_file(control_tgz_path, control_data)
    if not ok then return false, err end

    -- data.tar.xz
    info('Generate package data')
    local data_dir = fio.pathjoin(app_state.appfiles_dir, 'data')
    local data_tgz_path = fio.pathjoin(app_state.appfiles_dir, DEBIAN_DATA_ARCHIVE_NAME)
    local ok, err = utils.make_tree(data_dir)
    if not ok then return false, err end

    local distribution_dir = fio.pathjoin(data_dir, '/usr/share/tarantool/', app_state.name)
    local ok, err = form_distribution_dir(distribution_dir)
    if not ok then return false, err end

    local ok, err = form_systemd_dir(data_dir, opts)
    if not ok then return false, err end

    local ok, err = write_tmpfiles_conf(data_dir)
    if not ok then return false, err end

    info('Compress package data using TAR')
    local data, pack_data_err = check_output("cd %s && %s -cvJf - .", data_dir, tar)
    if data == nil then
        die('Failed to pack deb package files: %s', pack_data_err)
    end
    local ok, err = utils.write_file(data_tgz_path, data)
    if not ok then return false, err end

    -- pack .deb
    info('Pack DEB archive')
    local archive_files = table.concat({
        DEBIAN_BINARY_FILENAME,
        DEBIAN_CONTROL_ARCHIVE_NAME,
        DEBIAN_DATA_ARCHIVE_NAME,
    }, ' ')
    local ok, pack_deb_err = call(
        "cd %s && %s r %s %s",
        app_state.appfiles_dir, ar, deb_file_name, archive_files
    )
    if not ok then
        die('Failed to pack DEB package: %s', pack_deb_err)
    end

    local ok, err = utils.copyfile(fio.pathjoin(app_state.appfiles_dir, deb_file_name), app_state.dest_dir)
    if not ok then return false, err end

    return true
end

local function validate_base_dockerfile(dockerfile_content)
    local from_line

    for _, line in ipairs(dockerfile_content:split('\n')) do
        line = line:strip()
        -- skip comments and empty lines
        if not (line == '' or line:startswith('#')) then
            if not line:strip():lower():startswith('from') then
                return false, 'Base Dockerfile should be started with `FROM centos:8`'
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

local function pack_docker(opts)
    opts = opts or {}

    local docker = which('docker')
    if docker == nil then
        return false, "docker binary is required to pack docker image"
    end

    info("Packing docker in: %s", app_state.build_dir)

    local distribution_dir = fio.pathjoin(app_state.appfiles_dir, app_state.name)

    local ok, err = form_distribution_dir(distribution_dir)
    if not ok then return false, err end

    -- Construct runtime dockerfile
    local runtime_dockerfile_name = string.format('Dockerfile.%s', app_state.build_id)
    local runtime_dockerfile_path = fio.pathjoin(distribution_dir, runtime_dockerfile_name)
    local runtime_dockerfile_content, err = construct_runtime_image_dockerfile()
    if runtime_dockerfile_content == nil then
        return false, err
    end

    local ok, err = utils.write_file(runtime_dockerfile_path, runtime_dockerfile_content)
    if not ok then return false, err end

    -- Construct result image name
    local image_fullname
    if opts.tag ~= nil then
        image_fullname = opts.tag
    else
        image_fullname = string.format(
            '%s:%s',
            app_state.name,
            app_state.version_release
        )
    end

    -- Build result image
    info('Building docker image: %s', image_fullname)

    local create_build_image_command = utils.expand(BUILD_IMAGE_COMMAND_TEMPLATE, {
        docker = docker,
        build_dir = distribution_dir,
        image_fullname = image_fullname,
        dockerfile_name = runtime_dockerfile_path,
        docker_build_args = get_docker_build_args_string(),
    })

    local ok, err = call(create_build_image_command)
    if not ok then
        return false, string.format('Failed to build image: %s', err)
    end

    info('Result image tagged as: %s', image_fullname)

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

-- * -------------------------- Build helpers --------------------------

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

    if deprecated_build_flow_is_ised then
        warn(
            "Using `.cartridge.ignore` and `.cartridge.pre` files is deprecated in 1.3.0 " ..
                "and will be removed in 2.0.0"
        )
    end

    return deprecated_build_flow_is_ised
end

local function get_dockerfile_base_layers(dockerfile_path, default_layers)
    if dockerfile_path == nil then
        return default_layers
    end

    if not fio.path.exists(dockerfile_path) then
        die('Specified base dockerfile does not exists: %s', dockerfile_path)
    end

    info('Detected base Dockerfile %s', dockerfile_path)

    local dockerfile_content, err = utils.read_file(dockerfile_path)
    if dockerfile_content == nil then
        die('Failed to read base Dockerfile %s', err)
    end

    local ok, err = validate_base_dockerfile(dockerfile_content)
    if not ok then die('Base Dockerfile validation failed: %s', err) end

    info('Base Dockerfile is OK')
    return dockerfile_content
end

local function check_pack_state(state)
    local required_params = {
        'path', 'name', 'version', 'release', 'version', 'version_release',
        'dest_dir', 'deprecated_flow', 'tarantool_is_enterprise', 'build_dir',
        'build_id',
    }

    for _, p in ipairs(required_params) do
        if state[p] == nil then
            local err = string.format('Missed reqiured app_state parameter: %s', p)
            return false, err
        end
    end

    return true
end

-- * ------------------- Build dir --------------------

local function detect_and_create_build_dir(app_dir, build_id)
    -- By default, application build is performed in a temporary directory
    --   in `~/.cartridge/tmp/`
    -- User can specify build directory in CARTRIDGE_BUILDDIR env variable.
    -- There are two cases:
    -- - specified directory doesn't exist: we just create it and remove after the build
    -- - directory already exists:
    --   - ${CARTRIDGE_BUILDDIR}/build.cartridge will be the build directory
    --   - sub-directory build.cartridge is (re)created and used for application build
    --   - after the build, ${CARTRIDGE_BUILDDIR}/build.cartridge is removed

    local specified_build_dir = os.getenv('CARTRIDGE_BUILDDIR')

    local build_dir
    if specified_build_dir == nil then
        local build_dir_name = string.format(
            BUILD_DIRECTORY_NAME_TEMPLATE,
            build_id
        )
        build_dir = fio.pathjoin(CARTRIDGE_TMP_PATH, build_dir_name)
    else
        specified_build_dir = fio.abspath(specified_build_dir)
        -- specified build directory can't be project subdirectory
        if utils.is_subdirectory(specified_build_dir, app_dir) then
            die("Build directory can't be project subdirectory, specified: %s", specified_build_dir)
        end

        if not fio.path.exists(specified_build_dir) then
            build_dir = specified_build_dir
        else
            -- This little hack is used to prevent deletion of user files
            -- from the specified build directory on cleanup.
            -- Moreover, this subdirectory is definitely clean,
            -- so we'll have no problems
            if not fio.path.is_dir(specified_build_dir) then
                die("Specified build directory is not a directory: %s", specified_build_dir)
            end

            build_dir = fio.pathjoin(specified_build_dir, DEFAULT_BUILD_DIRECTORY_NAME)
        end
    end

    info('Build directory is set to %s', build_dir)

    if fio.path.exists(build_dir) then
        info('Build directory already exists. Cleaning it.')
        utils.remove_by_path(build_dir)
    end

    local ok, err = utils.make_tree(build_dir)
    if not ok then
        die('Failed co create build directory: %s', err)
    end

    return build_dir
end

local function remove_build_dir()
    info('Remove build directory %s', app_state.build_dir)
    local ok, err = utils.remove_by_path(app_state.build_dir)
    if not ok then
        warn('Failed to clean up build directory %s: %s', app_state.build_dir, err)
    end
end

local function detect_sdk_path(args)
    -- check that passed one option for SDK
    local sdk_params_are_right = utils.check_that_only_one_is_true({
        args.sdk_local or false,
        args.sdk_path ~= nil,
    })

    if not sdk_params_are_right then
        local err = (utils.remove_leading_spaces([=[
            For packing in docker you should specify one of:
            * --sdk-local: to use local SDK;;
            * --sdk-path: path to SDK
                (can be passed in environment variable TARANTOOL_SDK_PATH).
        ]=], 16))
        return nil, err
    end

    local sdk_path
    -- set sdk_path
    if args.sdk_local then
        sdk_path = get_tarantool_dir()
    elseif args.sdk_path ~= nil then
        sdk_path = args.sdk_path
    else
        return nil, format_internal_error('No SDK options passed')
    end

    return sdk_path
end

-- * --------------- Application packing ---------------

local cmd_pack = {
    name = 'pack',
    doc = 'Pack application into a distributable bundle',
    usage = utils.remove_leading_spaces([=[
        %s pack [options] <type> [<path>]

        Arguments
            type                      Distribution type to create
                                      Allowed types: %s

            path                      Path to the application
                                      Defaults to the current directory

        Options
            --name NAME               Application name
                                      By default, application name is taken
                                      from the application rockspec

            --version VERSION         Application version
                                      By default, version is discovered by git

            --unit-template PATH      Path to the template for systemd unit file
                                      Used for rpm and deb types

            --instantiated-unit-template PATH    Path to the template for systemd
                                                 instantiated unit file
                                                 Used for rpm and deb types

            --tag TAG                 Image tag
                                      Used for docker type

            --from PATH               Path to the base dockerfile for the runtime image
                                      Defaults to Dockerfile.cartridge in the project root
                                      Used for docker type

            --build-from PATH         Path to the base dockerfile for build image
                                      Defaults to Dockerfile.build.cartridge in the project root
                                      Used for docker type

            --sdk-local               Flag indicating that SDK from local machine should be
                                      installed on the image
                                      Used for docker type

            --sdk-path PATH           Path to SDK to be installed on the image
                                      Can be replaced with TARANTOOL_SDK_PATH environment
                                      variable (has lower priority)
                                      Used for docker type

        Packing to docker:
            If you use Tarantool Enterprise, it's required to specify one
            (and only one) of --sdk-local and --sdk-path options.

            You can pass additional arguments to `docker build` command using
            TARANTOOL_DOCKER_BUILD_ARGS env variable.
    ]=]):format(self_name, table.concat(available_distribution_types, ', '))
}

function cmd_pack.callback(args)
    if not fio.path.exists(args.path) then
        die("Specified path %s doesn't exist", args.path)
    end

    if not fio.path.is_dir(args.path) then
        die("Specified path %s is not a directory", args.path)
    end

    local name, version, release = detect_name_version_release(args.path, args.name, args.version)

    -- collect general application info
    app_state.path = fio.abspath(args.path)
    app_state.name = name
    app_state.version = version
    app_state.release = release
    app_state.version_release = string.format('%s-%s', version, release)

    app_state.tarantool_version = get_tarantool_version()

    -- collect pack-specific application info
    app_state.dest_dir = fio.cwd()
    app_state.deprecated_flow = check_if_deprecated_build_flow_is_ised(app_state.path)
    app_state.tarantool_is_enterprise = tarantool_is_enterprise()
    app_state.build_in_docker = (args.type == distribution_types.DOCKER) or args.use_docker
    app_state.base_image_fullname = string.format('%s-base', app_state.name)

    -- generate build ID
    app_state.build_id = utils.random_string()

    -- build directory structure:
    -- build_dir/
    --   app-files/               <- package files
    --     usr/share/tarantool/
    --     or
    --     appname/
    --   Dockerfile               <- additional files used for building the application
    app_state.build_dir = detect_and_create_build_dir(app_state.path, app_state.build_id)
    app_state.appfiles_dir = fio.pathjoin(app_state.build_dir, APPFILES_DIRNAME)
    local ok, err = utils.make_tree(app_state.appfiles_dir)
    if not ok then
        die('Failed to create directory for build application files: %s', err)
    end

    app_state.docker_build_args = args.docker_build_args
    app_state.build_sdk_dirname = string.format('sdk-%s', app_state.build_id)

    if app_state.build_in_docker then
        if which('docker') == nil then
            die('docker binary is required to build application in docker')
        end

        app_state.build_base_dockerfile_layers = get_dockerfile_base_layers(
            args.build_from, DEFAULT_BUILD_BASE_DOCKERFILE_LAYERS
        )
    end

    if args.type == distribution_types.DOCKER then
        app_state.runtime_base_dockerfile_layers = get_dockerfile_base_layers(
            args.from, DEFAULT_RUNTIME_BASE_DOCKERFILE_LAYERS
        )
    end

    if app_state.tarantool_is_enterprise and app_state.build_in_docker then
        local sdk_path, err = detect_sdk_path(args)
        if sdk_path == nil then die(err) end

        -- check that specified path is an existing directory
        if not fio.path.exists(sdk_path) then
            die('Specified SDK path does not exists: %s', sdk_path)
        end

        if not fio.path.is_dir(sdk_path) then
            die('Specified SDK path is not a directory: %s', sdk_path)
        end

        -- check that SDK directory contains tarantool and tarantoolctl binaries
        -- and they both are executable
        for _, binary in ipairs({'tarantool', 'tarantoolctl'}) do
            local sdk_binary_path = fio.pathjoin(sdk_path, binary)
            if not fio.path.exists(sdk_binary_path) then
                die('Specified SDK directory (%s) does not contain %s binary', sdk_path, binary)
            end

            if not is_executable(sdk_binary_path) then
                die('Specified SDK directory contains %s binary that is not executable', binary)
            end
        end

        -- save path to SDK in app_state
        app_state.sdk_path = sdk_path
    end

    local ok_state, err_state = check_pack_state(app_state)
    if not ok_state then
        die(format_internal_error(err_state))
    end

    local instantiated_unit_template
    if args.instantiated_unit_template then
        local err
        instantiated_unit_template, err = utils.read_file(args.instantiated_unit_template)
        if instantiated_unit_template == nil then return false, err end
    end

    local unit_template
    if args.unit_template then
        local err
        unit_template, err = utils.read_file(args.unit_template)
        if unit_template == nil then return false, err end
    end

    local opts
    if utils.array_contains({distribution_types.RPM, distribution_types.DEB}, args.type) then
        opts = {
            unit_template = unit_template,
            instantiated_unit_template = instantiated_unit_template
        }
    elseif args.type == distribution_types.DOCKER then
        opts = {
            tag = args.tag,
        }
    end

    local pack_handler = pack_handlers[args.type]
    if pack_handler == nil then
        local handler_err = string.format("Pack handler for %s distribution type not found", args.type)
        die(format_internal_error(handler_err))
    end

    local ok_pcall, res_pack, err_pack = pcall(pack_handler, opts)
    if not ok_pcall then
        warn('Failed to pack application')
        remove_build_dir()
        die(format_internal_error(res_pack))
    elseif not res_pack then
        warn('Failed to pack application')
        remove_build_dir()
        die('Failed to pack application: %s', err_pack)
    end


    -- clean build directory
    remove_build_dir()
    info('Packing application succeeded!')
end

function cmd_pack.parse(cmd_args)
    local args_schema = {
        args = {
            'type',
            'path',
        },
        opts = {
            name = 'string',
            version = 'string',
            instantiated_unit_template = 'string',
            unit_template = 'string',
            sdk_path = 'string',
            sdk_local = 'boolean',
            tag = 'string',
            from = 'string',
            build_from = 'string',
            use_docker = 'boolean',
        }
    }

    local args, err = argparse.parse(cmd_args, args_schema)

    if args == nil then
        die("Failed to parse args: %s", err)
    end

    args.use_docker = args.use_docker or false
    args.docker_build_args = os.getenv('TARANTOOL_DOCKER_BUILD_ARGS') or ''
    args.sdk_local = args.sdk_local or false
    args.sdk_path = args.sdk_path or os.getenv('TARANTOOL_SDK_PATH')

    if args.sdk_path ~= nil then
        args.sdk_path = fio.abspath(args.sdk_path)
    end

    if args.path == nil then
        args.path = fio.cwd()
    end

    if args.type == nil then
        die('Package type is required')
    end

    if not utils.array_contains(available_distribution_types, args.type) then
        die("Package type should be one of: %s",
                table.concat(available_distribution_types, ', '))
    end

    if args.type == distribution_types.DOCKER then
        if args.version ~= nil and args.tag ~= nil then
            die(
                'You can specify only one of --version and --tag options. ' ..
                'Run `cartridge pack --help` for details.'
            )
        end
    end

    if args.build_from == nil then
        local default_build_base_dockerfile_path = fio.pathjoin(
            args.path,
            DEFAULT_BUILD_BASE_DOCKERFILE_NAME
        )
        if fio.path.exists(default_build_base_dockerfile_path) then
            args.build_from = default_build_base_dockerfile_path
        end
    end

    if args.from == nil then
        local default_runtime_base_dockerfile_path = fio.pathjoin(
            args.path,
            DEFAULT_RUNTIME_BASE_DOCKERFILE_NAME
        )
        if fio.path.exists(default_runtime_base_dockerfile_path) then
            args.from = default_runtime_base_dockerfile_path
        end
    end

    return args
end

-- * ---------------- Application templating ----------------

local cmd_create = {
    name = 'create',
    doc = 'Create a new app from template',
    usage = utils.remove_leading_spaces([=[
        %s create [options] [<path>]

        Arguments
            path                   Directory to create the application in
                                   Defaults to the current directory

        Options
            --name NAME            Application name

            --template TEMPLATE    Application template
                                   Defaults to `cartridge` (no others currently supported)
    ]=]):format(self_name),
}

function cmd_create.parse(cmd_args)
    local args_schema = {
        args = {
            'path'
        },
        opts = {
            name = 'string',
            template = 'string'
        },
    }

    local args, err = argparse.parse(cmd_args, args_schema)

    if args == nil then
        die('Failed to parse args: %s', err)
    end

    return args
end

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


local function create_app_directory_and_init_git(dest_dir, template_name, app_name)
    assert(fio.path.exists(dest_dir))

    local ok, err = templates.instantiate(dest_dir, template_name, app_name)
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

            local ok, err = utils.write_file(fio.pathjoin(dest_dir, '.gitignore'), GITIGNORE)
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

local function remove_dest_dir(dest_dir)
    info('Remove dest directory %s', dest_dir)
    local ok, err = utils.remove_by_path(dest_dir)
    if not ok then
        warn('Failed to clean up build directory %s: %s', dest_dir, err)
    end
end

function cmd_create.callback(args)
    local path = args.path and fio.abspath(args.path) or fio.cwd()

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

    local ok, err = utils.make_tree(dest_dir)
    if not ok then
        die("Failed to create application directory: %s", err)
    end

    local ok_pcall, res_create, err_create = pcall(
        create_app_directory_and_init_git,
        dest_dir, template, name
    )
    if not ok_pcall then
        warn('Failed to create application')
        remove_dest_dir(dest_dir)
        die(format_internal_error(res_create))
    elseif not res_create then
        warn("Failed to create application...")
        remove_dest_dir(dest_dir)
        die('Failed to create application: %s', err_create)
    end

    info("Application successfully created in '%s'", dest_dir)
end

-- * ----------------- Build application -----------------

local cmd_build = {
    name = 'build',
    doc = 'Build application for local development',
    usage = utils.remove_leading_spaces([=[
        %s build [<path>]

        Arguments
            path                      Path to the application
                                      Defaults to the current directory
    ]=]):format(self_name),
}

function cmd_build.parse(cmd_args)
    local args_schema = {
        args = {
            'path'
        },
    }

    local args, err = argparse.parse(cmd_args, args_schema)

    if args == nil then
        die("Failed to parse args: %s", err)
    end

    if args.path == nil then
        args.path = fio.cwd()
    end

    return args
end

function cmd_build.callback(args)
    if not fio.path.exists(args.path) then
        die("Specified path %s doesn't exist", args.path)
    end

    if not fio.path.is_dir(args.path) then
        die("Specified path %s is not a directory", args.path)
    end

    -- collect application info
    app_state.path = fio.abspath(args.path)
    app_state.tarantool_is_enterprise = tarantool_is_enterprise()
    app_state.deprecated_flow = check_if_deprecated_build_flow_is_ised(app_state.path)

    info('Build application in %s', app_state.path)

    local ok, err = build_application(app_state.path)
    if not ok then
        die('Failed to build application: %s', err)
    end

    info('Application successfully built!')
end

-- * ---------------- Instance management ----------------

local cmd_start = {
    name = 'start',
    doc = 'Start a Tarantool instance(s)',
    usage = utils.remove_leading_spaces([=[
        %s start [APP_NAME[.INSTANCE_NAME]] [options]

        Default APP_NAME is parsed from ./*.rockspec filename.
        When INSTANCE_NAME is not provided it reads `cfg` file and starts all
        defined instances.

        Options
            --script FILE       Application's entry point.
                                Defaults to TARANTOOL_SCRIPT,
                                or ./init.lua when running from app's directory,
                                or :apps_path/:app_name/init.lua in multi-app env.

            --apps-path PATH    Path to apps directory when running in multi-app env.
                                Defaults to /usr/share/tarantool

            --run-dir DIR       Directory with pid and sock files
                                Defaults to TARANTOOL_RUN_DIR or /var/run/tarantool

            --cfg FILE          Cartridge instances config file.
                                Defaults to TARANTOOL_CFG or ./instances.yml

            --daemonize / -d    Start in background

        Default options can be overridden in ./.cartridge.yml or ~/.cartridge.yml,
        also options from .cartridge.yml can be overridden by corresponding
        TARANTOOL_* environment variables .
    ]=]):format(self_name),
}

-- Fetches app_name from .rockspec file.
-- Extracted from cartridge.argparse, but searches for rockspec only in the current
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
            from_file = yaml.decode(utils.read_file(path))
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

function cmd_start.parse(cmd_args)
    local args_schema = {
        opts = {
            script = 'string',
            apps_path = 'string',
            run_dir = 'string',
            cfg = 'string',
            daemonize = 'boolean',
            d = 'boolean',
            verbose = 'boolean',
        },
        args = {
            'instance_id',
        }
    }

    local result, err = argparse.parse(cmd_args, args_schema)

    if result == nil then
        die("Failed to parse args: %s", err)
    end

    if result.daemonize == nil then result.daemonize = result.d end
    local defaults = read_cartridge_defaults()
    for k, v in pairs(defaults) do
        result[k] = result[k] or v
    end

    local instance_id = (result.instance_id or ''):split('.')
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
        return yaml.decode(utils.read_file(path))
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
        local pid = tonumber(utils.read_file(self.pid_file))
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
    utils.write_file(self.pid_file, require('tarantool').pid(), tonumber('644', 8))
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
    utils.write_file(self.pid_file, pid, tonumber('644', 8))
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
    utils.write_file(self.pid_file, pid, tonumber('644', 8))
    fiber.create(log_pipes_forwarder, pipes, self.instance_name)
end

local cmd_stop = {
    name = 'stop',
    doc = 'Stop a Tarantool instance(s)',
    usage = utils.remove_leading_spaces([=[
        %s stop [APP_NAME[.INSTANCE_NAME]] [options]

        When INSTANCE_NAME is not provided it reads `cfg` file and stops all
        defined instances.

        These options from `start` command are supported:
            --run-dir DIR
            --cfg FILE
    ]=]):format(self_name),
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

    local pid = tonumber(utils.read_file(pid_file))
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
    cmd_create,
    cmd_pack,
    cmd_build,
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

    if utils.array_contains(arg, "--help") then
        if type(command.usage) == "string" then
            print(command.usage)
        else
            command.usage()
        end
        os.exit(0)
    end

    local args = command.parse(utils.array_slice(arg, 2))

    command.callback(args)
end

_G.app_state = app_state

return {
   matching = matching,
   main = main,
   dockerfile_constructors = {
       install_tarantool = construct_install_tarantool_dockerfile_part,
       build = construct_build_image_dockerfile,
       runtime = construct_runtime_image_dockerfile,
   },
   detect_sdk_path = detect_sdk_path,
}
