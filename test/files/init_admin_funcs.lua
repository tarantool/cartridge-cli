require('strict').on()

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

local console_sock = os.getenv('TARANTOOL_CONSOLE_SOCK')
if console_sock ~= nil then
    local console = require('console')
    console.listen('unix/:' .. console_sock)
end

-- this code is here to prevent building application with cartridge
-- for each `cartridge admin` test
local fio = require('fio')
local conf_path = 'instances.yml'
local file = fio.open(conf_path)
if file ~= nil then
    local app_name = os.getenv('TARANTOOL_APP_NAME')
    local instance_name = os.getenv('TARANTOOL_INSTANCE_NAME')

    assert(app_name ~= nil)
    assert(instance_name ~= nil)

    local workdir = string.format('tmp/data/%s.%s', app_name, instance_name)

    local cwd = fio.cwd()

    local instance_id = string.format('%s.%s', app_name, instance_name)

    local yaml = require('yaml')
    local conf = yaml.decode(file:read())

    for section_name, instance_conf in pairs(conf) do
        if section_name == instance_id then
            if instance_conf.advertise_uri ~= nil then
                box.cfg{
                    listen = instance_conf.advertise_uri,
                    memtx_dir = workdir,
                    wal_dir = workdir,
                }

                box.schema.user.passwd('admin', string.format('secret-cluster-cookie'))

                fio.chdir(cwd)
            end

            break
        end
    end
end

require('log').info('I am starting...')

-- register custom admin functions

local cli_admin = require('cartridge-cli-extensions.admin')

cli_admin.init()

local echo_user = {
    usage = 'echo_user usage',
    args = {
        username = {
            type = 'string',
            usage = 'username usage',
        },
        age = {
            type = 'number',
            usage = 'age usage',
        },
        loves_cakes = {
            type = 'boolean',
            usage = 'loves_cakes usage',
        },
    },
    call = function(opts)
        opts = opts or {}

        if opts.username == nil then
            return nil, "Please, pass --username flag, I need to know your name"
        end

        local res = {string.format('Hi, %s!', opts.username)}

        if opts.age ~= nil then
            table.insert(res, string.format("You are %s years old", opts.age))
        else
            table.insert(res, string.format("I don't know your age"))
        end

        if opts.loves_cakes then
            table.insert(res, string.format("I know that you like cakes!"))
        else
            table.insert(res, string.format("How can you not love cakes?"))
        end

        return res
    end,
}

local func_long_name = {
    usage = 'func_long_name usage',
    call = function() return {string.format('func_long_name was called')} end,
}

local func_no_args = {
    usage = 'func_no_args usage',
    call = function() return {string.format('func_no_args was called')} end,
}

local func_long_arg = {
    usage = 'func_long_arg usage',
    args = {
        long_arg = {
            usage = 'long_arg usage',
            type = 'string',
        }
    },
    call = function(opts)
        return {
            string.format('func_long_arg was called with %q arg', opts.long_arg),
        }
    end,
}

local func_rets_str = {
    usage = 'func_rets_str usage',
    call = function() return string.format('func_rets_str was called') end,
}

local func_rets_non_str = {
    usage = 'func_rets_non_str usage',
    call = function() return 666 end,
}

local func_conflicting = {
    usage = 'func_conflicting usage',
    args = {},
    call = function()
        return { string.format('func_conflicting shouldn\'t be called') }
    end,
}

local conflicting_names = {
    "name", "instance", "run_dir", "data_dir", "list", "help",
    "debug", "quiet", "verbose",
}

for _, argname in ipairs(conflicting_names) do
    func_conflicting.args[argname] = {
        usage = string.format('%s usage', argname),
        type = 'string',
    }
end

local func_rets_err = {
    usage = 'func_rets_err usage',
    call = function()
        return nil, string.format('Some horrible error')
    end,
}

local func_raises_err = {
    usage = 'func_raises_err usage',
    call = function()
        error('Some horrible error raised')
    end,
}

local func_print = {
    usage = 'func_print usage',
    args = {
        num = {
            usage = 'Iterations num',
            type = 'number',
        }
    },
    call = function(opts)
        for i=1,opts.num or 1 do
            print(string.format("Iteration %s (printed)", i))
            box.session.push(string.format("Iteration %s (pushed)", i))
        end

        return 'I am some great result'
    end,
}

assert(cli_admin.register('echo_user', echo_user.usage, echo_user.args, echo_user.call))
assert(cli_admin.register('func.long.name', func_long_name.usage, func_long_name.args, func_long_name.call))
assert(cli_admin.register('func_no_args', func_no_args.usage, func_no_args.args, func_no_args.call))
assert(cli_admin.register('func_long_arg', func_long_arg.usage, func_long_arg.args, func_long_arg.call))
assert(cli_admin.register('func_rets_str', func_rets_str.usage, func_rets_str.args, func_rets_str.call))
assert(cli_admin.register('func_rets_non_str', func_rets_non_str.usage, func_rets_non_str.args, func_rets_non_str.call))
assert(cli_admin.register('func_conflicting', func_conflicting.usage, func_conflicting.args, func_conflicting.call))
assert(cli_admin.register('func_rets_err', func_rets_err.usage, func_rets_err.args, func_rets_err.call))
assert(cli_admin.register('func_raises_err', func_raises_err.usage, func_raises_err.args, func_raises_err.call))
assert(cli_admin.register('func_print', func_print.usage, func_print.args, func_print.call))
